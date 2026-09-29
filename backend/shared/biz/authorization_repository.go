package biz

import (
	"context"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
)

type AuthorizationExistForCardRequest struct {
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

type AuthorizationFilters struct {
	Channel      enums.Channel
	IDs          []model.ID
	AccountIDs   []model.ID
	CardIDs      []model.ID
	Statuses     []enums.CardTransactionStatus
	MerchantName *string
	CreatedFrom  *time.Time
	CreatedTo    *time.Time
}

type AuthorizationListRequest struct {
	AuthorizationFilters
	Offset int
	Limit  *int
}

type AuthorizationCountRequest struct {
	AuthorizationFilters
}

type AuthorizationExistRequest struct {
	ID        model.ID
	Channel   enums.Channel
	AccountID model.ID
}

type AuthorizationFindRequest struct {
	ID        model.ID
	Channel   enums.Channel
	AccountID model.ID
}

type AuthorizationCreateRequest struct {
	Authorization *model.Authorization
}

type AuthorizationRepo interface {
	ExistForSimulation(context.Context, *AuthorizationSimulationExistRequest) (bool, error)
	FindForSimulation(context.Context, *AuthorizationSimulationFindRequest) (*model.Authorization, error)
	Create(context.Context, *AuthorizationCreateRequest) error
	ExistForCard(context.Context, *AuthorizationExistForCardRequest) (bool, error)
	FindByIDWithLock(context.Context, *AuthorizationFindByIDWithLockRequest) (*model.Authorization, error)
	List(context.Context, *AuthorizationListRequest) ([]*model.Authorization, error)
	Count(context.Context, *AuthorizationCountRequest) (int64, error)
	Exist(context.Context, *AuthorizationExistRequest) (bool, error)
	Find(context.Context, *AuthorizationFindRequest) (*model.Authorization, error)
}
