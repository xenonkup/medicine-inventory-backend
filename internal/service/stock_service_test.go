package service

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"

	"pharmacy-backend/internal/domain"
	"pharmacy-backend/internal/dto"
	"pharmacy-backend/internal/repository"
)

// --- in-memory mocks ---

// passthroughTx runs the function directly (no real transaction) so service
// logic can be tested without a database.
type passthroughTx struct{}

func (passthroughTx) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type mockLotRepo struct {
	lots map[uuid.UUID]*domain.Lot
}

func newMockLotRepo(lots ...*domain.Lot) *mockLotRepo {
	m := &mockLotRepo{lots: map[uuid.UUID]*domain.Lot{}}
	for _, l := range lots {
		if l.ID == uuid.Nil {
			l.ID = uuid.New()
		}
		cp := *l
		m.lots[l.ID] = &cp
	}
	return m
}

func (m *mockLotRepo) Create(_ context.Context, lot *domain.Lot) error {
	if lot.ID == uuid.Nil {
		lot.ID = uuid.New()
	}
	cp := *lot
	m.lots[lot.ID] = &cp
	return nil
}

func (m *mockLotRepo) Save(_ context.Context, lot *domain.Lot) error {
	cp := *lot
	m.lots[lot.ID] = &cp
	return nil
}

func (m *mockLotRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Lot, error) {
	if l, ok := m.lots[id]; ok {
		cp := *l
		return &cp, nil
	}
	return nil, domain.ErrLotNotFound
}

func (m *mockLotRepo) FindByMedicineAndLotNumber(_ context.Context, medicineID uuid.UUID, lotNumber string) (*domain.Lot, error) {
	for _, l := range m.lots {
		if l.MedicineID == medicineID && l.LotNumber == lotNumber {
			cp := *l
			return &cp, nil
		}
	}
	return nil, domain.ErrLotNotFound
}

func (m *mockLotRepo) FindAvailableForUpdate(_ context.Context, medicineID uuid.UUID) ([]domain.Lot, error) {
	var out []domain.Lot
	for _, l := range m.lots {
		if l.MedicineID == medicineID && l.QtyRemaining > 0 {
			out = append(out, *l)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ExpiryDate.Equal(out[j].ExpiryDate) {
			return out[i].ReceivedDate.Before(out[j].ReceivedDate)
		}
		return out[i].ExpiryDate.Before(out[j].ExpiryDate)
	})
	return out, nil
}

func (m *mockLotRepo) ListByMedicine(_ context.Context, medicineID uuid.UUID) ([]domain.Lot, error) {
	var out []domain.Lot
	for _, l := range m.lots {
		if l.MedicineID == medicineID {
			out = append(out, *l)
		}
	}
	return out, nil
}

func (m *mockLotRepo) SumRemaining(_ context.Context, medicineID uuid.UUID) (int, error) {
	sum := 0
	for _, l := range m.lots {
		if l.MedicineID == medicineID {
			sum += l.QtyRemaining
		}
	}
	return sum, nil
}

func (m *mockLotRepo) SumRemainingByMedicineIDs(_ context.Context, ids []uuid.UUID) (map[uuid.UUID]int, error) {
	res := map[uuid.UUID]int{}
	for _, l := range m.lots {
		res[l.MedicineID] += l.QtyRemaining
	}
	return res, nil
}

func (m *mockLotRepo) FindNearExpiry(_ context.Context, asOf time.Time, withinDays int) ([]domain.Lot, error) {
	cutoff := asOf.AddDate(0, 0, withinDays)
	var out []domain.Lot
	for _, l := range m.lots {
		if l.QtyRemaining > 0 && !l.ExpiryDate.After(cutoff) {
			out = append(out, *l)
		}
	}
	return out, nil
}

type mockTxnRepo struct {
	created []domain.StockTransaction
}

func (m *mockTxnRepo) Create(_ context.Context, txn *domain.StockTransaction) error {
	m.created = append(m.created, *txn)
	return nil
}

func (m *mockTxnRepo) List(_ context.Context, _ repository.TransactionFilter) ([]domain.StockTransaction, int64, error) {
	return m.created, int64(len(m.created)), nil
}

func (m *mockTxnRepo) CountSince(_ context.Context, _ time.Time) (int64, error) {
	return int64(len(m.created)), nil
}

type mockMedicineRepo struct {
	medicines map[uuid.UUID]*domain.Medicine
}

func (m *mockMedicineRepo) Create(context.Context, *domain.Medicine) error { return nil }
func (m *mockMedicineRepo) Update(context.Context, *domain.Medicine) error { return nil }
func (m *mockMedicineRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Medicine, error) {
	if med, ok := m.medicines[id]; ok {
		return med, nil
	}
	return nil, domain.ErrMedicineNotFound
}
func (m *mockMedicineRepo) List(context.Context, repository.MedicineFilter) ([]domain.Medicine, int64, error) {
	return nil, 0, nil
}

// --- helpers ---

