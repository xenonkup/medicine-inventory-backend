package dto

import (
	"time"

	"pharmacy-backend/internal/domain"
)

// UpdateSettingRequest is the body for PUT /settings/{key}.
type UpdateSettingRequest struct {
	Value string `json:"value" binding:"required"`
}

// SettingResponse is the public view of a setting.
type SettingResponse struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	UpdatedAt string `json:"updated_at"`
}

// NewSettingResponse maps a domain setting to its response shape.
func NewSettingResponse(s *domain.Setting) SettingResponse {
	return SettingResponse{
		Key:       s.Key,
		Value:     s.Value,
		UpdatedAt: s.UpdatedAt.Format(time.RFC3339),
	}
}
