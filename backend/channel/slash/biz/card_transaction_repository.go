package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
)

type CardTransactionCountRequest struct {
	AccountIDs       []model.ID
	Offset           int
	Limit            int
	IDs              []model.ID
	CardIDs          []model.ID
	AuthorizationIDs []model.ID
	Types            []enums.CardTransactionType
	Statuses         []enums.CardTransactionStatus
}

type CardTransactionExistByAccountIDRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type CardTransactionFindByAccountIDRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type CardTransactionListRequest struct {
	AccountIDs       []model.ID
	Offset           int
	Limit            int
	IDs              []model.ID
	CardIDs          []model.ID
	AuthorizationIDs []model.ID
	Types            []enums.CardTransactionType
	Statuses         []enums.CardTransactionStatus
}

type ListAuthorizationStagesRequest struct {
	AccountID model.ID
	ID        model.ID
}

type SlashCardTransactionRepository interface {
	ExistByAccountID(context.Context, *CardTransactionExistByAccountIDRequest) (bool, error)
	ListStages(context.Context, *ListAuthorizationStagesRequest) ([]*model.CardTransaction, error)
	Create(context.Context, *model.CardTransaction) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.CardTransaction, error)
	Count(context.Context, *CardTransactionCountRequest) (int64, error)
	List(context.Context, *CardTransactionListRequest) ([]*model.CardTransaction, error)
	FindByAccountID(context.Context, *CardTransactionFindByAccountIDRequest) (*model.CardTransaction, error)
}
