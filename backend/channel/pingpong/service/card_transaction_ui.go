package service

import (
	"context"
	"time"

	"generic-mock/channel/pingpong/biz"
	pingerrors "generic-mock/channel/pingpong/errors"
	common "generic-mock/enums"
	"generic-mock/model"

	"github.com/shopspring/decimal"
)

type UIListCardTransactionsRequest struct {
	UIListRequest
	UIListTimeRange
	ID              *model.ID                     `form:"id" binding:"omitempty,gt=0"`
	CardID          *model.ID                     `form:"card_id" binding:"omitempty,gt=0"`
	AuthorizationID *model.ID                     `form:"authorization_id" binding:"omitempty,gt=0"`
	Type            *common.CardTransactionType   `form:"transaction_type" binding:"omitempty,oneof=auth clear void refund verification fund_in fund_out"`
	Status          *common.CardTransactionStatus `form:"status" binding:"omitempty,oneof=pending authorized succeed failed void"`
}

type UICardTransactionData struct {
	ID              model.ID                     `json:"id"`
	AccountID       model.ID                     `json:"account_id"`
	AccountName     string                       `json:"account_name"`
	CardID          model.ID                     `json:"card_id"`
	AuthorizationID model.ID                     `json:"authorization_id"`
	Type            common.CardTransactionType   `json:"transaction_type"`
	Status          common.CardTransactionStatus `json:"status"`
	Amount          string                       `json:"amount"`
	Currency        common.Currency              `json:"currency"`
	MerchantName    string                       `json:"merchant_name"`
	MerchantMCC     string                       `json:"merchant_category_code"`
	TransactedAt    time.Time                    `json:"transacted_at"`
}

type UITransactionStageRequest struct {
	ID        model.ID                   `uri:"id" binding:"required,gt=0"`
	Type      common.CardTransactionType `json:"transaction_type" binding:"required,oneof=clear void refund"`
	Amount    *decimal.Decimal           `json:"amount"`
	RequestID string                     `json:"request_id" binding:"required"`
}

type UISimulateRefundRequest struct {
	CardID          *model.ID        `json:"card_id" binding:"omitempty,gt=0"`
	AuthorizationID *model.ID        `json:"authorization_id" binding:"omitempty,gt=0"`
	Amount          decimal.Decimal  `json:"amount"`
	Currency        *common.Currency `json:"currency" binding:"omitempty,oneof=USD"`
	MerchantName    *string          `json:"merchant_name" binding:"omitempty,min=1"`
	MerchantCountry *string          `json:"merchant_country" binding:"omitempty,min=1"`
	MerchantMCC     *string          `json:"merchant_category_code" binding:"omitempty,min=1"`
	RequestID       string           `json:"request_id" binding:"required"`
}

func toUICardTransactionData(item *model.CardTransaction) UICardTransactionData {
	authorizationID := item.AuthorizationID
	return UICardTransactionData{
		ID:              item.ID,
		AccountID:       item.AccountID,
		AccountName:     item.Account.GetName(),
		CardID:          item.CardID,
		AuthorizationID: authorizationID,
		Type:            item.Type,
		Status:          item.Status,
		Amount:          item.TxAmount.String(),
		Currency:        item.TxCurrency,
		MerchantName:    item.MerchantName,
		MerchantMCC:     item.MerchantMCC,
		TransactedAt:    item.CreatedAt,
	}
}

func (s *PingPongUIService) ListCardTransactions(ctx context.Context, req *UIListCardTransactionsRequest) (*UIPage[UICardTransactionData], error) {
	if err := req.UIListTimeRange.Validate(); err != nil {
		return nil, err
	}
	accountID := req.AccountID
	if accountID != nil && *accountID <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	id := req.ID
	if id != nil && *id <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	cardID := req.CardID
	if cardID != nil && *cardID <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	authorizationID := req.AuthorizationID
	if authorizationID != nil && *authorizationID <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	page, limit, err := req.PageRequest.resolvePagination()
	if err != nil {
		return nil, err
	}
	items, total, err := s.uc.ListCardTransactions(ctx, &biz.UIListCardTransactionsRequest{
		AccountID:       accountID,
		ID:              id,
		CardID:          cardID,
		AuthorizationID: authorizationID,
		Type:            req.Type,
		Status:          req.Status,
		CreatedFrom:     req.CreatedFrom,
		CreatedTo:       req.CreatedTo,
		Offset:          (page - 1) * limit,
		Limit:           limit,
	})
	if err != nil {
		return nil, err
	}
	result := &UIPage[UICardTransactionData]{
		Items: make([]UICardTransactionData, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		result.Items = append(result.Items, toUICardTransactionData(item))
	}
	return result, nil
}

func (s *PingPongUIService) ApplyTransactionStage(ctx context.Context, req *UITransactionStageRequest) (*Empty, error) {
	id := req.ID
	if id <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	if err := s.uc.ApplyTransactionStage(ctx, &biz.UITransactionStageRequest{
		ID:        id,
		Kind:      req.Type,
		Amount:    req.Amount,
		RequestID: req.RequestID,
	}); err != nil {
		return nil, err
	}
	return &Empty{}, nil
}

func (s *PingPongUIService) SimulateRefund(ctx context.Context, req *UISimulateRefundRequest) (*UICardTransactionData, error) {
	cardID := req.CardID
	if cardID != nil && *cardID <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	authorizationID := req.AuthorizationID
	if authorizationID != nil && *authorizationID <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	item, err := s.uc.SimulateRefund(ctx, &biz.UISimulateRefundRequest{
		CardID:          cardID,
		AuthorizationID: authorizationID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		MerchantName:    req.MerchantName,
		MerchantCountry: req.MerchantCountry,
		MerchantMCC:     req.MerchantMCC,
		RequestID:       req.RequestID,
	})
	if err != nil {
		return nil, err
	}
	result := toUICardTransactionData(item)
	return &result, nil
}
