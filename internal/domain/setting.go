package domain

import (
	"time"

	"github.com/google/uuid"
)

// Setting is a single configurable key/value (e.g. near_expiry_days). Keyed by
// its name rather than a UUID, since callers look settings up by key.
type Setting struct {
	Key       string     `gorm:"type:varchar(80);primaryKey" json:"key"`
	Value     string     `gorm:"type:text;not null" json:"value"`
	UpdatedBy *uuid.UUID `gorm:"type:uuid" json:"updated_by,omitempty"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// TableName pins the table name.
func (Setting) TableName() string { return "system_settings" }

// Known setting keys.
const (
	SettingNearExpiryDays = "near_expiry_days"
)
