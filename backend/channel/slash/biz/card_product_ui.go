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

func (u *SlashUIUsecase) getCardProductForUpdate(ctx context.Context, id model.ID) (*model.CardProduct, error) {
	if id == 0 {
		exists, err := u.cardProductRepository.ExistDefault(ctx)
		if err != nil {
			zap.S().Errorw("check slash default card product", "error", err)
			return nil, slasherrors.ErrDatabaseOperation
		}
		if !exists {
			return nil, slasherrors.ErrResourceNotFound
		}

		product, err := u.cardProductRepository.FindDefaultForUpdate(ctx)
		if err != nil {
			zap.S().Errorw("lock slash default card product", "error", err)
			return nil, slasherrors.ErrDatabaseOperation
		}

		return product, nil
	}

	if err := u.requireCardProduct(ctx, id); err != nil {
		return nil, err
	}
	product, err := u.cardProductRepository.FindByIDForUpdate(ctx, id)
	if err != nil {
		zap.S().Errorw("lock slash card product", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}

	return product, nil
}

func (u *SlashUIUsecase) requireCardProduct(ctx context.Context, id model.ID) error {
	exists, err := u.cardProductRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check slash card product", "error", err)
		return slasherrors.ErrDatabaseOperation
	}
	if !exists {
		return slasherrors.ErrResourceNotFound
	}

	return nil
}
