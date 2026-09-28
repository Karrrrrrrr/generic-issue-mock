package biz

import (
	"generic-mock/model"
	sharederrors "generic-mock/shared/errors"

	"github.com/shopspring/decimal"
)

type AuthorizationAccountingRequest struct {
	Wallet *model.Wallet
	Amount decimal.Decimal
}

func (req *AuthorizationAccountingRequest) Validate() error {
	if req == nil || req.Wallet == nil || !req.Amount.IsPositive() {
		return sharederrors.ErrInvalidSimulationRequest
	}
	return nil
}

type ClearingAccountingRequest struct {
	Wallet    *model.Wallet
	Amount    decimal.Decimal
	Remaining decimal.Decimal
}

func (req *ClearingAccountingRequest) Validate() error {
	if req == nil || req.Wallet == nil || !req.Amount.IsPositive() {
		return sharederrors.ErrInvalidSimulationRequest
	}
	return nil
}

type RefundAccountingRequest struct {
	Wallet *model.Wallet
	Amount decimal.Decimal
}

func (req *RefundAccountingRequest) Validate() error {
	if req == nil || req.Wallet == nil || !req.Amount.IsPositive() {
		return sharederrors.ErrInvalidSimulationRequest
	}
	return nil
}

type ReversalAccountingRequest struct {
	Wallet    *model.Wallet
	Amount    decimal.Decimal
	Remaining decimal.Decimal
}

func (req *ReversalAccountingRequest) Validate() error {
	if req == nil || req.Wallet == nil || !req.Amount.IsPositive() {
		return sharederrors.ErrInvalidSimulationRequest
	}
	return nil
}

type CardTransactionAccounting interface {
	Authorize(*AuthorizationAccountingRequest) error
	Clear(*ClearingAccountingRequest) error
	Refund(*RefundAccountingRequest) error
	Reverse(*ReversalAccountingRequest) error
}

type ClearingDebitAccounting struct{}

func (ClearingDebitAccounting) Authorize(req *AuthorizationAccountingRequest) error {
	return req.Validate()
}

func (ClearingDebitAccounting) Clear(req *ClearingAccountingRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}
	req.Wallet.Amount = req.Wallet.Amount.Sub(req.Amount)
	req.Wallet.Out = req.Wallet.Out.Add(req.Amount)
	return nil
}

func (ClearingDebitAccounting) Refund(req *RefundAccountingRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}
	req.Wallet.Amount = req.Wallet.Amount.Add(req.Amount)
	req.Wallet.In = req.Wallet.In.Add(req.Amount)
	return nil
}

func (ClearingDebitAccounting) Reverse(req *ReversalAccountingRequest) error {
	return req.Validate()
}

type AuthorizationHoldAccounting struct{}

func (AuthorizationHoldAccounting) Authorize(req *AuthorizationAccountingRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}
	if req.Wallet.Amount.Sub(req.Wallet.PendingOut).LessThan(req.Amount) {
		return sharederrors.ErrInsufficientCardBalance
	}
	req.Wallet.PendingOut = req.Wallet.PendingOut.Add(req.Amount)
	return nil
}

func (AuthorizationHoldAccounting) Clear(req *ClearingAccountingRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}
	release := decimal.Min(req.Amount, decimal.Max(req.Remaining, decimal.Zero))
	req.Wallet.PendingOut = req.Wallet.PendingOut.Sub(release)
	req.Wallet.Amount = req.Wallet.Amount.Sub(req.Amount)
	req.Wallet.Out = req.Wallet.Out.Add(req.Amount)
	return nil
}

func (AuthorizationHoldAccounting) Refund(req *RefundAccountingRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}
	req.Wallet.Amount = req.Wallet.Amount.Add(req.Amount)
	req.Wallet.In = req.Wallet.In.Add(req.Amount)
	return nil
}

func (AuthorizationHoldAccounting) Reverse(req *ReversalAccountingRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}
	release := decimal.Min(req.Amount, decimal.Max(req.Remaining, decimal.Zero))
	req.Wallet.PendingOut = req.Wallet.PendingOut.Sub(release)
	return nil
}
