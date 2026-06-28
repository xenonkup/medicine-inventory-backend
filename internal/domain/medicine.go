package domain

import "github.com/google/uuid"

// Medicine is a master record. Stock-on-hand is NOT stored here; it is derived
// from the sum of remaining lot quantities (added in Phase 3).
type Medicine struct {
	Base
	Code         string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	Name         string    `gorm:"type:varchar(150);not null" json:"name"`
	CategoryID   uuid.UUID `gorm:"type:uuid;not null;index" json:"category_id"`
	Category     *Category `gorm:"foreignKey:CategoryID;constraint:OnDelete:RESTRICT" json:"category,omitempty"`
	Unit         string    `gorm:"type:varchar(30);not null" json:"unit"`
	ReorderLevel int       `gorm:"not null;default:0;check:reorder_level >= 0" json:"reorder_level"`
	Description  *string   `gorm:"type:text" json:"description,omitempty"`
	IsActive     bool      `gorm:"not null;default:true" json:"is_active"`
}

// TableName pins the table name.
func (Medicine) TableName() string { return "medicines" }
