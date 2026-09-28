package service

import (
	"context"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/shared/biz"
	sharederrors "generic-mock/shared/errors"

	"github.com/shopspring/decimal"
)

type ListAccountsRequest struct {
	PageRequest
	ID *model.ID `form:"id" binding:"omitempty,gt=0"`
}

type AccountData struct {
	ID        model.ID        `json:"id"`
	Name      string          `json:"name"`
	WalletID  model.ID        `json:"wallet_id"`
	Available decimal.Decimal `json:"available"`
	Currency  enums.Currency  `json:"currency"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (s *Service) ListAccounts(ctx context.Context, req *ListAccountsRequest) (*Page[AccountData], error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	page, err := req.PageRequest.toBizPage()
	if err != nil {
		return nil, err
	}
	items, total, err := s.uc.ListAccounts(ctx, &biz.ListUIAccountsRequest{
		UIPageRequest: page,
		ID:            req.ID,
	})
	if err != nil {
		return nil, err
	}
	result := &Page[AccountData]{
		Items: make([]AccountData, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		result.Items = append(result.Items, toAccountData(item))
	}
	return result, nil
}

func toAccountData(item *model.Account) AccountData {
	available := decimal.Zero
	var currency enums.Currency
	if item.Wallet != nil {
		available = item.Wallet.Available
		currency = item.Wallet.Currency
	}
	return AccountData{
		ID:        item.ID,
		Name:      item.Name,
		WalletID:  item.WalletID,
		Available: available,
		Currency:  currency,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

type CreateAccountRequest struct {
	Name     string         `json:"name" binding:"required"`
	Currency enums.Currency `json:"currency" binding:"required"`
}

func (s *Service) CreateAccount(ctx context.Context, req *CreateAccountRequest) (*AccountData, error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	item, err := s.uc.CreateAccount(ctx, &biz.CreateUIAccountRequest{
		Name:     req.Name,
		Currency: req.Currency,
	})
	if err != nil {
		return nil, err
	}
	result := toAccountData(item)
	return &result, nil
}

type AdjustAccountRequest struct {
	AccountID model.ID        `json:"account_id" binding:"required,gt=0"`
	Currency  enums.Currency  `json:"currency" binding:"required"`
	Amount    decimal.Decimal `json:"amount" binding:"required"`
}

func (s *Service) AdjustAccount(ctx context.Context, req *AdjustAccountRequest) (*Empty, error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	err := s.uc.AdjustAccount(ctx, &biz.AdjustUIAccountRequest{
		AccountID: req.AccountID,
		Currency:  req.Currency,
		Amount:    req.Amount,
	})
	if err != nil {
		return nil, err
	}
	return &Empty{}, nil
}

type RenameAccountRequest struct {
	ID   model.ID `json:"id" binding:"required,gt=0"`
	Name string   `json:"name" binding:"required"`
}

func (s *Service) RenameAccount(ctx context.Context, req *RenameAccountRequest) (*AccountData, error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	item, err := s.uc.RenameAccount(ctx, &biz.RenameUIAccountRequest{
		ID:   req.ID,
		Name: req.Name,
	})
	if err != nil {
		return nil, err
	}
	result := toAccountData(item)
	return &result, nil
}
