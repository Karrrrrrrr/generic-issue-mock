package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
)

type AuthorizationRequest struct {
	AccountID       model.ID
	Channel         enums.Channel
	Card            *model.Card
	Authorization   *model.Authorization
	CardTransaction *model.CardTransaction
}

type AuthorizationDecision struct {
	Approved bool
	Reason   string
}

type AuthorizationRequester interface {
	RequestAuthorization(context.Context, *AuthorizationRequest) (*AuthorizationDecision, error)
}
