package service

import (
	"context"
	"time"

	"generic-mock/channel/pingpong/biz"
	"generic-mock/channel/pingpong/pkg/idconv"
	common "generic-mock/enums"
	"generic-mock/model"
)

type AccountData struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Balance   Number          `json:"balance"`
	Currency  common.Currency `json:"currency"`
	CreatedAt time.Time       `json:"created_at"`
}

type UICreateAccountRequest struct {
	Name string `json:"name" binding:"required"`
}

type UIAccountBalanceRequest struct {
	ID     string `uri:"id" binding:"required"`
	Amount Number `json:"amount"`
}

type UIResourceRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	ID        string `uri:"id" binding:"required"`
}

func toAccountData(item *model.Account) AccountData {
	result := AccountData{
		ID:        idconv.ToString(item.ID),
		Name:      item.Name,
		CreatedAt: item.CreatedAt,
	}
	if item.Wallet != nil {
		result.Balance = Number{item.Wallet.Amount.Sub(item.Wallet.PendingOut)}
		result.Currency = item.Wallet.Currency
	}
	return result
}

func (s *PingPongUIService) ListAccounts(ctx context.Context, req *PageRequest) (*UIPage[AccountData], error) {
	page, limit, err := req.resolvePagination()
	if err != nil {
		return nil, err
	}
	items, total, err := s.uc.ListAccounts(ctx, &biz.ListAccountsRequest{
		Offset: (page - 1) * limit,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}
	result := &UIPage[AccountData]{
		Items: make([]AccountData, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		result.Items = append(result.Items, toAccountData(item))
	}
	return result, nil
}

func (s *PingPongUIService) CreateAccount(ctx context.Context, req *UICreateAccountRequest) (*AccountData, error) {
	item, err := s.uc.CreateAccount(ctx, &biz.CreateAccountRequest{Name: req.Name})
	if err != nil {
		return nil, err
	}
	result := toAccountData(item)
	return &result, nil
}

func (s *PingPongUIService) AdjustAccountBalance(ctx context.Context, req *UIAccountBalanceRequest) (*Empty, error) {
	id, err := idconv.FromString(req.ID)
	if err != nil {
		return nil, err
	}
	if err := s.uc.AdjustAccount(ctx, &biz.AdjustAccountRequest{
		AccountID: id,
		Amount:    req.Amount.Decimal,
	}); err != nil {
		return nil, err
	}
	return &Empty{}, nil
}

func (req *UIResourceRequest) ParseAccountAndResourceIDs() (int64, int64, error) {
	accountID, err := idconv.FromString(req.AccountID)
	if err != nil {
		return 0, 0, err
	}
	id, err := idconv.FromString(req.ID)
	return accountID, id, err
}
