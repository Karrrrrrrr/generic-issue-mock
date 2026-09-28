package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
)

type CardExistRequest struct {
	ID        model.ID
	AccountID model.ID
	Channel   enums.Channel
}

type CardFindByIDWithLockRequest struct {
	ID        model.ID
	AccountID model.ID
	Channel   enums.Channel
}

type CardRepo interface {
	Create(context.Context, *model.Card) error
	Exist(context.Context, *CardExistRequest) (bool, error)
	FindByIDWithLock(context.Context, *CardFindByIDWithLockRequest) (*model.Card, error)
}
