package service

import (
	"context"
	common "generic-mock/enums"

	"generic-mock/channel/photonpay/biz"
	"generic-mock/channel/photonpay/pkg/idconv"
	"generic-mock/pkg/types"
)

type ListUITransactionsRequest struct {
	UIListRequest
	UIListTimeRange
	ID              *string                       `form:"id" binding:"omitempty,min=1"`
	CardID          *string                       `form:"card_id" binding:"omitempty,min=1"`
	AuthorizationID *string                       `form:"authorization_id" binding:"omitempty,min=1"`
	TransactionType *common.CardTransactionType   `form:"transaction_type" binding:"omitempty,oneof=auth clear void refund verification fund_in fund_out"`
	Status          *common.CardTransactionStatus `form:"status" binding:"omitempty,oneof=pending authorized succeed failed void"`
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
		Types:           types.PointerSlice(req.TransactionType),
		Statuses:        types.PointerSlice(req.Status),
	})
	if err != nil {
		return nil, err
	}
	return &UIListResponse[*UITransactionData]{
		TotalItems: int(total),
		Data:       types.BulkConvertSlice(items, photonPayUITransactionData),
	}, nil
}
