package biz

import (
	"context"

	"generic-mock/model"
)

type VirtualAccountListRequest struct {
	AccountIDs []model.ID
}

type VirtualAccountExistByAccountIDRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type VirtualAccountFindByAccountIDRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type SlashVirtualAccountRepository interface {
	ExistByAccountID(context.Context, *VirtualAccountExistByAccountIDRequest) (bool, error)
	ListVirtualAccounts(context.Context, *VirtualAccountListRequest) ([]*model.VirtualAccount, error)
	CreateVirtualAccount(context.Context, *model.VirtualAccount) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.VirtualAccount, error)
	FindByAccountID(context.Context, *VirtualAccountFindByAccountIDRequest) (*model.VirtualAccount, error)
}
