package biz

import (
	"context"
	common "generic-mock/enums"
	"time"

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

type CardTransactionListRequest struct {
	AccountIDs       []model.ID
	IDs              []model.ID
	CardIDs          []model.ID
	AuthorizationIDs []model.ID
	Types            []common.CardTransactionType
	Statuses         []common.CardTransactionStatus
	CreatedFrom      *time.Time
	CreatedTo        *time.Time
	Offset           int
	Limit            *int
}

type CardTransactionCountRequest struct {
	AccountIDs       []model.ID
	IDs              []model.ID
	CardIDs          []model.ID
	AuthorizationIDs []model.ID
	Types            []common.CardTransactionType
	Statuses         []common.CardTransactionStatus
	CreatedFrom      *time.Time
	CreatedTo        *time.Time
}

type CardTransactionExistsForSimulationRequest struct{ ID model.ID }
type CardTransactionFindForSimulationRequest struct{ ID model.ID }

type PingPongCardTransactionRepository interface {
	List(context.Context, *CardTransactionListRequest) ([]*model.CardTransaction, error)
	Count(context.Context, *CardTransactionCountRequest) (int64, error)
	ExistsForSimulation(context.Context, *CardTransactionExistsForSimulationRequest) (bool, error)
	FindForSimulation(context.Context, *CardTransactionFindForSimulationRequest) (*model.CardTransaction, error)
	ExistsRequest(context.Context, *CardTransactionRequestExistsRequest) (bool, error)
	FindRequest(context.Context, *CardTransactionRequestFindRequest) (*model.CardTransaction, error)

	Create(context.Context, *model.CardTransaction) error
}
