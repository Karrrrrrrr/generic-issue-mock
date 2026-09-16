package data

import (
	"context"

	"generic-mock/channel/photonpay/biz"
	"generic-mock/enums"
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
		Where(db.CardTransaction.Channel.Eq(string(enums.Channel_PhotonPay))).
		Order(db.CardTransaction.ID.Desc()).
		Offset(req.Offset).
		Limit(req.Limit).
		Find()
}

var _ biz.CardTransactionRepository = (*cardTransactionRepository)(nil)
