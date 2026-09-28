package service

import (
	"context"
	"time"

	"generic-mock/channel/pingpong/biz"
	pingerrors "generic-mock/channel/pingpong/errors"
	common "generic-mock/enums"
	"generic-mock/model"
)

type UIVirtualAccountIDData struct {
	VirtualAccountID model.ID `json:"virtual_account_id"`
}

type UIVirtualAccountData struct {
	WalletID    model.ID        `json:"wallet_id"`
	ID          model.ID        `json:"id"`
	AccountID   model.ID        `json:"account_id"`
	AccountName string          `json:"account_name"`
	Name        string          `json:"name"`
	Balance     Number          `json:"balance"`
	Currency    common.Currency `json:"currency"`
	CreatedAt   time.Time       `json:"created_at"`
}

type UICreateVirtualAccountRequest struct {
	AccountID model.ID `json:"account_id" binding:"required,gt=0"`
	Name      string   `json:"name" binding:"required"`
}

type UIFundVirtualAccountRequest struct {
	UIResourceRequest
	Amount    Number `json:"amount"`
	RequestID string `json:"request_id" binding:"required,max=36"`
}

func (s *PingPongUIService) ListVirtualAccounts(ctx context.Context, req *UIListRequest) (*UIPage[UIVirtualAccountData], error) {
	accountID := req.AccountID
	if accountID != nil && *accountID <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	page, limit, err := req.PageRequest.resolvePagination()
	if err != nil {
		return nil, err
	}
	items, total, err := s.uc.ListVirtualAccounts(ctx, &biz.UIListVirtualAccountsRequest{
		AccountID: accountID,
		Offset:    (page - 1) * limit,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}
	result := &UIPage[UIVirtualAccountData]{
		Items: make([]UIVirtualAccountData, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		if item.Wallet == nil {
			return nil, pingerrors.ErrInvalid
		}
		result.Items = append(result.Items, UIVirtualAccountData{
			ID:          item.ID,
			WalletID:    item.WalletID,
			AccountID:   item.AccountID,
			AccountName: item.Account.GetName(),
			Name:        item.Name,
			Balance:     Number{item.Wallet.Available},
			Currency:    item.Wallet.Currency,
			CreatedAt:   item.CreatedAt,
		})
	}
	return result, nil
}

func (s *PingPongUIService) CreateVirtualAccount(ctx context.Context, req *UICreateVirtualAccountRequest) (*UIVirtualAccountIDData, error) {
	accountID := req.AccountID
	if accountID <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	item, err := s.uc.CreateVirtualAccount(ctx, &biz.UICreateVirtualAccountRequest{
		AccountID: accountID,
		Name:      req.Name,
	})
	if err != nil {
		return nil, err
	}
	return &UIVirtualAccountIDData{VirtualAccountID: item.ID}, nil
}

func (s *PingPongUIService) FundVirtualAccount(ctx context.Context, req *UIFundVirtualAccountRequest) (*UIRecordData, error) {
	accountID, id := req.AccountID, req.ID
	if err := req.UIResourceRequest.Validate(); err != nil {
		return nil, err
	}
	item, err := s.uc.FundVirtualAccount(ctx, &biz.UIVirtualAccountFundingRequest{
		AccountID:        accountID,
		VirtualAccountID: id,
		Currency:         common.Currency_USD,
		Amount:           req.Amount.Decimal,
		RequestID:        &req.RequestID,
	})
	if err != nil {
		return nil, err
	}
	return &UIRecordData{RecordID: item.ID}, nil
}
