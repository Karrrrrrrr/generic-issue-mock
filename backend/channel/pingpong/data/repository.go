package data

import (
	"context"

	"generic-mock/channel/pingpong/biz"
	"generic-mock/internal/query"
	"generic-mock/pkg/gormx"

	"github.com/samber/do"
	"gorm.io/gorm"
)

type PingPongRepository struct{ db *gorm.DB }

func NewRepository(injector *do.Injector) (*PingPongRepository, error) {
	return &PingPongRepository{db: do.MustInvoke[*gorm.DB](injector)}, nil
}

func (repo *PingPongRepository) DB(ctx context.Context) *query.Query {
	return query.Use(gormx.DB(ctx, repo.db))
}

type transaction struct{ repo *PingPongRepository }

func NewTransaction(injector *do.Injector) (biz.PingPongTransaction, error) {
	return &transaction{repo: do.MustInvoke[*PingPongRepository](injector)}, nil
}

func (tx *transaction) InTx(ctx context.Context, run func(context.Context) error) error {
	return gormx.InTx(ctx, tx.repo.db, run)
}
