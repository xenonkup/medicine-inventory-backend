package service

import (
	"context"
	"time"

	"pharmacy-backend/internal/domain"
	"pharmacy-backend/internal/dto"
	"pharmacy-backend/internal/repository"
)

// ReportService produces aggregated reporting data.
type ReportService struct {
	txns repository.TransactionRepository
	lots repository.LotRepository
}

// NewReportService builds a ReportService.
func NewReportService(txns repository.TransactionRepository, lots repository.LotRepository) *ReportService {
	return &ReportService{txns: txns, lots: lots}
}

// Monthly returns the movement breakdown for the given year/month.
func (s *ReportService) Monthly(ctx context.Context, year int, month time.Month) (*dto.MonthlyReport, error) {
	from := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)

	aggs, err := s.txns.AggregateByType(ctx, from, to)
	if err != nil {
		return nil, err
	}

	report := &dto.MonthlyReport{Year: year, Month: int(month)}
	// Always present all three types (zero-filled) for stable charts.
	byType := map[domain.TxType]repository.TypeAggregate{}
	for _, a := range aggs {
		byType[a.Type] = a
	}
	for _, t := range []domain.TxType{domain.TxIn, domain.TxOut, domain.TxReturn} {
		a := byType[t]
		report.Movements = append(report.Movements, dto.MovementByType{
			Type:     string(t),
			Count:    a.Count,
			TotalQty: a.TotalQty,
		})
		switch t {
		case domain.TxIn:
			report.TotalIn = a.TotalQty
		case domain.TxOut:
			report.TotalOut = a.TotalQty
		case domain.TxReturn:
			report.TotalRet = a.TotalQty
		}
	}
	return report, nil
}

// StockByCategory returns derived stock totals per active category.
func (s *ReportService) StockByCategory(ctx context.Context) ([]dto.CategoryStockItem, error) {
	rows, err := s.lots.StockByCategory(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]dto.CategoryStockItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, dto.CategoryStockItem{Category: r.Category, Stock: r.Stock})
	}
	return items, nil
}
