package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
)

type AccountExistRequest struct {
	ID      model.ID
	Channel enums.Channel
}

type AccountRepo interface {
	Exist(context.Context, *AccountExistRequest) (bool, error)
}
