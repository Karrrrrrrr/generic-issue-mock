package biz

import (
	"context"

	photonpayerrors "generic-mock/channel/photonpay/errors"
	"generic-mock/model"

	"go.uber.org/zap"
)

func (u *PhotonPayOpenAPIUsecase) ListCardProducts(ctx context.Context) ([]*model.CardProduct, error) {
	products, err := u.cardProductRepo.List(ctx)
	if err != nil {
		zap.S().Errorw("list photonpay card products", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}

	return products, nil
}

func (u *PhotonPayOpenAPIUsecase) getCardProductByBinPrefix(
	ctx context.Context,
	prefix string,
) (*model.CardProduct, error) {
	exists, err := u.cardProductRepo.ExistByPrefix(ctx, prefix)
	if err != nil {
		zap.S().Errorw("check photonpay card product", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, photonpayerrors.ErrResourceNotFound
	}
	product, err := u.cardProductRepo.FindByPrefixForUpdate(ctx, prefix)
	if err != nil {
		zap.S().Errorw("lock photonpay card product", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}

	return product, nil
}
