package service

import (
	"context"

	"generic-mock/channel/slash/biz"
	slasherrors "generic-mock/channel/slash/errors"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"
)

type ListCardsRequest struct {
	ListRequest
	UIListTimeRange
	ID         *model.ID          `form:"id" binding:"omitempty,min=1"`
	CardNumber *string            `form:"card_number" binding:"omitempty,min=1"`
	CardStatus *common.CardStatus `form:"card_status" binding:"omitempty,oneof=inactive active freezing frozen deleting deleted"`
}

func (s *SlashUIService) ListCards(ctx context.Context, req *ListCardsRequest) (*ListResponse[*CardData], error) {
	accountID := req.AccountID
	if accountID != nil && *accountID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	id := req.ID
	if id != nil && *id <= 0 {
		return nil, slasherrors.ErrInvalidOperation
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
