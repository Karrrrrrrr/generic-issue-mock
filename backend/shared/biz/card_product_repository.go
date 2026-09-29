package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
)

type CardProductExistRequest struct {
	ID      model.ID
	Channel enums.Channel
}

type CardProductFindByIDWithLockRequest struct {
	ID      model.ID
	Channel enums.Channel
}

type CardProductUpdateSeqRequest struct {
	ID           model.ID
	Channel      enums.Channel
	NextSequence int64
}

type CardProductFilters struct {
	Channel enums.Channel
	IDs     []model.ID
}

type CardProductListRequest struct {
	CardProductFilters
	Offset int
	Limit  *int
}

type CardProductCountRequest struct {
	CardProductFilters
}

type CardProductRepo interface {
	Exist(context.Context, *CardProductExistRequest) (bool, error)
	FindByIDWithLock(context.Context, *CardProductFindByIDWithLockRequest) (*model.CardProduct, error)
	UpdateSeq(context.Context, *CardProductUpdateSeqRequest) error
	List(context.Context, *CardProductListRequest) ([]*model.CardProduct, error)
	Count(context.Context, *CardProductCountRequest) (int64, error)
}
