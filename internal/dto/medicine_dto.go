package dto

import "pharmacy-backend/internal/domain"

// CreateMedicineRequest is the body for POST /medicines.
type CreateMedicineRequest struct {
	Code         string  `json:"code" binding:"required,max=50"`
	Name         string  `json:"name" binding:"required,max=150"`
	CategoryID   string  `json:"category_id" binding:"required,uuid"`
	Unit         string  `json:"unit" binding:"required,max=30"`
	ReorderLevel int     `json:"reorder_level" binding:"gte=0"`
	Description  *string `json:"description"`
}

// UpdateMedicineRequest is the body for PUT /medicines/{id}.
type UpdateMedicineRequest struct {
	Code         string  `json:"code" binding:"required,max=50"`
	Name         string  `json:"name" binding:"required,max=150"`
	CategoryID   string  `json:"category_id" binding:"required,uuid"`
	Unit         string  `json:"unit" binding:"required,max=30"`
	ReorderLevel int     `json:"reorder_level" binding:"gte=0"`
	Description  *string `json:"description"`
	IsActive     *bool   `json:"is_active"`
}

// MedicineResponse is the public view of a medicine. StockOnHand is the derived
// balance (sum of remaining lots); it stays 0 until Phase 3 introduces lots.
type MedicineResponse struct {
	ID           string  `json:"id"`
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name,omitempty"`
	Unit         string  `json:"unit"`
	ReorderLevel int     `json:"reorder_level"`
	StockOnHand  int     `json:"stock_on_hand"`
	Description  *string `json:"description,omitempty"`
	IsActive     bool    `json:"is_active"`
}

// NewMedicineResponse maps a domain medicine to its response shape.
func NewMedicineResponse(m *domain.Medicine, stockOnHand int) MedicineResponse {
	resp := MedicineResponse{
		ID:           m.ID.String(),
		Code:         m.Code,
		Name:         m.Name,
		CategoryID:   m.CategoryID.String(),
		Unit:         m.Unit,
		ReorderLevel: m.ReorderLevel,
		StockOnHand:  stockOnHand,
		Description:  m.Description,
		IsActive:     m.IsActive,
	}
	if m.Category != nil {
		resp.CategoryName = m.Category.Name
	}
	return resp
}
