package biz

import (
	"context"

	"generic-mock/model"
)

type CardHolderCountRequest struct {
	AccountIDs []model.ID
	Offset     int
	Limit      int
}

type CardHolderListRequest struct {
	AccountIDs []model.ID
	Offset     int
	Limit      int
}

type SlashCardHolderRepository interface {
	Create(context.Context, *model.CardHolder) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.CardHolder, error)
	Count(context.Context, *CardHolderCountRequest) (int64, error)
	List(context.Context, *CardHolderListRequest) ([]*model.CardHolder, error)
	Save(context.Context, *model.CardHolder) error
}
