package service

import (
	"context"

	"generic-mock/channel/photonpay/biz"
	photonpayerrors "generic-mock/channel/photonpay/errors"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"
)

type ListUITransactionsRequest struct {
	UIListRequest
	UIListTimeRange
	ID              *model.ID                     `form:"id" binding:"omitempty,min=1"`
	CardID          *model.ID                     `form:"card_id" binding:"omitempty,min=1"`
	AuthorizationID *model.ID                     `form:"authorization_id" binding:"omitempty,min=1"`
	TransactionType *common.CardTransactionType   `form:"transaction_type" binding:"omitempty,oneof=auth clear void refund verification fund_in fund_out"`
	Status          *common.CardTransactionStatus `form:"status" binding:"omitempty,oneof=pending authorized succeed failed void"`
}

func (s *PhotonPayUIService) ListTransactions(ctx context.Context, req *ListUITransactionsRequest) (*UIListResponse[*UITransactionData], error) {
	accountID := req.AccountID
	if accountID != nil && *accountID <= 0 {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	id := req.ID
	if id != nil && *id <= 0 {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	cardID := req.CardID
	if cardID != nil && *cardID <= 0 {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	authorizationID := req.AuthorizationID
	if authorizationID != nil && *authorizationID <= 0 {
		return nil, photonpayerrors.ErrInvalidOperation
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
