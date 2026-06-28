package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"pharmacy-backend/internal/domain"
	"pharmacy-backend/internal/dto"
	"pharmacy-backend/internal/repository"
)

// CategoryService handles category master data.
type CategoryService struct {
	categories repository.CategoryRepository
}

// NewCategoryService builds a CategoryService.
func NewCategoryService(categories repository.CategoryRepository) *CategoryService {
	return &CategoryService{categories: categories}
}

// Create adds a new category.
func (s *CategoryService) Create(ctx context.Context, req dto.CreateCategoryRequest) (*domain.Category, error) {
	category := &domain.Category{
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
		IsActive:    true,
	}
	if err := s.categories.Create(ctx, category); err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrDuplicateCategory
		}
		return nil, err
	}
	return category, nil
}

// List returns a page of categories.
func (s *CategoryService) List(ctx context.Context, f repository.CategoryFilter) ([]domain.Category, int64, error) {
	return s.categories.List(ctx, f)
}

// GetByID returns a single category.
func (s *CategoryService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	return s.categories.FindByID(ctx, id)
}

// Update changes category fields.
func (s *CategoryService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateCategoryRequest) (*domain.Category, error) {
	category, err := s.categories.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	category.Name = strings.TrimSpace(req.Name)
	category.Description = req.Description
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}
	if err := s.categories.Update(ctx, category); err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrDuplicateCategory
		}
		return nil, err
	}
	return category, nil
}

// SoftDelete deactivates a category, refusing if it still has active medicines.
func (s *CategoryService) SoftDelete(ctx context.Context, id uuid.UUID) error {
	category, err := s.categories.FindByID(ctx, id)
	if err != nil {
		return err
	}
	count, err := s.categories.CountActiveMedicines(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return domain.ErrCategoryInUse
	}
	category.IsActive = false
	return s.categories.Update(ctx, category)
}
