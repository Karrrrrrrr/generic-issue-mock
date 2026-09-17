package biz

import (
	"context"

	"generic-mock/model"

	"go.uber.org/zap"
)

func (u *PhotonPayOpenAPIUsecase) ListCardProducts(ctx context.Context) ([]*model.CardProduct, error) {
	products, err := u.cardProductRepo.List(ctx)
	if err != nil {
		zap.S().Errorw("list photonpay card products", "error", err)
		return nil, ErrDatabaseOperation
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
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}
	product, err := u.cardProductRepo.FindByPrefixForUpdate(ctx, prefix)
	if err != nil {
		zap.S().Errorw("lock photonpay card product", "error", err)
		return nil, ErrDatabaseOperation
	}

	return product, nil
}
