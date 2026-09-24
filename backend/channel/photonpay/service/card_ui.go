package service

import (
	"context"

	"generic-mock/channel/photonpay/biz"
	channelEnums "generic-mock/channel/photonpay/enums"
	"generic-mock/channel/photonpay/pkg/idconv"
	"generic-mock/pkg/types"
)

type ListUICardsRequest struct {
	UIListRequest
	UIListTimeRange
	ID         *string                  `form:"id" binding:"omitempty,min=1"`
	CardNumber *string                  `form:"card_number" binding:"omitempty,min=1"`
	CardStatus *channelEnums.CardStatus `form:"card_status" binding:"omitempty,oneof=normal freezing frozen cancelled"`
}

func (s *PhotonPayUIService) ListCards(ctx context.Context, req *ListUICardsRequest) (*UIListResponse[*UICardData], error) {
	accountID, err := idconv.FromOptionalString(req.AccountID)
	if err != nil {
		return nil, err
	}
	id, err := idconv.FromOptionalString(req.ID)
	if err != nil {
		return nil, err
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
		Statuses:    uiCardStatuses(req.CardStatus),
	})
	if err != nil {
		return nil, err
	}
	return &UIListResponse[*UICardData]{
		TotalItems: int(total),
		Data:       types.BulkConvertSlice(items, photonPayUICardData),
	}, nil
}
