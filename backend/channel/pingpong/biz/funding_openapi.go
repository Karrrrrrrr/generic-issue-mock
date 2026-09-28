package biz

import (
	"context"

	pingerrors "generic-mock/channel/pingpong/errors"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/cardwallet"
	"generic-mock/pkg/types"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type CardFundingRequest struct {
	AccountID model.ID
	CardID    model.ID
	Withdraw  bool
	Amount    decimal.Decimal
	RequestID string
}

type VirtualAccountFundingRequest struct {
	AccountID              model.ID
	VirtualAccountID       model.ID
	TargetVirtualAccountID *model.ID
	Currency               common.Currency
	Amount                 decimal.Decimal
	RequestID              *string
}

type openAPIWalletTransferRequest struct {
	AccountID model.ID
	CardID    *model.ID
	SourceID  model.ID
	TargetID  model.ID
	Kind      common.WalletTransferKind
	Currency  common.Currency
	Amount    decimal.Decimal
	RequestID *string
}

func (uc *PingPongOpenAPIUsecase) FundCard(ctx context.Context, req *CardFundingRequest) (*model.WalletTransfer, error) {
	if req.RequestID == "" || !req.Amount.IsPositive() {
		return nil, pingerrors.ErrInvalid
	}
	var result *model.WalletTransfer
	err := uc.tx.InTx(ctx, func(ctx context.Context) error {
		if _, err := uc.lockAccount(ctx, req.AccountID); err != nil {
			return err
		}
		card, err := uc.GetCard(ctx, &GetCardRequest{
			AccountID: req.AccountID,
			ID:        req.CardID,
		})
		if err != nil {
			return err
		}
		fundingID, ok := cardwallet.FundingWalletID(card)
		if !ok || card.CardType != common.CardType_VirtualAccountSingle {
			return pingerrors.ErrInvalid
		}
		transfer := &openAPIWalletTransferRequest{
			AccountID: req.AccountID,
			CardID:    &card.ID,
			SourceID:  fundingID,
			TargetID:  card.WalletID,
			Kind:      common.WalletTransfer_CardTopUp,
			Currency:  card.CardCurrency,
			Amount:    req.Amount,
			RequestID: &req.RequestID,
		}
		if req.Withdraw {
			transfer.SourceID, transfer.TargetID = transfer.TargetID, transfer.SourceID
			transfer.Kind = common.WalletTransfer_CardWithdraw
		}
		previous, err := uc.findPreviousTransfer(ctx, transfer)
		if err != nil {
			return err
		}
		if previous != nil {
			result = previous
			return nil
		}
		if card.Status == common.CardStatus_Deleted {
			return pingerrors.ErrClosed
		}
		result, err = uc.transferWalletFunds(ctx, transfer)
		return err
	})
	return result, err
}

func (uc *PingPongOpenAPIUsecase) FundVirtualAccount(ctx context.Context, req *VirtualAccountFundingRequest) (*model.WalletTransfer, error) {
	if !req.Amount.IsPositive() || req.Currency != common.Currency_USD || (req.RequestID != nil && *req.RequestID == "") {
		return nil, pingerrors.ErrInvalid
	}
	if req.TargetVirtualAccountID == nil && req.RequestID == nil {
		return nil, pingerrors.ErrInvalid
	}
	var result *model.WalletTransfer
	err := uc.tx.InTx(ctx, func(ctx context.Context) error {
		account, err := uc.lockAccount(ctx, req.AccountID)
		if err != nil {
			return err
		}
		virtualAccount, err := uc.getVirtualAccount(ctx, &openAPIVirtualAccountReference{
			AccountID: req.AccountID,
			ID:        req.VirtualAccountID,
		})
		if err != nil {
			return err
		}
		transfer := &openAPIWalletTransferRequest{
			AccountID: req.AccountID,
			SourceID:  account.WalletID,
			TargetID:  virtualAccount.WalletID,
			Kind:      common.WalletTransfer_VirtualAccountTopUp,
			Currency:  req.Currency,
			Amount:    req.Amount,
		}
		if req.RequestID != nil {
			transfer.RequestID = req.RequestID
		}
		if req.TargetVirtualAccountID != nil {
			target, err := uc.getVirtualAccount(ctx, &openAPIVirtualAccountReference{
				AccountID: req.AccountID,
				ID:        *req.TargetVirtualAccountID,
			})
			if err != nil {
				return err
			}
			transfer.SourceID = virtualAccount.WalletID
			transfer.TargetID = target.WalletID
			transfer.Kind = common.WalletTransfer_VirtualAccountTransfer
		}
		previous, err := uc.findPreviousTransfer(ctx, transfer)
		if err != nil {
			return err
		}
		if previous != nil {
			result = previous
			return nil
		}
		result, err = uc.transferWalletFunds(ctx, transfer)
		return err
	})
	return result, err
}

func (uc *PingPongOpenAPIUsecase) findPreviousTransfer(ctx context.Context, req *openAPIWalletTransferRequest) (*model.WalletTransfer, error) {
	if req.RequestID == nil {
		return nil, nil
	}
	exists, err := uc.walletTransferRepo.ExistsRequest(ctx, &TransferRequestExistsRequest{
		AccountID: req.AccountID,
		RequestID: *req.RequestID,
	})
	if err != nil {
		zap.S().Errorw("check pingpong transfer request", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	if !exists {
		return nil, nil
	}
	previous, err := uc.walletTransferRepo.FindRequest(ctx, &TransferRequestFindRequest{
		AccountID: req.AccountID,
		RequestID: *req.RequestID,
	})
	if err != nil {
		zap.S().Errorw("find pingpong transfer request", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	if previous.SourceWalletID != req.SourceID || previous.TargetWalletID != req.TargetID || previous.Kind != req.Kind || previous.Currency != req.Currency || !previous.Amount.Equal(req.Amount) {
		return nil, pingerrors.ErrConflict
	}
	return previous, nil
}

func (uc *PingPongOpenAPIUsecase) transferWalletFunds(ctx context.Context, req *openAPIWalletTransferRequest) (*model.WalletTransfer, error) {
	if req.SourceID <= 0 || req.TargetID <= 0 || req.SourceID == req.TargetID {
		return nil, pingerrors.ErrInvalid
	}
	ids := []model.ID{req.SourceID, req.TargetID}
	if ids[0] > ids[1] {
		ids[0], ids[1] = ids[1], ids[0]
	}
	wallets := make(map[model.ID]*model.Wallet)
	for _, id := range ids {
		wallet, err := uc.walletRepo.Lock(ctx, &WalletLockRequest{
			AccountID: req.AccountID,
			ID:        id,
		})
		if err != nil {
			zap.S().Errorw("lock pingpong transfer wallet", "error", err)
			return nil, pingerrors.ErrDatabase
		}
		wallets[id] = wallet
	}
	source := wallets[req.SourceID]
	target := wallets[req.TargetID]
	if source.Currency != req.Currency || target.Currency != req.Currency {
		return nil, pingerrors.ErrInvalid
	}
	if source.Available.LessThan(req.Amount) {
		return nil, pingerrors.ErrInsufficient
	}
	result := &model.WalletTransfer{
		AccountID:      req.AccountID,
		Channel:        common.Channel_PingPong,
		RequestID:      types.Value(req.RequestID),
		Kind:           req.Kind,
		CardID:         req.CardID,
		SourceWalletID: source.ID,
		TargetWalletID: target.ID,
		Currency:       req.Currency,
		Amount:         req.Amount,
		SourceBefore:   source.Available,
		TargetBefore:   target.Available,
	}
	source.Available = source.Available.Sub(req.Amount)
	source.Out = source.Out.Add(req.Amount)
	target.Available = target.Available.Add(req.Amount)
	target.In = target.In.Add(req.Amount)
	result.SourceAfter = source.Available
	result.TargetAfter = target.Available
	for _, wallet := range []*model.Wallet{source, target} {
		if err := uc.walletRepo.Save(ctx, wallet); err != nil {
			zap.S().Errorw("save pingpong transfer wallet", "error", err)
			return nil, pingerrors.ErrDatabase
		}
	}
	if err := uc.walletTransferRepo.Create(ctx, result); err != nil {
		zap.S().Errorw("create pingpong transfer record", "error", err)
		return nil, pingerrors.ErrDatabase
	}
	return result, nil
}
