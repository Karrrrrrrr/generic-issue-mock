package service

import (
	"context"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/shared/biz"
	sharederrors "generic-mock/shared/errors"

	"github.com/shopspring/decimal"
)

type ListCardTransactionsRequest struct {
	PageRequest
	TimeRange
	ID              *model.ID                    `form:"id" binding:"omitempty,gt=0"`
	AccountID       *model.ID                    `form:"account_id" binding:"omitempty,gt=0"`
	CardID          *model.ID                    `form:"card_id"`
	AuthorizationID *model.ID                    `form:"authorization_id"`
	Status          *enums.CardTransactionStatus `form:"status"`
	Type            *enums.CardTransactionType   `form:"type"`
}

type CardTransactionData struct {
	ID                model.ID                    `json:"id"`
	AccountID         model.ID                    `json:"account_id"`
	AccountName       string                      `json:"account_name"`
	CardID            model.ID                    `json:"card_id"`
	AuthorizationID   model.ID                    `json:"authorization_id"`
	Type              enums.CardTransactionType   `json:"type"`
	Status            enums.CardTransactionStatus `json:"status"`
	Amount            decimal.Decimal             `json:"amount"`
	Currency          enums.Currency              `json:"currency"`
	MerchantName      string                      `json:"merchant_name"`
	MerchantMCC       string                      `json:"merchant_mcc"`
	MerchantCountry   string                      `json:"merchant_country"`
	AuthorizationCode string                      `json:"authorization_code"`
	AuthorizedAt      *time.Time                  `json:"authorized_at"`
	CreatedAt         time.Time                   `json:"created_at"`
}

func (s *Service) ListCardTransactions(ctx context.Context, req *ListCardTransactionsRequest) (*Page[CardTransactionData], error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	page, err := req.PageRequest.toBizPage()
	if err != nil {
		return nil, err
	}
	items, total, err := s.uc.ListCardTransactions(ctx, &biz.ListUICardTransactionsRequest{
		UIPageRequest:   page,
		ID:              req.ID,
		AccountID:       req.AccountID,
		CardID:          req.CardID,
		AuthorizationID: req.AuthorizationID,
		Status:          req.Status,
		Type:            req.Type,
		UITimeRange: biz.UITimeRange{
			CreatedFrom: req.CreatedFrom,
			CreatedTo:   req.CreatedTo,
		},
	})
	if err != nil {
		return nil, err
	}
	result := &Page[CardTransactionData]{
		Items: make([]CardTransactionData, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		result.Items = append(result.Items, toCardTransactionData(item))
	}
	return result, nil
}

func toCardTransactionData(item *model.CardTransaction) CardTransactionData {
	var authorizedAt *time.Time
	if item.Authorization != nil {
		authorizedAt = &item.Authorization.CreatedAt
	}
	return CardTransactionData{
		ID:                item.ID,
		AccountID:         item.AccountID,
		AccountName:       item.Account.GetName(),
		CardID:            item.CardID,
		AuthorizationID:   item.AuthorizationID,
		Type:              item.Type,
		Status:            item.Status,
		Amount:            item.TxAmount,
		Currency:          item.TxCurrency,
		MerchantName:      item.MerchantName,
		MerchantMCC:       item.MerchantMCC,
		MerchantCountry:   item.MerchantCountry,
		AuthorizationCode: item.AuthorizationCode,
		AuthorizedAt:      authorizedAt,
		CreatedAt:         item.CreatedAt,
	}
}

type GetCardTransactionRequest struct {
	ID        model.ID `uri:"id" binding:"required,gt=0"`
	AccountID model.ID `form:"account_id" binding:"required,gt=0"`
}

func (s *Service) GetCardTransaction(ctx context.Context, req *GetCardTransactionRequest) (*CardTransactionData, error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	item, err := s.uc.GetCardTransaction(ctx, &biz.GetUICardTransactionRequest{
		ID:        req.ID,
		AccountID: req.AccountID,
	})
	if err != nil {
		return nil, err
	}
	result := toCardTransactionData(item)
	return &result, nil
}
