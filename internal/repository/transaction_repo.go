package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"pharmacy-backend/internal/domain"
)

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

func (r *transactionRepository) List(ctx context.Context, f TransactionFilter) ([]domain.StockTransaction, int64, error) {
	q := dbFromCtx(ctx, r.db).Model(&domain.StockTransaction{})
	if f.MedicineID != nil {
		q = q.Where("medicine_id = ?", *f.MedicineID)
	}
	if f.Type != nil {
		q = q.Where("type = ?", *f.Type)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var txns []domain.StockTransaction
	err := q.Order("created_at DESC").Offset(f.Offset).Limit(f.Limit).Find(&txns).Error
	if err != nil {
		return nil, 0, err
	}
	return txns, total, nil
}

func (r *transactionRepository) CountSince(ctx context.Context, since time.Time) (int64, error) {
	var n int64
	err := dbFromCtx(ctx, r.db).
		Model(&domain.StockTransaction{}).
		Where("created_at >= ?", since).
		Count(&n).Error
	return n, err
}
