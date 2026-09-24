package biz

import (
	"context"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
)

type AuthorizationListBalancesRequest struct {
	Statuses     []enums.CardTransactionStatus
	AccountIDs   []model.ID
	IDs          []model.ID
	CardIDs      []model.ID
	MerchantName *string
	CreatedFrom  *time.Time
	CreatedTo    *time.Time
}

type FindAuthorizationDetailRequest struct {
	AccountID model.ID
	ID        model.ID
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
	FindAuthorizationDetail(context.Context, *FindAuthorizationDetailRequest) (*model.Authorization, error)
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
