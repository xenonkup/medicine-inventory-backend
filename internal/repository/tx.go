package repository

import (
	"context"

	"gorm.io/gorm"
)

// txKey is the context key under which an active *gorm.DB transaction is stored.
type txKey struct{}

// WithTx returns a context carrying the given transaction handle.
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// dbFromCtx returns the transaction bound to ctx if present, otherwise fallback.
// This lets repositories transparently participate in a service-managed
// transaction without changing their method signatures.
func dbFromCtx(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok && tx != nil {
		return tx.WithContext(ctx)
	}
	return fallback.WithContext(ctx)
}

// TxManager runs a function inside a database transaction. The function receives
// a context that carries the transaction, so repositories called within it use
// the same transaction. Either everything commits or everything rolls back.
type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type gormTxManager struct {
	db *gorm.DB
}

// NewTxManager builds a GORM-backed TxManager.
func NewTxManager(db *gorm.DB) TxManager {
	return &gormTxManager{db: db}
}

func (m *gormTxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		return fn(WithTx(ctx, tx))
	})
}
