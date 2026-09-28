package service

import (
	"context"
	"time"

	"generic-mock/channel/pingpong/biz"
	pingerrors "generic-mock/channel/pingpong/errors"
	common "generic-mock/enums"
	"generic-mock/model"
)

type AccountData struct {
	WalletID  model.ID        `json:"wallet_id"`
	ID        model.ID        `json:"id"`
	Name      string          `json:"name"`
	Balance   Number          `json:"balance"`
	Currency  common.Currency `json:"currency"`
	CreatedAt time.Time       `json:"created_at"`
}

type UIUpdateAccountRequest struct {
	ID   model.ID `uri:"id" binding:"required,gt=0"`
	Name string   `json:"name" binding:"required"`
}

func (s *PingPongUIService) UpdateAccount(ctx context.Context, req *UIUpdateAccountRequest) (*AccountData, error) {
	accountID := req.ID
	if accountID <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	item, err := s.uc.UpdateAccount(ctx, &biz.UpdateAccountRequest{
		AccountID: accountID,
		Name:      req.Name,
	})
	if err != nil {
		return nil, err
	}
	result := toAccountData(item)
	return &result, nil
}

type UICreateAccountRequest struct {
	Name string `json:"name" binding:"required"`
}

type UIAccountBalanceRequest struct {
	ID     model.ID `uri:"id" binding:"required,gt=0"`
	Amount Number   `json:"amount"`
}

type UIResourceRequest struct {
	AccountID model.ID `json:"account_id" binding:"required,gt=0"`
	ID        model.ID `uri:"id" binding:"required,gt=0"`
}

func toAccountData(item *model.Account) AccountData {
	result := AccountData{
		ID:        item.ID,
		WalletID:  item.WalletID,
		Name:      item.Name,
		CreatedAt: item.CreatedAt,
	}
	if item.Wallet != nil {
		result.Balance = Number{item.Wallet.Available}
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
	id := req.ID
	if id <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	if err := s.uc.AdjustAccount(ctx, &biz.AdjustAccountRequest{
		AccountID: id,
		Amount:    req.Amount.Decimal,
	}); err != nil {
		return nil, err
	}
	return &Empty{}, nil
}

func (req *UIResourceRequest) Validate() error {
	if req == nil || req.AccountID <= 0 || req.ID <= 0 {
		return pingerrors.ErrInvalid
	}
	return nil
}
