package biz

import (
	"context"

	"generic-mock/model"
)

type CardProductExistByIDRequest struct {
	ID model.ID
}

type CardProductFindByIDForUpdateRequest struct {
	ID model.ID
}

type CardProductRepository interface {
	ExistByID(context.Context, *CardProductExistByIDRequest) (bool, error)
	FindByIDForUpdate(context.Context, *CardProductFindByIDForUpdateRequest) (*model.CardProduct, error)
	ExistByPrefix(context.Context, string) (bool, error)
	FindByPrefixForUpdate(context.Context, string) (*model.CardProduct, error)
	List(context.Context) ([]*model.CardProduct, error)
	Save(context.Context, *model.CardProduct) error
}
