package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"pharmacy-backend/internal/domain"
	"pharmacy-backend/internal/dto"
	"pharmacy-backend/internal/repository"
)

// StockService orchestrates inventory movements. Every operation that touches
// both a lot's balance and the ledger runs inside a single transaction so a
// partial write is impossible.
type StockService struct {
	tx        repository.TxManager
	lots      repository.LotRepository
	txns      repository.TransactionRepository
	medicines repository.MedicineRepository
	now       func() time.Time
}

// NewStockService builds a StockService.
func NewStockService(
	tx repository.TxManager,
	lots repository.LotRepository,
	txns repository.TransactionRepository,
	medicines repository.MedicineRepository,
) *StockService {
	return &StockService{
		tx:        tx,
		lots:      lots,
		txns:      txns,
		medicines: medicines,
		now:       time.Now,
	}
}

// StockIn receives a batch. If a lot with the same (medicine, lot number)
// already exists, its quantities are increased; otherwise a new lot is created.
// A ledger IN entry is recorded.
func (s *StockService) StockIn(ctx context.Context, req dto.StockInRequest, userID uuid.UUID) (*dto.LotResponse, error) {
	medicineID, err := uuid.Parse(req.MedicineID)
	if err != nil {
		return nil, domain.ErrMedicineNotFound
	}
	medicine, err := s.medicines.FindByID(ctx, medicineID)
	if err != nil {
		return nil, err
	}
	if !medicine.IsActive {
		return nil, domain.ErrMedicineInactive
	}

	expiry, err := dto.ParseDate(req.ExpiryDate)
	if err != nil {
		return nil, domain.NewAppError("INVALID_DATE", 400, "expiry_date must be YYYY-MM-DD")
	}
	received := s.now()
	if req.ReceivedDate != nil {
		if d, derr := dto.ParseDate(*req.ReceivedDate); derr == nil {
			received = d
		}
	}

	var lot *domain.Lot
	err = s.tx.Do(ctx, func(ctx context.Context) error {
		existing, ferr := s.lots.FindByMedicineAndLotNumber(ctx, medicineID, req.LotNumber)
		switch ferr {
		case nil:
			existing.QtyReceived += req.Quantity
			existing.QtyRemaining += req.Quantity
			if uerr := s.lots.Save(ctx, existing); uerr != nil {
				return uerr
			}
			lot = existing
		case domain.ErrLotNotFound:
			lot = &domain.Lot{
				MedicineID:   medicineID,
				LotNumber:    req.LotNumber,
				ExpiryDate:   expiry,
				QtyReceived:  req.Quantity,
				QtyRemaining: req.Quantity,
				ReceivedDate: received,
			}
			if cerr := s.lots.Create(ctx, lot); cerr != nil {
				return cerr
			}
		default:
			return ferr
		}

		return s.txns.Create(ctx, &domain.StockTransaction{
			LotID:       lot.ID,
			MedicineID:  medicineID,
			Type:        domain.TxIn,
			Quantity:    req.Quantity,
			ReferenceNo: req.ReferenceNo,
			Note:        req.Note,
			CreatedByID: userID,
		})
	})
	if err != nil {
		return nil, err
	}

	resp := dto.NewLotResponse(lot)
	return &resp, nil
}

// StockOut dispenses a quantity using FEFO: lots are consumed in order of
// nearest expiry first. The whole request is rejected if total available stock
// is insufficient — no partial dispense.
func (s *StockService) StockOut(ctx context.Context, req dto.StockOutRequest, userID uuid.UUID) (*dto.StockOutResponse, error) {
	medicineID, err := uuid.Parse(req.MedicineID)
	if err != nil {
		return nil, domain.ErrMedicineNotFound
	}
	medicine, err := s.medicines.FindByID(ctx, medicineID)
	if err != nil {
		return nil, err
	}
	if !medicine.IsActive {
		return nil, domain.ErrMedicineInactive
	}

	allocations := make([]dto.Allocation, 0)
	err = s.tx.Do(ctx, func(ctx context.Context) error {
		lots, lerr := s.lots.FindAvailableForUpdate(ctx, medicineID)
		if lerr != nil {
			return lerr
		}

		total := 0
		for i := range lots {
			total += lots[i].QtyRemaining
		}
		if total < req.Quantity {
			return domain.ErrInsufficientStock
		}

		remaining := req.Quantity
		for i := range lots {
			if remaining == 0 {
				break
			}
			lot := &lots[i]
			take := lot.QtyRemaining
			if take > remaining {
				take = remaining
			}

			lot.QtyRemaining -= take
			if uerr := s.lots.Save(ctx, lot); uerr != nil {
				return uerr
			}
			if terr := s.txns.Create(ctx, &domain.StockTransaction{
				LotID:       lot.ID,
				MedicineID:  medicineID,
				Type:        domain.TxOut,
				Quantity:    take,
				ReferenceNo: req.ReferenceNo,
				Note:        req.Note,
				CreatedByID: userID,
			}); terr != nil {
				return terr
			}

			allocations = append(allocations, dto.Allocation{
				LotID:      lot.ID.String(),
				LotNumber:  lot.LotNumber,
				ExpiryDate: lot.ExpiryDate.Format("2006-01-02"),
				Deducted:   take,
			})
			remaining -= take
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &dto.StockOutResponse{
		MedicineID:    req.MedicineID,
		TotalQuantity: req.Quantity,
		Allocations:   allocations,
	}, nil
}

// Return puts stock back into a specific lot, provided the lot has not expired
// and the return would not exceed the lot's originally received quantity.
func (s *StockService) Return(ctx context.Context, req dto.StockReturnRequest, userID uuid.UUID) (*dto.LotResponse, error) {
	lotID, err := uuid.Parse(req.LotID)
	if err != nil {
		return nil, domain.ErrLotNotFound
	}

	var lot *domain.Lot
	err = s.tx.Do(ctx, func(ctx context.Context) error {
		found, ferr := s.lots.FindByID(ctx, lotID)
		if ferr != nil {
			return ferr
		}
		if found.IsExpired(s.now()) {
			return domain.ErrReturnAfterExpiry
		}
		if found.QtyRemaining+req.Quantity > found.QtyReceived {
			return domain.ErrReturnExceedsReceived
		}

		found.QtyRemaining += req.Quantity
		if uerr := s.lots.Save(ctx, found); uerr != nil {
			return uerr
		}
		lot = found

		return s.txns.Create(ctx, &domain.StockTransaction{
			LotID:       found.ID,
			MedicineID:  found.MedicineID,
			Type:        domain.TxReturn,
			Quantity:    req.Quantity,
			ReferenceNo: req.ReferenceNo,
			Note:        req.Note,
			CreatedByID: userID,
		})
	})
	if err != nil {
		return nil, err
	}

	resp := dto.NewLotResponse(lot)
	return &resp, nil
}

// LotsByMedicine returns all lots for a medicine (FEFO order).
func (s *StockService) LotsByMedicine(ctx context.Context, medicineID uuid.UUID) ([]domain.Lot, error) {
	return s.lots.ListByMedicine(ctx, medicineID)
}

// Transactions returns a page of ledger entries.
func (s *StockService) Transactions(ctx context.Context, f repository.TransactionFilter) ([]domain.StockTransaction, int64, error) {
	return s.txns.List(ctx, f)
}
