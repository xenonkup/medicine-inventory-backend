package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"pharmacy-backend/internal/domain"
)

type lotRepository struct {
	db *gorm.DB
}

// NewLotRepository builds a GORM-backed LotRepository.
func NewLotRepository(db *gorm.DB) LotRepository {
	return &lotRepository{db: db}
}

func (r *lotRepository) Create(ctx context.Context, lot *domain.Lot) error {
	return dbFromCtx(ctx, r.db).Create(lot).Error
}

func (r *lotRepository) Save(ctx context.Context, lot *domain.Lot) error {
	return dbFromCtx(ctx, r.db).Save(lot).Error
}

func (r *lotRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Lot, error) {
	var lot domain.Lot
	err := dbFromCtx(ctx, r.db).First(&lot, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrLotNotFound
	}
	if err != nil {
		return nil, err
	}
	return &lot, nil
}

func (r *lotRepository) FindByMedicineAndLotNumber(ctx context.Context, medicineID uuid.UUID, lotNumber string) (*domain.Lot, error) {
	var lot domain.Lot
	err := dbFromCtx(ctx, r.db).
		Where("medicine_id = ? AND lot_number = ?", medicineID, lotNumber).
		First(&lot).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrLotNotFound
	}
	if err != nil {
		return nil, err
	}
	return &lot, nil
}

func (r *lotRepository) FindAvailableForUpdate(ctx context.Context, medicineID uuid.UUID) ([]domain.Lot, error) {
	var lots []domain.Lot
	err := dbFromCtx(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("medicine_id = ? AND qty_remaining > 0", medicineID).
		Order("expiry_date ASC, received_date ASC").
		Find(&lots).Error
	return lots, err
}

func (r *lotRepository) ListByMedicine(ctx context.Context, medicineID uuid.UUID) ([]domain.Lot, error) {
	var lots []domain.Lot
	err := dbFromCtx(ctx, r.db).
		Where("medicine_id = ?", medicineID).
		Order("expiry_date ASC, received_date ASC").
		Find(&lots).Error
	return lots, err
}

func (r *lotRepository) SumRemaining(ctx context.Context, medicineID uuid.UUID) (int, error) {
	var sum int64
	err := dbFromCtx(ctx, r.db).
		Model(&domain.Lot{}).
		Where("medicine_id = ?", medicineID).
		Select("COALESCE(SUM(qty_remaining), 0)").
		Scan(&sum).Error
	return int(sum), err
}

func (r *lotRepository) SumRemainingByMedicineIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]int, error) {
	result := make(map[uuid.UUID]int, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	type row struct {
		MedicineID uuid.UUID
		Total      int
	}
	var rows []row
	err := dbFromCtx(ctx, r.db).
		Model(&domain.Lot{}).
		Select("medicine_id, COALESCE(SUM(qty_remaining), 0) AS total").
		Where("medicine_id IN ?", ids).
		Group("medicine_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, x := range rows {
		result[x.MedicineID] = x.Total
	}
	return result, nil
}

func (r *lotRepository) FindNearExpiry(ctx context.Context, asOf time.Time, withinDays int) ([]domain.Lot, error) {
	cutoff := asOf.AddDate(0, 0, withinDays)
	var lots []domain.Lot
	err := dbFromCtx(ctx, r.db).
		Preload("Medicine").
		Joins("JOIN medicines ON medicines.id = lots.medicine_id AND medicines.is_active = true").
		Where("lots.qty_remaining > 0 AND lots.expiry_date <= ?", cutoff).
		Order("lots.expiry_date ASC").
		Find(&lots).Error
	return lots, err
}
