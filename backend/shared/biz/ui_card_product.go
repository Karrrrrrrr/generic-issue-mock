package biz

import (
	"context"

	"generic-mock/model"
	"generic-mock/pkg/types"
	sharederrors "generic-mock/shared/errors"

	"go.uber.org/zap"
)

type ListUICardProductsRequest struct {
	UIPageRequest
	ID *model.ID
}

func (req *ListUICardProductsRequest) Validate() error {
	if req == nil || !validUIIDs([]*model.ID{req.ID}) {
		return sharederrors.ErrInvalidUIRequest
	}
	return req.UIPageRequest.Validate()
}

type UICardProducts interface {
	ListCardProducts(context.Context, *ListUICardProductsRequest) ([]*model.CardProduct, int64, error)
}

func (uc *ui) ListCardProducts(ctx context.Context, req *ListUICardProductsRequest) ([]*model.CardProduct, int64, error) {
	if err := req.Validate(); err != nil {
		return nil, 0, err
	}
	filters := CardProductFilters{
		Channel: uc.channel,
		IDs:     types.PointerSlice(req.ID),
	}
	items, err := uc.cardProductRepo.List(ctx, &CardProductListRequest{
		CardProductFilters: filters,
		Offset:             req.Offset,
		Limit:              req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list shared UI card_product", "channel", uc.channel, "error", err)
		return nil, 0, sharederrors.ErrDatabaseOperation
	}
	total, err := uc.cardProductRepo.Count(ctx, &CardProductCountRequest{CardProductFilters: filters})
	if err != nil {
		zap.S().Errorw("count shared UI card_product", "channel", uc.channel, "error", err)
		return nil, 0, sharederrors.ErrDatabaseOperation
	}
	return items, total, nil
}
