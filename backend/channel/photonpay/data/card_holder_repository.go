package data

import (
	"context"

	"generic-mock/channel/photonpay/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do"
)

type cardHolderRepository struct {
	repository *Repository
}

func NewCardHolderRepository(injector *do.Injector) (biz.CardHolderRepository, error) {
	return &cardHolderRepository{
		repository: do.MustInvoke[*Repository](injector),
	}, nil
}

func (r *cardHolderRepository) Create(ctx context.Context, holder *model.CardHolder) error {
	return r.repository.DB(ctx).CardHolder.WithContext(ctx).Create(holder)
}

func (r *cardHolderRepository) ExistCardHolderByID(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.CardHolder.WithContext(ctx).
		Where(
			db.CardHolder.ID.Eq(id),
			db.CardHolder.Channel.Eq(string(enums.Channel_PhotonPay)),
		).
		Count()

	return count > 0, err
}

func (r *cardHolderRepository) FindCardHolderByID(ctx context.Context, id model.ID) (*model.CardHolder, error) {
	db := r.repository.DB(ctx)

	return db.CardHolder.WithContext(ctx).
		Where(
			db.CardHolder.ID.Eq(id),
			db.CardHolder.Channel.Eq(string(enums.Channel_PhotonPay)),
		).
		First()
}

func (r *cardHolderRepository) Save(ctx context.Context, holder *model.CardHolder) error {
	return r.repository.DB(ctx).CardHolder.WithContext(ctx).Save(holder)
}

func (r *cardHolderRepository) List(ctx context.Context, req *biz.ListRequest) ([]*model.CardHolder, error) {
	db := r.repository.DB(ctx)

	return db.CardHolder.WithContext(ctx).
		Where(db.CardHolder.Channel.Eq(string(enums.Channel_PhotonPay))).
		Order(db.CardHolder.ID.Desc()).
		Offset(req.Offset).
		Limit(req.Limit).
		Find()
}

var _ biz.CardHolderRepository = (*cardHolderRepository)(nil)
