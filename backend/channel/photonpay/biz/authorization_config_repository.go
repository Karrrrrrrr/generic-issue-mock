package biz

import (
	"context"

	"generic-mock/model"
)

type AuthorizationConfigRepository interface {
	Create(context.Context, *model.AuthorizationConfig) error
	ExistByAccountID(context.Context, model.ID) (bool, error)
	FindByAccountID(context.Context, model.ID) (*model.AuthorizationConfig, error)
	Save(context.Context, *model.AuthorizationConfig) error
}
