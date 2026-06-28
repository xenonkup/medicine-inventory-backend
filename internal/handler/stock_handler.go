package handler

import (
	"github.com/gin-gonic/gin"

	"pharmacy-backend/internal/domain"
	"pharmacy-backend/internal/dto"
	"pharmacy-backend/internal/middleware"
	"pharmacy-backend/internal/repository"
	"pharmacy-backend/internal/service"
	"pharmacy-backend/pkg/response"
)

// StockHandler exposes inventory movement endpoints.
type StockHandler struct {
	stock *service.StockService
}

// NewStockHandler builds a StockHandler.
func NewStockHandler(stock *service.StockService) *StockHandler {
	return &StockHandler{stock: stock}
}

// StockIn godoc
// @Summary  Receive stock into a lot
// @Tags     Inventory
// @Security BearerAuth
// @Router   /stock/in [post]
func (h *StockHandler) StockIn(c *gin.Context) {
	var req dto.StockInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}
	lot, err := h.stock.StockIn(c.Request.Context(), req, middleware.UserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, lot)
}

// StockOut godoc
// @Summary  Dispense stock (FEFO)
// @Tags     Inventory
// @Security BearerAuth
// @Router   /stock/out [post]
func (h *StockHandler) StockOut(c *gin.Context) {
	var req dto.StockOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}
	res, err := h.stock.StockOut(c.Request.Context(), req, middleware.UserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, res)
}

// Return godoc
// @Summary  Return stock into a lot
// @Tags     Inventory
// @Security BearerAuth
// @Router   /stock/return [post]
func (h *StockHandler) Return(c *gin.Context) {
	var req dto.StockReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}
	lot, err := h.stock.Return(c.Request.Context(), req, middleware.UserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, lot)
}

// Transactions godoc
// @Summary  List ledger entries
// @Tags     Inventory
// @Security BearerAuth
// @Router   /stock/transactions [get]
func (h *StockHandler) Transactions(c *gin.Context) {
	page, pageSize := paginationParams(c)
	filter := repository.TransactionFilter{
		Offset: (page - 1) * pageSize,
		Limit:  pageSize,
	}
	if raw := c.Query("medicine_id"); raw != "" {
		if id, ok := parseUUID(raw); ok {
			filter.MedicineID = &id
		}
	}
	if raw := c.Query("type"); raw != "" {
		t := domain.TxType(raw)
		filter.Type = &t
	}

	txns, total, err := h.stock.Transactions(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, err)
		return
	}
	out := make([]dto.TransactionResponse, 0, len(txns))
	for i := range txns {
		out = append(out, dto.NewTransactionResponse(&txns[i]))
	}
	response.Paginated(c, out, page, pageSize, total)
}

// LotsByMedicine godoc
// @Summary  List lots of a medicine
// @Tags     Inventory
// @Security BearerAuth
// @Router   /medicines/{id}/lots [get]
func (h *StockHandler) LotsByMedicine(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	lots, err := h.stock.LotsByMedicine(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	out := make([]dto.LotResponse, 0, len(lots))
	for i := range lots {
		out = append(out, dto.NewLotResponse(&lots[i]))
	}
	response.OK(c, out)
}
