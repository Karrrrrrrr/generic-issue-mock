package biz

import (
	"context"

	"generic-mock/model"
)

type WebhookRecordCountRequest struct {
	AccountIDs []model.ID
}

type WebhookRecordListRequest struct {
	AccountIDs []model.ID
	Offset     int
	Limit      int
}

type PayndaWebhookDeliveryResult struct {
	StatusCode      int
	ResponseBody    string
	RequestHeaders  []byte
	ResponseHeaders []byte
}

type PayndaWebhookDeliveryRequest struct {
	TargetURL      string
	Payload        []byte
	RequestHeaders []byte
	Category       string
}

type PayndaWebhookRecordRepository interface {
	Create(context.Context, *model.WebhookRecord) error
	Exist(context.Context, model.ID) (bool, error)
	Find(context.Context, model.ID) (*model.WebhookRecord, error)
	Count(context.Context, *WebhookRecordCountRequest) (int64, error)
	List(context.Context, *WebhookRecordListRequest) ([]*model.WebhookRecord, error)
	Save(context.Context, *model.WebhookRecord) error
}

type PayndaWebhookClient interface {
	Deliver(context.Context, *PayndaWebhookDeliveryRequest) (*PayndaWebhookDeliveryResult, error)
}
