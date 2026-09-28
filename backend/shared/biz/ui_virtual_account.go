package biz

import (
	"context"
	"strings"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"
	sharederrors "generic-mock/shared/errors"

	"go.uber.org/zap"
)

type ListUIVirtualAccountsRequest struct {
	UIPageRequest
	ID        *model.ID
	AccountID *model.ID
}

func (req *ListUIVirtualAccountsRequest) Validate() error {
	if req == nil || !validUIIDs([]*model.ID{req.ID, req.AccountID}) {
		return sharederrors.ErrInvalidUIRequest
	}
	return req.UIPageRequest.Validate()
}

type UIVirtualAccounts interface {
	CreateVirtualAccount(context.Context, *CreateUIVirtualAccountRequest) (*model.VirtualAccount, error)

	ListVirtualAccounts(context.Context, *ListUIVirtualAccountsRequest) ([]*model.VirtualAccount, int64, error)
}

func (uc *ui) ListVirtualAccounts(ctx context.Context, req *ListUIVirtualAccountsRequest) ([]*model.VirtualAccount, int64, error) {
	if err := req.Validate(); err != nil {
		return nil, 0, err
	}
	filters := VirtualAccountFilters{
		Channel:    uc.channel,
		IDs:        types.PointerSlice(req.ID),
		AccountIDs: types.PointerSlice(req.AccountID),
	}
	items, err := uc.virtualAccountRepo.List(ctx, &VirtualAccountListRequest{
		VirtualAccountFilters: filters,
		Offset:                req.Offset,
		Limit:                 req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list shared UI virtual_account", "channel", uc.channel, "error", err)
		return nil, 0, sharederrors.ErrDatabaseOperation
	}
	total, err := uc.virtualAccountRepo.Count(ctx, &VirtualAccountCountRequest{VirtualAccountFilters: filters})
	if err != nil {
		zap.S().Errorw("count shared UI virtual_account", "channel", uc.channel, "error", err)
		return nil, 0, sharederrors.ErrDatabaseOperation
	}
	return items, total, nil
}

type CreateUIVirtualAccountRequest struct {
	AccountID model.ID
	Name      string
	Currency  enums.Currency
}

func (req *CreateUIVirtualAccountRequest) Validate() error {
	if req == nil || req.AccountID <= 0 || strings.TrimSpace(req.Name) == "" || !validUICurrency(req.Currency) {
		return sharederrors.ErrInvalidUIRequest
	}
	return nil
}

func (uc *ui) CreateVirtualAccount(ctx context.Context, req *CreateUIVirtualAccountRequest) (*model.VirtualAccount, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var result *model.VirtualAccount
	err := uc.tx.InTx(ctx, func(ctx context.Context) error {
		account, err := uc.lockUIAccount(ctx, req.AccountID)
		if err != nil {
			return err
		}
		result, err = uc.createVirtualAccount(ctx, req)
		if err != nil {
			return err
		}
		result.Account = account
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (uc *ui) createVirtualAccount(ctx context.Context, req *CreateUIVirtualAccountRequest) (*model.VirtualAccount, error) {
	wallet := &model.Wallet{
		AccountID: req.AccountID,
		Channel:   uc.channel,
		Type:      enums.WalletType_VirtualAccount,
		Currency:  req.Currency,
	}
	if err := uc.walletRepo.Create(ctx, &WalletCreateRequest{Wallet: wallet}); err != nil {
		zap.S().Errorw("create shared UI virtual account wallet", "error", err)
		return nil, sharederrors.ErrDatabaseOperation
	}
	result := &model.VirtualAccount{
		AccountID: req.AccountID,
		Channel:   uc.channel,
		WalletID:  wallet.ID,
		Wallet:    wallet,
		Name:      strings.TrimSpace(req.Name),
	}
	if err := uc.virtualAccountRepo.Create(ctx, &VirtualAccountCreateRequest{VirtualAccount: result}); err != nil {
		zap.S().Errorw("create shared UI virtual account", "error", err)
		return nil, sharederrors.ErrDatabaseOperation
	}
	return result, nil
}
