package service

import (
	"context"
	"time"

	"generic-mock/channel/pingpong/biz"
	"generic-mock/channel/pingpong/pkg/idconv"
	common "generic-mock/enums"
	"generic-mock/model"

	"github.com/shopspring/decimal"
)

type UIListCardTransactionsRequest struct {
	UIListRequest
	UIListTimeRange
	ID              *string                       `form:"id"`
	CardID          *string                       `form:"card_id"`
	AuthorizationID *string                       `form:"authorization_id"`
	Type            *common.CardTransactionType   `form:"transaction_type" binding:"omitempty,oneof=auth clear void refund verification fund_in fund_out"`
	Status          *common.CardTransactionStatus `form:"status" binding:"omitempty,oneof=pending authorized succeed failed void"`
}

type UICardTransactionData struct {
	ID              string                       `json:"id"`
	AccountID       string                       `json:"account_id"`
	AccountName     string                       `json:"account_name"`
	CardID          string                       `json:"card_id"`
	AuthorizationID string                       `json:"authorization_id"`
	Type            common.CardTransactionType   `json:"transaction_type"`
	Status          common.CardTransactionStatus `json:"status"`
	Amount          string                       `json:"amount"`
	Currency        common.Currency              `json:"currency"`
	MerchantName    string                       `json:"merchant_name"`
	MerchantMCC     string                       `json:"merchant_category_code"`
	TransactedAt    time.Time                    `json:"transacted_at"`
}

type UITransactionStageRequest struct {
	ID        string                     `uri:"id" binding:"required"`
	Type      common.CardTransactionType `json:"transaction_type" binding:"required,oneof=clear void refund"`
	Amount    *decimal.Decimal           `json:"amount"`
	RequestID string                     `json:"request_id" binding:"required"`
}

type UISimulateRefundRequest struct {
	CardID          *string          `json:"card_id"`
	AuthorizationID *string          `json:"authorization_id"`
	Amount          decimal.Decimal  `json:"amount"`
	Currency        *common.Currency `json:"currency" binding:"omitempty,oneof=USD"`
	MerchantName    *string          `json:"merchant_name" binding:"omitempty,min=1"`
	MerchantCountry *string          `json:"merchant_country" binding:"omitempty,min=1"`
	MerchantMCC     *string          `json:"merchant_category_code" binding:"omitempty,min=1"`
	RequestID       string           `json:"request_id" binding:"required"`
}

func toUICardTransactionData(item *model.CardTransaction) UICardTransactionData {
	authorizationID := ""
	if item.AuthorizationID > 0 {
		authorizationID = idconv.ToString(item.AuthorizationID)
	}
	return UICardTransactionData{
		ID:              idconv.ToString(item.ID),
		AccountID:       idconv.ToString(item.AccountID),
		AccountName:     item.Account.GetName(),
		CardID:          idconv.ToString(item.CardID),
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
	id, err := idconv.FromString(req.ID)
	if err != nil {
		return nil, err
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
	cardID, err := idconv.FromOptionalString(req.CardID)
	if err != nil {
		return nil, err
	}
	authorizationID, err := idconv.FromOptionalString(req.AuthorizationID)
	if err != nil {
		return nil, err
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
