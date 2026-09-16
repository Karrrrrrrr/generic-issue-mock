package data

import (
	"context"
	"errors"

	"generic-mock/channel/photonpay/biz"
	"generic-mock/internal/query"
	"generic-mock/model"

	"gorm.io/gorm"
)

type Repository struct {
	query *query.Query
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		query: query.Use(db),
	}
}

func (r *Repository) Create(ctx context.Context, holder *model.CardHolder) error {
	return r.query.CardHolder.WithContext(ctx).Create(holder)
}

func (r *Repository) FindCardHolderByID(ctx context.Context, id model.ID) (*model.CardHolder, error) {
	holder, err := r.query.CardHolder.WithContext(ctx).
		Where(r.query.CardHolder.ID.Eq(id)).
		First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, biz.ErrNotFound
	}
	return holder, err
}

func (r *Repository) Save(ctx context.Context, holder *model.CardHolder) error {
	return r.query.CardHolder.WithContext(ctx).Save(holder)
}

func (r *Repository) List(ctx context.Context, offset int, limit int) ([]*model.CardHolder, error) {
	return r.query.CardHolder.WithContext(ctx).
		Order(r.query.CardHolder.ID.Asc()).
		Offset(offset).
		Limit(limit).
		Find()
}

func (r *Repository) CreateCard(ctx context.Context, card *model.Card) error {
	return r.query.Card.WithContext(ctx).Create(card)
}

func (r *Repository) FindCardByID(ctx context.Context, id model.ID) (*model.Card, error) {
	card, err := r.query.Card.WithContext(ctx).
		Where(r.query.Card.ID.Eq(id)).
		First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, biz.ErrNotFound
	}
	return card, err
}

func (r *Repository) FindByRequestID(ctx context.Context, requestID string) (*model.Card, error) {
	card, err := r.query.Card.WithContext(ctx).
		Where(r.query.Card.RequestID.Eq(requestID)).
		First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, biz.ErrNotFound
	}
	return card, err
}

func (r *Repository) SaveCard(ctx context.Context, card *model.Card) error {
	return r.query.Card.WithContext(ctx).Save(card)
}

func (r *Repository) ListTransactions(ctx context.Context, offset int, limit int) ([]*model.CardTransaction, error) {
	return r.query.CardTransaction.WithContext(ctx).
		Order(r.query.CardTransaction.ID.Asc()).
		Offset(offset).
		Limit(limit).
		Find()
}

var _ biz.CardHolderRepository = (*Repository)(nil)
var _ biz.CardRepository = (*Repository)(nil)
var _ biz.TransactionRepository = (*Repository)(nil)
