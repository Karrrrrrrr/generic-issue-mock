package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
)

type CardHolderExistRequest struct {
	ID        model.ID
	AccountID model.ID
	Channel   enums.Channel
}

type CardHolderFilters struct {
	Channel    enums.Channel
	IDs        []model.ID
	AccountIDs []model.ID
}

type CardHolderListRequest struct {
	CardHolderFilters
	Offset int
	Limit  *int
}

type CardHolderCountRequest struct {
	CardHolderFilters
}

type CardHolderCreateRequest struct {
	CardHolder *model.CardHolder
}

type CardHolderRepo interface {
	Create(context.Context, *CardHolderCreateRequest) error
	Exist(context.Context, *CardHolderExistRequest) (bool, error)
	List(context.Context, *CardHolderListRequest) ([]*model.CardHolder, error)
	Count(context.Context, *CardHolderCountRequest) (int64, error)
}