func date(y int, mo time.Month, d int) time.Time {
	return time.Date(y, mo, d, 0, 0, 0, 0, time.UTC)
}

func newStockServiceFixture(medicineID uuid.UUID, lots ...*domain.Lot) (*StockService, *mockLotRepo, *mockTxnRepo) {
	lotRepo := newMockLotRepo(lots...)
	txnRepo := &mockTxnRepo{}
	medRepo := &mockMedicineRepo{medicines: map[uuid.UUID]*domain.Medicine{
		medicineID: {Base: domain.Base{ID: medicineID}, Code: "M1", Name: "Med", IsActive: true},
	}}
	svc := NewStockService(passthroughTx{}, lotRepo, txnRepo, medRepo)
	return svc, lotRepo, txnRepo
}

// --- tests ---

func TestStockOut_FEFO_AllocatesEarliestExpiryFirst(t *testing.T) {
	medID := uuid.New()
	lotEarly := &domain.Lot{Base: domain.Base{ID: uuid.New()}, MedicineID: medID, LotNumber: "C", ExpiryDate: date(2026, 7, 1), QtyReceived: 10, QtyRemaining: 10, ReceivedDate: date(2026, 1, 1)}
	lotMid := &domain.Lot{Base: domain.Base{ID: uuid.New()}, MedicineID: medID, LotNumber: "A", ExpiryDate: date(2026, 9, 1), QtyReceived: 20, QtyRemaining: 20, ReceivedDate: date(2026, 1, 1)}
	lotLate := &domain.Lot{Base: domain.Base{ID: uuid.New()}, MedicineID: medID, LotNumber: "B", ExpiryDate: date(2026, 12, 1), QtyReceived: 50, QtyRemaining: 50, ReceivedDate: date(2026, 1, 1)}

	svc, lotRepo, txnRepo := newStockServiceFixture(medID, lotEarly, lotMid, lotLate)

	res, err := svc.StockOut(context.Background(), dto.StockOutRequest{
		MedicineID: medID.String(), Quantity: 25,
	}, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.Allocations) != 2 {
		t.Fatalf("expected 2 allocations, got %d", len(res.Allocations))
	}
	// Earliest expiry (lot C) consumed fully first, then lot A.
	if res.Allocations[0].LotNumber != "C" || res.Allocations[0].Deducted != 10 {
		t.Errorf("first allocation should be C:10, got %s:%d", res.Allocations[0].LotNumber, res.Allocations[0].Deducted)
	}
	if res.Allocations[1].LotNumber != "A" || res.Allocations[1].Deducted != 15 {
		t.Errorf("second allocation should be A:15, got %s:%d", res.Allocations[1].LotNumber, res.Allocations[1].Deducted)
	}

	if got := lotRepo.lots[lotEarly.ID].QtyRemaining; got != 0 {
		t.Errorf("lot C remaining should be 0, got %d", got)
	}
	if got := lotRepo.lots[lotMid.ID].QtyRemaining; got != 5 {
		t.Errorf("lot A remaining should be 5, got %d", got)
	}
	if got := lotRepo.lots[lotLate.ID].QtyRemaining; got != 50 {
		t.Errorf("lot B remaining should be untouched 50, got %d", got)
	}
	if len(txnRepo.created) != 2 {
		t.Errorf("expected 2 OUT ledger entries, got %d", len(txnRepo.created))
	}
	for _, tx := range txnRepo.created {
		if tx.Type != domain.TxOut {
			t.Errorf("expected OUT type, got %s", tx.Type)
		}
	}
}

func TestStockOut_InsufficientStock_NoMutation(t *testing.T) {
	medID := uuid.New()
	lot := &domain.Lot{Base: domain.Base{ID: uuid.New()}, MedicineID: medID, LotNumber: "A", ExpiryDate: date(2026, 9, 1), QtyReceived: 30, QtyRemaining: 30, ReceivedDate: date(2026, 1, 1)}
	svc, lotRepo, txnRepo := newStockServiceFixture(medID, lot)

	_, err := svc.StockOut(context.Background(), dto.StockOutRequest{
		MedicineID: medID.String(), Quantity: 31,
	}, uuid.New())
	if err != domain.ErrInsufficientStock {
		t.Fatalf("expected ErrInsufficientStock, got %v", err)
	}
	if got := lotRepo.lots[lot.ID].QtyRemaining; got != 30 {
		t.Errorf("stock must be unchanged on insufficient, got %d", got)
	}
	if len(txnRepo.created) != 0 {
		t.Errorf("no ledger entries expected, got %d", len(txnRepo.created))
	}
}

