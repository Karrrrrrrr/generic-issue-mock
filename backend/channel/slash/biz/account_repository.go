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

type SlashAccountRepository interface {
	Create(context.Context, *model.Account) error
	Save(context.Context, *model.Account) error
	Exist(context.Context, model.ID) (bool, error)
	Find(context.Context, model.ID) (*model.Account, error)
	Count(context.Context, *AccountCountRequest) (int64, error)
	List(context.Context, *AccountListRequest) ([]*model.Account, error)
}
