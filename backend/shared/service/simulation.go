package service

import (
	"context"
	"encoding/json"
	"errors"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/shared/biz"
	sharederrors "generic-mock/shared/errors"

	"github.com/shopspring/decimal"
)

type SimulationData struct {
	AuthorizationID     model.ID                 `json:"authorization_id"`
	Transaction         *CardTransactionData     `json:"transaction,omitempty"`
	Remaining           decimal.Decimal          `json:"remaining"`
	Replayed            bool                     `json:"replayed"`
	AuthorizationResult *AuthorizationResultData `json:"authorization_result,omitempty"`
	NotificationError   *string                  `json:"notification_error,omitempty"`
}

func toSimulationData(result *biz.CardTransactionSimulationResult) *SimulationData {
	transaction := toCardTransactionData(result.CardTransaction)
	data := &SimulationData{
		AuthorizationID: result.CardTransaction.AuthorizationID,
		Transaction:     &transaction,
		Remaining:       result.Remaining,
		Replayed:        result.Replayed,
	}
	if result.AuthorizationResult != nil {
		data.AuthorizationResult = toAuthorizationResultData(result.AuthorizationResult)
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

type AuthorizationResultData struct {
	Attempted       bool            `json:"attempted"`
	Approved        bool            `json:"approved"`
	FailedSide      string          `json:"failed_side"`
	Reason          string          `json:"reason"`
	Message         string          `json:"message"`
	AccountID       model.ID        `json:"account_id"`
	CardID          model.ID        `json:"card_id"`
	Amount          decimal.Decimal `json:"amount"`
	Currency        enums.Currency  `json:"currency"`
	MerchantName    string          `json:"merchant_name"`
	MerchantCountry string          `json:"merchant_country"`
	MerchantMCC     string          `json:"merchant_mcc"`
	TargetURL       string          `json:"target_url,omitempty"`
	StatusCode      int             `json:"status_code,omitempty"`
	RequestPayload  json.RawMessage `json:"request_payload,omitempty"`
	RequestHeaders  json.RawMessage `json:"request_headers,omitempty"`
	ResponseBody    string          `json:"response_body,omitempty"`
	ResponseHeaders json.RawMessage `json:"response_headers,omitempty"`
}

func toAuthorizationResultData(result *biz.AuthorizationResult) *AuthorizationResultData {
	if result == nil {
		return nil
	}
	data := &AuthorizationResultData{
		Attempted:       result.Attempted,
		Approved:        result.Approved,
		FailedSide:      string(result.FailureSide),
		Reason:          result.Reason,
		Message:         result.Message,
		AccountID:       result.AccountID,
		CardID:          result.CardID,
		Amount:          result.Amount,
		Currency:        result.Currency,
		MerchantName:    result.MerchantName,
		MerchantCountry: result.MerchantCountry,
		MerchantMCC:     result.MerchantMCC,
	}
	if result.Exchange == nil {
		return data
	}
	data.TargetURL = result.Exchange.TargetURL
	data.StatusCode = result.Exchange.StatusCode
	data.RequestPayload = jsonRawMessage(result.Exchange.RequestPayload)
	data.RequestHeaders = jsonRawMessage(result.Exchange.RequestHeaders)
	data.ResponseBody = result.Exchange.ResponseBody
	data.ResponseHeaders = jsonRawMessage(result.Exchange.ResponseHeaders)
	return data
}

func toAuthorizationFailureSimulationData(err *biz.AuthorizationFailureError) *SimulationData {
	return &SimulationData{
		AuthorizationResult: toAuthorizationResultData(err.Result),
	}
}

func jsonRawMessage(value []byte) json.RawMessage {
	if !json.Valid(value) {
		return nil
	}
	return json.RawMessage(value)
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
		var authorizationErr *biz.AuthorizationFailureError
		if errors.As(err, &authorizationErr) {
			return toAuthorizationFailureSimulationData(authorizationErr), nil
		}
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
