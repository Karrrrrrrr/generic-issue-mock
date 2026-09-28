package service

import (
	"context"
	"generic-mock/pkg/types"
	"time"

	"generic-mock/channel/slash/biz"
	"generic-mock/channel/slash/pkg/idconv"
	common "generic-mock/enums"

	"github.com/shopspring/decimal"
)

type ListAuthorizationBalancesRequest struct {
	UIListTimeRange
	Status       *common.CardTransactionStatus `form:"status" binding:"omitempty,oneof=pending authorized succeed failed void"`
	AccountID    *string                       `form:"account_id"`
	ID           *string                       `form:"id" binding:"omitempty,min=1"`
	CardID       *string                       `form:"card_id" binding:"omitempty,min=1"`
	MerchantName *string                       `form:"merchant_name" binding:"omitempty,min=1"`
}

func (s *SlashUIService) ListAuthorizationBalances(ctx context.Context, req *ListAuthorizationBalancesRequest) (*[]AuthorizationBalanceData, error) {
	accountID, err := idconv.FromOptionalUUID(req.AccountID)
	if err != nil {
		return nil, err
	}
	id, err := idconv.FromOptionalUUID(req.ID)
	if err != nil {
		return nil, err
	}
	cardID, err := idconv.FromOptionalUUID(req.CardID)
	if err != nil {
		return nil, err
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
	AccountID    string                       `json:"account_id"`
	ID           string                       `json:"id"`
	CardID       string                       `json:"card_id"`
	Currency     common.Currency              `json:"currency"`
	Amount       string                       `json:"amount"`
	Settled      string                       `json:"settled"`
	Remaining    string                       `json:"remaining"`
	MerchantName string                       `json:"merchant_name"`
	CreatedAt    time.Time                    `json:"created_at"`
}

type ClearAuthorizationRequest struct {
	ID     string          `uri:"id" binding:"required"`
	Amount decimal.Decimal `json:"amount"`
}
type ClearAuthorizationData struct {
	ID string `json:"id"`
}

func (s *SlashUIService) ClearAuthorization(ctx context.Context, req *ClearAuthorizationRequest) (*ClearAuthorizationData, error) {
	authID, err := idconv.FromUUID(req.ID)
	if err != nil {
		return nil, err
	}
	item, err := s.usecase.ClearAuthorization(ctx, &biz.ClearAuthorizationRequest{
		Notificator: s,
		ID:          authID,
		Amount:      req.Amount,
	})
	if err != nil {
		return nil, err
	}
	return &ClearAuthorizationData{ID: idconv.ToUUID(item.ID)}, nil
}

func authorizationBalanceData(item *biz.AuthorizationBalance) AuthorizationBalanceData {
	auth := item.Authorization
	return AuthorizationBalanceData{
		AccountID:    idconv.ToUUID(auth.AccountID),
		AccountName:  uiAccountName(auth.Account),
		ID:           idconv.ToUUID(auth.ID),
		CardID:       idconv.ToUUID(auth.CardID),
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
