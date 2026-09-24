package biz

import (
	"context"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
)

type WebhookRecordFilters struct {
	AccountIDs  []model.ID
	Events      []string
	Statuses    []enums.WebhookDeliveryStatus
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

type WebhookRecordListRequest struct {
	WebhookRecordFilters
	Offset int
	Limit  int
}

type WebhookRecordCountRequest struct {
	WebhookRecordFilters
}

type WebhookRecordExistRequest struct {
	AccountID model.ID
	ID        model.ID
}

type WebhookRecordFindRequest struct {
	AccountID model.ID
	ID        model.ID
}

type SlashWebhookDeliveryRequest struct {
	TargetURL      string
	Payload        []byte
	RequestHeaders []byte
}

type SlashWebhookDeliveryResult struct {
	StatusCode      int
	ResponseBody    string
	RequestHeaders  []byte
	ResponseHeaders []byte
}

type SlashWebhookRecordRepository interface {
	List(context.Context, *WebhookRecordListRequest) ([]*model.WebhookRecord, error)
	Count(context.Context, *WebhookRecordCountRequest) (int64, error)
	Exist(context.Context, *WebhookRecordExistRequest) (bool, error)
	Find(context.Context, *WebhookRecordFindRequest) (*model.WebhookRecord, error)
	Create(context.Context, *model.WebhookRecord) error
	Save(context.Context, *model.WebhookRecord) error
}

type SlashWebhookClient interface {
	Deliver(context.Context, *SlashWebhookDeliveryRequest) (*SlashWebhookDeliveryResult, error)
}
