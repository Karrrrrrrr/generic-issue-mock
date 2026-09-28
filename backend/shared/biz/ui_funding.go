package biz

import (
	"context"
	"sort"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/cardwallet"
	"generic-mock/pkg/types"
	sharederrors "generic-mock/shared/errors"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type UIFunding interface {
	FundCard(context.Context, *UIFundCardRequest) (*model.WalletTransfer, error)
	FundVirtualAccount(context.Context, *UIFundVirtualAccountRequest) (*model.WalletTransfer, error)
}

type UIFundCardRequest struct {
	AccountID model.ID
	CardID    model.ID
	Kind      enums.WalletTransferKind
	Amount    decimal.Decimal
	RequestID *string
}

func (req *UIFundCardRequest) Validate() error {
	if req == nil || req.AccountID <= 0 || req.CardID <= 0 || !req.Amount.IsPositive() || !validUIOptionalText(req.RequestID) ||
		(req.Kind != enums.WalletTransfer_CardTopUp && req.Kind != enums.WalletTransfer_CardWithdraw) {
		return sharederrors.ErrInvalidUIRequest
	}
	return nil
}

type UIFundVirtualAccountRequest struct {
	AccountID              model.ID
	VirtualAccountID       model.ID
	TargetVirtualAccountID *model.ID
	Withdraw               *bool
	Amount                 decimal.Decimal
	RequestID              *string
}

func (req *UIFundVirtualAccountRequest) Validate() error {
	if req == nil || req.AccountID <= 0 || req.VirtualAccountID <= 0 || !req.Amount.IsPositive() ||
		!validUIIDs([]*model.ID{req.TargetVirtualAccountID}) || !validUIOptionalText(req.RequestID) ||
		(req.TargetVirtualAccountID != nil && (req.VirtualAccountID == *req.TargetVirtualAccountID || types.Value(req.Withdraw))) {
		return sharederrors.ErrInvalidUIRequest
	}
	return nil
}

type uiFundingVirtualAccountRequest struct {
	ID        model.ID
	AccountID model.ID
}

type uiTransferRequest struct {
	SourceWalletType enums.WalletType
	TargetWalletType enums.WalletType
	AccountID        model.ID
	CardID           *model.ID
	SourceWalletID   model.ID
	TargetWalletID   model.ID
	Currency         enums.Currency
	Amount           decimal.Decimal
	Kind             enums.WalletTransferKind
	RequestID        *string
}

func (uc *ui) FundCard(ctx context.Context, req *UIFundCardRequest) (*model.WalletTransfer, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var result *model.WalletTransfer
	err := uc.tx.InTx(ctx, func(ctx context.Context) error {
		account, err := uc.lockUIAccount(ctx, req.AccountID)
		if err != nil {
			return err
		}
		exists, err := uc.cardRepo.Exist(ctx, &CardExistRequest{
			ID:        req.CardID,
			AccountID: account.ID,
			Channel:   uc.channel,
		})
		if err != nil {
			zap.S().Errorw("check shared UI funding card", "error", err)
			return sharederrors.ErrDatabaseOperation
		}
		if !exists {
			return sharederrors.ErrCardNotFound
		}
		card, err := uc.cardRepo.FindByIDWithLock(ctx, &CardFindByIDWithLockRequest{
			ID:        req.CardID,
			AccountID: account.ID,
			Channel:   uc.channel,
		})
		if err != nil {
			zap.S().Errorw("lock shared UI funding card", "error", err)
			return sharederrors.ErrDatabaseOperation
		}
		if card.Status != enums.CardStatus_Active {
			return sharederrors.ErrCardNotActive
		}
		fundingID, ok := cardwallet.FundingWalletID(card)
		if !ok {
			return sharederrors.ErrInvalidWallet
		}
		cardWalletType := enums.WalletType_Card
		fundingWalletType := enums.WalletType_Account
		if card.CardType == enums.CardType_Share {
			cardWalletType = enums.WalletType_VirtualAccount
		}
		if card.CardType == enums.CardType_VirtualAccountSingle {
			fundingWalletType = enums.WalletType_VirtualAccount
		}

		transfer := &uiTransferRequest{
			SourceWalletType: fundingWalletType,
			TargetWalletType: cardWalletType,
			AccountID:        account.ID,
			CardID:           &card.ID,
			SourceWalletID:   fundingID,
			TargetWalletID:   card.WalletID,
			Currency:         card.CardCurrency,
			Amount:           req.Amount,
			Kind:             req.Kind,
			RequestID:        req.RequestID,
		}
		if req.Kind == enums.WalletTransfer_CardWithdraw {
			transfer.SourceWalletType = cardWalletType
			transfer.TargetWalletType = fundingWalletType
			transfer.SourceWalletID = card.WalletID
			transfer.TargetWalletID = fundingID
		}
		result, err = uc.transferUIWalletFunds(ctx, transfer)
		if err == nil {
			result.Account = account
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (uc *ui) FundVirtualAccount(ctx context.Context, req *UIFundVirtualAccountRequest) (*model.WalletTransfer, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var result *model.WalletTransfer
	err := uc.tx.InTx(ctx, func(ctx context.Context) error {
		account, err := uc.lockUIAccount(ctx, req.AccountID)
		if err != nil {
			return err
		}
		virtualAccount, err := uc.findUIVirtualAccount(ctx, &uiFundingVirtualAccountRequest{
			ID:        req.VirtualAccountID,
			AccountID: req.AccountID,
		})
		if err != nil {
			return err
		}
		if virtualAccount.Wallet == nil || virtualAccount.Wallet.Type != enums.WalletType_VirtualAccount {
			return sharederrors.ErrInvalidWallet
		}
		transfer := &uiTransferRequest{
			SourceWalletType: enums.WalletType_Account,
			TargetWalletType: enums.WalletType_VirtualAccount,
			AccountID:        account.ID,
			SourceWalletID:   account.WalletID,
			TargetWalletID:   virtualAccount.WalletID,
			Currency:         virtualAccount.Wallet.Currency,
			Amount:           req.Amount,
			Kind:             enums.WalletTransfer_VirtualAccountTopUp,
			RequestID:        req.RequestID,
		}
		if types.Value(req.Withdraw) {
			transfer.SourceWalletType = enums.WalletType_VirtualAccount
			transfer.TargetWalletType = enums.WalletType_Account
			transfer.SourceWalletID = virtualAccount.WalletID
			transfer.TargetWalletID = account.WalletID
			transfer.Kind = enums.WalletTransfer_VirtualAccountTransfer
		}
		if req.TargetVirtualAccountID != nil {
			target, err := uc.findUIVirtualAccount(ctx, &uiFundingVirtualAccountRequest{
				ID:        *req.TargetVirtualAccountID,
				AccountID: req.AccountID,
			})
			if err != nil {
				return err
			}
			transfer.SourceWalletID = virtualAccount.WalletID
			transfer.TargetWalletID = target.WalletID
			transfer.SourceWalletType = enums.WalletType_VirtualAccount
			transfer.TargetWalletType = enums.WalletType_VirtualAccount
			transfer.Kind = enums.WalletTransfer_VirtualAccountTransfer
		}
		result, err = uc.transferUIWalletFunds(ctx, transfer)
		if err == nil {
			result.Account = account
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (uc *ui) findUIVirtualAccount(ctx context.Context, req *uiFundingVirtualAccountRequest) (*model.VirtualAccount, error) {
	exists, err := uc.virtualAccountRepo.Exist(ctx, &VirtualAccountExistRequest{
		ID:        req.ID,
		AccountID: req.AccountID,
		Channel:   uc.channel,
	})
	if err != nil {
		zap.S().Errorw("check shared UI funding virtual account", "error", err)
		return nil, sharederrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, sharederrors.ErrVirtualAccountNotFound
	}
	item, err := uc.virtualAccountRepo.Find(ctx, &VirtualAccountFindRequest{
		ID:        req.ID,
		AccountID: req.AccountID,
		Channel:   uc.channel,
	})
	if err != nil {
		zap.S().Errorw("find shared UI funding virtual account", "error", err)
		return nil, sharederrors.ErrDatabaseOperation
	}
	return item, nil
}

func (uc *ui) transferUIWalletFunds(ctx context.Context, req *uiTransferRequest) (*model.WalletTransfer, error) {
	if req.SourceWalletID <= 0 || req.TargetWalletID <= 0 || req.SourceWalletID == req.TargetWalletID {
		return nil, sharederrors.ErrInvalidWallet
	}
	if req.RequestID != nil {
		exists, err := uc.walletTransferRepo.ExistByRequestID(ctx, &WalletTransferExistByRequestIDRequest{
			AccountID: req.AccountID,
			Channel:   uc.channel,
			RequestID: *req.RequestID,
		})
		if err != nil {
			zap.S().Errorw("check shared UI funding request", "error", err)
			return nil, sharederrors.ErrDatabaseOperation
		}
		if exists {
			previous, err := uc.walletTransferRepo.FindByRequestID(ctx, &WalletTransferFindByRequestIDRequest{
				AccountID: req.AccountID,
				Channel:   uc.channel,
				RequestID: *req.RequestID,
			})
			if err != nil {
				zap.S().Errorw("find shared UI funding request", "error", err)
				return nil, sharederrors.ErrDatabaseOperation
			}
			if previous.SourceWalletID != req.SourceWalletID || previous.TargetWalletID != req.TargetWalletID ||
				previous.Currency != req.Currency || previous.Kind != req.Kind || !previous.Amount.Equal(req.Amount) ||
				(previous.CardID == nil) != (req.CardID == nil) || types.Value(previous.CardID) != types.Value(req.CardID) {
				return nil, sharederrors.ErrUITransferConflict
			}
			return previous, nil
		}
	}
	ids := []model.ID{req.SourceWalletID, req.TargetWalletID}
	sort.Slice(ids, func(left, right int) bool { return ids[left] < ids[right] })
	wallets := make(map[model.ID]*model.Wallet, len(ids))
	for _, walletID := range ids {
		exists, err := uc.walletRepo.Exist(ctx, &WalletExistRequest{
			ID:        walletID,
			AccountID: req.AccountID,
			Channel:   uc.channel,
		})
		if err != nil {
			zap.S().Errorw("check shared UI transfer wallet", "error", err)
			return nil, sharederrors.ErrDatabaseOperation
		}
		if !exists {
			return nil, sharederrors.ErrWalletNotFound
		}
		wallet, err := uc.walletRepo.FindByIDWithLock(ctx, &WalletFindByIDWithLockRequest{
			ID:        walletID,
			AccountID: req.AccountID,
			Channel:   uc.channel,
		})
		if err != nil {
			zap.S().Errorw("lock shared UI transfer wallet", "error", err)
			return nil, sharederrors.ErrDatabaseOperation
		}
		if wallet.Currency != req.Currency {
			return nil, sharederrors.ErrWalletCurrencyMismatch
		}
		expectedType := req.SourceWalletType
		if walletID == req.TargetWalletID {
			expectedType = req.TargetWalletType
		}
		if wallet.Type != expectedType {
			return nil, sharederrors.ErrInvalidWallet
		}
		wallets[walletID] = wallet
	}
	source := wallets[req.SourceWalletID]
	target := wallets[req.TargetWalletID]
	if err := uc.balanceChanger.ChangeBalanceSimple(ctx, &ChangeBalanceSimpleReq{
		AccountID:      req.AccountID,
		Channel:        uc.channel,
		WalletID:       source.ID,
		Currency:       req.Currency,
		Amount:         req.Amount.Neg(),
		CheckAvailable: true,
	}); err != nil {
		return nil, err
	}
	if err := uc.balanceChanger.ChangeBalanceSimple(ctx, &ChangeBalanceSimpleReq{
		AccountID:      req.AccountID,
		Channel:        uc.channel,
		WalletID:       target.ID,
		Currency:       req.Currency,
		Amount:         req.Amount,
		CheckAvailable: false,
	}); err != nil {
		return nil, err
	}
	result := &model.WalletTransfer{
		AccountID:      req.AccountID,
		Channel:        uc.channel,
		CardID:         req.CardID,
		RequestID:      types.Value(req.RequestID),
		SourceWalletID: source.ID,
		TargetWalletID: target.ID,
		Currency:       req.Currency,
		Amount:         req.Amount,
		Kind:           req.Kind,
		SourceBefore:   source.Available,
		SourceAfter:    source.Available.Sub(req.Amount),
		TargetBefore:   target.Available,
		TargetAfter:    target.Available.Add(req.Amount),
	}
	if err := uc.walletTransferRepo.Create(ctx, &WalletTransferCreateRequest{WalletTransfer: result}); err != nil {
		zap.S().Errorw("create shared UI funding order", "error", err)
		return nil, sharederrors.ErrDatabaseOperation
	}
	return result, nil
}
