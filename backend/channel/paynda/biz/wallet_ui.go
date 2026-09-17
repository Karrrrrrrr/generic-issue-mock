package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type MoveFundsRequest struct {
	AccountID model.ID
	SourceID  model.ID
	TargetID  model.ID
	Amount    decimal.Decimal
}

func (u *PayndaUIUsecase) ListFunds(ctx context.Context, accountID *model.ID) ([]*model.Wallet, error) {
	items, err := u.walletRepository.ListWallets(ctx, &WalletListRequest{
		AccountIDs: types.PointerSlice(accountID),
	})
	if err != nil {
		zap.S().Errorw("list paynda wallets", "error", err)
		return nil, ErrDatabaseOperation
	}
	return items, nil
}

func (u *PayndaUIUsecase) MoveFunds(ctx context.Context, req *MoveFundsRequest) error {
	if req.AccountID == 0 || !req.Amount.IsPositive() || req.SourceID == req.TargetID {
		return ErrInvalidOperation
	}
	return u.transaction.InTx(ctx, func(ctx context.Context) error {
		exists, err := u.accountRepository.ExistByID(ctx, req.AccountID)
		if err != nil {
			zap.S().Errorw("check paynda funding account", "error", err)
			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}
		account, err := u.accountRepository.FindByID(ctx, req.AccountID)
		if err != nil {
			zap.S().Errorw("find paynda funding account", "error", err)
			return ErrDatabaseOperation
		}
		if account.WalletID == 0 || (req.SourceID != account.WalletID && req.TargetID != account.WalletID) {
			return ErrInvalidOperation
		}
		wallets := make(map[model.ID]*model.Wallet)
		ids := []model.ID{
			req.SourceID,
			req.TargetID,
		}
		if ids[0] > ids[1] {
			ids[0], ids[1] = ids[1], ids[0]
		}
		for _, walletID := range ids {
			if walletID == 0 {
				continue
			}
			exists, err := u.walletRepository.WalletExists(ctx, &ExistWalletRequest{
				AccountID: req.AccountID,
				ID:        walletID,
			})
			if err != nil {
				zap.S().Errorw("check paynda transfer wallet", "error", err)
				return ErrDatabaseOperation
			}
			if !exists {
				return ErrResourceNotFound
			}
			wallet, err := u.walletRepository.LockWallet(ctx, &LockWalletRequest{
				AccountID: req.AccountID,
				ID:        walletID,
			})
			if err != nil {
				zap.S().Errorw("lock paynda transfer wallet", "error", err)
				return ErrDatabaseOperation
			}
			wallets[walletID] = wallet
			if wallet.Type != enums.WalletType_Account && wallet.Type != enums.WalletType_Card {
				return ErrInvalidOperation
			}
		}
		source, target := wallets[req.SourceID], wallets[req.TargetID]
		if source == nil && target.Type != enums.WalletType_Account || target == nil && source.Type != enums.WalletType_Account {
			return ErrInvalidOperation
		}
		if source != nil && target != nil && source.Currency != target.Currency {
			return ErrInvalidOperation
		}
		if source != nil {
			if source.Amount.LessThan(req.Amount) {
				return ErrInvalidOperation
			}
			source.Amount = source.Amount.Sub(req.Amount)
			source.Out = source.Out.Add(req.Amount)
			if err := u.walletRepository.SaveWallet(ctx, source); err != nil {
				zap.S().Errorw("debit paynda wallet", "error", err)
				return ErrDatabaseOperation
			}
		}
		if target != nil {
			target.Amount = target.Amount.Add(req.Amount)
			target.In = target.In.Add(req.Amount)
			if err := u.walletRepository.SaveWallet(ctx, target); err != nil {
				zap.S().Errorw("credit paynda wallet", "error", err)
				return ErrDatabaseOperation
			}
		}
		return nil
	})
}
