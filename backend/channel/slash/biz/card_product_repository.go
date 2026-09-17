package biz

import (
	"context"

	"generic-mock/model"
)

type SlashCardProductRepository interface {
	ExistByID(context.Context, model.ID) (bool, error)
	FindByIDForUpdate(context.Context, model.ID) (*model.CardProduct, error)
	ExistDefault(context.Context) (bool, error)
	FindDefaultForUpdate(context.Context) (*model.CardProduct, error)
	List(context.Context) ([]*model.CardProduct, error)
	Save(context.Context, *model.CardProduct) error
}
