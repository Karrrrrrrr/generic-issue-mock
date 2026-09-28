package service

import (
	"context"
	"time"

	"generic-mock/channel/photonpay/biz"
	photon "generic-mock/channel/photonpay/enums"
	photonpayerrors "generic-mock/channel/photonpay/errors"
	"generic-mock/channel/photonpay/pkg/idconv"
	common "generic-mock/enums"

	"github.com/shopspring/decimal"
)

type ListAuthorizationBalancesRequest struct {
	UIListTimeRange
	Status       *photon.AuthorizationStatus `form:"status" binding:"omitempty,oneof=pending authorized declined void"`
	AccountID    *string                     `form:"account_id"`
	ID           *string                     `form:"id" binding:"omitempty,min=1"`
	CardID       *string                     `form:"card_id" binding:"omitempty,min=1"`
	MerchantName *string                     `form:"merchant_name" binding:"omitempty,min=1"`
}

func (s *PhotonPayUIService) ListAuthorizationBalances(ctx context.Context, req *ListAuthorizationBalancesRequest) (*[]AuthorizationBalanceData, error) {
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
	statuses := uiAuthorizationStatuses(req.Status)
	if req.Status != nil && len(statuses) == 0 {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	items, err := s.usecase.ListAuthorizationBalances(ctx, &biz.ListAuthorizationBalancesRequest{
		AccountID:    accountID,
		Statuses:     statuses,
		ID:           id,
		CardID:       cardID,
		MerchantName: req.MerchantName,
		CreatedFrom:  req.CreatedFrom,
		CreatedTo:    req.CreatedTo,
	})
	if err != nil {
		return nil, err
	}
	result := make([]AuthorizationBalanceData, 0, len(items))
	for _, item := range items {
		result = append(result, authorizationBalanceData(item))
	}
	return &result, nil
}

type AuthorizationBalanceData struct {
	Status       photon.AuthorizationStatus `json:"status"`
	Reversed     string                     `json:"reversed"`
	Refunded     string                     `json:"refunded"`
	AccountName  string                     `json:"account_name"`
	AccountID    string                     `json:"account_id"`
	ID           string                     `json:"id"`
	CardID       string                     `json:"card_id"`
	Currency     common.Currency            `json:"currency"`
	Amount       string                     `json:"amount"`
	Settled      string                     `json:"settled"`
	Remaining    string                     `json:"remaining"`
	MerchantName string                     `json:"merchant_name"`
	CreatedAt    time.Time                  `json:"created_at"`
}

type ClearAuthorizationRequest struct {
	ManagementAccountRequest
	ID     string          `uri:"id" binding:"required"`
	Amount decimal.Decimal `json:"amount"`
}
type ClearAuthorizationData struct {
	ID string `json:"id"`
}

func (s *PhotonPayUIService) ClearAuthorization(ctx context.Context, req *ClearAuthorizationRequest) (*ClearAuthorizationData, error) {
	accountID, err := idconv.FromAccountString(req.AccountID)
	if err != nil {
		return nil, err
	}
	authID, err := idconv.FromString(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.ClearAuthorization(ctx, &biz.ClearAuthorizationRequest{
		AccountID: accountID,
		ID:        authID,
		Amount:    req.Amount,
	})
	if err != nil {
		return nil, err
	}

	return &ClearAuthorizationData{ID: idconv.ToString(item.ID)}, nil
}

func authorizationBalanceData(item *biz.AuthorizationBalance) AuthorizationBalanceData {
	auth := item.Authorization
	return AuthorizationBalanceData{
		AccountID:    idconv.ToString(auth.AccountID),
		AccountName:  uiAccountName(auth.Account),
		ID:           idconv.ToString(auth.ID),
		CardID:       idconv.ToString(auth.CardID),
		Status:       photon.AuthorizationStatusFromGeneric(auth.Status),
		Currency:     auth.Currency,
		Amount:       auth.Amount.String(),
		Settled:      item.Settled.String(),
		Reversed:     item.Reversed.String(),
		Refunded:     item.Refunded.String(),
		Remaining:    item.Remaining.String(),
		MerchantName: auth.MerchantName,
		CreatedAt:    auth.CreatedAt,
	}
}

func uiAuthorizationStatuses(value *photon.AuthorizationStatus) []common.CardTransactionStatus {
	if value == nil {
		return nil
	}
	var result []common.CardTransactionStatus
	for _, status := range []common.CardTransactionStatus{
		common.TransactionStatus_PENDING,
		common.TransactionStatus_AUTHORIZED,
		common.TransactionStatus_SUCCEED,
		common.TransactionStatus_FAILED,
		common.TransactionStatus_VOID,
	} {
		if photon.AuthorizationStatusFromGeneric(status) == *value {
			result = append(result, status)
		}
	}
	return result
}
