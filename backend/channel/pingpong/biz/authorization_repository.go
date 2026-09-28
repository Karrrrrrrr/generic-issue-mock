package biz

import (
	"context"
	"time"

	common "generic-mock/enums"
	"generic-mock/model"
)

type AuthorizationExistsRequest struct {
	AccountID model.ID
	ID        model.ID
}

type AuthorizationFindRequest struct {
	AccountID model.ID
	ID        model.ID
}

type AuthorizationListRequest struct {
	IDs          []model.ID
	AccountIDs   []model.ID
	CardIDs      []model.ID
	Statuses     []common.CardTransactionStatus
	MerchantName *string
	CreatedFrom  *time.Time
	CreatedTo    *time.Time
	Offset       int
	Limit        *int
}

type AuthorizationCountRequest struct {
	IDs          []model.ID
	AccountIDs   []model.ID
	CardIDs      []model.ID
	Statuses     []common.CardTransactionStatus
	MerchantName *string
	CreatedFrom  *time.Time
	CreatedTo    *time.Time
}

type PingPongAuthorizationRepository interface {
	Exists(context.Context, *AuthorizationExistsRequest) (bool, error)
	Find(context.Context, *AuthorizationFindRequest) (*model.Authorization, error)

	Create(context.Context, *model.Authorization) error

	List(context.Context, *AuthorizationListRequest) ([]*model.Authorization, error)
	Count(context.Context, *AuthorizationCountRequest) (int64, error)
}
