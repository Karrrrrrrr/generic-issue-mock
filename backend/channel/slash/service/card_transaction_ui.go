package service

import (
	"context"

	"generic-mock/channel/slash/biz"
	channelEnums "generic-mock/channel/slash/enums"
	"generic-mock/channel/slash/pkg/idconv"
	"generic-mock/pkg/types"
)

type ListTransactionsRequest struct {
	ListRequest
	UIListTimeRange
	ID              *string                         `form:"id" binding:"omitempty,min=1"`
	CardID          *string                         `form:"card_id" binding:"omitempty,min=1"`
	AuthorizationID *string                         `form:"authorization_id" binding:"omitempty,min=1"`
	TransactionType *channelEnums.TransactionType   `form:"transaction_type" binding:"omitempty,oneof=auth clear void refund"`
	Status          *channelEnums.TransactionStatus `form:"status" binding:"omitempty,oneof=pending posted failed"`
}

func (s *SlashUIService) ListTransactions(ctx context.Context, req *ListTransactionsRequest) (*ListResponse[*TransactionData], error) {
	accountID, err := idconv.FromOptionalUUID(req.AccountID)
	if err != nil {
		return nil, err
	}
	id, err := idconv.FromOptionalUUID(req.ID)
	if err != nil {
		return nil, err
	}
	cardID, err := idconv.FromOptionalUUID(req.CardID)
	if err != nil {
		return nil, err
	}
	authorizationID, err := idconv.FromOptionalUUID(req.AuthorizationID)
	if err != nil {
		return nil, err
	}
	page, size := types.NormalizePagination(types.Value(req.PageNumber), types.Value(req.PageSize))
	items, total, err := s.usecase.ListCardTransactions(ctx, &biz.ListCardTransactionsRequest{
		AccountID:       accountID,
		ID:              id,
		Offset:          (page - 1) * size,
		Limit:           size,
		CreatedFrom:     req.CreatedFrom,
		CreatedTo:       req.CreatedTo,
		CardID:          cardID,
		AuthorizationID: authorizationID,
		Types:           uiTransactionTypes(req.TransactionType),
		Statuses:        uiTransactionStatuses(req.Status),
	})
	if err != nil {
		return nil, err
	}
	return &ListResponse[*TransactionData]{
		TotalItems: total,
		Data:       types.BulkConvertSlice(items, transactionData),
	}, nil
}
