package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"pharmacy-backend/internal/service"
	"pharmacy-backend/pkg/response"
)

// ReportHandler exposes reporting endpoints.
type ReportHandler struct {
	reports *service.ReportService
}

// NewReportHandler builds a ReportHandler.
func NewReportHandler(reports *service.ReportService) *ReportHandler {
	return &ReportHandler{reports: reports}
}

// Monthly godoc
// @Summary  Monthly movement report
// @Tags     Reports
// @Security BearerAuth
// @Router   /reports/monthly [get]
func (h *ReportHandler) Monthly(c *gin.Context) {
	now := time.Now()
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
	if month < 1 || month > 12 {
		month = int(now.Month())
	}
	if year < 2000 || year > 3000 {
		year = now.Year()
	}

	res, err := h.reports.Monthly(c.Request.Context(), year, time.Month(month))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, res)
}

// Movements godoc
// @Summary  Movement report for a date range
// @Tags     Reports
// @Security BearerAuth
// @Router   /reports/movements [get]
func (h *ReportHandler) Movements(c *gin.Context) {
	const layout = "2006-01-02"
	now := time.Now()

	from, err := time.ParseInLocation(layout, c.Query("from"), time.UTC)
	if err != nil {
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	}
	to, err := time.ParseInLocation(layout, c.Query("to"), time.UTC)
	if err != nil {
		to = now
	}
	// Make `to` inclusive of the whole day.
	to = to.AddDate(0, 0, 1)
	if to.Before(from) {
		from, to = to.AddDate(0, 0, -1), from.AddDate(0, 0, 1)
	}

	res, rerr := h.reports.Range(c.Request.Context(), from, to)
	if rerr != nil {
		response.Error(c, rerr)
		return
	}
	response.OK(c, res)
}

// StockByCategory godoc
// @Summary  Derived stock grouped by category
// @Tags     Reports
// @Security BearerAuth
// @Router   /reports/stock-by-category [get]
func (h *ReportHandler) StockByCategory(c *gin.Context) {
	res, err := h.reports.StockByCategory(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, res)
}
