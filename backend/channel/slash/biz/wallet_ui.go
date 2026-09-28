package biz

import (
	"context"

	slasherrors "generic-mock/channel/slash/errors"
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

func (u *SlashUIUsecase) ListFunds(ctx context.Context, accountID *model.ID) ([]*model.Wallet, error) {
	items, err := u.walletRepository.ListWallets(ctx, &WalletListRequest{
		AccountIDs: types.PointerSlice(accountID),
	})
	if err != nil {
		zap.S().Errorw("list slash wallets", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	return items, nil
}

func (u *SlashUIUsecase) MoveFunds(ctx context.Context, req *MoveFundsRequest) error {
	if req.AccountID == 0 || !req.Amount.IsPositive() || req.SourceID == req.TargetID {
		return slasherrors.ErrInvalidOperation
	}
	return u.transaction.InTx(ctx, func(ctx context.Context) error {
		exists, err := u.accountRepository.Exist(ctx, req.AccountID)
		if err != nil {
			zap.S().Errorw("check slash funding account", "error", err)
			return slasherrors.ErrDatabaseOperation
		}
		if !exists {
			return slasherrors.ErrResourceNotFound
		}
		account, err := u.accountRepository.Find(ctx, req.AccountID)
		if err != nil {
			zap.S().Errorw("find slash funding account", "error", err)
			return slasherrors.ErrDatabaseOperation
		}
		if account.WalletID == 0 || (req.SourceID != account.WalletID && req.TargetID != account.WalletID) {
			return slasherrors.ErrInvalidOperation
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
				zap.S().Errorw("check slash transfer wallet", "error", err)
				return slasherrors.ErrDatabaseOperation
			}
			if !exists {
				return slasherrors.ErrResourceNotFound
			}
			wallet, err := u.walletRepository.LockWallet(ctx, &LockWalletRequest{
				AccountID: req.AccountID,
				ID:        walletID,
			})
			if err != nil {
				zap.S().Errorw("lock slash transfer wallet", "error", err)
				return slasherrors.ErrDatabaseOperation
			}
			wallets[walletID] = wallet
		}
		source, target := wallets[req.SourceID], wallets[req.TargetID]
		if source == nil && target.Type != enums.WalletType_Account || target == nil && source.Type != enums.WalletType_Account {
			return slasherrors.ErrInvalidOperation
		}
		if source != nil && target != nil && source.Currency != target.Currency {
			return slasherrors.ErrInvalidOperation
		}
		if source != nil {
			if source.Available.LessThan(req.Amount) {
				return slasherrors.ErrInvalidOperation
			}
			source.Available = source.Available.Sub(req.Amount)
			source.Out = source.Out.Add(req.Amount)
			if err := u.walletRepository.SaveWallet(ctx, source); err != nil {
				zap.S().Errorw("debit slash wallet", "error", err)
				return slasherrors.ErrDatabaseOperation
			}
		}
		if target != nil {
			target.Available = target.Available.Add(req.Amount)
			target.In = target.In.Add(req.Amount)
			if err := u.walletRepository.SaveWallet(ctx, target); err != nil {
				zap.S().Errorw("credit slash wallet", "error", err)
				return slasherrors.ErrDatabaseOperation
			}
		}
		return nil
	})
}
