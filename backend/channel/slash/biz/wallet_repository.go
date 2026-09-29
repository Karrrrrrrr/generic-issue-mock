package biz

import (
	"context"

	"generic-mock/model"
)

type WalletListRequest struct {
	AccountIDs []model.ID
}

type WalletListByAccountIDForUpdateRequest struct {
	AccountID model.ID
	IDs       []model.ID
}

type LockWalletRequest struct {
	AccountID model.ID
	ID        model.ID
}

type ExistWalletRequest struct {
	AccountID model.ID
	ID        model.ID
}

type SlashWalletRepository interface {
	ListWallets(context.Context, *WalletListRequest) ([]*model.Wallet, error)
	WalletExists(context.Context, *ExistWalletRequest) (bool, error)
	LockWallet(context.Context, *LockWalletRequest) (*model.Wallet, error)
	SaveWallet(context.Context, *model.Wallet) error
	Create(context.Context, *model.Wallet) error
	FindByIDForUpdate(context.Context, model.ID) (*model.Wallet, error)
	Save(context.Context, *model.Wallet) error
	ListByAccountIDForUpdate(context.Context, *WalletListByAccountIDForUpdateRequest) ([]*model.Wallet, error)
}
