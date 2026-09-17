package data

import (
	"context"

	"generic-mock/channel/paynda/biz"
	"generic-mock/pkg/gormx"

	"github.com/samber/do"
)

type transaction struct {
	repository *PayndaRepository
}

var _ biz.PayndaTransaction = (*transaction)(nil)

func NewPayndaTransaction(injector *do.Injector) (biz.PayndaTransaction, error) {
	return &transaction{
		repository: do.MustInvoke[*PayndaRepository](injector),
	}, nil
}

func (t *transaction) InTx(ctx context.Context, fn func(context.Context) error) error {
	return gormx.InTx(ctx, t.repository.db, fn)
}
