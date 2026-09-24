package service

import (
	"context"

	photon "generic-mock/channel/photonpay/enums"
	common "generic-mock/enums"
	timeTypes "generic-mock/pkg/types/time"
)

type OpenAPIIssuingHistoryData struct {
	CardID            string                 `json:"cardId"`
	CardType          photon.CardType        `json:"cardType"`
	CardFormFactor    photon.CardFormFactor  `json:"cardFormFactor"`
	CreatedAt         timeTypes.ISODateTime  `json:"createdAt"`
	Status            photon.OperationStatus `json:"status"`
	ActualFeeAmount   string                 `json:"actualFeeAmount"`
	ActualFeeCurrency common.Currency        `json:"actualFeeCurrency"`
	MaskCardNo        string                 `json:"maskCardNo"`
}

func (service *PhotonPayOpenAPIService) IssuingHistory(ctx context.Context, req *ListCardsRequest) (*[]OpenAPIIssuingHistoryData, error) {
	cards, err := service.ListCards(ctx, req)
	if err != nil {
		return nil, err
	}
	items := make([]OpenAPIIssuingHistoryData, 0, len(*cards))
	for _, card := range *cards {
		items = append(items, OpenAPIIssuingHistoryData{
			CardID:            card.CardID,
			CardType:          card.CardType,
			CardFormFactor:    card.CardFormFactor,
			CreatedAt:         card.CreatedAt,
			Status:            photon.OperationStatus_Succeed,
			ActualFeeAmount:   "0",
			ActualFeeCurrency: card.CardCurrency,
			MaskCardNo:        card.MaskCardNo,
		})
	}
	return &items, nil
}

type OpenAPICardFundsData struct {
	CardID                    string                `json:"cardId"`
	CreatedAt                 timeTypes.ISODateTime `json:"createdAt"`
	Amount                    float64               `json:"amount"`
	CardBalance               float64               `json:"cardBalance"`
	AvailableTransactionLimit float64               `json:"availableTransactionLimit"`
	ChangeAmount              float64               `json:"changeAmount"`
	FeeAmount                 float64               `json:"feeAmount"`
	CardCurrency              common.Currency       `json:"cardCurrency"`
	CardFormFactor            photon.CardFormFactor `json:"cardFormFactor"`
	MaskCardNo                string                `json:"maskCardNo"`
}

func (service *PhotonPayOpenAPIService) CardFundsHistory(ctx context.Context, req *ListCardsRequest) (*[]OpenAPICardFundsData, error) {
	cards, err := service.ListCards(ctx, req)
	if err != nil {
		return nil, err
	}
	items := make([]OpenAPICardFundsData, 0, len(*cards))
	for _, card := range *cards {
		items = append(items, OpenAPICardFundsData{
			CardID:         card.CardID,
			CreatedAt:      card.CreatedAt,
			CardCurrency:   card.CardCurrency,
			CardFormFactor: card.CardFormFactor,
			MaskCardNo:     card.MaskCardNo,
		})
	}
	return &items, nil
}

type OpenAPIBillingAddressRequest struct {
	CardIDRequest
	BillingAddress    *string `json:"billingAddress"`    // Invalid: card-specific billing overrides are not persisted.
	BillingCity       *string `json:"billingCity"`       // Invalid: card-specific billing overrides are not persisted.
	BillingCountry    *string `json:"billingCountry"`    // Invalid: card-specific billing overrides are not persisted.
	BillingPostalCode *string `json:"billingPostalCode"` // Invalid: card-specific billing overrides are not persisted.
	BillingState      *string `json:"billingState"`      // Invalid: card-specific billing overrides are not persisted.
}

func (service *PhotonPayOpenAPIService) EditCardBillingAddress(ctx context.Context, req *OpenAPIBillingAddressRequest) (*struct{}, error) {
	_, err := service.CardDetail(ctx, &req.CardIDRequest)
	return &struct{}{}, err
}
