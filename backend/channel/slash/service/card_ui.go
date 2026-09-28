package service

import (
	"context"
	common "generic-mock/enums"

	"generic-mock/channel/slash/biz"
	"generic-mock/channel/slash/pkg/idconv"
	"generic-mock/pkg/types"
)

type ListCardsRequest struct {
	ListRequest
	UIListTimeRange
	ID         *string            `form:"id" binding:"omitempty,min=1"`
	CardNumber *string            `form:"card_number" binding:"omitempty,min=1"`
	CardStatus *common.CardStatus `form:"card_status" binding:"omitempty,oneof=inactive active freezing frozen deleting deleted"`
}

func (s *SlashUIService) ListCards(ctx context.Context, req *ListCardsRequest) (*ListResponse[*CardData], error) {
	accountID, err := idconv.FromOptionalUUID(req.AccountID)
	if err != nil {
		return nil, err
	}
	id, err := idconv.FromOptionalUUID(req.ID)
	if err != nil {
		return nil, err
	}
	page, size := types.NormalizePagination(types.Value(req.PageNumber), types.Value(req.PageSize))
	items, total, err := s.usecase.ListCards(ctx, &biz.ListCardsRequest{
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
	return &ListResponse[*CardData]{
		TotalItems: total,
		Data:       types.BulkConvertSlice(items, cardData),
	}, nil
}
