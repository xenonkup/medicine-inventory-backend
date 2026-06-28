package repository

import (
	"context"

	"github.com/google/uuid"
	"pharmacy-backend/internal/domain"
)

// UserRepository abstracts persistence for users so services can be tested
// against a mock without a live database.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	List(ctx context.Context, offset, limit int) ([]domain.User, int64, error)
	CountAll(ctx context.Context) (int64, error)
}

// CategoryFilter narrows a category listing.
type CategoryFilter struct {
	Search     string
	ActiveOnly bool
	Offset     int
	Limit      int
}

// CategoryRepository abstracts persistence for categories.
type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	Update(ctx context.Context, category *domain.Category) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Category, error)
	List(ctx context.Context, f CategoryFilter) ([]domain.Category, int64, error)
	CountActiveMedicines(ctx context.Context, categoryID uuid.UUID) (int64, error)
}

// MedicineFilter narrows a medicine listing.
type MedicineFilter struct {
	Search     string
	CategoryID *uuid.UUID
	ActiveOnly bool
	Offset     int
	Limit      int
}

// MedicineRepository abstracts persistence for medicines.
type MedicineRepository interface {
	Create(ctx context.Context, medicine *domain.Medicine) error
	Update(ctx context.Context, medicine *domain.Medicine) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Medicine, error)
	List(ctx context.Context, f MedicineFilter) ([]domain.Medicine, int64, error)
}
