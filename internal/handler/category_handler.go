package handler

import (
	"github.com/gin-gonic/gin"

	"pharmacy-backend/internal/dto"
	"pharmacy-backend/internal/repository"
	"pharmacy-backend/internal/service"
	"pharmacy-backend/pkg/response"
)

// CategoryHandler exposes category endpoints.
type CategoryHandler struct {
	categories *service.CategoryService
}

// NewCategoryHandler builds a CategoryHandler.
func NewCategoryHandler(categories *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categories: categories}
}

// List godoc
// @Summary  List categories
// @Tags     Categories
// @Security BearerAuth
// @Router   /categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	page, pageSize := paginationParams(c)
	filter := repository.CategoryFilter{
		Search:     c.Query("search"),
		ActiveOnly: c.Query("active") == "true",
		Offset:     (page - 1) * pageSize,
		Limit:      pageSize,
	}
	categories, total, err := h.categories.List(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, err)
		return
	}
	out := make([]dto.CategoryResponse, 0, len(categories))
	for i := range categories {
		out = append(out, dto.NewCategoryResponse(&categories[i]))
	}
	response.Paginated(c, out, page, pageSize, total)
}

// Create godoc
// @Summary  Create category
// @Tags     Categories
// @Security BearerAuth
// @Router   /categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var req dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}
	category, err := h.categories.Create(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, dto.NewCategoryResponse(category))
}

// Get godoc
// @Summary  Get category by id
// @Tags     Categories
// @Security BearerAuth
// @Router   /categories/{id} [get]
func (h *CategoryHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	category, err := h.categories.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.NewCategoryResponse(category))
}

// Update godoc
// @Summary  Update category
// @Tags     Categories
// @Security BearerAuth
// @Router   /categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}
	category, err := h.categories.Update(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.NewCategoryResponse(category))
}

// Delete godoc
// @Summary  Soft-delete category
// @Tags     Categories
// @Security BearerAuth
// @Router   /categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.categories.SoftDelete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"message": "category deactivated"})
}
