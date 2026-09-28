package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
)

type AuthorizationExistRequest struct {
	ID        model.ID
	AccountID model.ID
	Channel   enums.Channel
	CardID    model.ID
}

type AuthorizationFindByIDWithLockRequest struct {
	ID        model.ID
	AccountID model.ID
	Channel   enums.Channel
	CardID    model.ID
}

type AuthorizationRepo interface {
	Create(context.Context, *model.Authorization) error
	Exist(context.Context, *AuthorizationExistRequest) (bool, error)
	FindByIDWithLock(context.Context, *AuthorizationFindByIDWithLockRequest) (*model.Authorization, error)
}
