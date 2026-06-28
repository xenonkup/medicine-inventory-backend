package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"pharmacy-backend/internal/domain"
	"pharmacy-backend/internal/dto"
	"pharmacy-backend/internal/repository"
)

// MedicineService handles medicine master data.
type MedicineService struct {
	medicines  repository.MedicineRepository
	categories repository.CategoryRepository
	lots       repository.LotRepository
}

// NewMedicineService builds a MedicineService.
func NewMedicineService(
	medicines repository.MedicineRepository,
	categories repository.CategoryRepository,
	lots repository.LotRepository,
) *MedicineService {
	return &MedicineService{medicines: medicines, categories: categories, lots: lots}
}

// StockOnHand returns a map of medicine id -> derived stock (sum of lot
// remaining) for the given medicines.
func (s *MedicineService) StockOnHand(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]int, error) {
	return s.lots.SumRemainingByMedicineIDs(ctx, ids)
}

// Create adds a new medicine after validating its category.
func (s *MedicineService) Create(ctx context.Context, req dto.CreateMedicineRequest) (*domain.Medicine, error) {
	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, domain.ErrCategoryNotFound
	}
	if _, err := s.categories.FindByID(ctx, categoryID); err != nil {
		return nil, err
	}

	medicine := &domain.Medicine{
		Code:         strings.TrimSpace(req.Code),
		Name:         strings.TrimSpace(req.Name),
		CategoryID:   categoryID,
		Unit:         strings.TrimSpace(req.Unit),
		ReorderLevel: req.ReorderLevel,
		Description:  req.Description,
		IsActive:     true,
	}
	if err := s.medicines.Create(ctx, medicine); err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrDuplicateMedicineCode
		}
		return nil, err
	}
	return s.medicines.FindByID(ctx, medicine.ID)
}

// List returns a page of medicines.
func (s *MedicineService) List(ctx context.Context, f repository.MedicineFilter) ([]domain.Medicine, int64, error) {
	return s.medicines.List(ctx, f)
}

// GetByID returns a single medicine.
func (s *MedicineService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Medicine, error) {
	return s.medicines.FindByID(ctx, id)
}

// Update changes medicine fields after validating its category.
func (s *MedicineService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateMedicineRequest) (*domain.Medicine, error) {
	medicine, err := s.medicines.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, domain.ErrCategoryNotFound
	}
	if _, err := s.categories.FindByID(ctx, categoryID); err != nil {
		return nil, err
	}

	medicine.Code = strings.TrimSpace(req.Code)
	medicine.Name = strings.TrimSpace(req.Name)
	medicine.CategoryID = categoryID
	medicine.Unit = strings.TrimSpace(req.Unit)
	medicine.ReorderLevel = req.ReorderLevel
	medicine.Description = req.Description
	if req.IsActive != nil {
		medicine.IsActive = *req.IsActive
	}
	if err := s.medicines.Update(ctx, medicine); err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrDuplicateMedicineCode
		}
		return nil, err
	}
	return s.medicines.FindByID(ctx, medicine.ID)
}

// SoftDelete deactivates a medicine.
func (s *MedicineService) SoftDelete(ctx context.Context, id uuid.UUID) error {
	medicine, err := s.medicines.FindByID(ctx, id)
	if err != nil {
		return err
	}
	medicine.IsActive = false
	return s.medicines.Update(ctx, medicine)
}
