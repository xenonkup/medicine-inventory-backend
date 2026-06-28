package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"pharmacy-backend/internal/dto"
	"pharmacy-backend/internal/repository"
	"pharmacy-backend/internal/service"
	"pharmacy-backend/pkg/response"
)

// MedicineHandler exposes medicine endpoints.
type MedicineHandler struct {
	medicines *service.MedicineService
}

// NewMedicineHandler builds a MedicineHandler.
func NewMedicineHandler(medicines *service.MedicineService) *MedicineHandler {
	return &MedicineHandler{medicines: medicines}
}

// List godoc
// @Summary  List medicines
// @Tags     Medicines
// @Security BearerAuth
// @Router   /medicines [get]
func (h *MedicineHandler) List(c *gin.Context) {
	page, pageSize := paginationParams(c)
	filter := repository.MedicineFilter{
		Search:     c.Query("search"),
		ActiveOnly: c.Query("active") == "true",
		Offset:     (page - 1) * pageSize,
		Limit:      pageSize,
	}
	if raw := c.Query("category_id"); raw != "" {
		if cid, err := uuid.Parse(raw); err == nil {
			filter.CategoryID = &cid
		}
	}

	ctx := c.Request.Context()
	medicines, total, err := h.medicines.List(ctx, filter)
	if err != nil {
		response.Error(c, err)
		return
	}

	ids := make([]uuid.UUID, 0, len(medicines))
	for i := range medicines {
		ids = append(ids, medicines[i].ID)
	}
	stock, err := h.medicines.StockOnHand(ctx, ids)
	if err != nil {
		response.Error(c, err)
		return
	}

	out := make([]dto.MedicineResponse, 0, len(medicines))
	for i := range medicines {
		out = append(out, dto.NewMedicineResponse(&medicines[i], stock[medicines[i].ID]))
	}
	response.Paginated(c, out, page, pageSize, total)
}

// Create godoc
// @Summary  Create medicine
// @Tags     Medicines
// @Security BearerAuth
// @Router   /medicines [post]
func (h *MedicineHandler) Create(c *gin.Context) {
	var req dto.CreateMedicineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}
	medicine, err := h.medicines.Create(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, dto.NewMedicineResponse(medicine, 0))
}

// Get godoc
// @Summary  Get medicine by id
// @Tags     Medicines
// @Security BearerAuth
// @Router   /medicines/{id} [get]
func (h *MedicineHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	medicine, err := h.medicines.GetByID(ctx, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	stock, err := h.medicines.StockOnHand(ctx, []uuid.UUID{medicine.ID})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.NewMedicineResponse(medicine, stock[medicine.ID]))
}

// Update godoc
// @Summary  Update medicine
// @Tags     Medicines
// @Security BearerAuth
// @Router   /medicines/{id} [put]
func (h *MedicineHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.UpdateMedicineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}
	medicine, err := h.medicines.Update(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.NewMedicineResponse(medicine, 0))
}

// Delete godoc
// @Summary  Soft-delete medicine
// @Tags     Medicines
// @Security BearerAuth
// @Router   /medicines/{id} [delete]
func (h *MedicineHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.medicines.SoftDelete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"message": "medicine deactivated"})
}
