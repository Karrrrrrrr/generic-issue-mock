package data

import (
	"context"

	"generic-mock/internal/query"
	"generic-mock/pkg/gormx"

	"github.com/samber/do"
	"gorm.io/gorm"
)

type PayndaRepository struct {
	db *gorm.DB
}

func NewPayndaRepository(injector *do.Injector) (*PayndaRepository, error) {
	return &PayndaRepository{
		db: do.MustInvoke[*gorm.DB](injector),
	}, nil
}

func (r *PayndaRepository) DB(ctx context.Context) *query.Query {
	return query.Use(gormx.DB(ctx, r.db))
}
