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

type WalletUpdateBalanceRequest struct {
	ID         model.ID
	AccountID  model.ID
	Channel    enums.Channel
	Available  decimal.Decimal
	PendingOut decimal.Decimal
	In         decimal.Decimal
	Out        decimal.Decimal
}

type WalletRepo interface {
	Create(context.Context, *model.Wallet) error
	Exist(context.Context, *WalletExistRequest) (bool, error)
	FindByIDWithLock(context.Context, *WalletFindByIDWithLockRequest) (*model.Wallet, error)
	UpdateBalance(context.Context, *WalletUpdateBalanceRequest) error
}
