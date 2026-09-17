package biz

import (
	"context"

	"generic-mock/model"
)

type CardExistByIDRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type CardExistByLastOperationRequestIDRequest struct {
	AccountID model.ID
	RequestID string
}

type CardExistByRequestIDRequest struct {
	AccountID model.ID
	RequestID string
}

type CardFindByIDRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type CardFindByLastOperationRequestIDRequest struct {
	AccountID model.ID
	RequestID string
}

type CardFindByRequestIDRequest struct {
	AccountID model.ID
	RequestID string
}

type FindCardRequest struct {
	AccountID model.ID
	ID        model.ID
}

type CardListRequest struct {
	AccountIDs []model.ID
	Offset     int
	Limit      int
}

type PayndaCardRepository interface {
	FindCard(context.Context, *FindCardRequest) (*model.Card, error)
	Create(context.Context, *model.Card) error
	ExistByID(context.Context, *CardExistByIDRequest) (bool, error)
	ExistByRequestID(context.Context, *CardExistByRequestIDRequest) (bool, error)
	ExistByLastOperationRequestID(context.Context, *CardExistByLastOperationRequestIDRequest) (bool, error)
	FindByID(context.Context, *CardFindByIDRequest) (*model.Card, error)
	FindByRequestID(context.Context, *CardFindByRequestIDRequest) (*model.Card, error)
	FindByLastOperationRequestID(context.Context, *CardFindByLastOperationRequestIDRequest) (*model.Card, error)
	List(context.Context, *CardListRequest) ([]*model.Card, error)
	Save(context.Context, *model.Card) error
}
