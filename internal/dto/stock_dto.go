package dto

import (
	"time"

	"pharmacy-backend/internal/domain"
)

const dateLayout = "2006-01-02"

// StockInRequest receives a batch into a (new or existing) lot.
type StockInRequest struct {
	MedicineID   string  `json:"medicine_id" binding:"required,uuid"`
	LotNumber    string  `json:"lot_number" binding:"required,max=60"`
	ExpiryDate   string  `json:"expiry_date" binding:"required"` // YYYY-MM-DD
	Quantity     int     `json:"quantity" binding:"required,gt=0"`
	ReceivedDate *string `json:"received_date"`                  // YYYY-MM-DD, defaults today
	ReferenceNo  *string `json:"reference_no"`
	Note         *string `json:"note"`
}

// StockOutRequest dispenses a quantity, allocated across lots via FEFO.
type StockOutRequest struct {
	MedicineID  string  `json:"medicine_id" binding:"required,uuid"`
	Quantity    int     `json:"quantity" binding:"required,gt=0"`
	ReferenceNo *string `json:"reference_no"`
	Note        *string `json:"note"`
}

// StockReturnRequest returns a quantity back into a specific lot.
type StockReturnRequest struct {
	LotID       string  `json:"lot_id" binding:"required,uuid"`
	Quantity    int     `json:"quantity" binding:"required,gt=0"`
	ReferenceNo *string `json:"reference_no"`
	Note        *string `json:"note"`
}

// LotResponse is the public view of a lot.
type LotResponse struct {
	ID           string `json:"id"`
	MedicineID   string `json:"medicine_id"`
	LotNumber    string `json:"lot_number"`
	ExpiryDate   string `json:"expiry_date"`
	QtyReceived  int    `json:"qty_received"`
	QtyRemaining int    `json:"qty_remaining"`
	ReceivedDate string `json:"received_date"`
}

// Allocation describes how much was taken from one lot during a FEFO stock-out.
type Allocation struct {
	LotID      string `json:"lot_id"`
	LotNumber  string `json:"lot_number"`
	ExpiryDate string `json:"expiry_date"`
	Deducted   int    `json:"deducted"`
}

// StockOutResponse summarises a FEFO dispense.
type StockOutResponse struct {
	MedicineID    string       `json:"medicine_id"`
	TotalQuantity int          `json:"total_quantity"`
	Allocations   []Allocation `json:"allocations"`
}

// TransactionResponse is the public view of a ledger entry.
type TransactionResponse struct {
	ID          string  `json:"id"`
	LotID       string  `json:"lot_id"`
	MedicineID  string  `json:"medicine_id"`
	Type        string  `json:"type"`
	Quantity    int     `json:"quantity"`
	ReferenceNo *string `json:"reference_no,omitempty"`
	Note        *string `json:"note,omitempty"`
	CreatedBy   string  `json:"created_by"`
	CreatedAt   string  `json:"created_at"`
}

// ParseDate parses a YYYY-MM-DD string into a time.Time (UTC midnight).
func ParseDate(s string) (time.Time, error) {
	return time.ParseInLocation(dateLayout, s, time.UTC)
}

// NewLotResponse maps a domain lot to its response shape.
func NewLotResponse(l *domain.Lot) LotResponse {
	return LotResponse{
		ID:           l.ID.String(),
		MedicineID:   l.MedicineID.String(),
		LotNumber:    l.LotNumber,
		ExpiryDate:   l.ExpiryDate.Format(dateLayout),
		QtyReceived:  l.QtyReceived,
		QtyRemaining: l.QtyRemaining,
		ReceivedDate: l.ReceivedDate.Format(dateLayout),
	}
}

// NewTransactionResponse maps a domain ledger entry to its response shape.
func NewTransactionResponse(t *domain.StockTransaction) TransactionResponse {
	return TransactionResponse{
		ID:          t.ID.String(),
		LotID:       t.LotID.String(),
		MedicineID:  t.MedicineID.String(),
		Type:        string(t.Type),
		Quantity:    t.Quantity,
		ReferenceNo: t.ReferenceNo,
		Note:        t.Note,
		CreatedBy:   t.CreatedByID.String(),
		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
	}
}
