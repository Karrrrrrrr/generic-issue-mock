package service

import (
	"context"
	"time"

	photon "generic-mock/channel/photonpay/enums"
	photonpayerrors "generic-mock/channel/photonpay/errors"
	"generic-mock/channel/photonpay/pkg/idconv"
	common "generic-mock/enums"
	"generic-mock/pkg/types"
	timeTypes "generic-mock/pkg/types/time"
)

type OpenAPIAccountHistoryRequest struct {
	OpenAPIAccountRequest
	TransactedAtStart time.Time `form:"transactedAtStart" time_format:"2006-01-02T15:04:05" time_utc:"1" binding:"required"`
	TransactedAtEnd   time.Time `form:"transactedAtEnd" time_format:"2006-01-02T15:04:05" time_utc:"1" binding:"required"`
}

type OpenAPIAccountHistoryData struct {
	AccountNo    string                `json:"accountNo"`
	Currency     common.Currency       `json:"currency"`
	Amount       float64               `json:"amount"`
	BalanceFund  float64               `json:"balanceFund"`
	TransactedAt timeTypes.ISODateTime `json:"transactedAt"`
}

func (service *PhotonPayOpenAPIService) AccountHistory(ctx context.Context, req *OpenAPIAccountHistoryRequest) (*OpenAPIAccountHistoryData, error) {
	accountID, err := idconv.FromAccountString(req.Token)
	if err != nil {
		return nil, err
	}
	return &OpenAPIAccountHistoryData{
		AccountNo:    idconv.ToString(accountID),
		Currency:     common.Currency_USD,
		TransactedAt: timeTypes.ISODateTime(time.Now().UTC()),
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
	AccountID              string                `json:"accountId"`
	RequestID              string                `json:"requestId"`
	ArrivalAmount          float64               `json:"arrivalAmount"`
	ArrivalAmountCurrency  common.Currency       `json:"arrivalAmountCurrency"`
	RechargeAmount         float64               `json:"rechargeAmount"`
	RechargeCurrency       common.Currency       `json:"rechargeCurrency"`
	RechargeFee            float64               `json:"rechargeFee"`
	RechargeFeeCurrency    common.Currency       `json:"rechargeFeeCurrency"`
	ExchangeRate           float64               `json:"exchangeRate"`
	EffectiveQuotationTime int                   `json:"effectiveQuotationTime"`
	QuotedAt               timeTypes.ISODateTime `json:"quotedAt"`
}

func (service *PhotonPayOpenAPIService) PreRecharge(ctx context.Context, req *OpenAPIPreRechargeRequest) (*OpenAPIPreRechargeData, error) {
	accountID, err := idconv.FromAccountString(req.Token)
	if err != nil {
		return nil, err
	}
	selectedAccountID, err := idconv.FromAccountString(req.AccountID)
	if err != nil {
		return nil, err
	}
	if accountID != selectedAccountID {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	if _, err := service.CardDetail(ctx, &CardIDRequest{
		OpenAPIAccountRequest: req.OpenAPIAccountRequest,
		CardID:                req.CardID,
	}); err != nil {
		return nil, err
	}
	if (req.RechargeAmount == nil) == (req.ArrivalAmount == nil) {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	amount := types.Value(req.RechargeAmount)
	if req.ArrivalAmount != nil {
		amount = *req.ArrivalAmount
	}
	if amount <= 0 {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	return &OpenAPIPreRechargeData{
		AccountID:              idconv.ToString(accountID),
		RequestID:              req.RequestID,
		ArrivalAmount:          amount,
		ArrivalAmountCurrency:  common.Currency_USD,
		RechargeAmount:         amount,
		RechargeCurrency:       common.Currency_USD,
		RechargeFeeCurrency:    common.Currency_USD,
		ExchangeRate:           1,
		EffectiveQuotationTime: 60,
		QuotedAt:               timeTypes.ISODateTime(time.Now().UTC()),
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
	CreatedAt             timeTypes.ISODateTime  `json:"createdAt"`
}

func (service *PhotonPayOpenAPIService) Recharge(ctx context.Context, req *OpenAPIRechargeRequest) (*OpenAPIRechargeData, error) {
	if _, err := idconv.FromAccountString(req.Token); err != nil {
		return nil, err
	}
	result := &OpenAPIRechargeData{
		ArrivalAmountCurrency: common.Currency_USD,
		RechargeCurrency:      common.Currency_USD,
		ExchangeRate:          1,
		Status:                photon.OperationStatus_Succeed,
		CreatedAt:             timeTypes.ISODateTime(time.Now().UTC()),
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
		return nil, photonpayerrors.ErrInvalidOperation
	}
	return result, nil
}
