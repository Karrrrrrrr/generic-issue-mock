package biz

import (
	"context"

	"generic-mock/model"
)

type SlashWebhookDeliveryRequest struct {
	TargetURL string
	Payload   []byte
}

type SlashWebhookDeliveryResult struct {
	StatusCode      int
	ResponseBody    string
	RequestHeaders  []byte
	ResponseHeaders []byte
}

type SlashWebhookRecordRepository interface {
	Create(context.Context, *model.WebhookRecord) error
	Save(context.Context, *model.WebhookRecord) error
}

type SlashWebhookClient interface {
	Deliver(context.Context, *SlashWebhookDeliveryRequest) (*SlashWebhookDeliveryResult, error)
}
