package service

import (
	"context"
	"time"

	"generic-mock/channel/photonpay/biz"
	photonpayerrors "generic-mock/channel/photonpay/errors"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"github.com/shopspring/decimal"
)

type ListAuthorizationBalancesRequest struct {
	UIListTimeRange
	Status       *common.CardTransactionStatus `form:"status" binding:"omitempty,oneof=pending authorized succeed failed void"`
	AccountID    *model.ID                     `form:"account_id" binding:"omitempty,gt=0"`
	ID           *model.ID                     `form:"id" binding:"omitempty,min=1"`
	CardID       *model.ID                     `form:"card_id" binding:"omitempty,min=1"`
	MerchantName *string                       `form:"merchant_name" binding:"omitempty,min=1"`
}

func (s *PhotonPayUIService) ListAuthorizationBalances(ctx context.Context, req *ListAuthorizationBalancesRequest) (*[]AuthorizationBalanceData, error) {
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
	items, err := s.usecase.ListAuthorizationBalances(ctx, &biz.ListAuthorizationBalancesRequest{
		AccountID:    accountID,
		Statuses:     types.PointerSlice(req.Status),
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
	Status       common.CardTransactionStatus `json:"status"`
	Reversed     string                       `json:"reversed"`
	Refunded     string                       `json:"refunded"`
	AccountName  string                       `json:"account_name"`
	AccountID    model.ID                     `json:"account_id"`
	ID           model.ID                     `json:"id"`
	CardID       model.ID                     `json:"card_id"`
	Currency     common.Currency              `json:"currency"`
	Amount       string                       `json:"amount"`
	Settled      string                       `json:"settled"`
	Remaining    string                       `json:"remaining"`
	MerchantName string                       `json:"merchant_name"`
	CreatedAt    time.Time                    `json:"created_at"`
}

type ClearAuthorizationRequest struct {
	ID     model.ID        `uri:"id" binding:"required,gt=0"`
	Amount decimal.Decimal `json:"amount"`
}
type ClearAuthorizationData struct {
	ID model.ID `json:"id"`
}

func (s *PhotonPayUIService) ClearAuthorization(ctx context.Context, req *ClearAuthorizationRequest) (*ClearAuthorizationData, error) {
	authID := req.ID
	if authID <= 0 {
		return nil, photonpayerrors.ErrInvalidOperation
	}
	item, err := s.usecase.ClearAuthorization(ctx, &biz.ClearAuthorizationRequest{
		ID:     authID,
		Amount: req.Amount,
	})
	if err != nil {
		return nil, err
	}

	return &ClearAuthorizationData{ID: item.ID}, nil
}

func authorizationBalanceData(item *biz.AuthorizationBalance) AuthorizationBalanceData {
	auth := item.Authorization
	return AuthorizationBalanceData{
		AccountID:    auth.AccountID,
		AccountName:  uiAccountName(auth.Account),
		ID:           auth.ID,
		CardID:       auth.CardID,
		Status:       auth.Status,
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
