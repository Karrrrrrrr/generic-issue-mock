package biz

import (
	"context"

	"generic-mock/model"
)

type CardHolderExistCardHolderByAccountIDRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type CardHolderFindCardHolderByAccountIDRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type CardHolderListRequest struct {
	AccountIDs []model.ID
	Offset     int
	Limit      int
}

type CardHolderCountRequest struct {
	AccountIDs []model.ID
}

type CardHolderRepository interface {
	Count(context.Context, *CardHolderCountRequest) (int64, error)
	Create(context.Context, *model.CardHolder) error
	ExistCardHolderByID(context.Context, model.ID) (bool, error)
	FindCardHolderByID(context.Context, model.ID) (*model.CardHolder, error)
	ExistCardHolderByAccountID(context.Context, *CardHolderExistCardHolderByAccountIDRequest) (bool, error)
	FindCardHolderByAccountID(context.Context, *CardHolderFindCardHolderByAccountIDRequest) (*model.CardHolder, error)
	Save(context.Context, *model.CardHolder) error
	List(context.Context, *CardHolderListRequest) ([]*model.CardHolder, error)
}
