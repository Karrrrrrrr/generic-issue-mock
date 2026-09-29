package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"

	"github.com/shopspring/decimal"
)

type WalletExistRequest struct {
	ID        model.ID
	AccountID model.ID
	Channel   enums.Channel
}

type WalletFindByIDWithLockRequest struct {
	ID        model.ID
	AccountID model.ID
	Channel   enums.Channel
}

type WalletListByIDsWithLockRequest struct {
	IDs       []model.ID
	AccountID model.ID
	Channel   enums.Channel
}

type WalletUpdateBalanceRequest struct {
	ID         model.ID
	AccountID  model.ID
	Channel    enums.Channel
	Available  decimal.Decimal
	PendingOut decimal.Decimal
	In         decimal.Decimal
	Out        decimal.Decimal
}

type WalletFilters struct {
	Channel    enums.Channel
	IDs        []model.ID
	AccountIDs []model.ID
	Types      []enums.WalletType
}

type WalletListRequest struct {
	WalletFilters
	Offset int
	Limit  int
}

type WalletCountRequest struct {
	WalletFilters
}

type WalletCreateRequest struct {
	Wallet *model.Wallet
}

type WalletRepo interface {
	Create(context.Context, *WalletCreateRequest) error
	Exist(context.Context, *WalletExistRequest) (bool, error)
	FindByIDWithLock(context.Context, *WalletFindByIDWithLockRequest) (*model.Wallet, error)
	ListByIDsWithLock(context.Context, *WalletListByIDsWithLockRequest) ([]*model.Wallet, error)
	UpdateBalance(context.Context, *WalletUpdateBalanceRequest) error
	List(context.Context, *WalletListRequest) ([]*model.Wallet, error)
	Count(context.Context, *WalletCountRequest) (int64, error)
}
