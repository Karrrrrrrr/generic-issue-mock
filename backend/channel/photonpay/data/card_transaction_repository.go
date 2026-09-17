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

func (r *cardTransactionRepository) Create(ctx context.Context, item *model.CardTransaction) error {
	return r.repository.DB(ctx).CardTransaction.WithContext(ctx).Create(item)
}

func (r *cardTransactionRepository) ExistByID(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.CardTransaction.WithContext(ctx).Where(
		db.CardTransaction.ID.Eq(id),
		db.CardTransaction.Channel.Eq(string(enums.Channel_PhotonPay)),
	).Count()

	return count > 0, err
}

func (r *cardTransactionRepository) FindByID(ctx context.Context, id model.ID) (*model.CardTransaction, error) {
	db := r.repository.DB(ctx)

	return db.CardTransaction.WithContext(ctx).Where(
		db.CardTransaction.ID.Eq(id),
		db.CardTransaction.Channel.Eq(string(enums.Channel_PhotonPay)),
	).First()
}

func (r *cardTransactionRepository) ListTransactions(ctx context.Context, req *biz.ListRequest) ([]*model.CardTransaction, error) {
	db := r.repository.DB(ctx)
	query := db.CardTransaction.WithContext(ctx).Where(db.CardTransaction.Channel.Eq(string(enums.Channel_PhotonPay)))
	if req.AccountID != 0 {
		query = query.Where(db.CardTransaction.AccountID.Eq(req.AccountID))
	}
	return query.
		Order(db.CardTransaction.ID.Desc()).
		Offset(req.Offset).
		Limit(req.Limit).
		Find()
}

var _ biz.CardTransactionRepository = (*cardTransactionRepository)(nil)
