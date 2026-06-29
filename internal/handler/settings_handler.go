package handler

import (
	"github.com/gin-gonic/gin"

	"pharmacy-backend/internal/dto"
	"pharmacy-backend/internal/middleware"
	"pharmacy-backend/internal/service"
	"pharmacy-backend/pkg/response"
)

// SettingsHandler exposes system settings endpoints (Admin).
type SettingsHandler struct {
	settings *service.SettingsService
}

// NewSettingsHandler builds a SettingsHandler.
func NewSettingsHandler(settings *service.SettingsService) *SettingsHandler {
	return &SettingsHandler{settings: settings}
}

// List godoc
// @Summary  List system settings
// @Tags     Settings
// @Security BearerAuth
// @Router   /settings [get]
func (h *SettingsHandler) List(c *gin.Context) {
	settings, err := h.settings.List(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	out := make([]dto.SettingResponse, 0, len(settings))
	for i := range settings {
		out = append(out, dto.NewSettingResponse(&settings[i]))
	}
	response.OK(c, out)
}

// Update godoc
// @Summary  Update a system setting
// @Tags     Settings
// @Security BearerAuth
// @Router   /settings/{key} [put]
func (h *SettingsHandler) Update(c *gin.Context) {
	key := c.Param("key")
	var req dto.UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}
	setting, err := h.settings.Set(c.Request.Context(), key, req.Value, middleware.UserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.NewSettingResponse(setting))
}
