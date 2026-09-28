package service

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/shared/biz"
	sharederrors "generic-mock/shared/errors"

	"github.com/shopspring/decimal"
)

type ListWalletsRequest struct {
	PageRequest
	ID        *model.ID         `form:"id" binding:"omitempty,gt=0"`
	AccountID *model.ID         `form:"account_id" binding:"omitempty,gt=0"`
	Type      *enums.WalletType `form:"type"`
}

type WalletData struct {
	ID          model.ID         `json:"id"`
	AccountID   model.ID         `json:"account_id"`
	AccountName string           `json:"account_name"`
	Type        enums.WalletType `json:"type"`
	Currency    enums.Currency   `json:"currency"`
	Available   decimal.Decimal  `json:"available"`
	PendingOut  decimal.Decimal  `json:"pending_out"`
	In          decimal.Decimal  `json:"in"`
	Out         decimal.Decimal  `json:"out"`
}

func (s *Service) ListWallets(ctx context.Context, req *ListWalletsRequest) (*Page[WalletData], error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	page, err := req.PageRequest.toBizPage()
	if err != nil {
		return nil, err
	}
	items, total, err := s.uc.ListWallets(ctx, &biz.ListUIWalletsRequest{
		UIPageRequest: page,
		ID:            req.ID,
		AccountID:     req.AccountID,
		Type:          req.Type,
	})
	if err != nil {
		return nil, err
	}
	result := &Page[WalletData]{
		Items: make([]WalletData, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		result.Items = append(result.Items, toWalletData(item))
	}
	return result, nil
}

func toWalletData(item *model.Wallet) WalletData {
	return WalletData{
		ID:          item.ID,
		AccountID:   item.AccountID,
		AccountName: item.Account.GetName(),
		Type:        item.Type,
		Currency:    item.Currency,
		Available:   item.Available,
		PendingOut:  item.PendingOut,
		In:          item.In,
		Out:         item.Out,
	}
}
