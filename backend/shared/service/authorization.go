package service

import (
	"context"
	"encoding/json"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/shared/biz"
	sharederrors "generic-mock/shared/errors"

	"github.com/shopspring/decimal"
)

type ListAuthorizationsRequest struct {
	PageRequest
	TimeRange
	ID           *model.ID                    `form:"id" binding:"omitempty,gt=0"`
	AccountID    *model.ID                    `form:"account_id" binding:"omitempty,gt=0"`
	CardID       *model.ID                    `form:"card_id"`
	Status       *enums.CardTransactionStatus `form:"status"`
	MerchantName *string                      `form:"merchant_name"`
}

type AuthorizationData struct {
	ID                model.ID                    `json:"id"`
	AccountID         model.ID                    `json:"account_id"`
	AccountName       string                      `json:"account_name"`
	CardID            model.ID                    `json:"card_id"`
	Status            enums.CardTransactionStatus `json:"status"`
	Currency          enums.Currency              `json:"currency"`
	Amount            decimal.Decimal             `json:"amount"`
	Settled           decimal.Decimal             `json:"settled"`
	Reversed          decimal.Decimal             `json:"reversed"`
	Refunded          decimal.Decimal             `json:"refunded"`
	Remaining         decimal.Decimal             `json:"remaining"`
	MerchantName      string                      `json:"merchant_name"`
	MerchantCountry   string                      `json:"merchant_country"`
	MerchantMCC       string                      `json:"merchant_mcc"`
	AuthorizationCode string                      `json:"authorization_code"`
	CreatedAt         time.Time                   `json:"created_at"`
}

func (s *Service) ListAuthorizations(ctx context.Context, req *ListAuthorizationsRequest) (*Page[AuthorizationData], error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	page, err := req.PageRequest.toBizPage()
	if err != nil {
		return nil, err
	}
	items, total, err := s.uc.ListAuthorizations(ctx, &biz.ListUIAuthorizationsRequest{
		UIPageRequest: page,
		ID:            req.ID,
		AccountID:     req.AccountID,
		CardID:        req.CardID,
		Status:        req.Status,
		MerchantName:  req.MerchantName,
		UITimeRange: biz.UITimeRange{
			CreatedFrom: req.CreatedFrom,
			CreatedTo:   req.CreatedTo,
		},
	})
	if err != nil {
		return nil, err
	}
	result := &Page[AuthorizationData]{
		Items: make([]AuthorizationData, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		result.Items = append(result.Items, toAuthorizationData(item))
	}
	return result, nil
}

func toAuthorizationData(item *model.Authorization) AuthorizationData {
	amounts := biz.CalculateUIAuthorizationAmounts(item)
	return AuthorizationData{
		ID:                item.ID,
		AccountID:         item.AccountID,
		AccountName:       item.Account.GetName(),
		CardID:            item.CardID,
		Status:            item.Status,
		Currency:          item.Currency,
		Amount:            item.Amount,
		Settled:           amounts.Settled,
		Reversed:          amounts.Reversed,
		Refunded:          amounts.Refunded,
		Remaining:         amounts.Remaining,
		MerchantName:      item.MerchantName,
		MerchantCountry:   item.MerchantCountry,
		MerchantMCC:       item.MerchantMCC,
		AuthorizationCode: item.AuthorizationCode,
		CreatedAt:         item.CreatedAt,
	}
}

type GetAuthorizationDetailRequest struct {
	ID        model.ID `form:"id" binding:"required,gt=0"`
	AccountID model.ID `form:"account_id" binding:"required,gt=0"`
}

type AuthorizationDetailData struct {
	AuthorizationData
	Card         CardData              `json:"card"`
	Transactions []CardTransactionData `json:"transactions"`
	RawPayload   json.RawMessage       `json:"raw_payload,omitempty"`
}

func (s *Service) GetAuthorizationDetail(ctx context.Context, req *GetAuthorizationDetailRequest) (*AuthorizationDetailData, error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	detail, err := s.uc.GetAuthorizationDetail(ctx, &biz.GetUIAuthorizationDetailRequest{
		ID:        req.ID,
		AccountID: req.AccountID,
	})
	if err != nil {
		return nil, err
	}
	result := &AuthorizationDetailData{
		AuthorizationData: toAuthorizationData(detail.Authorization),
		Card:              toCardData(detail.Card),
		Transactions:      make([]CardTransactionData, 0, len(detail.Authorization.CardTransactions)),
	}
	for _, stage := range detail.Authorization.CardTransactions {
		data := toCardTransactionData(stage)
		data.AccountName = detail.Authorization.Account.GetName()
		data.AuthorizedAt = &detail.Authorization.CreatedAt
		result.Transactions = append(result.Transactions, data)
	}
	if json.Valid(detail.Authorization.RawPayload) {
		result.RawPayload = json.RawMessage(detail.Authorization.RawPayload)
	}
	return result, nil
}
