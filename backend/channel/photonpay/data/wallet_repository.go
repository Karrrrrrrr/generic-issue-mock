package data

import (
	"context"

	"generic-mock/channel/photonpay/biz"
	"generic-mock/model"

	"github.com/samber/do"
)

type walletRepository struct{ repository *Repository }

func NewWalletRepository(injector *do.Injector) (biz.WalletRepository, error) {
	return &walletRepository{repository: do.MustInvoke[*Repository](injector)}, nil
}

func (r *walletRepository) Create(ctx context.Context, item *model.Wallet) error {
	return r.repository.DB(ctx).Wallet.WithContext(ctx).Create(item)
}

var _ biz.WalletRepository = (*walletRepository)(nil)
