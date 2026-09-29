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

type CardTransactionExistsForSimulationRequest struct{ ID model.ID }
type CardTransactionFindForSimulationRequest struct{ ID model.ID }

type PingPongCardTransactionRepository interface {
	ExistsForSimulation(context.Context, *CardTransactionExistsForSimulationRequest) (bool, error)
	FindForSimulation(context.Context, *CardTransactionFindForSimulationRequest) (*model.CardTransaction, error)
	ExistsRequest(context.Context, *CardTransactionRequestExistsRequest) (bool, error)
	FindRequest(context.Context, *CardTransactionRequestFindRequest) (*model.CardTransaction, error)

	Create(context.Context, *model.CardTransaction) error
}
