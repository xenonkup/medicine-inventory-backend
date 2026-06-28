package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"pharmacy-backend/internal/domain"
)

type medicineRepository struct {
	db *gorm.DB
}

// NewMedicineRepository builds a GORM-backed MedicineRepository.
func NewMedicineRepository(db *gorm.DB) MedicineRepository {
	return &medicineRepository{db: db}
}

func (r *medicineRepository) Create(ctx context.Context, medicine *domain.Medicine) error {
	return r.db.WithContext(ctx).Create(medicine).Error
}

func (r *medicineRepository) Update(ctx context.Context, medicine *domain.Medicine) error {
	return r.db.WithContext(ctx).Save(medicine).Error
}

func (r *medicineRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Medicine, error) {
	var medicine domain.Medicine
	err := r.db.WithContext(ctx).Preload("Category").First(&medicine, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrMedicineNotFound
	}
	if err != nil {
		return nil, err
	}
	return &medicine, nil
}

func (r *medicineRepository) List(ctx context.Context, f MedicineFilter) ([]domain.Medicine, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.Medicine{})
	if f.ActiveOnly {
		q = q.Where("is_active = ?", true)
	}
	if f.CategoryID != nil {
		q = q.Where("category_id = ?", *f.CategoryID)
	}
	if f.Search != "" {
		like := "%" + f.Search + "%"
		q = q.Where("name ILIKE ? OR code ILIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var medicines []domain.Medicine
	err := q.Preload("Category").
		Order("name ASC").
		Offset(f.Offset).Limit(f.Limit).
		Find(&medicines).Error
	if err != nil {
		return nil, 0, err
	}
	return medicines, total, nil
}
