package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"pharmacy-backend/internal/domain"
)

type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository builds a GORM-backed CategoryRepository.
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *domain.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *categoryRepository) Update(ctx context.Context, category *domain.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *categoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	var category domain.Category
	err := r.db.WithContext(ctx).First(&category, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrCategoryNotFound
	}
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) List(ctx context.Context, f CategoryFilter) ([]domain.Category, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.Category{})
	if f.ActiveOnly {
		q = q.Where("is_active = ?", true)
	}
	if f.Search != "" {
		q = q.Where("name ILIKE ?", "%"+f.Search+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var categories []domain.Category
	err := q.Order("name ASC").Offset(f.Offset).Limit(f.Limit).Find(&categories).Error
	if err != nil {
		return nil, 0, err
	}
	return categories, total, nil
}

func (r *categoryRepository) CountActiveMedicines(ctx context.Context, categoryID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Medicine{}).
		Where("category_id = ? AND is_active = ?", categoryID, true).
		Count(&count).Error
	return count, err
}
