package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
)

type WebhookConfigFilters struct {
	Channel    enums.Channel
	AccountIDs []model.ID
	IDs        []model.ID
	Events     []string
}

type WebhookConfigListRequest struct {
	WebhookConfigFilters
	Offset int
	Limit  int
}

type WebhookConfigCountRequest struct {
	WebhookConfigFilters
}

type WebhookConfigExistRequest struct {
	AccountID model.ID
	Channel   enums.Channel
	ID        model.ID
}

type WebhookConfigFindRequest struct {
	AccountID model.ID
	Channel   enums.Channel
	ID        model.ID
}

type WebhookConfigDeleteRequest struct {
	AccountID model.ID
	Channel   enums.Channel
	ID        model.ID
}

type WebhookConfigUpdateRequest struct {
	AccountID model.ID
	Channel   enums.Channel
	ID        model.ID
	TargetURL string
	Enabled   bool
}

type WebhookConfigCreateRequest struct {
	WebhookConfig *model.WebhookConfig
}

type WebhookConfigRepo interface {
	List(context.Context, *WebhookConfigListRequest) ([]*model.WebhookConfig, error)
	Count(context.Context, *WebhookConfigCountRequest) (int64, error)
	Exist(context.Context, *WebhookConfigExistRequest) (bool, error)
	Find(context.Context, *WebhookConfigFindRequest) (*model.WebhookConfig, error)
	Create(context.Context, *WebhookConfigCreateRequest) error
	Update(context.Context, *WebhookConfigUpdateRequest) error
	Delete(context.Context, *WebhookConfigDeleteRequest) error
}
