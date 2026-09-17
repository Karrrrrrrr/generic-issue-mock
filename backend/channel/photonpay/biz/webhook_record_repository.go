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

type WebhookDeliveryResult struct {
	StatusCode      int
	ResponseBody    string
	RequestHeaders  []byte
	ResponseHeaders []byte
}

type PhotonPayWebhookDeliveryRequest struct {
	TargetURL      string
	Payload        []byte
	RequestHeaders []byte
	NotifyCategory string
	NotifyType     string
	PublishedAt    string
}

type WebhookRecordRepository interface {
	Create(context.Context, *model.WebhookRecord) error
	Exist(context.Context, model.ID) (bool, error)
	Find(context.Context, model.ID) (*model.WebhookRecord, error)
	Count(context.Context, *WebhookRecordCountRequest) (int64, error)
	List(context.Context, *WebhookRecordListRequest) ([]*model.WebhookRecord, error)
	Save(context.Context, *model.WebhookRecord) error
}

type WebhookClient interface {
	Deliver(context.Context, *PhotonPayWebhookDeliveryRequest) (*WebhookDeliveryResult, error)
}
