package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
)

type CardHolderExistRequest struct {
	ID        model.ID
	AccountID model.ID
	Channel   enums.Channel
}

type CardHolderRepo interface {
	Create(context.Context, *model.CardHolder) error
	Exist(context.Context, *CardHolderExistRequest) (bool, error)
}
