package biz

import (
	"context"

	photonpayerrors "generic-mock/channel/photonpay/errors"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/cardwallet"
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
		return nil, photonpayerrors.ErrInvalidOperation
	}

	var card *model.Card
	err := u.transaction.InTx(ctx, func(txCtx context.Context) error {
		var err error
		card, err = u.getCard(txCtx, req.CardID)
		if err != nil {
			return err
		}
		if card.WalletID == 0 {
			return photonpayerrors.ErrResourceNotFound
		}

		account, err := u.accountRepo.Find(txCtx, card.AccountID)
		if err != nil {
			zap.S().Errorw("find photonpay UI card account", "error", err)

			return photonpayerrors.ErrDatabaseOperation
		}
		card.Account = account
		fundingWalletID, ok := cardwallet.FundingWalletID(card)
		if !ok {
			return photonpayerrors.ErrInvalidOperation
		}
		walletIDs := []model.ID{fundingWalletID, card.WalletID}
		if walletIDs[0] > walletIDs[1] {
			walletIDs[0], walletIDs[1] = walletIDs[1], walletIDs[0]
		}
		wallets := make(map[model.ID]*model.Wallet)
		for _, walletID := range walletIDs {
			wallet, err := u.walletRepo.LockWallet(txCtx, &LockWalletRequest{
				AccountID: card.AccountID,
				ID:        walletID,
			})
			if err != nil {
				zap.S().Errorw("lock photonpay UI funding wallet", "error", err)
				return photonpayerrors.ErrDatabaseOperation
			}
			wallets[walletID] = wallet
		}
		source := wallets[fundingWalletID]
		target := wallets[card.WalletID]
		if source.Currency != target.Currency || target.Currency != card.CardCurrency {
			return photonpayerrors.ErrInvalidOperation
		}
		if card.CardType == enums.CardType_VirtualAccountSingle &&
			(source.Type != enums.WalletType_VirtualAccount || target.Type != enums.WalletType_Card) {
			return photonpayerrors.ErrInvalidOperation
		}
		if source.Available.LessThan(req.Amount) {
			return photonpayerrors.ErrInvalidOperation
		}
		source.Available = source.Available.Sub(req.Amount)
		source.Out = source.Out.Add(req.Amount)
		target.Available = target.Available.Add(req.Amount)
		target.In = target.In.Add(req.Amount)
		if err := u.walletRepo.Save(txCtx, source); err != nil {
			zap.S().Errorw("save photonpay UI funding source wallet", "error", err)

			return photonpayerrors.ErrDatabaseOperation
		}
		if err := u.walletRepo.Save(txCtx, target); err != nil {
			zap.S().Errorw("save photonpay UI card wallet", "error", err)

			return photonpayerrors.ErrDatabaseOperation
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
		return nil, photonpayerrors.ErrDatabaseOperation
	}
	return items, nil
}

func (u *PhotonPayUIUsecase) MoveFunds(ctx context.Context, req *MoveFundsRequest) error {
	if req.AccountID == 0 || !req.Amount.IsPositive() || req.SourceID == req.TargetID {
		return photonpayerrors.ErrInvalidOperation
	}
	return u.transaction.InTx(ctx, func(ctx context.Context) error {
		exists, err := u.accountRepo.Exist(ctx, req.AccountID)
		if err != nil {
			zap.S().Errorw("check photonpay funding account", "error", err)
			return photonpayerrors.ErrDatabaseOperation
		}
		if !exists {
			return photonpayerrors.ErrResourceNotFound
		}
		account, err := u.accountRepo.Find(ctx, req.AccountID)
		if err != nil {
			zap.S().Errorw("find photonpay funding account", "error", err)
			return photonpayerrors.ErrDatabaseOperation
		}
		if account.WalletID == 0 || (req.SourceID != account.WalletID && req.TargetID != account.WalletID) {
			return photonpayerrors.ErrInvalidOperation
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
				return photonpayerrors.ErrDatabaseOperation
			}
			if !exists {
				return photonpayerrors.ErrResourceNotFound
			}
			wallet, err := u.walletRepo.LockWallet(ctx, &LockWalletRequest{
				AccountID: req.AccountID,
				ID:        walletID,
			})
			if err != nil {
				zap.S().Errorw("lock photonpay transfer wallet", "error", err)
				return photonpayerrors.ErrDatabaseOperation
			}
			wallets[walletID] = wallet
		}
		source, target := wallets[req.SourceID], wallets[req.TargetID]
		if source == nil && target.Type != enums.WalletType_Account || target == nil && source.Type != enums.WalletType_Account {
			return photonpayerrors.ErrInvalidOperation
		}
		if source != nil && target != nil && source.Currency != target.Currency {
			return photonpayerrors.ErrInvalidOperation
		}
		if source != nil {
			if source.Available.LessThan(req.Amount) {
				return photonpayerrors.ErrInvalidOperation
			}
			source.Available = source.Available.Sub(req.Amount)
			source.Out = source.Out.Add(req.Amount)
			if err := u.walletRepo.SaveWallet(ctx, source); err != nil {
				zap.S().Errorw("debit photonpay wallet", "error", err)
				return photonpayerrors.ErrDatabaseOperation
			}
		}
		if target != nil {
			target.Available = target.Available.Add(req.Amount)
			target.In = target.In.Add(req.Amount)
			if err := u.walletRepo.SaveWallet(ctx, target); err != nil {
				zap.S().Errorw("credit photonpay wallet", "error", err)
				return photonpayerrors.ErrDatabaseOperation
			}
		}
		return nil
	})
}
