package service

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/shared/biz"
	sharederrors "generic-mock/shared/errors"

	"github.com/shopspring/decimal"
)

type SimulationData struct {
	AuthorizationID   model.ID            `json:"authorization_id"`
	Transaction       CardTransactionData `json:"transaction"`
	Remaining         decimal.Decimal     `json:"remaining"`
	Replayed          bool                `json:"replayed"`
	NotificationError *string             `json:"notification_error,omitempty"`
}

func toSimulationData(result *biz.CardTransactionSimulationResult) *SimulationData {
	data := &SimulationData{
		AuthorizationID: result.CardTransaction.AuthorizationID,
		Transaction:     toCardTransactionData(result.CardTransaction),
		Remaining:       result.Remaining,
		Replayed:        result.Replayed,
	}
	if result.Authorization != nil {
		data.Transaction.AuthorizedAt = &result.Authorization.CreatedAt
	}
	if result.NotificationError != nil {
		message := result.NotificationError.Error()
		data.NotificationError = &message
	}
	return data
}

type SimulateAuthorizationRequest struct {
	CardID          model.ID        `json:"card_id"`
	Amount          decimal.Decimal `json:"amount"`
	Currency        enums.Currency  `json:"currency"`
	MerchantName    *string         `json:"merchant_name"`
	MerchantCountry *string         `json:"merchant_country"`
	MerchantMCC     *string         `json:"merchant_mcc"`
	RequestID       *string         `json:"request_id"`
}

func (s *Service) SimulateAuthorization(ctx context.Context, req *SimulateAuthorizationRequest) (*SimulationData, error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	result, err := s.uc.SimulateAuthorization(ctx, &biz.UISimulateAuthorizationRequest{
		CardID:          req.CardID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		MerchantName:    req.MerchantName,
		MerchantCountry: req.MerchantCountry,
		MerchantMCC:     req.MerchantMCC,
		RequestID:       req.RequestID,
	})
	if err != nil {
		return nil, err
	}
	return toSimulationData(result), nil
}

type SimulateClearingRequest struct {
	AuthorizationID model.ID        `json:"authorization_id"`
	Amount          decimal.Decimal `json:"amount"`
	RequestID       *string         `json:"request_id"`
}

func (s *Service) SimulateClearing(ctx context.Context, req *SimulateClearingRequest) (*SimulationData, error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	result, err := s.uc.SimulateClearing(ctx, &biz.UISimulateClearingRequest{
		AuthorizationID: req.AuthorizationID,
		Amount:          req.Amount,
		RequestID:       req.RequestID,
	})
	if err != nil {
		return nil, err
	}
	return toSimulationData(result), nil
}

type SimulateReversalRequest struct {
	AuthorizationID model.ID        `json:"authorization_id"`
	Amount          decimal.Decimal `json:"amount"`
	RequestID       *string         `json:"request_id"`
}

func (s *Service) SimulateReversal(ctx context.Context, req *SimulateReversalRequest) (*SimulationData, error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	result, err := s.uc.SimulateReversal(ctx, &biz.UISimulateReversalRequest{
		AuthorizationID: req.AuthorizationID,
		Amount:          req.Amount,
		RequestID:       req.RequestID,
	})
	if err != nil {
		return nil, err
	}
	return toSimulationData(result), nil
}

type SimulateRefundRequest struct {
	AuthorizationID *model.ID       `json:"authorization_id"`
	CardID          *model.ID       `json:"card_id"`
	Amount          decimal.Decimal `json:"amount"`
	Currency        *enums.Currency `json:"currency"`
	MerchantName    *string         `json:"merchant_name"`
	MerchantCountry *string         `json:"merchant_country"`
	MerchantMCC     *string         `json:"merchant_mcc"`
	RequestID       *string         `json:"request_id"`
}

func (s *Service) SimulateRefund(ctx context.Context, req *SimulateRefundRequest) (*SimulationData, error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	result, err := s.uc.SimulateRefund(ctx, &biz.UISimulateRefundRequest{
		AuthorizationID: req.AuthorizationID,
		CardID:          req.CardID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		MerchantName:    req.MerchantName,
		MerchantCountry: req.MerchantCountry,
		MerchantMCC:     req.MerchantMCC,
		RequestID:       req.RequestID,
	})
	if err != nil {
		return nil, err
	}
	return toSimulationData(result), nil
}

type ApplyTransactionStageRequest struct {
	ID        model.ID                  `json:"id"`
	Type      enums.CardTransactionType `json:"type"`
	Amount    *decimal.Decimal          `json:"amount"`
	RequestID *string                   `json:"request_id"`
}

func (s *Service) ApplyTransactionStage(ctx context.Context, req *ApplyTransactionStageRequest) (*SimulationData, error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	result, err := s.uc.ApplyTransactionStage(ctx, &biz.UIApplyTransactionStageRequest{
		ID:        req.ID,
		Type:      req.Type,
		Amount:    req.Amount,
		RequestID: req.RequestID,
	})
	if err != nil {
		return nil, err
	}
	return toSimulationData(result), nil
}
