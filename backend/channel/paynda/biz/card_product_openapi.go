package biz

import (
	"context"

	payndaerrors "generic-mock/channel/paynda/errors"
	"generic-mock/model"

	"go.uber.org/zap"
)

func (u *PayndaOpenAPIUsecase) ListCardProducts(
	ctx context.Context,
) ([]*model.CardProduct, error) {
	items, err := u.cardProductRepository.List(ctx)
	if err != nil {
		zap.S().Errorw("list paynda card products", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}

	return items, nil
}

func (u *PayndaOpenAPIUsecase) requireCardProduct(
	ctx context.Context,
	id model.ID,
) error {
	exists, err := u.cardProductRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check paynda card product", "error", err)
		return payndaerrors.ErrDatabaseOperation
	}
	if !exists {
		return payndaerrors.ErrResourceNotFound
	}

	return nil
}
