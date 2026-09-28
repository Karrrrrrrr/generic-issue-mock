package data

import (
	"context"

	"generic-mock/internal/query"
	"generic-mock/pkg/gormx"
	"generic-mock/shared/biz"
	sharederrors "generic-mock/shared/errors"

	"github.com/samber/do"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(injector *do.Injector) (*Repository, error) {
	return &Repository{db: do.MustInvoke[*gorm.DB](injector)}, nil
}

func (repo *Repository) DB(ctx context.Context) *query.Query {
	return query.Use(gormx.DB(ctx, repo.db))
}

type transaction struct {
	repo *Repository
}

func NewTransaction(injector *do.Injector) (biz.Transaction, error) {
	return &transaction{repo: do.MustInvoke[*Repository](injector)}, nil
}

func (tx *transaction) InTx(ctx context.Context, run func(context.Context) error) error {
	if gormx.DB(ctx, tx.repo.db) != tx.repo.db {
		return sharederrors.ErrNestedTransaction
	}
	return gormx.InTx(ctx, tx.repo.db, run)
}
