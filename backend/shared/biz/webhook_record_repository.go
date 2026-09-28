package biz

import (
	"context"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
)

type WebhookRecordFilters struct {
	Channel     enums.Channel
	AccountIDs  []model.ID
	IDs         []model.ID
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
	Channel   enums.Channel
	ID        model.ID
}

type WebhookRecordFindRequest struct {
	AccountID model.ID
	Channel   enums.Channel
	ID        model.ID
}

type WebhookRecordCreateRequest struct {
	Record *model.WebhookRecord
}

type WebhookRecordExistBySourceRequest struct {
	AccountID       model.ID
	Channel         enums.Channel
	WebhookConfigID model.ID
	Event           string
	SourceID        string
}

type WebhookRecordUpdateDeliveryRequest struct {
	AccountID       model.ID
	Channel         enums.Channel
	ID              model.ID
	Status          enums.WebhookDeliveryStatus
	StatusCode      int
	ResponseBody    string
	ResponseHeaders []byte
	DeliveredAt     *time.Time
	ErrorMessage    string
}

type WebhookRecordRepo interface {
	Create(context.Context, *WebhookRecordCreateRequest) error
	ExistBySource(context.Context, *WebhookRecordExistBySourceRequest) (bool, error)
	UpdateDelivery(context.Context, *WebhookRecordUpdateDeliveryRequest) error
	List(context.Context, *WebhookRecordListRequest) ([]*model.WebhookRecord, error)
	Count(context.Context, *WebhookRecordCountRequest) (int64, error)
	Exist(context.Context, *WebhookRecordExistRequest) (bool, error)
	Find(context.Context, *WebhookRecordFindRequest) (*model.WebhookRecord, error)
}
