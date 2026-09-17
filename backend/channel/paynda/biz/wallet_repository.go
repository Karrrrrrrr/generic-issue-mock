package biz

import (
	"context"

	"generic-mock/model"
)

type WalletListRequest struct {
	AccountIDs []model.ID
}

type WalletExistByIDRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type WalletFindByIDRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type WalletFindByIDForUpdateRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type LockWalletRequest struct {
	AccountID model.ID
	ID        model.ID
}

type ExistWalletRequest struct {
	AccountID model.ID
	ID        model.ID
}

type PayndaWalletRepository interface {
	ListWallets(context.Context, *WalletListRequest) ([]*model.Wallet, error)
	WalletExists(context.Context, *ExistWalletRequest) (bool, error)
	LockWallet(context.Context, *LockWalletRequest) (*model.Wallet, error)
	SaveWallet(context.Context, *model.Wallet) error
	Create(context.Context, *model.Wallet) error
	ExistByID(context.Context, *WalletExistByIDRequest) (bool, error)
	FindByID(context.Context, *WalletFindByIDRequest) (*model.Wallet, error)
	FindByIDForUpdate(context.Context, *WalletFindByIDForUpdateRequest) (*model.Wallet, error)
	Save(context.Context, *model.Wallet) error
}
