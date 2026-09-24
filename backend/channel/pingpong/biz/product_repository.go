package biz

import (
	"context"

	"generic-mock/model"
)

type ProductExistsRequest struct {
	ID model.ID
}

type ProductLockRequest struct {
	ID model.ID
}

type ProductListRequest struct {
	IDs    []model.ID
	Offset int
	Limit  *int
}

type PingPongProductRepository interface {
	Exists(context.Context, *ProductExistsRequest) (bool, error)

	Lock(context.Context, *ProductLockRequest) (*model.CardProduct, error)
	Save(context.Context, *model.CardProduct) error
	List(context.Context, *ProductListRequest) ([]*model.CardProduct, error)
}
