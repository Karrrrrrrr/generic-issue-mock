package biz

import (
	"context"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
)

type CardTransactionExistByAccountIDRequest struct {
	AccountID model.ID
	ID        model.ID
}

type CardTransactionExistByRequestIDRequest struct {
	AccountID model.ID
	RequestID string
}

type CardTransactionFindByAccountIDRequest struct {
	AccountID model.ID
	ID        model.ID
}

type CardTransactionFindByRequestIDRequest struct {
	AccountID model.ID
	RequestID string
}

type CardTransactionListRequest struct {
	AccountIDs     []model.ID
	Offset         int
	Limit          int
	CardIDs        []model.ID
	StartCreatedAt *time.Time
	EndCreatedAt   *time.Time
	Types          []enums.CardTransactionType
}

type ListAuthorizationStagesRequest struct {
	AccountID model.ID
	ID        model.ID
}

type PayndaCardTransactionRepository interface {
	ListStages(context.Context, *ListAuthorizationStagesRequest) ([]*model.CardTransaction, error)
	Create(context.Context, *model.CardTransaction) error
	Save(context.Context, *model.CardTransaction) error
	ExistByID(context.Context, model.ID) (bool, error)
	ExistByRequestID(context.Context, *CardTransactionExistByRequestIDRequest) (bool, error)
	FindByID(context.Context, model.ID) (*model.CardTransaction, error)
	ExistByAccountID(context.Context, *CardTransactionExistByAccountIDRequest) (bool, error)
	FindByAccountID(context.Context, *CardTransactionFindByAccountIDRequest) (*model.CardTransaction, error)
	FindByRequestID(context.Context, *CardTransactionFindByRequestIDRequest) (*model.CardTransaction, error)
	List(context.Context, *CardTransactionListRequest) ([]*model.CardTransaction, error)
}
