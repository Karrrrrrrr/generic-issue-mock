package service

import (
	"context"

	"generic-mock/channel/photonpay/biz"
	channelEnums "generic-mock/channel/photonpay/enums"
	"generic-mock/channel/photonpay/pkg/idconv"
	"generic-mock/pkg/types"
)

type ListUITransactionsRequest struct {
	UIListRequest
	UIListTimeRange
	ID              *string                         `form:"id" binding:"omitempty,min=1"`
	CardID          *string                         `form:"card_id" binding:"omitempty,min=1"`
	AuthorizationID *string                         `form:"authorization_id" binding:"omitempty,min=1"`
	TransactionType *channelEnums.TransactionType   `form:"transaction_type" binding:"omitempty,oneof=auth clear void refund"`
	Status          *channelEnums.TransactionStatus `form:"status" binding:"omitempty,oneof=pending authorized succeed failed void"`
}

func (s *PhotonPayUIService) ListTransactions(ctx context.Context, req *ListUITransactionsRequest) (*UIListResponse[*UITransactionData], error) {
	accountID, err := idconv.FromOptionalString(req.AccountID)
	if err != nil {
		return nil, err
	}
	id, err := idconv.FromOptionalString(req.ID)
	if err != nil {
		return nil, err
	}
	cardID, err := idconv.FromOptionalString(req.CardID)
	if err != nil {
		return nil, err
	}
	authorizationID, err := idconv.FromOptionalString(req.AuthorizationID)
	if err != nil {
		return nil, err
	}
	page, size := types.NormalizePagination(types.Value(req.PageNumber), types.Value(req.PageSize))
	items, total, err := s.usecase.ListTransactions(ctx, &biz.ListUITransactionsRequest{
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
	return &UIListResponse[*UITransactionData]{
		TotalItems: int(total),
		Data:       types.BulkConvertSlice(items, photonPayUITransactionData),
	}, nil
}
