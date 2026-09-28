package biz

import (
	"context"

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

type CardTransactionRepo interface {
	Create(context.Context, *model.CardTransaction) error
	ExistByRequestID(context.Context, *CardTransactionExistByRequestIDRequest) (bool, error)
	FindByRequestID(context.Context, *CardTransactionFindByRequestIDRequest) (*model.CardTransaction, error)
	ListStages(context.Context, *CardTransactionListStagesRequest) ([]*model.CardTransaction, error)
}
