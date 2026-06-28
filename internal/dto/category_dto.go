package dto

import "pharmacy-backend/internal/domain"

// CreateCategoryRequest is the body for POST /categories.
type CreateCategoryRequest struct {
	Name        string  `json:"name" binding:"required,max=100"`
	Description *string `json:"description"`
}

// UpdateCategoryRequest is the body for PUT /categories/{id}.
type UpdateCategoryRequest struct {
	Name        string  `json:"name" binding:"required,max=100"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}

// CategoryResponse is the public view of a category.
type CategoryResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	IsActive    bool    `json:"is_active"`
}

// NewCategoryResponse maps a domain category to its response shape.
func NewCategoryResponse(c *domain.Category) CategoryResponse {
	return CategoryResponse{
		ID:          c.ID.String(),
		Name:        c.Name,
		Description: c.Description,
		IsActive:    c.IsActive,
	}
}
