package service

import (
	"context"
	"encoding/json"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/shared/biz"
	sharederrors "generic-mock/shared/errors"
)

type WebhookRecordData struct {
	ID              model.ID                    `json:"id"`
	AccountID       model.ID                    `json:"account_id"`
	AccountName     string                      `json:"account_name"`
	Event           string                      `json:"event"`
	TargetURL       string                      `json:"target_url"`
	CreatedAt       time.Time                   `json:"created_at"`
	WebhookConfigID model.ID                    `json:"webhook_config_id"`
	SourceID        string                      `json:"source_id"`
	Payload         json.RawMessage             `json:"payload"`
	RequestHeaders  json.RawMessage             `json:"request_headers"`
	ResponseHeaders json.RawMessage             `json:"response_headers"`
	ResponseBody    string                      `json:"response_body"`
	StatusCode      int                         `json:"status_code"`
	Status          enums.WebhookDeliveryStatus `json:"status"`
	AttemptCount    int                         `json:"attempt_count"`
	DeliveredAt     *time.Time                  `json:"delivered_at"`
	ErrorMessage    string                      `json:"error_message"`
}

type ListWebhookRecordsRequest struct {
	PageRequest
	TimeRange
	Status    *enums.WebhookDeliveryStatus `form:"status"`
	ID        *model.ID                    `form:"id" binding:"omitempty,gt=0"`
	AccountID *model.ID                    `form:"account_id" binding:"omitempty,gt=0"`
	Event     *string                      `form:"event" binding:"omitempty,min=1"`
}

func (s *Service) ListWebhookRecords(ctx context.Context, req *ListWebhookRecordsRequest) (*Page[WebhookRecordData], error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	page, err := req.PageRequest.toBizPage()
	if err != nil {
		return nil, err
	}
	items, total, err := s.uc.ListWebhookRecords(ctx, &biz.ListUIWebhookRecordsRequest{
		UIPageRequest: page,
		ID:            req.ID,
		AccountID:     req.AccountID,
		Event:         req.Event,
		Status:        req.Status,
		UITimeRange: biz.UITimeRange{
			CreatedFrom: req.CreatedFrom,
			CreatedTo:   req.CreatedTo,
		},
	})
	if err != nil {
		return nil, err
	}
	result := &Page[WebhookRecordData]{
		Items: make([]WebhookRecordData, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		result.Items = append(result.Items, toWebhookRecordData(item))
	}
	return result, nil
}

func toWebhookRecordData(item *model.WebhookRecord) WebhookRecordData {
	return WebhookRecordData{
		ID:              item.ID,
		AccountID:       item.AccountID,
		AccountName:     item.Account.GetName(),
		Event:           item.Event,
		TargetURL:       item.TargetURL,
		CreatedAt:       item.CreatedAt,
		WebhookConfigID: item.WebhookConfigID,
		SourceID:        item.SourceID,
		Payload:         json.RawMessage(item.Payload),
		RequestHeaders:  json.RawMessage(item.RequestHeaders),
		ResponseHeaders: json.RawMessage(item.ResponseHeaders),
		ResponseBody:    item.ResponseBody,
		StatusCode:      item.StatusCode,
		Status:          item.Status,
		AttemptCount:    item.AttemptCount,
		DeliveredAt:     item.DeliveredAt,
		ErrorMessage:    item.ErrorMessage,
	}
}

type GetWebhookRecordRequest struct {
	ID        model.ID `form:"id" binding:"required,gt=0"`
	AccountID model.ID `form:"account_id" binding:"required,gt=0"`
}

func (s *Service) GetWebhookRecord(ctx context.Context, req *GetWebhookRecordRequest) (*WebhookRecordData, error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	item, err := s.uc.GetWebhookRecord(ctx, &biz.GetUIWebhookRecordRequest{
		ID:        req.ID,
		AccountID: req.AccountID,
	})
	if err != nil {
		return nil, err
	}
	result := toWebhookRecordData(item)
	return &result, nil
}

type ReplayWebhookRecordRequest struct {
	ID        model.ID `json:"id" binding:"required,gt=0"`
	AccountID model.ID `json:"account_id" binding:"required,gt=0"`
}

func (s *Service) ReplayWebhookRecord(ctx context.Context, req *ReplayWebhookRecordRequest) (*WebhookRecordData, error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	item, err := s.uc.ReplayWebhookRecord(ctx, &biz.ReplayUIWebhookRecordRequest{
		ID:        req.ID,
		AccountID: req.AccountID,
	})
	if err != nil {
		return nil, err
	}
	result := toWebhookRecordData(item)
	return &result, nil
}
