package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"pharmacy-backend/internal/domain"
)

// TransactionWithDetails is a flat struct for Scan — embedded structs confuse GORM's column mapper.
type TransactionWithDetails struct {
	ID          uuid.UUID     `gorm:"column:id"`
	LotID       uuid.UUID     `gorm:"column:lot_id"`
	MedicineID  uuid.UUID     `gorm:"column:medicine_id"`
	Type        domain.TxType `gorm:"column:type"`
	Quantity    int           `gorm:"column:quantity"`
	ReferenceNo *string       `gorm:"column:reference_no"`
	Note        *string       `gorm:"column:note"`
	CreatedByID uuid.UUID     `gorm:"column:created_by_id"`
	CreatedAt   time.Time     `gorm:"column:created_at"`
	MedicineName string       `gorm:"column:medicine_name"`
	LotNumber    string       `gorm:"column:lot_number"`
}

type transactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository builds a GORM-backed TransactionRepository.
func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, txn *domain.StockTransaction) error {
	return dbFromCtx(ctx, r.db).Create(txn).Error
}

func (r *transactionRepository) List(ctx context.Context, f TransactionFilter) ([]TransactionWithDetails, int64, error) {
	base := dbFromCtx(ctx, r.db).Model(&domain.StockTransaction{})
	if f.MedicineID != nil {
		base = base.Where("stock_transactions.medicine_id = ?", *f.MedicineID)
	}
	if f.Type != nil {
		base = base.Where("stock_transactions.type = ?", *f.Type)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []TransactionWithDetails
	err := base.
		Joins("JOIN medicines ON medicines.id = stock_transactions.medicine_id").
		Joins("JOIN lots ON lots.id = stock_transactions.lot_id").
		Select("stock_transactions.*, medicines.name AS medicine_name, lots.lot_number AS lot_number").
		Order("stock_transactions.created_at DESC").
		Offset(f.Offset).Limit(f.Limit).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *transactionRepository) CountSince(ctx context.Context, since time.Time) (int64, error) {
	var n int64
	err := dbFromCtx(ctx, r.db).
		Model(&domain.StockTransaction{}).
		Where("created_at >= ?", since).
		Count(&n).Error
	return n, err
}

func (r *transactionRepository) AggregateByType(ctx context.Context, from, to time.Time) ([]TypeAggregate, error) {
	var rows []TypeAggregate
	err := dbFromCtx(ctx, r.db).
		Model(&domain.StockTransaction{}).
		Select("type, COUNT(*) AS count, COALESCE(SUM(quantity), 0) AS total_qty").
		Where("created_at >= ? AND created_at < ?", from, to).
		Group("type").
		Scan(&rows).Error
	return rows, err
}
