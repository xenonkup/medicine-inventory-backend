package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"pharmacy-backend/internal/domain"
)

// UserRepository abstracts persistence for users so services can be tested
// against a mock without a live database.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	List(ctx context.Context, offset, limit int) ([]domain.User, int64, error)
	CountAll(ctx context.Context) (int64, error)
}

// CategoryFilter narrows a category listing.
type CategoryFilter struct {
	Search     string
	ActiveOnly bool
	Offset     int
	Limit      int
}

// CategoryRepository abstracts persistence for categories.
type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	Update(ctx context.Context, category *domain.Category) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Category, error)
	List(ctx context.Context, f CategoryFilter) ([]domain.Category, int64, error)
	CountActiveMedicines(ctx context.Context, categoryID uuid.UUID) (int64, error)
}

// MedicineFilter narrows a medicine listing.
type MedicineFilter struct {
	Search     string
	CategoryID *uuid.UUID
	ActiveOnly bool
	Offset     int
	Limit      int
}

// MedicineRepository abstracts persistence for medicines.
type MedicineRepository interface {
	Create(ctx context.Context, medicine *domain.Medicine) error
	Update(ctx context.Context, medicine *domain.Medicine) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Medicine, error)
	List(ctx context.Context, f MedicineFilter) ([]domain.Medicine, int64, error)
}

// LotRepository abstracts persistence for lots. Reads/writes honour any
// transaction carried on the context (see WithTx).
type LotRepository interface {
	Create(ctx context.Context, lot *domain.Lot) error
	Save(ctx context.Context, lot *domain.Lot) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Lot, error)
	FindByMedicineAndLotNumber(ctx context.Context, medicineID uuid.UUID, lotNumber string) (*domain.Lot, error)
	// FindAvailableForUpdate returns lots with remaining stock for a medicine,
	// ordered for FEFO (earliest expiry first), locked FOR UPDATE.
	FindAvailableForUpdate(ctx context.Context, medicineID uuid.UUID) ([]domain.Lot, error)
	ListByMedicine(ctx context.Context, medicineID uuid.UUID) ([]domain.Lot, error)
	SumRemaining(ctx context.Context, medicineID uuid.UUID) (int, error)
	SumRemainingByMedicineIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]int, error)
	// FindNearExpiry returns lots of active medicines that still have stock and
	// expire on or before asOf + withinDays, earliest expiry first (Medicine preloaded).
	FindNearExpiry(ctx context.Context, asOf time.Time, withinDays int) ([]domain.Lot, error)
	// StockByCategory returns derived stock totals grouped by active category.
	StockByCategory(ctx context.Context) ([]CategoryStock, error)
}

// TransactionFilter narrows a ledger listing.
type TransactionFilter struct {
	MedicineID *uuid.UUID
	Type       *domain.TxType
	Offset     int
	Limit      int
}

// TypeAggregate is a per-type rollup of ledger entries over a period.
type TypeAggregate struct {
	Type     domain.TxType
	Count    int64
	TotalQty int64
}

// TransactionRepository abstracts persistence for the append-only ledger.
type TransactionRepository interface {
	Create(ctx context.Context, txn *domain.StockTransaction) error
	List(ctx context.Context, f TransactionFilter) ([]TransactionWithDetails, int64, error)
	CountSince(ctx context.Context, since time.Time) (int64, error)
	AggregateByType(ctx context.Context, from, to time.Time) ([]TypeAggregate, error)
}

// CategoryStock is the derived stock total for one category.
type CategoryStock struct {
	Category string
	Stock    int
}

// SettingRepository abstracts persistence for key/value system settings.
type SettingRepository interface {
	Get(ctx context.Context, key string) (*domain.Setting, error)
	List(ctx context.Context) ([]domain.Setting, error)
	Upsert(ctx context.Context, setting *domain.Setting) error
}
