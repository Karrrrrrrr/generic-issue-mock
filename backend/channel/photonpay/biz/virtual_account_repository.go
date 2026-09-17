package biz

import (
	"context"

	"generic-mock/model"
)

type VirtualAccountListRequest struct {
	AccountIDs []model.ID
}

type VirtualAccountRepository interface {
	ListVirtualAccounts(context.Context, *VirtualAccountListRequest) ([]*model.VirtualAccount, error)
	Create(context.Context, *model.VirtualAccount) error
	FindByAccountID(context.Context, model.ID) (*model.VirtualAccount, error)
}
