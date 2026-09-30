package biz

import (
	"context"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
)

type CardTransactionExistByRequestIDRequest struct {
	AccountID model.ID
	Channel   enums.Channel
	RequestID string
}

type CardTransactionFindByRequestIDRequest struct {
	AccountID model.ID
	Channel   enums.Channel
	RequestID string
}

type CardTransactionListStagesRequest struct {
	AccountID       model.ID
	Channel         enums.Channel
	CardID          model.ID
	AuthorizationID model.ID
}

type CardTransactionFilters struct {
	Channel           enums.Channel
	IDs               []model.ID
	AccountIDs        []model.ID
	CardIDs           []model.ID
	VirtualAccountIDs []model.ID
	AuthorizationIDs  []model.ID
	RequestIDs        []string
	Statuses          []enums.CardTransactionStatus
	Types             []enums.CardTransactionType
	CreatedFrom       *time.Time
	CreatedTo         *time.Time
}

type CardTransactionListRequest struct {
	CardTransactionFilters
	Offset int
	Limit  *int
}

type CardTransactionCountRequest struct {
	CardTransactionFilters
}

type CardTransactionExistRequest struct {
	ID        model.ID
	Channel   enums.Channel
	AccountID model.ID
}

type CardTransactionFindRequest struct {
	ID        model.ID
	Channel   enums.Channel
	AccountID model.ID
}

type CardTransactionExistForSimulationRequest struct {
	ID      model.ID
	Channel enums.Channel
}

type CardTransactionFindForSimulationRequest struct {
	ID      model.ID
	Channel enums.Channel
}

type CardTransactionCreateRequest struct {
	CardTransaction *model.CardTransaction
}

type CardTransactionSaveRequest struct {
	CardTransaction *model.CardTransaction
}

type CardTransactionRepo interface {
	Create(context.Context, *CardTransactionCreateRequest) error
	Save(context.Context, *CardTransactionSaveRequest) error
	ExistByRequestID(context.Context, *CardTransactionExistByRequestIDRequest) (bool, error)
	FindByRequestID(context.Context, *CardTransactionFindByRequestIDRequest) (*model.CardTransaction, error)
	ListStages(context.Context, *CardTransactionListStagesRequest) ([]*model.CardTransaction, error)
	List(context.Context, *CardTransactionListRequest) ([]*model.CardTransaction, error)
	Count(context.Context, *CardTransactionCountRequest) (int64, error)
	Exist(context.Context, *CardTransactionExistRequest) (bool, error)
	Find(context.Context, *CardTransactionFindRequest) (*model.CardTransaction, error)
	ExistForSimulation(context.Context, *CardTransactionExistForSimulationRequest) (bool, error)
	FindForSimulation(context.Context, *CardTransactionFindForSimulationRequest) (*model.CardTransaction, error)
}
