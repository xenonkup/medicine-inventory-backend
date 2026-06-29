package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"pharmacy-backend/internal/domain"
	"pharmacy-backend/internal/dto"
	"pharmacy-backend/internal/repository"
)

// DashboardService computes dashboard KPIs and alert lists.
type DashboardService struct {
	medicines       repository.MedicineRepository
	lots            repository.LotRepository
	txns            repository.TransactionRepository
	settings        *SettingsService
	defaultExpiryDays int
	now             func() time.Time
}

// NewDashboardService builds a DashboardService. The near-expiry window is read
// from settings at request time, falling back to defaultExpiryDays.
func NewDashboardService(
	medicines repository.MedicineRepository,
	lots repository.LotRepository,
	txns repository.TransactionRepository,
	settings *SettingsService,
	defaultExpiryDays int,
) *DashboardService {
	return &DashboardService{
		medicines:         medicines,
		lots:              lots,
		txns:              txns,
		settings:          settings,
		defaultExpiryDays: defaultExpiryDays,
		now:               time.Now,
	}
}

// nearExpiryDays resolves the configured window (settings override default).
func (s *DashboardService) nearExpiryDays(ctx context.Context) int {
	return s.settings.GetInt(ctx, domain.SettingNearExpiryDays, s.defaultExpiryDays)
}

func startOfToday(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// Summary returns the dashboard KPI counts.
func (s *DashboardService) Summary(ctx context.Context) (*dto.DashboardSummary, error) {
	now := s.now()

	_, totalMedicines, err := s.medicines.List(ctx, repository.MedicineFilter{ActiveOnly: true, Limit: 1})
	if err != nil {
		return nil, err
	}

	nearExpiry, err := s.NearExpiry(ctx)
	if err != nil {
		return nil, err
	}

	lowStock, err := s.LowStock(ctx)
	if err != nil {
		return nil, err
	}

	todayMovements, err := s.txns.CountSince(ctx, startOfToday(now))
	if err != nil {
		return nil, err
	}

	return &dto.DashboardSummary{
		TotalMedicines:  totalMedicines,
		NearExpiryCount: len(nearExpiry),
		LowStockCount:   len(lowStock),
		TodayMovements:  todayMovements,
		NearExpiryDays:  s.nearExpiryDays(ctx),
	}, nil
}

// NearExpiry returns lots expiring within the configured window.
func (s *DashboardService) NearExpiry(ctx context.Context) ([]dto.NearExpiryItem, error) {
	now := startOfToday(s.now())
	lots, err := s.lots.FindNearExpiry(ctx, now, s.nearExpiryDays(ctx))
	if err != nil {
		return nil, err
	}

	items := make([]dto.NearExpiryItem, 0, len(lots))
	for i := range lots {
		l := &lots[i]
		name := ""
		if l.Medicine != nil {
			name = l.Medicine.Name
		}
		daysLeft := int(l.ExpiryDate.Sub(now).Hours() / 24)
		items = append(items, dto.NearExpiryItem{
			MedicineID:   l.MedicineID.String(),
			MedicineName: name,
			LotNumber:    l.LotNumber,
			ExpiryDate:   l.ExpiryDate.Format("2006-01-02"),
			QtyRemaining: l.QtyRemaining,
			DaysLeft:     daysLeft,
		})
	}
	return items, nil
}

// LowStock returns active medicines whose derived stock is at or below their
// reorder level.
func (s *DashboardService) LowStock(ctx context.Context) ([]dto.LowStockItem, error) {
	medicines, _, err := s.medicines.List(ctx, repository.MedicineFilter{ActiveOnly: true, Limit: 1000})
	if err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, 0, len(medicines))
	for i := range medicines {
		ids = append(ids, medicines[i].ID)
	}
	stock, err := s.lots.SumRemainingByMedicineIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	items := make([]dto.LowStockItem, 0)
	for i := range medicines {
		m := &medicines[i]
		onHand := stock[m.ID]
		if onHand <= m.ReorderLevel {
			items = append(items, dto.LowStockItem{
				MedicineID:   m.ID.String(),
				Code:         m.Code,
				Name:         m.Name,
				Unit:         m.Unit,
				StockOnHand:  onHand,
				ReorderLevel: m.ReorderLevel,
			})
		}
	}
	return items, nil
}
