package gormx

import (
	"context"

	"gorm.io/gorm"
)

type transactionContextKey struct{}

func InTx(ctx context.Context, db *gorm.DB, fn func(context.Context) error) error {
	return DB(ctx, db).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, transactionContextKey{}, tx))
	})
}

func DB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	tx, ok := ctx.Value(transactionContextKey{}).(*gorm.DB)
	if ok {
		return tx
	}

	return defaultDB
}
