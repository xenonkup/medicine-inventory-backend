package domain

import "github.com/google/uuid"

// TxType is the direction of a stock movement.
type TxType string

const (
	TxIn     TxType = "IN"
	TxOut    TxType = "OUT"
	TxReturn TxType = "RETURN"
)

// StockTransaction is one row of the append-only movement ledger. Quantity is
// always positive; direction is given by Type. IN and RETURN increase a lot's
// remaining quantity, OUT decreases it. Rows are never updated or deleted —
// corrections are new opposing entries.
type StockTransaction struct {
	Base
	LotID       uuid.UUID `gorm:"type:uuid;not null;index" json:"lot_id"`
	MedicineID  uuid.UUID `gorm:"type:uuid;not null;index" json:"medicine_id"`
	Type        TxType    `gorm:"type:varchar(10);not null;index" json:"type"`
	Quantity    int       `gorm:"not null;check:quantity > 0" json:"quantity"`
	ReferenceNo *string   `gorm:"type:varchar(80)" json:"reference_no,omitempty"`
	Note        *string   `gorm:"type:text" json:"note,omitempty"`
	CreatedByID uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
}

// TableName pins the table name.
func (StockTransaction) TableName() string { return "stock_transactions" }
