package biz

import (
	"context"

	"generic-mock/model"
)

type VirtualAccountExistsRequest struct {
	AccountID model.ID
	ID        model.ID
}

type VirtualAccountFindRequest struct {
	AccountID model.ID
	ID        model.ID
}

type VirtualAccountListRequest struct {
	IDs        []model.ID
	AccountIDs []model.ID
	Offset     int
	Limit      *int
}

type VirtualAccountCountRequest struct {
	IDs        []model.ID
	AccountIDs []model.ID
}

type PingPongVirtualAccountRepository interface {
	Exists(context.Context, *VirtualAccountExistsRequest) (bool, error)
	Find(context.Context, *VirtualAccountFindRequest) (*model.VirtualAccount, error)

	Create(context.Context, *model.VirtualAccount) error

	List(context.Context, *VirtualAccountListRequest) ([]*model.VirtualAccount, error)
	Count(context.Context, *VirtualAccountCountRequest) (int64, error)
}
