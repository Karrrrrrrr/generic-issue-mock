package biz

import (
	"context"

	slasherrors "generic-mock/channel/slash/errors"
	"generic-mock/model"

	"go.uber.org/zap"
)

type CardProductInfo struct {
	Product *model.CardProduct
}

func (u *SlashUIUsecase) ListCardProducts(ctx context.Context) ([]*CardProductInfo, error) {
	products, err := u.cardProductRepository.List(ctx)
	if err != nil {
		zap.S().Errorw("list slash card products", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	items := make([]*CardProductInfo, 0, len(products))
	for _, product := range products {
		items = append(items, &CardProductInfo{
			Product: product,
		})
	}

	return items, nil
}
