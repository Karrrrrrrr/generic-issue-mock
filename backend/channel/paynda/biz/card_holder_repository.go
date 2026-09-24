package biz

import (
	"context"

	"generic-mock/model"
)

type CardHolderExistByAccountIDRequest struct {
	AccountID model.ID
	ID        model.ID
}

type CardHolderFindByAccountIDRequest struct {
	AccountID model.ID
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

type PayndaCardHolderRepository interface {
	Count(context.Context, *CardHolderCountRequest) (int64, error)
	Create(context.Context, *model.CardHolder) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.CardHolder, error)
	ExistByAccountID(context.Context, *CardHolderExistByAccountIDRequest) (bool, error)
	FindByAccountID(context.Context, *CardHolderFindByAccountIDRequest) (*model.CardHolder, error)
	List(context.Context, *CardHolderListRequest) ([]*model.CardHolder, error)
	Save(context.Context, *model.CardHolder) error
}
