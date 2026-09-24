package biz

import (
	"context"

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
	IDs        []model.ID
	AccountIDs []model.ID
	CardIDs    []model.ID
	Offset     int
	Limit      *int
}

type AuthorizationCountRequest struct {
	IDs        []model.ID
	AccountIDs []model.ID
	CardIDs    []model.ID
}

type PingPongAuthorizationRepository interface {
	Exists(context.Context, *AuthorizationExistsRequest) (bool, error)
	Find(context.Context, *AuthorizationFindRequest) (*model.Authorization, error)

	Create(context.Context, *model.Authorization) error

	List(context.Context, *AuthorizationListRequest) ([]*model.Authorization, error)
	Count(context.Context, *AuthorizationCountRequest) (int64, error)
}
