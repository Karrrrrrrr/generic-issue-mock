package biz

import (
	"context"

	"generic-mock/model"
)

type WebhookConfigListRequest struct {
	AccountIDs []model.ID
}

type WebhookConfigRepository interface {
	Create(context.Context, *model.WebhookConfig) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.WebhookConfig, error)
	List(context.Context, *WebhookConfigListRequest) ([]*model.WebhookConfig, error)
	Save(context.Context, *model.WebhookConfig) error
	Delete(context.Context, model.ID) error
}
