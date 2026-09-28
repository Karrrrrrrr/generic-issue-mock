package data

import (
	"context"

	"generic-mock/internal/query"
	"generic-mock/pkg/gormx"

	"github.com/samber/do/v2"
	"gorm.io/gorm"
)

type PingPongRepository struct{ db *gorm.DB }

func NewRepository(injector do.Injector) (*PingPongRepository, error) {
	return &PingPongRepository{db: do.MustInvoke[*gorm.DB](injector)}, nil
}

func (repo *PingPongRepository) DB(ctx context.Context) *query.Query {
	return query.Use(gormx.DB(ctx, repo.db))
}
