package service

import (
	"context"

	"generic-mock/channel/paynda/biz"
	channelEnums "generic-mock/channel/paynda/enums"
	"generic-mock/channel/paynda/pkg/idconv"
	"generic-mock/pkg/types"
)

type ListUICardsRequest struct {
	PayndaUIListRequest
	UIListTimeRange
	ID         *string                  `form:"id" binding:"omitempty,min=1"`
	CardNumber *string                  `form:"card_number" binding:"omitempty,min=1"`
	CardStatus *channelEnums.CardStatus `form:"card_status" binding:"omitempty,oneof=ACTIVE FROZEN DELETED"`
}

func (s *PayndaUIService) ListCards(ctx context.Context, req *ListUICardsRequest) (*PayndaUIListResponse[*PayndaUICardData], error) {
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
	return &PayndaUIListResponse[*PayndaUICardData]{
		TotalItems: int(total),
		Data:       types.BulkConvertSlice(items, payndaUICardData),
	}, nil
}
