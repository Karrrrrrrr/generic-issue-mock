package biz

import (
	"context"

	"generic-mock/model"
)

type WebhookConfigListByAccountIDsRequest struct {
	AccountIDs []model.ID
}

type PayndaWebhookConfigRepository interface {
	Create(context.Context, *model.WebhookConfig) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.WebhookConfig, error)
	ListByAccountIDs(context.Context, *WebhookConfigListByAccountIDsRequest) ([]*model.WebhookConfig, error)
	Save(context.Context, *model.WebhookConfig) error
	Delete(context.Context, model.ID) error
}
