package biz

import (
	"context"

	"generic-mock/model"
)

type CardProductRepository interface {
	ExistByPrefix(context.Context, string) (bool, error)
	FindByPrefixForUpdate(context.Context, string) (*model.CardProduct, error)
	List(context.Context) ([]*model.CardProduct, error)
	Save(context.Context, *model.CardProduct) error
}
