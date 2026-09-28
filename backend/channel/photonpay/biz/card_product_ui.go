package biz

import (
	"context"

	photonpayerrors "generic-mock/channel/photonpay/errors"
	"generic-mock/model"

	"go.uber.org/zap"
)

func (u *PhotonPayUIUsecase) ListCardProducts(ctx context.Context) ([]*model.CardProduct, error) {
	products, err := u.cardProductRepo.List(ctx)
	if err != nil {
		zap.S().Errorw("list photonpay UI card products", "error", err)
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	return products, nil
}
