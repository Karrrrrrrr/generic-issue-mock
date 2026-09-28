package biz

import (
	"context"
	"strings"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"
	sharederrors "generic-mock/shared/errors"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type ListUIAccountsRequest struct {
	UIPageRequest
	ID *model.ID
}

func (req *ListUIAccountsRequest) Validate() error {
	if req == nil || !validUIIDs([]*model.ID{req.ID}) {
		return sharederrors.ErrInvalidUIRequest
	}
	return req.UIPageRequest.Validate()
}

type UIAccounts interface {
	CreateAccount(context.Context, *CreateUIAccountRequest) (*model.Account, error)
	RenameAccount(context.Context, *RenameUIAccountRequest) (*model.Account, error)
	AdjustAccount(context.Context, *AdjustUIAccountRequest) error

	ListAccounts(context.Context, *ListUIAccountsRequest) ([]*model.Account, int64, error)
}

func (uc *ui) ListAccounts(ctx context.Context, req *ListUIAccountsRequest) ([]*model.Account, int64, error) {
	if err := req.Validate(); err != nil {
		return nil, 0, err
	}
	filters := AccountFilters{
		Channel: uc.channel,
		IDs:     types.PointerSlice(req.ID),
	}
	items, err := uc.accountRepo.List(ctx, &AccountListRequest{
		AccountFilters: filters,
		Offset:         req.Offset,
		Limit:          req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list shared UI account", "channel", uc.channel, "error", err)
		return nil, 0, sharederrors.ErrDatabaseOperation
	}
	total, err := uc.accountRepo.Count(ctx, &AccountCountRequest{AccountFilters: filters})
	if err != nil {
		zap.S().Errorw("count shared UI account", "channel", uc.channel, "error", err)
		return nil, 0, sharederrors.ErrDatabaseOperation
	}
	return items, total, nil
}

type CreateUIAccountRequest struct {
	Name     string
	Currency enums.Currency
}

func (req *CreateUIAccountRequest) Validate() error {
	if req == nil || strings.TrimSpace(req.Name) == "" || !validUICurrency(req.Currency) {
		return sharederrors.ErrInvalidUIRequest
	}
	return nil
}

type RenameUIAccountRequest struct {
	ID   model.ID
	Name string
}

func (req *RenameUIAccountRequest) Validate() error {
	if req == nil || req.ID <= 0 || strings.TrimSpace(req.Name) == "" {
		return sharederrors.ErrInvalidUIRequest
	}
	return nil
}

type AdjustUIAccountRequest struct {
	AccountID model.ID
	Currency  enums.Currency
	Amount    decimal.Decimal
}

func (req *AdjustUIAccountRequest) Validate() error {
	if req == nil || req.AccountID <= 0 || !validUICurrency(req.Currency) || req.Amount.IsZero() {
		return sharederrors.ErrInvalidUIRequest
	}
	return nil
}

func (uc *ui) CreateAccount(ctx context.Context, req *CreateUIAccountRequest) (*model.Account, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	account := &model.Account{
		Channel: uc.channel,
		Name:    strings.TrimSpace(req.Name),
	}
	err := uc.tx.InTx(ctx, func(ctx context.Context) error {
		if err := uc.accountRepo.Create(ctx, &AccountCreateRequest{Account: account}); err != nil {
			zap.S().Errorw("create shared UI account", "error", err)
			return sharederrors.ErrDatabaseOperation
		}
		wallet := &model.Wallet{
			AccountID: account.ID,
			Channel:   uc.channel,
			Currency:  req.Currency,
			Type:      enums.WalletType_Account,
		}
		if err := uc.walletRepo.Create(ctx, &WalletCreateRequest{Wallet: wallet}); err != nil {
			zap.S().Errorw("create shared UI account wallet", "error", err)
			return sharederrors.ErrDatabaseOperation
		}
		if err := uc.accountRepo.SetWallet(ctx, &AccountSetWalletRequest{
			ID:       account.ID,
			Channel:  uc.channel,
			WalletID: wallet.ID,
		}); err != nil {
			zap.S().Errorw("assign shared UI account wallet", "error", err)
			return sharederrors.ErrDatabaseOperation
		}
		account.WalletID = wallet.ID
		account.Wallet = wallet
		if uc.createAccountVirtualAccount {
			_, err := uc.createVirtualAccount(ctx, &CreateUIVirtualAccountRequest{
				AccountID: account.ID,
				Name:      account.Name,
				Currency:  req.Currency,
			})
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (uc *ui) RenameAccount(ctx context.Context, req *RenameUIAccountRequest) (*model.Account, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var result *model.Account
	err := uc.tx.InTx(ctx, func(ctx context.Context) error {
		if _, err := uc.lockUIAccount(ctx, req.ID); err != nil {
			return err
		}
		if err := uc.accountRepo.Rename(ctx, &AccountRenameRequest{
			ID:      req.ID,
			Channel: uc.channel,
			Name:    strings.TrimSpace(req.Name),
		}); err != nil {
			zap.S().Errorw("rename shared UI account", "error", err)
			return sharederrors.ErrDatabaseOperation
		}
		var err error
		result, err = uc.accountRepo.Find(ctx, &AccountFindRequest{
			ID:      req.ID,
			Channel: uc.channel,
		})
		if err != nil {
			zap.S().Errorw("load renamed shared UI account", "error", err)
			return sharederrors.ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (uc *ui) AdjustAccount(ctx context.Context, req *AdjustUIAccountRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}
	return uc.tx.InTx(ctx, func(ctx context.Context) error {
		account, err := uc.lockUIAccount(ctx, req.AccountID)
		if err != nil {
			return err
		}
		return uc.balanceChanger.ChangeBalanceSimple(ctx, &ChangeBalanceSimpleReq{
			AccountID:      account.ID,
			Channel:        uc.channel,
			WalletID:       account.WalletID,
			Currency:       req.Currency,
			Amount:         req.Amount,
			CheckAvailable: true,
		})
	})
}

func (uc *ui) lockUIAccount(ctx context.Context, accountID model.ID) (*model.Account, error) {
	exists, err := uc.accountRepo.Exist(ctx, &AccountExistRequest{
		ID:      accountID,
		Channel: uc.channel,
	})
	if err != nil {
		zap.S().Errorw("check shared UI account", "error", err)
		return nil, sharederrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, sharederrors.ErrAccountNotFound
	}
	account, err := uc.accountRepo.FindByIDWithLock(ctx, &AccountFindByIDWithLockRequest{
		ID:      accountID,
		Channel: uc.channel,
	})
	if err != nil {
		zap.S().Errorw("lock shared UI account", "error", err)
		return nil, sharederrors.ErrDatabaseOperation
	}
	return account, nil
}
