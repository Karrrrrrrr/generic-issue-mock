package service

import (
	"context"

	"generic-mock/channel/pingpong/biz"
	pingerrors "generic-mock/channel/pingpong/errors"
	"generic-mock/channel/pingpong/pkg/idconv"
	common "generic-mock/enums"
)

type CreateBudgetRequest struct {
	OpenAPIRequest
	BudgetName string `json:"budget_name" binding:"required"`
}

type BudgetIDData struct {
	BudgetID string `json:"budget_id"`
}

type BudgetBalanceRequest struct {
	OpenAPIRequest
	BudgetID *string `form:"budget_id"`
}

type BudgetBalance struct {
	Balance    Number          `json:"balance"`
	Currency   common.Currency `json:"currency"`
	BudgetID   string          `json:"budget_id"`
	BudgetName string          `json:"budget_name"`
}

type BudgetBalancesData struct {
	BalanceList []BudgetBalance `json:"balance_list"`
}

func (s *PingPongOpenAPIService) CreateBudget(ctx context.Context, req *CreateBudgetRequest) (*BudgetIDData, error) {
	accountID, err := s.resolveAccountID(ctx, &req.OpenAPIRequest)
	if err != nil {
		return nil, err
	}
	item, err := s.uc.CreateVirtualAccount(ctx, &biz.CreateVirtualAccountRequest{
		AccountID: accountID,
		Name:      req.BudgetName,
	})
	if err != nil {
		return nil, err
	}
	return &BudgetIDData{BudgetID: idconv.ToString(item.ID)}, nil
}

func (s *PingPongOpenAPIService) ListBudgetBalances(ctx context.Context, req *BudgetBalanceRequest) (*BudgetBalancesData, error) {
	accountID, err := s.resolveAccountID(ctx, &req.OpenAPIRequest)
	if err != nil {
		return nil, err
	}
	virtualAccountID, err := idconv.FromOptionalString(req.BudgetID)
	if err != nil {
		return nil, err
	}
	items, err := s.uc.VirtualAccountBalances(ctx, &biz.VirtualAccountBalancesRequest{
		AccountID: accountID,
		ID:        virtualAccountID,
	})
	if err != nil {
		return nil, err
	}
	result := &BudgetBalancesData{BalanceList: make([]BudgetBalance, 0, len(items))}
	for _, item := range items {
		if item.Wallet == nil {
			return nil, pingerrors.ErrInvalid
		}
		result.BalanceList = append(result.BalanceList, BudgetBalance{
			BudgetID:   idconv.ToString(item.ID),
			BudgetName: item.Name,
			Currency:   item.Wallet.Currency,
			Balance:    Number{item.Wallet.Amount.Sub(item.Wallet.PendingOut)},
		})
	}
	return result, nil
}
