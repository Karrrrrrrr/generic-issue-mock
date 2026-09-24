package biz

import (
	"context"
	"strings"

	pingerrors "generic-mock/channel/pingpong/errors"
	common "generic-mock/enums"
	"generic-mock/model"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type CreateAccountRequest struct{ Name string }
type AdjustAccountRequest struct {
	AccountID model.ID
	Amount    decimal.Decimal
}

type ListAccountsRequest struct {
	Offset int
	Limit  int
}

func (uc *PingPongUIUsecase) CreateAccount(ctx context.Context, req *CreateAccountRequest) (*model.Account, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, pingerrors.ErrInvalid
	}
	account := &model.Account{
		Channel: common.Channel_PingPong,
		Name:    strings.TrimSpace(req.Name),
	}
	err := uc.tx.InTx(ctx, func(ctx context.Context) error {
		if err := uc.accountRepo.Create(ctx, account); err != nil {
			zap.S().Errorw("create pingpong account", "error", err)
			return pingerrors.ErrDatabase
		}
		wallet := &model.Wallet{
			AccountID: account.ID,
			Channel:   common.Channel_PingPong,
			Type:      common.WalletType_Account,
			Currency:  common.Currency_USD,
		}
		if err := uc.walletRepo.Create(ctx, wallet); err != nil {
			zap.S().Errorw("create pingpong account wallet", "error", err)
			return pingerrors.ErrDatabase
		}
		account.WalletID = wallet.ID
		account.Wallet = wallet
		if err := uc.accountRepo.Save(ctx, account); err != nil {
			zap.S().Errorw("assign pingpong account wallet", "error", err)
			return pingerrors.ErrDatabase
		}
		return nil
	})
	return account, err
}

func (uc *PingPongUIUsecase) ListAccounts(ctx context.Context, req *ListAccountsRequest) ([]*model.Account, int64, error) {
	items, err := uc.accountRepo.List(ctx, &AccountListRequest{
		Offset: req.Offset,
		Limit:  &req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list pingpong accounts", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	total, err := uc.accountRepo.Count(ctx, &AccountCountRequest{})
	if err != nil {
		zap.S().Errorw("count pingpong accounts", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	return items, total, nil
}

func (uc *PingPongUIUsecase) AdjustAccount(ctx context.Context, req *AdjustAccountRequest) error {
	if req.Amount.IsZero() {
		return pingerrors.ErrInvalid
	}
	return uc.tx.InTx(ctx, func(ctx context.Context) error {
		account, err := uc.lockAccount(ctx, req.AccountID)
		if err != nil {
			return err
		}
		wallet, err := uc.walletRepo.Lock(ctx, &WalletLockRequest{
			AccountID: account.ID,
			ID:        account.WalletID,
		})
		if err != nil {
			zap.S().Errorw("lock pingpong account balance", "error", err)
			return pingerrors.ErrDatabase
		}
		if wallet.Amount.Add(req.Amount).LessThan(wallet.PendingOut) {
			return pingerrors.ErrInsufficient
		}
		wallet.Amount = wallet.Amount.Add(req.Amount)
		if req.Amount.IsPositive() {
			wallet.In = wallet.In.Add(req.Amount)
		} else {
			wallet.Out = wallet.Out.Sub(req.Amount)
		}
		if err := uc.walletRepo.Save(ctx, wallet); err != nil {
			zap.S().Errorw("adjust pingpong account balance", "error", err)
			return pingerrors.ErrDatabase
		}
		return nil
	})
}

func (uc *PingPongUIUsecase) lockAccount(ctx context.Context, accountID model.ID) (*model.Account, error) {
	if accountID <= 0 {
		return nil, pingerrors.ErrInvalid
	}
	exists, err := uc.accountRepo.Exists(ctx, &AccountExistsRequest{ID: accountID})
	if err != nil {
		zap.S().Errorw("check pingpong account", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	if !exists {
		return nil, pingerrors.ErrNotFound
	}
	account, err := uc.accountRepo.Lock(ctx, &AccountLockRequest{ID: accountID})
	if err != nil {
		zap.S().Errorw("lock pingpong account", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	return account, nil
}
