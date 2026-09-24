package biz

import (
	"context"

	"generic-mock/model"
)

type CardTransactionRequestExistsRequest struct {
	AccountID model.ID
	RequestID string
}

type CardTransactionRequestFindRequest struct {
	AccountID model.ID
	RequestID string
}

type PingPongCardTransactionRepository interface {
	ExistsRequest(context.Context, *CardTransactionRequestExistsRequest) (bool, error)
	FindRequest(context.Context, *CardTransactionRequestFindRequest) (*model.CardTransaction, error)

	Create(context.Context, *model.CardTransaction) error
}
