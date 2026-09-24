package biz

import (
	"context"

	"generic-mock/model"
)

type WalletLockRequest struct {
	AccountID model.ID
	ID        model.ID
}

type PingPongWalletRepository interface {
	Lock(context.Context, *WalletLockRequest) (*model.Wallet, error)
	Create(context.Context, *model.Wallet) error
	Save(context.Context, *model.Wallet) error
}
