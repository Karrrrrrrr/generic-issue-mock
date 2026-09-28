package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
	sharederrors "generic-mock/shared/errors"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type UISimulation interface {
	SimulateAuthorization(context.Context, *UISimulateAuthorizationRequest) (*CardTransactionSimulationResult, error)
	SimulateClearing(context.Context, *UISimulateClearingRequest) (*CardTransactionSimulationResult, error)
	SimulateReversal(context.Context, *UISimulateReversalRequest) (*CardTransactionSimulationResult, error)
	SimulateRefund(context.Context, *UISimulateRefundRequest) (*CardTransactionSimulationResult, error)
	ApplyTransactionStage(context.Context, *UIApplyTransactionStageRequest) (*CardTransactionSimulationResult, error)
}

type UISimulateAuthorizationRequest struct {
	CardID          model.ID
	Amount          decimal.Decimal
	Currency        enums.Currency
	MerchantName    *string
	MerchantCountry *string
	MerchantMCC     *string
	RequestID       *string
}

func (uc *ui) SimulateAuthorization(ctx context.Context, req *UISimulateAuthorizationRequest) (*CardTransactionSimulationResult, error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	if uc.notificator == nil {
		return nil, sharederrors.ErrUIUnsupported
	}
	return uc.simulator.SimulateAuthorization(ctx, &SimulateAuthorizationReq{
		Channel:         uc.channel,
		Notificator:     uc.notificator,
		CardID:          req.CardID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		MerchantName:    req.MerchantName,
		MerchantCountry: req.MerchantCountry,
		MerchantMCC:     req.MerchantMCC,
		RequestID:       req.RequestID,
	})
}

type UISimulateClearingRequest struct {
	AuthorizationID model.ID
	Amount          decimal.Decimal
	RequestID       *string
}

func (uc *ui) SimulateClearing(ctx context.Context, req *UISimulateClearingRequest) (*CardTransactionSimulationResult, error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	if uc.notificator == nil {
		return nil, sharederrors.ErrUIUnsupported
	}
	return uc.simulator.SimulateClearing(ctx, &SimulateClearingReq{
		Channel:         uc.channel,
		Notificator:     uc.notificator,
		AuthorizationID: req.AuthorizationID,
		Amount:          req.Amount,
		RequestID:       req.RequestID,
	})
}

type UISimulateReversalRequest struct {
	AuthorizationID model.ID
	Amount          decimal.Decimal
	RequestID       *string
}

func (uc *ui) SimulateReversal(ctx context.Context, req *UISimulateReversalRequest) (*CardTransactionSimulationResult, error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	if uc.notificator == nil {
		return nil, sharederrors.ErrUIUnsupported
	}
	return uc.simulator.SimulateReversal(ctx, &SimulateReversalReq{
		Channel:         uc.channel,
		Notificator:     uc.notificator,
		Status:          enums.TransactionStatus_VOID,
		AuthorizationID: req.AuthorizationID,
		Amount:          req.Amount,
		RequestID:       req.RequestID,
	})
}

type UISimulateRefundRequest struct {
	AuthorizationID *model.ID
	CardID          *model.ID
	Amount          decimal.Decimal
	Currency        *enums.Currency
	MerchantName    *string
	MerchantCountry *string
	MerchantMCC     *string
	RequestID       *string
}

func (uc *ui) SimulateRefund(ctx context.Context, req *UISimulateRefundRequest) (*CardTransactionSimulationResult, error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	if uc.notificator == nil {
		return nil, sharederrors.ErrUIUnsupported
	}
	return uc.simulator.SimulateRefund(ctx, &SimulateRefundReq{
		Channel:         uc.channel,
		Notificator:     uc.notificator,
		AuthorizationID: req.AuthorizationID,
		CardID:          req.CardID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		MerchantName:    req.MerchantName,
		MerchantCountry: req.MerchantCountry,
		MerchantMCC:     req.MerchantMCC,
		RequestID:       req.RequestID,
	})
}

type UIApplyTransactionStageRequest struct {
	ID        model.ID
	Type      enums.CardTransactionType
	Amount    *decimal.Decimal
	RequestID *string
}

func (req *UIApplyTransactionStageRequest) Validate() error {
	if req == nil || req.ID <= 0 || (req.Amount != nil && !req.Amount.IsPositive()) || !validUIOptionalText(req.RequestID) {
		return sharederrors.ErrInvalidUIRequest
	}
	switch req.Type {
	case enums.CardTransactionType_CLEAR, enums.CardTransactionType_VOID, enums.CardTransactionType_REFUND:
		return nil
	default:
		return sharederrors.ErrInvalidUIRequest
	}
}

func (uc *ui) ApplyTransactionStage(ctx context.Context, req *UIApplyTransactionStageRequest) (*CardTransactionSimulationResult, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if uc.notificator == nil {
		return nil, sharederrors.ErrUIUnsupported
	}
	exists, err := uc.cardTransactionRepo.ExistForSimulation(ctx, &CardTransactionExistForSimulationRequest{
		ID:      req.ID,
		Channel: uc.channel,
	})
	if err != nil {
		zap.S().Errorw("check shared UI simulation transaction", "error", err)
		return nil, sharederrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, sharederrors.ErrUITransactionNotFound
	}
	origin, err := uc.cardTransactionRepo.FindForSimulation(ctx, &CardTransactionFindForSimulationRequest{
		ID:      req.ID,
		Channel: uc.channel,
	})
	if err != nil {
		zap.S().Errorw("find shared UI simulation transaction", "error", err)
		return nil, sharederrors.ErrDatabaseOperation
	}
	if origin.AuthorizationID > 0 {
		authorization, err := uc.GetAuthorization(ctx, &GetUIAuthorizationRequest{
			ID:        origin.AuthorizationID,
			AccountID: origin.AccountID,
		})
		if err != nil {
			return nil, err
		}
		if authorization.CardID != origin.CardID || authorization.Currency != origin.TxCurrency {
			return nil, sharederrors.ErrInvalidAuthorization
		}
	} else {
		if _, err := uc.GetCard(ctx, &GetUICardRequest{
			ID:        origin.CardID,
			AccountID: origin.AccountID,
		}); err != nil {
			return nil, err
		}
	}
	amount := origin.TxAmount
	if req.Amount != nil {
		amount = *req.Amount
	}
	if req.Type != enums.CardTransactionType_REFUND &&
		(origin.AuthorizationID <= 0 || origin.Type != enums.CardTransactionType_AUTH || origin.Status != enums.TransactionStatus_AUTHORIZED) {
		return nil, sharederrors.ErrInvalidAuthorization
	}
	switch req.Type {
	case enums.CardTransactionType_CLEAR:
		return uc.SimulateClearing(ctx, &UISimulateClearingRequest{
			AuthorizationID: origin.AuthorizationID,
			Amount:          amount,
			RequestID:       req.RequestID,
		})
	case enums.CardTransactionType_VOID:
		return uc.SimulateReversal(ctx, &UISimulateReversalRequest{
			AuthorizationID: origin.AuthorizationID,
			Amount:          amount,
			RequestID:       req.RequestID,
		})
	default:
		refund := &UISimulateRefundRequest{
			CardID:    &origin.CardID,
			Currency:  &origin.TxCurrency,
			Amount:    amount,
			RequestID: req.RequestID,
		}
		if origin.AuthorizationID > 0 {
			refund.AuthorizationID = &origin.AuthorizationID
		}
		return uc.SimulateRefund(ctx, refund)
	}
}
