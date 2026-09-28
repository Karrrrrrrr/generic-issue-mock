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

type ListWalletTransfersRequest struct {
	PageRequest
	TimeRange
	ID        *model.ID                 `form:"id" binding:"omitempty,gt=0"`
	AccountID *model.ID                 `form:"account_id" binding:"omitempty,gt=0"`
	CardID    *model.ID                 `form:"card_id"`
	Kind      *enums.WalletTransferKind `form:"kind"`
}

type WalletTransferData struct {
	ID             model.ID                 `json:"id"`
	AccountID      model.ID                 `json:"account_id"`
	AccountName    string                   `json:"account_name"`
	CardID         *model.ID                `json:"card_id"`
	SourceWalletID model.ID                 `json:"source_wallet_id"`
	TargetWalletID model.ID                 `json:"target_wallet_id"`
	Kind           enums.WalletTransferKind `json:"kind"`
	Currency       enums.Currency           `json:"currency"`
	Amount         decimal.Decimal          `json:"amount"`
	RequestID      string                   `json:"request_id"`
	CreatedAt      time.Time                `json:"created_at"`
}

func (s *Service) ListWalletTransfers(ctx context.Context, req *ListWalletTransfersRequest) (*Page[WalletTransferData], error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	page, err := req.PageRequest.toBizPage()
	if err != nil {
		return nil, err
	}
	items, total, err := s.uc.ListWalletTransfers(ctx, &biz.ListUIWalletTransfersRequest{
		UIPageRequest: page,
		ID:            req.ID,
		AccountID:     req.AccountID,
		CardID:        req.CardID,
		Kind:          req.Kind,
		UITimeRange: biz.UITimeRange{
			CreatedFrom: req.CreatedFrom,
			CreatedTo:   req.CreatedTo,
		},
	})
	if err != nil {
		return nil, err
	}
	result := &Page[WalletTransferData]{
		Items: make([]WalletTransferData, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		result.Items = append(result.Items, toWalletTransferData(item))
	}
	return result, nil
}

func toWalletTransferData(item *model.WalletTransfer) WalletTransferData {
	return WalletTransferData{
		ID:             item.ID,
		AccountID:      item.AccountID,
		AccountName:    item.Account.GetName(),
		CardID:         item.CardID,
		SourceWalletID: item.SourceWalletID,
		TargetWalletID: item.TargetWalletID,
		Kind:           item.Kind,
		Currency:       item.Currency,
		Amount:         item.Amount,
		RequestID:      item.RequestID,
		CreatedAt:      item.CreatedAt,
	}
}

type FundCardRequest struct {
	CardID    model.ID                 `json:"card_id" binding:"required,gt=0"`
	AccountID model.ID                 `json:"account_id" binding:"required,gt=0"`
	Kind      enums.WalletTransferKind `json:"kind" binding:"required,oneof=card_top_up card_withdraw"`
	Amount    decimal.Decimal          `json:"amount" binding:"required"`
	RequestID *string                  `json:"request_id" binding:"omitempty,min=1"`
}

func (s *Service) FundCard(ctx context.Context, req *FundCardRequest) (*WalletTransferData, error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	item, err := s.uc.FundCard(ctx, &biz.UIFundCardRequest{
		CardID:    req.CardID,
		AccountID: req.AccountID,
		Kind:      req.Kind,
		Amount:    req.Amount,
		RequestID: req.RequestID,
	})
	if err != nil {
		return nil, err
	}
	result := toWalletTransferData(item)
	return &result, nil
}

type FundVirtualAccountRequest struct {
	VirtualAccountID       model.ID        `json:"virtual_account_id" binding:"required,gt=0"`
	AccountID              model.ID        `json:"account_id" binding:"required,gt=0"`
	TargetVirtualAccountID *model.ID       `json:"target_virtual_account_id" binding:"omitempty,gt=0"`
	Withdraw               *bool           `json:"withdraw"`
	Amount                 decimal.Decimal `json:"amount" binding:"required"`
	RequestID              *string         `json:"request_id" binding:"omitempty,min=1"`
}

func (s *Service) FundVirtualAccount(ctx context.Context, req *FundVirtualAccountRequest) (*WalletTransferData, error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	item, err := s.uc.FundVirtualAccount(ctx, &biz.UIFundVirtualAccountRequest{
		VirtualAccountID:       req.VirtualAccountID,
		AccountID:              req.AccountID,
		TargetVirtualAccountID: req.TargetVirtualAccountID,
		Withdraw:               req.Withdraw,
		Amount:                 req.Amount,
		RequestID:              req.RequestID,
	})
	if err != nil {
		return nil, err
	}
	result := toWalletTransferData(item)
	return &result, nil
}
