package service

import (
	"context"
	"time"

	"generic-mock/channel/photonpay/biz"
	photon "generic-mock/channel/photonpay/enums"
	common "generic-mock/enums"
	"generic-mock/pkg/types"
)

type OpenAPIAccountHistoryRequest struct {
	OpenAPIAccountRequest
	TransactedAtStart string `form:"transactedAtStart" binding:"required"`
	TransactedAtEnd   string `form:"transactedAtEnd" binding:"required"`
}

type OpenAPIAccountHistoryData struct {
	AccountNo    string          `json:"accountNo"`
	Currency     common.Currency `json:"currency"`
	Amount       float64         `json:"amount"`
	BalanceFund  float64         `json:"balanceFund"`
	TransactedAt string          `json:"transactedAt"`
}

func (service *PhotonPayOpenAPIService) AccountHistory(ctx context.Context, req *OpenAPIAccountHistoryRequest) (*OpenAPIAccountHistoryData, error) {
	accountID, err := service.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	return &OpenAPIAccountHistoryData{
		AccountNo:    photonPayIDString(accountID),
		Currency:     common.Currency_USD,
		TransactedAt: time.Now().UTC().Format("2006-01-02T15:04:05"),
	}, nil
}

type OpenAPIPreRechargeRequest struct {
	OpenAPIAccountRequest
	MemberID       *string  `form:"memberId"` // Invalid: the token selects the account; Matrix members are unsupported.
	AccountID      string   `form:"accountId" binding:"required"`
	CardID         string   `form:"cardId" binding:"required"`
	RequestID      string   `form:"requestId" binding:"required"`
	RechargeAmount *float64 `form:"rechargeAmount"`
	ArrivalAmount  *float64 `form:"arrivalAmount"`
}

type OpenAPIPreRechargeData struct {
	AccountID              string          `json:"accountId"`
	RequestID              string          `json:"requestId"`
	ArrivalAmount          float64         `json:"arrivalAmount"`
	ArrivalAmountCurrency  common.Currency `json:"arrivalAmountCurrency"`
	RechargeAmount         float64         `json:"rechargeAmount"`
	RechargeCurrency       common.Currency `json:"rechargeCurrency"`
	RechargeFee            float64         `json:"rechargeFee"`
	RechargeFeeCurrency    common.Currency `json:"rechargeFeeCurrency"`
	ExchangeRate           float64         `json:"exchangeRate"`
	EffectiveQuotationTime int             `json:"effectiveQuotationTime"`
	QuotedAt               string          `json:"quotedAt"`
}

func (service *PhotonPayOpenAPIService) PreRecharge(ctx context.Context, req *OpenAPIPreRechargeRequest) (*OpenAPIPreRechargeData, error) {
	accountID, err := service.accountID(&req.OpenAPIAccountRequest)
	if err != nil {
		return nil, err
	}
	selectedAccountID, err := photonPayAccountID(req.AccountID)
	if err != nil {
		return nil, err
	}
	if accountID != selectedAccountID {
		return nil, biz.ErrInvalidOperation
	}
	if _, err := service.CardDetail(ctx, &CardIDRequest{
		OpenAPIAccountRequest: req.OpenAPIAccountRequest,
		CardID:                req.CardID,
	}); err != nil {
		return nil, err
	}
	if (req.RechargeAmount == nil) == (req.ArrivalAmount == nil) {
		return nil, biz.ErrInvalidOperation
	}
	amount := types.Value(req.RechargeAmount)
	if req.ArrivalAmount != nil {
		amount = *req.ArrivalAmount
	}
	if amount <= 0 {
		return nil, biz.ErrInvalidOperation
	}
	return &OpenAPIPreRechargeData{
		AccountID:              photonPayIDString(accountID),
		RequestID:              req.RequestID,
		ArrivalAmount:          amount,
		ArrivalAmountCurrency:  common.Currency_USD,
		RechargeAmount:         amount,
		RechargeCurrency:       common.Currency_USD,
		RechargeFeeCurrency:    common.Currency_USD,
		ExchangeRate:           1,
		EffectiveQuotationTime: 60,
		QuotedAt:               time.Now().UTC().Format("2006-01-02T15:04:05"),
	}, nil
}

type OpenAPIRechargeRequest struct {
	OpenAPIAccountRequest
	MemberID     *string  `json:"memberId"` // Invalid: the token selects the account; Matrix members are unsupported.
	RequestID    string   `json:"requestId" binding:"required"`
	CardID       *string  `json:"cardId"`
	ReturnAmount *float64 `json:"returnAmount"` // Invalid: recharge quotes and reversals are protocol-only; balances are managed through UI.
}

type OpenAPIRechargeData struct {
	CardID                string                 `json:"cardId,omitempty"`
	CardBalance           float64                `json:"cardBalance"`
	ArrivalAmount         float64                `json:"arrivalAmount"`
	ArrivalAmountCurrency common.Currency        `json:"arrivalAmountCurrency"`
	RechargeAmount        float64                `json:"rechargeAmount"`
	RechargeCurrency      common.Currency        `json:"rechargeCurrency"`
	ExchangeRate          float64                `json:"exchangeRate"`
	Status                photon.OperationStatus `json:"status"`
	CreatedAt             string                 `json:"createdAt"`
}

func (service *PhotonPayOpenAPIService) Recharge(ctx context.Context, req *OpenAPIRechargeRequest) (*OpenAPIRechargeData, error) {
	if _, err := service.accountID(&req.OpenAPIAccountRequest); err != nil {
		return nil, err
	}
	result := &OpenAPIRechargeData{
		ArrivalAmountCurrency: common.Currency_USD,
		RechargeCurrency:      common.Currency_USD,
		ExchangeRate:          1,
		Status:                photon.OperationStatus_Succeed,
		CreatedAt:             time.Now().UTC().Format("2006-01-02T15:04:05"),
	}
	if req.CardID != nil {
		card, err := service.CardDetail(ctx, &CardIDRequest{
			OpenAPIAccountRequest: req.OpenAPIAccountRequest,
			CardID:                *req.CardID,
		})
		if err != nil {
			return nil, err
		}
		result.CardID = card.CardID
	}
	if req.ReturnAmount != nil && *req.ReturnAmount <= 0 {
		return nil, biz.ErrInvalidOperation
	}
	return result, nil
}
