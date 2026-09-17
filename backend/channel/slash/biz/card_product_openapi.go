package biz

import (
	"context"

	"generic-mock/model"

	"go.uber.org/zap"
)

func (u *SlashOpenAPIUsecase) ListCardProducts(ctx context.Context) ([]*model.CardProduct, error) {
	items, err := u.cardProductRepository.List(ctx)
	if err != nil {
		zap.S().Errorw("list slash openapi card products", "error", err)

		return nil, ErrDatabaseOperation
	}

	return items, nil
}
