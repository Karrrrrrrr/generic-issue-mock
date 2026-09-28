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

type ListVirtualAccountsRequest struct {
	PageRequest
	ID        *model.ID `form:"id" binding:"omitempty,gt=0"`
	AccountID *model.ID `form:"account_id" binding:"omitempty,gt=0"`
}

type VirtualAccountData struct {
	ID          model.ID        `json:"id"`
	AccountID   model.ID        `json:"account_id"`
	AccountName string          `json:"account_name"`
	Name        string          `json:"name"`
	WalletID    model.ID        `json:"wallet_id"`
	Available   decimal.Decimal `json:"available"`
	Currency    enums.Currency  `json:"currency"`
	CreatedAt   time.Time       `json:"created_at"`
}

func (s *Service) ListVirtualAccounts(ctx context.Context, req *ListVirtualAccountsRequest) (*Page[VirtualAccountData], error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	page, err := req.PageRequest.toBizPage()
	if err != nil {
		return nil, err
	}
	items, total, err := s.uc.ListVirtualAccounts(ctx, &biz.ListUIVirtualAccountsRequest{
		UIPageRequest: page,
		ID:            req.ID,
		AccountID:     req.AccountID,
	})
	if err != nil {
		return nil, err
	}
	result := &Page[VirtualAccountData]{
		Items: make([]VirtualAccountData, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		result.Items = append(result.Items, toVirtualAccountData(item))
	}
	return result, nil
}

func toVirtualAccountData(item *model.VirtualAccount) VirtualAccountData {
	available := decimal.Zero
	var currency enums.Currency
	if item.Wallet != nil {
		available = item.Wallet.Available
		currency = item.Wallet.Currency
	}
	return VirtualAccountData{
		ID:          item.ID,
		AccountID:   item.AccountID,
		AccountName: item.Account.GetName(),
		Name:        item.Name,
		WalletID:    item.WalletID,
		Available:   available,
		Currency:    currency,
		CreatedAt:   item.CreatedAt,
	}
}

type CreateVirtualAccountRequest struct {
	AccountID model.ID       `json:"account_id" binding:"required,gt=0"`
	Name      string         `json:"name" binding:"required"`
	Currency  enums.Currency `json:"currency" binding:"required"`
}

func (s *Service) CreateVirtualAccount(ctx context.Context, req *CreateVirtualAccountRequest) (*VirtualAccountData, error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	item, err := s.uc.CreateVirtualAccount(ctx, &biz.CreateUIVirtualAccountRequest{
		AccountID: req.AccountID,
		Name:      req.Name,
		Currency:  req.Currency,
	})
	if err != nil {
		return nil, err
	}
	result := toVirtualAccountData(item)
	return &result, nil
}
