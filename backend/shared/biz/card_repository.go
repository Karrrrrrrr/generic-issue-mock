package biz

import (
	"context"
	"time"

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

type CardFilters struct {
	Channel     enums.Channel
	IDs         []model.ID
	AccountIDs  []model.ID
	Statuses    []enums.CardStatus
	CardNumber  *string
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

type CardListRequest struct {
	CardFilters
	Offset int
	Limit  int
}

type CardCountRequest struct {
	CardFilters
}

type CardFindRequest struct {
	ID        model.ID
	Channel   enums.Channel
	AccountID model.ID
}

type CardUpdateStatusRequest struct {
	ID        model.ID
	Channel   enums.Channel
	AccountID model.ID
	Status    enums.CardStatus
}

type CardCreateRequest struct {
	Card *model.Card
}

type CardRepo interface {
	ExistForSimulation(context.Context, *CardSimulationExistRequest) (bool, error)
	FindForSimulation(context.Context, *CardSimulationFindRequest) (*model.Card, error)
	Create(context.Context, *CardCreateRequest) error
	Exist(context.Context, *CardExistRequest) (bool, error)
	FindByIDWithLock(context.Context, *CardFindByIDWithLockRequest) (*model.Card, error)
	List(context.Context, *CardListRequest) ([]*model.Card, error)
	Count(context.Context, *CardCountRequest) (int64, error)
	Find(context.Context, *CardFindRequest) (*model.Card, error)
	UpdateStatus(context.Context, *CardUpdateStatusRequest) error
}
