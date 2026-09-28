package data

import (
	"context"

	"generic-mock/model"
	"generic-mock/shared/biz"

	"github.com/samber/do"
)

type cardRepository struct {
	*Repository
}

func NewCardRepository(injector *do.Injector) (biz.CardRepo, error) {
	return &cardRepository{Repository: do.MustInvoke[*Repository](injector)}, nil
}

func (repo *cardRepository) Create(ctx context.Context, card *model.Card) error {
	return repo.DB(ctx).Card.WithContext(ctx).Create(card)
}
