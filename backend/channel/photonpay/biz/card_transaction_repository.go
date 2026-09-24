package biz

import (
	"context"
	"generic-mock/enums"
	"time"

	"generic-mock/model"
)

type CardTransactionExistByAccountIDRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type CardTransactionFindByAccountIDRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type ListAuthorizationStagesRequest struct {
	AccountID model.ID
	ID        model.ID
}

type CardTransactionListTransactionsRequest struct {
	AccountIDs       []model.ID
	Offset           int
	Limit            int
	IDs              []model.ID
	CreatedFrom      *time.Time
	CreatedTo        *time.Time
	Statuses         []enums.CardTransactionStatus
	CardIDs          []model.ID
	AuthorizationIDs []model.ID
	Types            []enums.CardTransactionType
}

type CardTransactionCountRequest struct {
	AccountIDs       []model.ID
	IDs              []model.ID
	CreatedFrom      *time.Time
	CreatedTo        *time.Time
	Statuses         []enums.CardTransactionStatus
	CardIDs          []model.ID
	AuthorizationIDs []model.ID
	Types            []enums.CardTransactionType
}

type CardTransactionRepository interface {
	Count(context.Context, *CardTransactionCountRequest) (int64, error)
	ListStages(context.Context, *ListAuthorizationStagesRequest) ([]*model.CardTransaction, error)
	Create(context.Context, *model.CardTransaction) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.CardTransaction, error)
	ExistByAccountID(context.Context, *CardTransactionExistByAccountIDRequest) (bool, error)
	FindByAccountID(context.Context, *CardTransactionFindByAccountIDRequest) (*model.CardTransaction, error)
	ListTransactions(context.Context, *CardTransactionListTransactionsRequest) ([]*model.CardTransaction, error)
}
