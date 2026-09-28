package errors

import (
	stderrors "errors"

	sharederrors "generic-mock/shared/errors"
)

func FromSimulation(err error) error {
	switch {
	case err == nil:
		return nil
	case stderrors.Is(err, sharederrors.ErrAccountNotFound),
		stderrors.Is(err, sharederrors.ErrCardNotFound),
		stderrors.Is(err, sharederrors.ErrAuthorizationNotFound),
		stderrors.Is(err, sharederrors.ErrWalletNotFound):
		return ErrResourceNotFound
	case stderrors.Is(err, sharederrors.ErrInvalidSimulationRequest),
		stderrors.Is(err, sharederrors.ErrCardNotActive),
		stderrors.Is(err, sharederrors.ErrInvalidAuthorization),
		stderrors.Is(err, sharederrors.ErrInvalidWallet),
		stderrors.Is(err, sharederrors.ErrInvalidBalanceChange),
		stderrors.Is(err, sharederrors.ErrWalletCurrencyMismatch),
		stderrors.Is(err, sharederrors.ErrInsufficientCardBalance),
		stderrors.Is(err, sharederrors.ErrInsufficientAvailableBalance),
		stderrors.Is(err, sharederrors.ErrInsufficientFrozenBalance),
		stderrors.Is(err, sharederrors.ErrSimulationRequestConflict):
		return ErrInvalidOperation
	default:
		return ErrDatabaseOperation
	}
}
