package registry

import (
	"context"

	"gorm.io/gorm"
)

type contextKey string

const txKey contextKey = "gorm_tx"

// db returns the database connection with context already attached
func (engine *Engine) db(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return tx.WithContext(ctx) // Transaction with context
	}
	return engine.DatabaseClient.WithContext(ctx) // Default DB with context
}

// Transaction executes a function within a database transaction
func (engine *Engine) Transaction(ctx context.Context, fn func(context.Context) error) error {
	return engine.DatabaseClient.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Put transaction in context
		txCtx := context.WithValue(ctx, txKey, tx)
		return fn(txCtx)
	})
}