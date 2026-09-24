package biz

import (
	"context"

	pingerrors "generic-mock/channel/pingpong/errors"
	"generic-mock/model"

	"go.uber.org/zap"
)

func (uc *PingPongOpenAPIUsecase) Products(ctx context.Context) ([]*model.CardProduct, error) {
	items, err := uc.cardProductRepo.List(ctx, &ProductListRequest{})
	if err != nil {
		zap.S().Errorw("list pingpong products", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	return items, nil
}