func TestStockOut_ExactTotalAcrossLots(t *testing.T) {
	medID := uuid.New()
	a := &domain.Lot{Base: domain.Base{ID: uuid.New()}, MedicineID: medID, LotNumber: "A", ExpiryDate: date(2026, 7, 1), QtyReceived: 10, QtyRemaining: 10, ReceivedDate: date(2026, 1, 1)}
	b := &domain.Lot{Base: domain.Base{ID: uuid.New()}, MedicineID: medID, LotNumber: "B", ExpiryDate: date(2026, 8, 1), QtyReceived: 15, QtyRemaining: 15, ReceivedDate: date(2026, 1, 1)}
	svc, lotRepo, _ := newStockServiceFixture(medID, a, b)

	res, err := svc.StockOut(context.Background(), dto.StockOutRequest{MedicineID: medID.String(), Quantity: 25}, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Allocations) != 2 {
		t.Fatalf("expected 2 allocations, got %d", len(res.Allocations))
	}
	if lotRepo.lots[a.ID].QtyRemaining != 0 || lotRepo.lots[b.ID].QtyRemaining != 0 {
		t.Errorf("both lots should be drained to 0")
	}
}

func TestReturn_AfterExpiry_Rejected(t *testing.T) {
	medID := uuid.New()
	lot := &domain.Lot{Base: domain.Base{ID: uuid.New()}, MedicineID: medID, LotNumber: "A", ExpiryDate: date(2026, 1, 1), QtyReceived: 10, QtyRemaining: 5, ReceivedDate: date(2025, 1, 1)}
	svc, lotRepo, _ := newStockServiceFixture(medID, lot)
	svc.now = func() time.Time { return date(2026, 6, 29) } // after expiry

	_, err := svc.Return(context.Background(), dto.StockReturnRequest{LotID: lot.ID.String(), Quantity: 2}, uuid.New())
	if err != domain.ErrReturnAfterExpiry {
		t.Fatalf("expected ErrReturnAfterExpiry, got %v", err)
	}
	if lotRepo.lots[lot.ID].QtyRemaining != 5 {
		t.Errorf("remaining must be unchanged after rejected return")
	}
}

func TestReturn_ExceedsReceived_Rejected(t *testing.T) {
	medID := uuid.New()
	lot := &domain.Lot{Base: domain.Base{ID: uuid.New()}, MedicineID: medID, LotNumber: "A", ExpiryDate: date(2026, 12, 1), QtyReceived: 10, QtyRemaining: 9, ReceivedDate: date(2026, 1, 1)}
	svc, _, _ := newStockServiceFixture(medID, lot)
	svc.now = func() time.Time { return date(2026, 6, 29) }

	_, err := svc.Return(context.Background(), dto.StockReturnRequest{LotID: lot.ID.String(), Quantity: 2}, uuid.New())
	if err != domain.ErrReturnExceedsReceived {
		t.Fatalf("expected ErrReturnExceedsReceived, got %v", err)
	}
}

func TestReturn_Success_RestoresAndLogs(t *testing.T) {
	medID := uuid.New()
	lot := &domain.Lot{Base: domain.Base{ID: uuid.New()}, MedicineID: medID, LotNumber: "A", ExpiryDate: date(2026, 12, 1), QtyReceived: 10, QtyRemaining: 4, ReceivedDate: date(2026, 1, 1)}
	svc, lotRepo, txnRepo := newStockServiceFixture(medID, lot)
	svc.now = func() time.Time { return date(2026, 6, 29) }

	_, err := svc.Return(context.Background(), dto.StockReturnRequest{LotID: lot.ID.String(), Quantity: 3}, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := lotRepo.lots[lot.ID].QtyRemaining; got != 7 {
		t.Errorf("remaining should be 7, got %d", got)
	}
	if len(txnRepo.created) != 1 || txnRepo.created[0].Type != domain.TxReturn {
		t.Errorf("expected one RETURN ledger entry")
	}
}

func TestStockIn_NewLot_AndMergeExisting(t *testing.T) {
	medID := uuid.New()
	svc, lotRepo, txnRepo := newStockServiceFixture(medID)

	// First receipt creates a lot.
	_, err := svc.StockIn(context.Background(), dto.StockInRequest{
		MedicineID: medID.String(), LotNumber: "LOT-1", ExpiryDate: "2027-01-01", Quantity: 100,
	}, uuid.New())
	if err != nil {
		t.Fatalf("stock-in (new) failed: %v", err)
	}
	if len(lotRepo.lots) != 1 {
		t.Fatalf("expected 1 lot, got %d", len(lotRepo.lots))
	}

	// Second receipt of the same lot number merges quantities.
	_, err = svc.StockIn(context.Background(), dto.StockInRequest{
		MedicineID: medID.String(), LotNumber: "LOT-1", ExpiryDate: "2027-01-01", Quantity: 50,
	}, uuid.New())
	if err != nil {
		t.Fatalf("stock-in (merge) failed: %v", err)
	}
	if len(lotRepo.lots) != 1 {
		t.Fatalf("merge should not create a new lot, got %d lots", len(lotRepo.lots))
	}
	for _, l := range lotRepo.lots {
		if l.QtyReceived != 150 || l.QtyRemaining != 150 {
			t.Errorf("merged lot should have 150/150, got %d/%d", l.QtyReceived, l.QtyRemaining)
		}
	}
	if len(txnRepo.created) != 2 {
		t.Errorf("expected 2 IN ledger entries, got %d", len(txnRepo.created))
	}
}
