package data

import (
	"context"

	"generic-mock/channel/photonpay/biz"
	"generic-mock/model"

	"github.com/samber/do"
)

type cardTransactionRepository struct {
	repository *Repository
}

func NewCardTransactionRepository(injector *do.Injector) (biz.CardTransactionRepository, error) {
	return &cardTransactionRepository{
		repository: do.MustInvoke[*Repository](injector),
	}, nil
}

func (r *cardTransactionRepository) ListTransactions(ctx context.Context, req *biz.ListRequest) ([]*model.CardTransaction, error) {
	db := r.repository.DB(ctx)

	return db.CardTransaction.WithContext(ctx).
		Order(db.CardTransaction.ID.Desc()).
		Offset(req.Offset).
		Limit(req.Limit).
		Find()
}

var _ biz.CardTransactionRepository = (*cardTransactionRepository)(nil)
