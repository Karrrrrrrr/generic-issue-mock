package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
	sharederrors "generic-mock/shared/errors"

	"github.com/shopspring/decimal"
)

type AuthorizationFailureSide string

const (
	AuthorizationFailureSideMock       AuthorizationFailureSide = "mock"
	AuthorizationFailureSideThirdParty AuthorizationFailureSide = "third_party"
)

type AuthorizationExchange struct {
	TargetURL       string
	StatusCode      int
	RequestPayload  []byte
	RequestHeaders  []byte
	ResponseBody    string
	ResponseHeaders []byte
}

type AuthorizationResult struct {
	Attempted       bool
	Approved        bool
	FailureSide     AuthorizationFailureSide
	Reason          string
	Message         string
	AccountID       model.ID
	CardID          model.ID
	Amount          decimal.Decimal
	Currency        enums.Currency
	MerchantName    string
	MerchantCountry string
	MerchantMCC     string
	Exchange        *AuthorizationExchange
}

type AuthorizationRequest struct {
	AccountID       model.ID
	Channel         enums.Channel
	Card            *model.Card
	Authorization   *model.Authorization
	CardTransaction *model.CardTransaction
}

type AuthorizationDecision struct {
	Approved    bool
	FailureSide AuthorizationFailureSide
	Reason      string
	Message     string
	Exchange    *AuthorizationExchange
}

type AuthorizationRequester interface {
	RequestAuthorization(context.Context, *AuthorizationRequest) (*AuthorizationDecision, error)
}

type AuthorizationFailureError struct {
	Result *AuthorizationResult
	Cause  error
}

func NewAuthorizationFailureError(
	result *AuthorizationResult,
	cause error,
) *AuthorizationFailureError {
	if cause == nil {
		cause = sharederrors.ErrAuthorizationRequestFailed
	}
	return &AuthorizationFailureError{
		Result: result,
		Cause:  cause,
	}
}

func (e *AuthorizationFailureError) Error() string {
	if e == nil {
		return ""
	}
	if e.Result != nil && e.Result.Message != "" {
		return e.Result.Message
	}
	return e.Cause.Error()
}

func (e *AuthorizationFailureError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}
