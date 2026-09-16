package data

import (
	"context"

	"generic-mock/channel/photonpay/biz"
	"generic-mock/enums"
	"generic-mock/model"

	"github.com/samber/do"
)

type cardRepository struct {
	repository *Repository
}

func NewCardRepository(injector *do.Injector) (biz.CardRepository, error) {
	return &cardRepository{
		repository: do.MustInvoke[*Repository](injector),
	}, nil
}

func (r *cardRepository) CreateCard(ctx context.Context, card *model.Card) error {
	return r.repository.DB(ctx).Card.WithContext(ctx).Create(card)
}

func (r *cardRepository) ExistCardByID(ctx context.Context, id model.ID) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.Card.WithContext(ctx).
		Where(
			db.Card.ID.Eq(id),
			db.Card.Channel.Eq(string(enums.Channel_PhotonPay)),
		).
		Count()

	return count > 0, err
}

func (r *cardRepository) FindCardByID(ctx context.Context, id model.ID) (*model.Card, error) {
	db := r.repository.DB(ctx)

	return db.Card.WithContext(ctx).
		Where(
			db.Card.ID.Eq(id),
			db.Card.Channel.Eq(string(enums.Channel_PhotonPay)),
		).
		First()
}

func (r *cardRepository) ExistCardByRequestID(ctx context.Context, requestID string) (bool, error) {
	db := r.repository.DB(ctx)
	count, err := db.Card.WithContext(ctx).
		Where(
			db.Card.RequestID.Eq(requestID),
			db.Card.Channel.Eq(string(enums.Channel_PhotonPay)),
		).
		Count()

	return count > 0, err
}

func (r *cardRepository) FindByRequestID(ctx context.Context, requestID string) (*model.Card, error) {
	db := r.repository.DB(ctx)

	return db.Card.WithContext(ctx).
		Where(
			db.Card.RequestID.Eq(requestID),
			db.Card.Channel.Eq(string(enums.Channel_PhotonPay)),
		).
		First()
}

func (r *cardRepository) SaveCard(ctx context.Context, card *model.Card) error {
	return r.repository.DB(ctx).Card.WithContext(ctx).Save(card)
}

var _ biz.CardRepository = (*cardRepository)(nil)
