package biz

import (
	"context"

	"generic-mock/model"
)

type WalletRepo interface {
	Create(context.Context, *model.Wallet) error
}
