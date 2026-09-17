package biz

import (
	"context"

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
	AccountIDs []model.ID
	Offset     int
	Limit      int
}

type CardTransactionRepository interface {
	ListStages(context.Context, *ListAuthorizationStagesRequest) ([]*model.CardTransaction, error)
	Create(context.Context, *model.CardTransaction) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.CardTransaction, error)
	ExistByAccountID(context.Context, *CardTransactionExistByAccountIDRequest) (bool, error)
	FindByAccountID(context.Context, *CardTransactionFindByAccountIDRequest) (*model.CardTransaction, error)
	ListTransactions(context.Context, *CardTransactionListTransactionsRequest) ([]*model.CardTransaction, error)
}
