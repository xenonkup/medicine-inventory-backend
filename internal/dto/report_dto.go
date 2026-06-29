package dto

// MovementByType is one row of the monthly movement breakdown.
type MovementByType struct {
	Type     string `json:"type"`
	Count    int64  `json:"count"`
	TotalQty int64  `json:"total_qty"`
}

// MonthlyReport summarises stock movements for a given month.
type MonthlyReport struct {
	Year      int              `json:"year"`
	Month     int              `json:"month"`
	Movements []MovementByType `json:"movements"`
	TotalIn   int64            `json:"total_in"`
	TotalOut  int64            `json:"total_out"`
	TotalRet  int64            `json:"total_return"`
}

// CategoryStockItem is the derived stock total for one category (for charts).
type CategoryStockItem struct {
	Category string `json:"category"`
	Stock    int    `json:"stock"`
}
