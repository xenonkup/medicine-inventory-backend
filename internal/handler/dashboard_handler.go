package handler

import (
	"github.com/gin-gonic/gin"

	"pharmacy-backend/internal/service"
	"pharmacy-backend/pkg/response"
)

// DashboardHandler exposes dashboard and alert endpoints.
type DashboardHandler struct {
	dashboard *service.DashboardService
}

// NewDashboardHandler builds a DashboardHandler.
func NewDashboardHandler(dashboard *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboard: dashboard}
}

// Summary godoc
// @Summary  Dashboard KPI summary
// @Tags     Dashboard
// @Security BearerAuth
// @Router   /dashboard/summary [get]
func (h *DashboardHandler) Summary(c *gin.Context) {
	res, err := h.dashboard.Summary(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, res)
}

// NearExpiry godoc
// @Summary  Lots near expiry
// @Tags     Dashboard
// @Security BearerAuth
// @Router   /dashboard/near-expiry [get]
func (h *DashboardHandler) NearExpiry(c *gin.Context) {
	res, err := h.dashboard.NearExpiry(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, res)
}

// LowStock godoc
// @Summary  Medicines at/below reorder level
// @Tags     Dashboard
// @Security BearerAuth
// @Router   /dashboard/low-stock [get]
func (h *DashboardHandler) LowStock(c *gin.Context) {
	res, err := h.dashboard.LowStock(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, res)
}
