package data

import (
	"context"

	"generic-mock/model"
	"generic-mock/shared/biz"

	"github.com/samber/do"
)

type walletRepository struct {
	*Repository
}

func NewWalletRepository(injector *do.Injector) (biz.WalletRepo, error) {
	return &walletRepository{Repository: do.MustInvoke[*Repository](injector)}, nil
}

func (repo *walletRepository) Create(ctx context.Context, wallet *model.Wallet) error {
	return repo.DB(ctx).Wallet.WithContext(ctx).Create(wallet)
}
