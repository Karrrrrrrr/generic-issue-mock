package data

import (
	"context"

	"generic-mock/channel/pingpong/biz"
	"generic-mock/model"

	"github.com/samber/do"
)

type cardTransactionRepository struct{ *PingPongRepository }

func NewCardTransactionRepository(injector *do.Injector) (biz.PingPongCardTransactionRepository, error) {
	return &cardTransactionRepository{PingPongRepository: do.MustInvoke[*PingPongRepository](injector)}, nil
}

func (repo *cardTransactionRepository) Create(ctx context.Context, item *model.CardTransaction) error {
	return repo.DB(ctx).CardTransaction.WithContext(ctx).Create(item)
}
