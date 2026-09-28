package data

import (
	"context"

	"generic-mock/internal/query"
	"generic-mock/pkg/gormx"

	"github.com/samber/do/v2"
	"gorm.io/gorm"
)

type SlashRepository struct {
	db *gorm.DB
}

func NewRepository(injector do.Injector) (*SlashRepository, error) {
	return &SlashRepository{
		db: do.MustInvoke[*gorm.DB](injector),
	}, nil
}

func (r *SlashRepository) DB(ctx context.Context) *query.Query {
	return query.Use(gormx.DB(ctx, r.db))
}
