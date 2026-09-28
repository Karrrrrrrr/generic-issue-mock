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

type CardSimulationExistRequest struct {
	ID      model.ID
	Channel enums.Channel
}

type CardSimulationFindRequest struct {
	ID      model.ID
	Channel enums.Channel
}

type CardRepo interface {
	ExistForSimulation(context.Context, *CardSimulationExistRequest) (bool, error)
	FindForSimulation(context.Context, *CardSimulationFindRequest) (*model.Card, error)
	Create(context.Context, *model.Card) error
	Exist(context.Context, *CardExistRequest) (bool, error)
	FindByIDWithLock(context.Context, *CardFindByIDWithLockRequest) (*model.Card, error)
}
