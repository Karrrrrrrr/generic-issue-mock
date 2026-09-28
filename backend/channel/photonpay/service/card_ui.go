package service

import (
	"context"

	"generic-mock/channel/photonpay/biz"
	photonpayerrors "generic-mock/channel/photonpay/errors"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"
)

type ListUICardsRequest struct {
	UIListRequest
	UIListTimeRange
	ID         *model.ID          `form:"id" binding:"omitempty,min=1"`
	CardNumber *string            `form:"card_number" binding:"omitempty,min=1"`
	CardStatus *common.CardStatus `form:"card_status" binding:"omitempty,oneof=inactive active freezing frozen deleting deleted"`
}

func (s *PhotonPayUIService) ListCards(ctx context.Context, req *ListUICardsRequest) (*UIListResponse[*UICardData], error) {
	accountID := req.AccountID
	if accountID != nil && *accountID <= 0 {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	id := req.ID
	if id != nil && *id <= 0 {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	page, size := types.NormalizePagination(types.Value(req.PageNumber), types.Value(req.PageSize))
	items, total, err := s.usecase.ListCards(ctx, &biz.ListUICardsRequest{
		AccountID:   accountID,
		ID:          id,
		Offset:      (page - 1) * size,
		Limit:       size,
		CreatedFrom: req.CreatedFrom,
		CreatedTo:   req.CreatedTo,
		CardNumber:  req.CardNumber,
		Statuses:    types.PointerSlice(req.CardStatus),
	})
	if err != nil {
		return nil, err
	}
	return &UIListResponse[*UICardData]{
		TotalItems: int(total),
		Data:       types.BulkConvertSlice(items, photonPayUICardData),
	}, nil
}
