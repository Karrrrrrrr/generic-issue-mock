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

type AuthorizationSimulationExistRequest struct {
	ID      model.ID
	Channel enums.Channel
}

type AuthorizationSimulationFindRequest struct {
	ID      model.ID
	Channel enums.Channel
}

type AuthorizationRepo interface {
	ExistForSimulation(context.Context, *AuthorizationSimulationExistRequest) (bool, error)
	FindForSimulation(context.Context, *AuthorizationSimulationFindRequest) (*model.Authorization, error)
	Create(context.Context, *model.Authorization) error
	Exist(context.Context, *AuthorizationExistRequest) (bool, error)
	FindByIDWithLock(context.Context, *AuthorizationFindByIDWithLockRequest) (*model.Authorization, error)
}
