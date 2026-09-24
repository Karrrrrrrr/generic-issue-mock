package service

import (
	"context"

	"generic-mock/channel/paynda/biz"
	paynda "generic-mock/channel/paynda/enums"
)

type OpenAPICardControlRequest struct {
	PayndaCardRequest
	Period           *paynda.CardControlPeriod `json:"period"`           // Invalid: velocity controls are not persisted.
	TransactionCount *int64                    `json:"transactionCount"` // Invalid: velocity controls are not persisted.
	Amount           *string                   `json:"amount"`           // Invalid: velocity controls are not persisted.
}

type OpenAPICardControlData struct {
	Period           paynda.CardControlPeriod `json:"period"`
	TransactionCount int64                    `json:"transactionCount"`
	Amount           string                   `json:"amount"`
}

func (service *PayndaOpenAPIService) GetCardControls(ctx context.Context, req *OpenAPICardControlRequest) (*[]OpenAPICardControlData, error) {
	if _, err := service.GetCard(ctx, &req.PayndaCardRequest); err != nil {
		return nil, err
	}
	items := []OpenAPICardControlData{{
		Period: paynda.CardControlPeriodDay,
		Amount: "0",
	}}
	return &items, nil
}

func (service *PayndaOpenAPIService) UpdateCardControls(ctx context.Context, req *OpenAPICardControlRequest) (*struct{}, error) {
	_, err := service.GetCardControls(ctx, req)
	return &struct{}{}, err
}

type OpenAPICardBalanceUpdateRequest struct {
	BalanceAccountID string `uri:"balanceAccountId" binding:"required"`
	CardID           string `json:"cardId" binding:"required"`
	RequestID        string `form:"requestId" binding:"required"`
	Amount           string `json:"amount" binding:"required"`
}

func (service *PayndaOpenAPIService) UpdateCardBalance(ctx context.Context, req *OpenAPICardBalanceUpdateRequest) (*PayndaCardBalanceData, error) {
	_, err := service.TransferCardBalance(ctx, &PayndaCardBalanceTransferRequest{
		BalanceAccountID: req.BalanceAccountID,
		CardID:           req.CardID,
		RequestID:        req.RequestID,
		Amount:           req.Amount,
		Type:             paynda.TransferType_In,
	})
	if err != nil {
		return nil, err
	}
	return service.GetCardBalance(ctx, &PayndaCardRequest{
		BalanceAccountID: req.BalanceAccountID,
		CardID:           req.CardID,
	})
}

type OpenAPICardholderWalletRequest struct {
	BalanceAccountID string                       `uri:"balanceAccountId" binding:"required"`
	CardholderID     string                       `uri:"cardholderId" json:"cardholderId" binding:"required"`
	Type             *paynda.BalanceOperationType `json:"type"`   // Invalid: independent cardholder wallets are not persisted.
	Amount           *string                      `json:"amount"` // Invalid: independent cardholder wallets are not persisted.
}

type OpenAPICardholderWalletData struct {
	PayndaBalanceAccountWalletData
	CardholderID string `json:"cardholderId"`
}

func (service *PayndaOpenAPIService) CardholderWallet(ctx context.Context, req *OpenAPICardholderWalletRequest) (*OpenAPICardholderWalletData, error) {
	accountID, err := payndaAccountID(req.BalanceAccountID)
	if err != nil {
		return nil, err
	}
	holderID, err := payndaID(req.CardholderID)
	if err != nil {
		return nil, err
	}
	if _, err := service.usecase.GetCardHolder(ctx, &biz.PayndaResourceRequest{
		AccountID: accountID,
		ID:        holderID,
	}); err != nil {
		return nil, err
	}
	item, err := service.usecase.GetAccountWallet(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return &OpenAPICardholderWalletData{
		PayndaBalanceAccountWalletData: *payndaBalanceAccountWalletData(item, req.BalanceAccountID),
		CardholderID:                   payndaIDString(holderID),
	}, nil
}

func (service *PayndaOpenAPIService) ListCardholderWallets(ctx context.Context, req *OpenAPICardholderWalletRequest) (*[]*OpenAPICardholderWalletData, error) {
	item, err := service.CardholderWallet(ctx, req)
	if err != nil {
		return nil, err
	}
	items := []*OpenAPICardholderWalletData{item}
	return &items, nil
}

func (service *PayndaOpenAPIService) DeleteCardholder(ctx context.Context, req *OpenAPICardholderWalletRequest) (*struct{}, error) {
	_, err := service.CardholderWallet(ctx, req)
	return &struct{}{}, err
}
