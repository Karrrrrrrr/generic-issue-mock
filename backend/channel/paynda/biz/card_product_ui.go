package biz

import (
	"context"

	payndaerrors "generic-mock/channel/paynda/errors"
	"generic-mock/model"

	"go.uber.org/zap"
)

func (u *PayndaUIUsecase) ListCardProducts(ctx context.Context) ([]*model.CardProduct, error) {
	products, err := u.cardProductRepository.List(ctx)
	if err != nil {
		zap.S().Errorw("list paynda UI card products", "error", err)
		return nil, payndaerrors.ErrDatabaseOperation
	}
	return products, nil
}
