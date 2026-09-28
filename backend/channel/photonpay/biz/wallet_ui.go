package biz

import (
	"context"
	"sort"

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
		return photonpayerrors.ErrInvalidOperation
	}
	return nil
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
		if card.Status != enums.CardStatus_Active {
			return photonpayerrors.ErrCardNotActive
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
		if account.WalletID == 0 || (sourceID != account.WalletID && targetID != account.WalletID) {
			return photonpayerrors.ErrInvalidOperation
		}
		cards, err := u.cardRepo.ListForFunding(ctx, &CardListForFundingRequest{
			AccountID: req.AccountID,
			WalletIDs: walletIDs,
		})
		if err != nil {
			zap.S().Errorw("lock photonpay funding cards", "error", err)
			return photonpayerrors.ErrDatabaseOperation
		}
		selectedCardFound := req.CardID == nil
		for _, card := range cards {
			selected := req.CardID != nil && card.ID == *req.CardID
			if selected {
				selectedCardFound = true
			}
			if (selected || card.CardType != enums.CardType_Share) && card.Status != enums.CardStatus_Active {
				return photonpayerrors.ErrCardNotActive
			}
		}
		if !selectedCardFound {
			return photonpayerrors.ErrResourceNotFound
		}
		wallets := make(map[model.ID]*model.Wallet)
		for _, walletID := range walletIDs {
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
		source, target := wallets[sourceID], wallets[targetID]
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
