package biz

import (
	"context"

	"generic-mock/model"
)

type AccountExistsRequest struct {
	ID model.ID
}

type AccountFindRequest struct {
	ID model.ID
}

type AccountLockRequest struct {
	ID model.ID
}

type AccountListRequest struct {
	IDs    []model.ID
	Offset int
	Limit  *int
}

type AccountCountRequest struct {
	IDs []model.ID
}

type PingPongAccountRepository interface {
	Exists(context.Context, *AccountExistsRequest) (bool, error)
	Find(context.Context, *AccountFindRequest) (*model.Account, error)
	Lock(context.Context, *AccountLockRequest) (*model.Account, error)
	Create(context.Context, *model.Account) error
	Save(context.Context, *model.Account) error
	List(context.Context, *AccountListRequest) ([]*model.Account, error)
	Count(context.Context, *AccountCountRequest) (int64, error)
}
