package biz

import (
	"context"
	"time"

	"generic-mock/model"
)

type AuthorizationListBalancesRequest struct {
	AccountIDs   []model.ID
	IDs          []model.ID
	CardIDs      []model.ID
	MerchantName *string
	CreatedFrom  *time.Time
	CreatedTo    *time.Time
}

type ExistAuthorizationRequest struct {
	AccountID model.ID
	ID        model.ID
}

type AuthorizationListRequest struct {
	AccountIDs []model.ID
	Offset     int
	Limit      int
}

type LockAuthorizationRequest struct {
	AccountID model.ID
	ID        model.ID
}

type AuthorizationRepository interface {
	ListAuthorizations(context.Context, *AuthorizationListBalancesRequest) ([]*model.Authorization, error)
	AuthorizationExists(context.Context, *ExistAuthorizationRequest) (bool, error)
	LockAuthorization(context.Context, *LockAuthorizationRequest) (*model.Authorization, error)
	Create(context.Context, *model.Authorization) error
	List(context.Context, *AuthorizationListRequest) ([]*model.Authorization, error)
}
