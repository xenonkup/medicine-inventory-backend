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

	movements, totalIn, totalOut, totalRet, err := s.movements(ctx, from, to)
	if err != nil {
		return nil, err
	}
	return &dto.MonthlyReport{
		Year:      year,
		Month:     int(month),
		Movements: movements,
		TotalIn:   totalIn,
		TotalOut:  totalOut,
		TotalRet:  totalRet,
	}, nil
}

// Range returns the movement breakdown for an arbitrary [from, to) date range.
// `to` is treated as inclusive by the caller (the handler adds one day).
func (s *ReportService) Range(ctx context.Context, from, to time.Time) (*dto.MovementReport, error) {
	movements, totalIn, totalOut, totalRet, err := s.movements(ctx, from, to)
	if err != nil {
		return nil, err
	}
	return &dto.MovementReport{
		From:      from.Format("2006-01-02"),
		To:        to.AddDate(0, 0, -1).Format("2006-01-02"),
		Movements: movements,
		TotalIn:   totalIn,
		TotalOut:  totalOut,
		TotalRet:  totalRet,
	}, nil
}

// movements aggregates ledger entries in [from, to) into the three movement
// types (zero-filled for stable charts) plus per-type totals.
func (s *ReportService) movements(ctx context.Context, from, to time.Time) ([]dto.MovementByType, int64, int64, int64, error) {
	aggs, err := s.txns.AggregateByType(ctx, from, to)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	byType := map[domain.TxType]repository.TypeAggregate{}
	for _, a := range aggs {
		byType[a.Type] = a
	}

	var movements []dto.MovementByType
	var totalIn, totalOut, totalRet int64
	for _, t := range []domain.TxType{domain.TxIn, domain.TxOut, domain.TxReturn} {
		a := byType[t]
		movements = append(movements, dto.MovementByType{
			Type:     string(t),
			Count:    a.Count,
			TotalQty: a.TotalQty,
		})
		switch t {
		case domain.TxIn:
			totalIn = a.TotalQty
		case domain.TxOut:
			totalOut = a.TotalQty
		case domain.TxReturn:
			totalRet = a.TotalQty
		}
	}
	return movements, totalIn, totalOut, totalRet, nil
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
