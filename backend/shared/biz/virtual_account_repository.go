package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
)

type VirtualAccountExistRequest struct {
	ID        model.ID
	AccountID model.ID
	Channel   enums.Channel
}

type VirtualAccountFindRequest struct {
	ID        model.ID
	AccountID model.ID
	Channel   enums.Channel
}

type VirtualAccountFilters struct {
	Channel    enums.Channel
	IDs        []model.ID
	AccountIDs []model.ID
}

type VirtualAccountListRequest struct {
	VirtualAccountFilters
	Offset int
	Limit  *int
}

type VirtualAccountCountRequest struct {
	VirtualAccountFilters
}

type VirtualAccountCreateRequest struct {
	VirtualAccount *model.VirtualAccount
}

type VirtualAccountRepo interface {
	Exist(context.Context, *VirtualAccountExistRequest) (bool, error)
	Find(context.Context, *VirtualAccountFindRequest) (*model.VirtualAccount, error)
	List(context.Context, *VirtualAccountListRequest) ([]*model.VirtualAccount, error)
	Count(context.Context, *VirtualAccountCountRequest) (int64, error)
	Create(context.Context, *VirtualAccountCreateRequest) error
}
