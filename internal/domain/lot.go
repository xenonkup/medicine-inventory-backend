package domain

import (
	"time"

	"github.com/google/uuid"
)

// Lot is a received batch of a medicine. Each lot carries its own expiry and a
// running balance (QtyRemaining). Stock-on-hand for a medicine is the sum of
// its lots' QtyRemaining. This is the backbone of FEFO and near-expiry alerts.
type Lot struct {
	Base
	MedicineID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_lot_medicine_number;index" json:"medicine_id"`
	Medicine     *Medicine `gorm:"foreignKey:MedicineID;constraint:OnDelete:RESTRICT" json:"medicine,omitempty"`
	LotNumber    string    `gorm:"type:varchar(60);not null;uniqueIndex:idx_lot_medicine_number" json:"lot_number"`
	ExpiryDate   time.Time `gorm:"type:date;not null;index" json:"expiry_date"`
	QtyReceived  int       `gorm:"not null;check:qty_received > 0" json:"qty_received"`
	QtyRemaining int       `gorm:"not null;check:qty_remaining >= 0" json:"qty_remaining"`
	ReceivedDate time.Time `gorm:"type:date;not null" json:"received_date"`
}

// TableName pins the table name.
func (Lot) TableName() string { return "lots" }

// IsExpired reports whether the lot is past its expiry date (date-only compare).
func (l *Lot) IsExpired(now time.Time) bool {
	return l.ExpiryDate.Before(truncateDay(now))
}

func truncateDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
