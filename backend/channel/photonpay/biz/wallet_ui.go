package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type UIFundCardRequest struct {
	CardID model.ID
	Amount decimal.Decimal
}

type MoveFundsRequest struct {
	AccountID model.ID
	SourceID  model.ID
	TargetID  model.ID
	Amount    decimal.Decimal
}

func DefaultBalance() decimal.Decimal {
	return decimal.NewFromInt(1_000_000)
}

func (u *PhotonPayUIUsecase) FundCard(ctx context.Context, req *UIFundCardRequest) (*model.Card, error) {
	if !req.Amount.IsPositive() {
		return nil, ErrInvalidOperation
	}

	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		var err error
		card, err = u.getCard(txCtx, req.CardID)
		if err != nil {
			return err
		}
		if card.WalletID == 0 {
			return ErrResourceNotFound
		}

		account, err := u.accountRepo.Find(txCtx, card.AccountID)
		if err != nil {
			zap.S().Errorw("find photonpay UI card account", "error", err)

			return ErrDatabaseOperation
		}
		accountID := card.AccountID
		source, err := u.walletRepo.FindByIDForUpdate(txCtx, &WalletFindByIDForUpdateRequest{
			AccountID: &accountID,
			ID:        account.WalletID,
		})
		if err != nil {
			zap.S().Errorw("lock photonpay UI account wallet", "error", err)

			return ErrDatabaseOperation
		}
		target, err := u.walletRepo.FindByIDForUpdate(txCtx, &WalletFindByIDForUpdateRequest{
			AccountID: &accountID,
			ID:        card.WalletID,
		})
		if err != nil {
			zap.S().Errorw("lock photonpay UI card wallet", "error", err)

			return ErrDatabaseOperation
		}
		if source.Amount.LessThan(req.Amount) {
			return ErrInvalidOperation
		}
		source.Amount = source.Amount.Sub(req.Amount)
		source.Out = source.Out.Add(req.Amount)
		target.Amount = target.Amount.Add(req.Amount)
		target.In = target.In.Add(req.Amount)
		if err := u.walletRepo.Save(txCtx, source); err != nil {
			zap.S().Errorw("save photonpay UI account wallet", "error", err)

			return ErrDatabaseOperation
		}
		if err := u.walletRepo.Save(txCtx, target); err != nil {
			zap.S().Errorw("save photonpay UI card wallet", "error", err)

			return ErrDatabaseOperation
		}
		card.Wallet = target

		return nil
	})
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (u *PhotonPayUIUsecase) ListFunds(ctx context.Context, accountID *model.ID) ([]*model.Wallet, error) {
	items, err := u.walletRepo.ListWallets(ctx, &WalletListRequest{
		AccountIDs: types.PointerSlice(accountID),
	})
	if err != nil {
		zap.S().Errorw("list photonpay wallets", "error", err)
		return nil, ErrDatabaseOperation
	}
	return items, nil
}

func (u *PhotonPayUIUsecase) MoveFunds(ctx context.Context, req *MoveFundsRequest) error {
	if req.AccountID == 0 || !req.Amount.IsPositive() || req.SourceID == req.TargetID {
		return ErrInvalidOperation
	}
	return u.transaction.InTx(ctx, func(ctx context.Context) error {
		exists, err := u.accountRepo.Exist(ctx, req.AccountID)
		if err != nil {
			zap.S().Errorw("check photonpay funding account", "error", err)
			return ErrDatabaseOperation
		}
		if !exists {
			return ErrResourceNotFound
		}
		account, err := u.accountRepo.Find(ctx, req.AccountID)
		if err != nil {
			zap.S().Errorw("find photonpay funding account", "error", err)
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
			exists, err := u.walletRepo.WalletExists(ctx, &ExistWalletRequest{
				AccountID: req.AccountID,
				ID:        walletID,
			})
			if err != nil {
				zap.S().Errorw("check photonpay transfer wallet", "error", err)
				return ErrDatabaseOperation
			}
			if !exists {
				return ErrResourceNotFound
			}
			wallet, err := u.walletRepo.LockWallet(ctx, &LockWalletRequest{
				AccountID: req.AccountID,
				ID:        walletID,
			})
			if err != nil {
				zap.S().Errorw("lock photonpay transfer wallet", "error", err)
				return ErrDatabaseOperation
			}
			wallets[walletID] = wallet
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
			if err := u.walletRepo.SaveWallet(ctx, source); err != nil {
				zap.S().Errorw("debit photonpay wallet", "error", err)
				return ErrDatabaseOperation
			}
		}
		if target != nil {
			target.Amount = target.Amount.Add(req.Amount)
			target.In = target.In.Add(req.Amount)
			if err := u.walletRepo.SaveWallet(ctx, target); err != nil {
				zap.S().Errorw("credit photonpay wallet", "error", err)
				return ErrDatabaseOperation
			}
		}
		return nil
	})
}
