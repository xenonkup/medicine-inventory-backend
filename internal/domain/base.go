package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base is embedded by every entity to provide a UUID primary key and timestamps.
// The UUID is generated in the application layer so we don't depend on a specific
// PostgreSQL extension being enabled.
type Base struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate assigns a new UUID if one was not provided.
func (b *Base) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
