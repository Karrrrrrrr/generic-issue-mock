package data

import (
	"context"

	"generic-mock/channel/slash/biz"
	"generic-mock/pkg/gormx"

	"github.com/samber/do"
)

type transaction struct {
	repository *SlashRepository
}

func NewTransaction(injector *do.Injector) (biz.SlashTransaction, error) {
	return &transaction{
		repository: do.MustInvoke[*SlashRepository](injector),
	}, nil
}

func (t *transaction) InTx(ctx context.Context, fn func(context.Context) error) error {
	return gormx.InTx(ctx, t.repository.db, fn)
}

var _ biz.SlashTransaction = (*transaction)(nil)
