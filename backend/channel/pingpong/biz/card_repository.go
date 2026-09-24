package biz

import (
	"context"

	common "generic-mock/enums"
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

type CardListRequest struct {
	IDs        []model.ID
	AccountIDs []model.ID
	RequestIDs []string
	Statuses   []common.CardStatus
	Offset     int
	Limit      *int
}

type CardCountRequest struct {
	IDs        []model.ID
	AccountIDs []model.ID
	RequestIDs []string
	Statuses   []common.CardStatus
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
	List(context.Context, *CardListRequest) ([]*model.Card, error)
	Count(context.Context, *CardCountRequest) (int64, error)
}
