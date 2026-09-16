package data

import (
	"context"

	"generic-mock/internal/query"
	"generic-mock/pkg/gormx"

	"github.com/samber/do"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(injector *do.Injector) (*Repository, error) {
	return &Repository{
		db: do.MustInvoke[*gorm.DB](injector),
	}, nil
}

func (r *Repository) DB(ctx context.Context) *query.Query {
	return query.Use(gormx.DB(ctx, r.db))
}
