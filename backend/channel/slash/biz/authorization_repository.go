package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
)

type AuthorizationListBalancesRequest struct {
	AccountIDs []model.ID
}

type ExistAuthorizationRequest struct {
	AccountID model.ID
	ID        model.ID
}

type AuthorizationCountRequest struct {
	AccountIDs []model.ID
	Offset     int
	Limit      int
	IDs        []model.ID
	CardIDs    []model.ID
	Statuses   []enums.CardTransactionStatus
}

type AuthorizationListRequest struct {
	AccountIDs []model.ID
	Offset     int
	Limit      int
	IDs        []model.ID
	CardIDs    []model.ID
	Statuses   []enums.CardTransactionStatus
}

type LockAuthorizationRequest struct {
	AccountID model.ID
	ID        model.ID
}

type SlashAuthorizationRepository interface {
	ListAuthorizations(context.Context, *AuthorizationListBalancesRequest) ([]*model.Authorization, error)
	AuthorizationExists(context.Context, *ExistAuthorizationRequest) (bool, error)
	LockAuthorization(context.Context, *LockAuthorizationRequest) (*model.Authorization, error)
	Create(context.Context, *model.Authorization) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.Authorization, error)
	Count(context.Context, *AuthorizationCountRequest) (int64, error)
	List(context.Context, *AuthorizationListRequest) ([]*model.Authorization, error)
	Save(context.Context, *model.Authorization) error
}
