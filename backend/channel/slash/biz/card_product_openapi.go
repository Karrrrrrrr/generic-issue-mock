package biz

import (
	"context"

	slasherrors "generic-mock/channel/slash/errors"
	"generic-mock/model"

	"go.uber.org/zap"
)

func (u *SlashOpenAPIUsecase) ListCardProducts(ctx context.Context) ([]*model.CardProduct, error) {
	items, err := u.cardProductRepository.List(ctx)
	if err != nil {
		zap.S().Errorw("list slash openapi card products", "error", err)

		return nil, slasherrors.ErrDatabaseOperation
	}

	return items, nil
}
