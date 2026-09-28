package biz

import (
	"context"
	"errors"

	"generic-mock/enums"
	"generic-mock/model"
	sharederrors "generic-mock/shared/errors"

	"github.com/samber/do/v2"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type BalanceChangeReq struct {
	AccountID       model.ID
	Channel         enums.Channel
	WalletID        model.ID
	Currency        enums.Currency
	AvailableDelta  decimal.Decimal
	PendingOutDelta decimal.Decimal
	InDelta         decimal.Decimal
	OutDelta        decimal.Decimal
	CheckAvailable  bool
}

func (req *BalanceChangeReq) Validate() error {
	if req == nil || req.AccountID <= 0 || req.Channel == "" || req.WalletID <= 0 || req.Currency == "" {
		return sharederrors.ErrInvalidBalanceChange
	}
	if req.AvailableDelta.IsZero() && req.PendingOutDelta.IsZero() && req.InDelta.IsZero() && req.OutDelta.IsZero() {
		return sharederrors.ErrInvalidBalanceChange
	}
	return nil
}

type ChangeBalanceSimpleReq struct {
	AccountID      model.ID
	Channel        enums.Channel
	WalletID       model.ID
	Currency       enums.Currency
	Amount         decimal.Decimal
	CheckAvailable bool
}

func (req *ChangeBalanceSimpleReq) Validate() error {
	if req == nil || req.AccountID <= 0 || req.Channel == "" || req.WalletID <= 0 ||
		req.Currency == "" || req.Amount.IsZero() {
		return sharederrors.ErrInvalidBalanceChange
	}
	return nil
}

type TccBalanceChangeReq struct {
	AccountID      model.ID
	Channel        enums.Channel
	WalletID       model.ID
	Currency       enums.Currency
	Amount         decimal.Decimal
	CheckAvailable bool
}

func (req *TccBalanceChangeReq) Validate() error {
	if req == nil || req.AccountID <= 0 || req.Channel == "" || req.WalletID <= 0 ||
		req.Currency == "" || !req.Amount.IsPositive() {
		return sharederrors.ErrInvalidBalanceChange
	}
	return nil
}

type ConfirmBalanceChangeReq struct {
	AccountID      model.ID
	Channel        enums.Channel
	WalletID       model.ID
	Currency       enums.Currency
	Amount         decimal.Decimal
	ReleaseAmount  decimal.Decimal
	CheckAvailable bool
}

func (req *ConfirmBalanceChangeReq) Validate() error {
	if req == nil || req.AccountID <= 0 || req.Channel == "" || req.WalletID <= 0 ||
		req.Currency == "" || !req.Amount.IsPositive() || req.ReleaseAmount.IsNegative() ||
		req.ReleaseAmount.GreaterThan(req.Amount) {
		return sharederrors.ErrInvalidBalanceChange
	}
	return nil
}

type BalanceChanger interface {
	ChangeBalance(context.Context, *BalanceChangeReq) error
	TryBalanceChange(context.Context, *TccBalanceChangeReq) error
	ConfirmBalanceChange(context.Context, *ConfirmBalanceChangeReq) error
	CancelBalanceChange(context.Context, *TccBalanceChangeReq) error
	ChangeBalanceSimple(context.Context, *ChangeBalanceSimpleReq) error
}

type balanceChanger struct {
	walletRepo WalletRepo
	tx         Transaction
}

var _ BalanceChanger = (*balanceChanger)(nil)

func NewBalanceChanger(injector do.Injector) (BalanceChanger, error) {
	return &balanceChanger{
		walletRepo: do.MustInvoke[WalletRepo](injector),
		tx:         do.MustInvoke[Transaction](injector),
	}, nil
}

func (changer *balanceChanger) ChangeBalance(ctx context.Context, req *BalanceChangeReq) error {
	if err := req.Validate(); err != nil {
		return err
	}
	if changer.tx.IsInTx(ctx) {
		return changer.changeBalance(ctx, req)
	}
	var operationErr error
	err := changer.tx.InTx(ctx, func(ctx context.Context) error {
		operationErr = changer.changeBalance(ctx, req)
		return operationErr
	})
	if err != nil {
		if operationErr != nil {
			return operationErr
		}
		if errors.Is(err, sharederrors.ErrNestedTransaction) {
			return err
		}
		zap.S().Errorw("run shared balance change transaction",
			"account_id", req.AccountID,
			"channel", req.Channel,
			"wallet_id", req.WalletID,
			"error", err,
		)
		return sharederrors.ErrDatabaseOperation
	}
	return nil
}

func (changer *balanceChanger) TryBalanceChange(ctx context.Context, req *TccBalanceChangeReq) error {
	if err := req.Validate(); err != nil {
		return err
	}
	return changer.ChangeBalance(ctx, &BalanceChangeReq{
		AccountID:       req.AccountID,
		Channel:         req.Channel,
		WalletID:        req.WalletID,
		Currency:        req.Currency,
		AvailableDelta:  req.Amount.Neg(),
		PendingOutDelta: req.Amount,
		CheckAvailable:  req.CheckAvailable,
	})
}

func (changer *balanceChanger) ConfirmBalanceChange(ctx context.Context, req *ConfirmBalanceChangeReq) error {
	if err := req.Validate(); err != nil {
		return err
	}
	return changer.ChangeBalance(ctx, &BalanceChangeReq{
		AccountID:       req.AccountID,
		Channel:         req.Channel,
		WalletID:        req.WalletID,
		Currency:        req.Currency,
		AvailableDelta:  req.ReleaseAmount.Sub(req.Amount),
		PendingOutDelta: req.ReleaseAmount.Neg(),
		OutDelta:        req.Amount,
		CheckAvailable:  req.CheckAvailable,
	})
}

func (changer *balanceChanger) CancelBalanceChange(ctx context.Context, req *TccBalanceChangeReq) error {
	if err := req.Validate(); err != nil {
		return err
	}
	return changer.ChangeBalance(ctx, &BalanceChangeReq{
		AccountID:       req.AccountID,
		Channel:         req.Channel,
		WalletID:        req.WalletID,
		Currency:        req.Currency,
		AvailableDelta:  req.Amount,
		PendingOutDelta: req.Amount.Neg(),
		CheckAvailable:  req.CheckAvailable,
	})
}

func (changer *balanceChanger) ChangeBalanceSimple(ctx context.Context, req *ChangeBalanceSimpleReq) error {
	if err := req.Validate(); err != nil {
		return err
	}
	inDelta := decimal.Zero
	outDelta := decimal.Zero
	if req.Amount.IsPositive() {
		inDelta = req.Amount
	} else {
		outDelta = req.Amount.Neg()
	}
	return changer.ChangeBalance(ctx, &BalanceChangeReq{
		AccountID:      req.AccountID,
		Channel:        req.Channel,
		WalletID:       req.WalletID,
		Currency:       req.Currency,
		AvailableDelta: req.Amount,
		InDelta:        inDelta,
		OutDelta:       outDelta,
		CheckAvailable: req.CheckAvailable,
	})
}

func (changer *balanceChanger) changeBalance(ctx context.Context, req *BalanceChangeReq) error {
	exists, err := changer.walletRepo.Exist(ctx, &WalletExistRequest{
		ID:        req.WalletID,
		AccountID: req.AccountID,
		Channel:   req.Channel,
	})
	if err != nil {
		zap.S().Errorw("check shared balance change wallet",
			"account_id", req.AccountID,
			"channel", req.Channel,
			"wallet_id", req.WalletID,
			"error", err,
		)
		return sharederrors.ErrDatabaseOperation
	}
	if !exists {
		return sharederrors.ErrWalletNotFound
	}
	wallet, err := changer.walletRepo.FindByIDWithLock(ctx, &WalletFindByIDWithLockRequest{
		ID:        req.WalletID,
		AccountID: req.AccountID,
		Channel:   req.Channel,
	})
	if err != nil {
		zap.S().Errorw("lock shared balance change wallet",
			"account_id", req.AccountID,
			"channel", req.Channel,
			"wallet_id", req.WalletID,
			"error", err,
		)
		return sharederrors.ErrDatabaseOperation
	}
	if wallet.Currency != req.Currency {
		return sharederrors.ErrWalletCurrencyMismatch
	}
	available := wallet.Available.Add(req.AvailableDelta)
	pendingOut := wallet.PendingOut.Add(req.PendingOutDelta)
	inAmount := wallet.In.Add(req.InDelta)
	outAmount := wallet.Out.Add(req.OutDelta)
	if req.CheckAvailable && available.IsNegative() {
		return sharederrors.ErrInsufficientAvailableBalance
	}
	if pendingOut.IsNegative() {
		return sharederrors.ErrInsufficientFrozenBalance
	}
	if inAmount.IsNegative() || outAmount.IsNegative() {
		return sharederrors.ErrInvalidBalanceChange
	}
	if err := changer.walletRepo.UpdateBalance(ctx, &WalletUpdateBalanceRequest{
		ID:         req.WalletID,
		AccountID:  req.AccountID,
		Channel:    req.Channel,
		Available:  available,
		PendingOut: pendingOut,
		In:         inAmount,
		Out:        outAmount,
	}); err != nil {
		zap.S().Errorw("update shared wallet balance",
			"account_id", req.AccountID,
			"channel", req.Channel,
			"wallet_id", req.WalletID,
			"error", err,
		)
		return sharederrors.ErrDatabaseOperation
	}
	return nil
}
