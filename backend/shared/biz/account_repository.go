package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
)

type AccountExistRequest struct {
	ID      model.ID
	Channel enums.Channel
}

type AccountFindByIDWithLockRequest struct {
	ID      model.ID
	Channel enums.Channel
}

type AccountFilters struct {
	Channel enums.Channel
	IDs     []model.ID
	Name    *string
}

type AccountListRequest struct {
	AccountFilters
	Offset int
	Limit  *int
}

type AccountCountRequest struct {
	AccountFilters
}

type AccountFindRequest struct {
	ID      model.ID
	Channel enums.Channel
}

type AccountCreateRequest struct {
	Account *model.Account
}

type AccountRenameRequest struct {
	ID      model.ID
	Channel enums.Channel
	Name    string
}

type AccountSetWalletRequest struct {
	ID       model.ID
	Channel  enums.Channel
	WalletID model.ID
}

type AccountRepo interface {
	Exist(context.Context, *AccountExistRequest) (bool, error)
	FindByIDWithLock(context.Context, *AccountFindByIDWithLockRequest) (*model.Account, error)
	List(context.Context, *AccountListRequest) ([]*model.Account, error)
	Count(context.Context, *AccountCountRequest) (int64, error)
	Find(context.Context, *AccountFindRequest) (*model.Account, error)
	Create(context.Context, *AccountCreateRequest) error
	Rename(context.Context, *AccountRenameRequest) error
	SetWallet(context.Context, *AccountSetWalletRequest) error
}
