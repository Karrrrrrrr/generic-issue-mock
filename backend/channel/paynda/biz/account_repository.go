package biz

import (
	"context"

	"generic-mock/model"
)

type AccountCountRequest struct {
	IDs []model.ID
}

type AccountListRequest struct {
	IDs    []model.ID
	Offset int
	Limit  int
}

type PayndaAccountRepository interface {
	Create(context.Context, *model.Account) error
	Save(context.Context, *model.Account) error
	ExistByChannel(context.Context) (bool, error)
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.Account, error)
	FindByChannel(context.Context) (*model.Account, error)
	Count(context.Context, *AccountCountRequest) (int64, error)
	List(context.Context, *AccountListRequest) ([]*model.Account, error)
}
