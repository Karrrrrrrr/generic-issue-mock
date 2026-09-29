package biz

import (
	"context"

	"generic-mock/model"
)

type CardExistsRequest struct {
	AccountID model.ID
	ID        model.ID
}

type CardFindRequest struct {
	AccountID model.ID
	ID        model.ID
}

type CardRequestExistsRequest struct {
	AccountID model.ID
	RequestID string
}

type CardRequestFindRequest struct {
	AccountID model.ID
	RequestID string
}

type PingPongCardRepository interface {
	ExistsRequest(context.Context, *CardRequestExistsRequest) (bool, error)
	FindRequest(context.Context, *CardRequestFindRequest) (*model.Card, error)
	Exists(context.Context, *CardExistsRequest) (bool, error)
	Find(context.Context, *CardFindRequest) (*model.Card, error)

	Create(context.Context, *model.Card) error
	Save(context.Context, *model.Card) error
}
