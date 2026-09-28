package biz

import (
	"context"
	"sort"

	slasherrors "generic-mock/channel/slash/errors"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type MoveFundsRequest struct {
	CardID    *model.ID
	AccountID model.ID
	SourceID  *model.ID
	TargetID  *model.ID
	Amount    decimal.Decimal
}

func (req *MoveFundsRequest) Validate() error {
	if req == nil || req.AccountID <= 0 || !req.Amount.IsPositive() ||
		(req.CardID != nil && *req.CardID <= 0) ||
		(req.SourceID != nil && *req.SourceID <= 0) ||
		(req.TargetID != nil && *req.TargetID <= 0) ||
		(req.SourceID == nil && req.TargetID == nil) ||
		(req.SourceID != nil && req.TargetID != nil && *req.SourceID == *req.TargetID) {
		return slasherrors.ErrInvalidOperation
	}
	return nil
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
	if err := req.Validate(); err != nil {
		return err
	}
	sourceID := types.Value(req.SourceID)
	targetID := types.Value(req.TargetID)
	walletIDs := make([]model.ID, 0, 2)
	if req.SourceID != nil {
		walletIDs = append(walletIDs, *req.SourceID)
	}
	if req.TargetID != nil {
		walletIDs = append(walletIDs, *req.TargetID)
	}
	sort.Slice(walletIDs, func(left, right int) bool {
		return walletIDs[left] < walletIDs[right]
	})

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
		if account.WalletID == 0 || (sourceID != account.WalletID && targetID != account.WalletID) {
			return slasherrors.ErrInvalidOperation
		}
		cards, err := u.cardRepository.ListForFunding(ctx, &CardListForFundingRequest{
			AccountID: req.AccountID,
			WalletIDs: walletIDs,
		})
		if err != nil {
			zap.S().Errorw("lock slash funding cards", "error", err)
			return slasherrors.ErrDatabaseOperation
		}
		selectedCardFound := req.CardID == nil
		for _, card := range cards {
			selected := req.CardID != nil && card.ID == *req.CardID
			if selected {
				selectedCardFound = true
			}
			if (selected || card.CardType != enums.CardType_Share) && card.Status != enums.CardStatus_Active {
				return slasherrors.ErrCardNotActive
			}
		}
		if !selectedCardFound {
			return slasherrors.ErrResourceNotFound
		}
		wallets := make(map[model.ID]*model.Wallet)
		for _, walletID := range walletIDs {
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
		source, target := wallets[sourceID], wallets[targetID]
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
