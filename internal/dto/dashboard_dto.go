package dto

// DashboardSummary holds the KPI counts shown on the dashboard.
type DashboardSummary struct {
	TotalMedicines  int64 `json:"total_medicines"`
	NearExpiryCount int   `json:"near_expiry_count"`
	LowStockCount   int   `json:"low_stock_count"`
	TodayMovements  int64 `json:"today_movements"`
	NearExpiryDays  int   `json:"near_expiry_days"`
}

// NearExpiryItem is one lot approaching its expiry date.
type NearExpiryItem struct {
	MedicineID   string `json:"medicine_id"`
	MedicineName string `json:"medicine_name"`
	LotNumber    string `json:"lot_number"`
	ExpiryDate   string `json:"expiry_date"`
	QtyRemaining int    `json:"qty_remaining"`
	DaysLeft     int    `json:"days_left"`
}

// LowStockItem is a medicine at or below its reorder level.
type LowStockItem struct {
	MedicineID   string `json:"medicine_id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	Unit         string `json:"unit"`
	StockOnHand  int    `json:"stock_on_hand"`
	ReorderLevel int    `json:"reorder_level"`
}
