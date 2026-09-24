package service

import (
	"context"

	"generic-mock/channel/slash/biz"
	"generic-mock/channel/slash/pkg/idconv"
)

type ListAuthorizationBalancesRequest struct {
	UIListTimeRange
	AccountID    *string `form:"account_id"`
	ID           *string `form:"id" binding:"omitempty,min=1"`
	CardID       *string `form:"card_id" binding:"omitempty,min=1"`
	MerchantName *string `form:"merchant_name" binding:"omitempty,min=1"`
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
		auth := item.Authorization
		result = append(result, AuthorizationBalanceData{
			AccountID:    idconv.ToUUID(auth.AccountID),
			AccountName:  uiAccountName(auth.Account),
			ID:           idconv.ToUUID(auth.ID),
			CardID:       idconv.ToUUID(auth.CardID),
			Currency:     auth.Currency,
			Amount:       auth.Amount.String(),
			Settled:      item.Settled.String(),
			Remaining:    item.Remaining.String(),
			MerchantName: auth.MerchantName,
			CreatedAt:    auth.CreatedAt,
		})
	}
	return &result, nil
}
