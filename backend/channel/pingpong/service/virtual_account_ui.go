package service

import (
	"context"
	"time"

	"generic-mock/channel/pingpong/biz"
	pingerrors "generic-mock/channel/pingpong/errors"
	"generic-mock/channel/pingpong/pkg/idconv"
	common "generic-mock/enums"
)

type UIVirtualAccountIDData struct {
	VirtualAccountID string `json:"virtual_account_id"`
}

type UIVirtualAccountData struct {
	ID          string          `json:"id"`
	AccountID   string          `json:"account_id"`
	AccountName string          `json:"account_name"`
	Name        string          `json:"name"`
	Balance     Number          `json:"balance"`
	Currency    common.Currency `json:"currency"`
	CreatedAt   time.Time       `json:"created_at"`
}

type UICreateVirtualAccountRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type UIFundVirtualAccountRequest struct {
	UIResourceRequest
	Amount    Number `json:"amount"`
	RequestID string `json:"request_id" binding:"required,max=36"`
}

func (s *PingPongUIService) ListVirtualAccounts(ctx context.Context, req *UIListRequest) (*UIPage[UIVirtualAccountData], error) {
	accountID, err := idconv.FromOptionalString(req.AccountID)
	if err != nil {
		return nil, err
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
			ID:          idconv.ToString(item.ID),
			AccountID:   idconv.ToString(item.AccountID),
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
	accountID, err := idconv.FromString(req.AccountID)
	if err != nil {
		return nil, err
	}
	item, err := s.uc.CreateVirtualAccount(ctx, &biz.UICreateVirtualAccountRequest{
		AccountID: accountID,
		Name:      req.Name,
	})
	if err != nil {
		return nil, err
	}
	return &UIVirtualAccountIDData{VirtualAccountID: idconv.ToString(item.ID)}, nil
}

func (s *PingPongUIService) FundVirtualAccount(ctx context.Context, req *UIFundVirtualAccountRequest) (*RecordData, error) {
	accountID, id, err := req.UIResourceRequest.ParseAccountAndResourceIDs()
	if err != nil {
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
	return &RecordData{RecordID: idconv.ToString(item.ID)}, nil
}
