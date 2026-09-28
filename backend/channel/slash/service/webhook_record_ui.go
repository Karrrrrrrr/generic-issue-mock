package service

import (
	"context"
	"time"

	"generic-mock/channel/slash/biz"
	slash "generic-mock/channel/slash/enums"
	slasherrors "generic-mock/channel/slash/errors"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"
)

type ListWebhookRecordsRequest struct {
	ListRequest
	Event       *slash.WebhookEvent           `form:"event" binding:"omitempty,oneof=aggregated_transaction.create aggregated_transaction.update card_creation.event card.update card.delete"`
	Status      *common.WebhookDeliveryStatus `form:"status" binding:"omitempty,oneof=pending succeeded failed"`
	CreatedFrom *time.Time                    `form:"created_from" time_format:"2006-01-02T15:04:05Z07:00" time_utc:"1"`
	CreatedTo   *time.Time                    `form:"created_to" time_format:"2006-01-02T15:04:05Z07:00" time_utc:"1"`
}

type ReplayWebhookRecordRequest struct {
	ID        model.ID `uri:"id" binding:"required,gt=0"`
	AccountID model.ID `json:"account_id" binding:"required,gt=0"`
}

type WebhookRecordData struct {
	ID              model.ID                     `json:"id"`
	AccountID       model.ID                     `json:"account_id"`
	AccountName     string                       `json:"account_name"`
	Event           slash.WebhookEvent           `json:"event"`
	TargetURL       string                       `json:"target_url"`
	SourceID        string                       `json:"source_id"`
	Payload         string                       `json:"payload"`
	RequestHeaders  string                       `json:"request_headers"`
	ResponseBody    string                       `json:"response_body"`
	ResponseHeaders string                       `json:"response_headers"`
	StatusCode      int                          `json:"status_code"`
	Status          common.WebhookDeliveryStatus `json:"status"`
	AttemptCount    int                          `json:"attempt_count"`
	DeliveredAt     *time.Time                   `json:"delivered_at"`
	ErrorMessage    string                       `json:"error_message"`
	CreatedAt       time.Time                    `json:"created_at"`
}

func (s *SlashUIService) ListWebhookRecords(ctx context.Context, req *ListWebhookRecordsRequest) (*ListResponse[*WebhookRecordData], error) {
	accountID := req.AccountID
	if accountID != nil && *accountID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	offset, limit := pagination(types.Value(req.PageNumber), types.Value(req.PageSize))
	items, total, err := s.webhookUsecase.ListRecords(ctx, &biz.ListWebhookRecordsRequest{
		AccountID:   accountID,
		Event:       req.Event,
		Status:      req.Status,
		CreatedFrom: req.CreatedFrom,
		CreatedTo:   req.CreatedTo,
		Offset:      offset,
		Limit:       limit,
	})
	if err != nil {
		return nil, err
	}
	return &ListResponse[*WebhookRecordData]{
		TotalItems: total,
		Data:       types.BulkConvertSlice(items, webhookRecordData),
	}, nil
}

func (s *SlashUIService) ReplayWebhookRecord(ctx context.Context, req *ReplayWebhookRecordRequest) (*WebhookRecordData, error) {
	accountID := req.AccountID
	if accountID <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	id := req.ID
	if id <= 0 {
		return nil, slasherrors.ErrInvalidOperation
	}
	item, err := s.webhookUsecase.ReplayRecord(ctx, &biz.ReplayWebhookRecordRequest{
		AccountID: accountID,
		ID:        id,
	})
	if err != nil {
		return nil, err
	}
	return webhookRecordData(item), nil
}

func webhookRecordData(item *model.WebhookRecord) *WebhookRecordData {
	return &WebhookRecordData{
		ID:              item.ID,
		AccountID:       item.AccountID,
		AccountName:     uiAccountName(item.Account),
		Event:           slash.WebhookEvent(item.Event),
		TargetURL:       item.TargetURL,
		SourceID:        item.SourceID,
		Payload:         string(item.Payload),
		RequestHeaders:  string(item.RequestHeaders),
		ResponseBody:    item.ResponseBody,
		ResponseHeaders: string(item.ResponseHeaders),
		StatusCode:      item.StatusCode,
		Status:          item.Status,
		AttemptCount:    item.AttemptCount,
		DeliveredAt:     item.DeliveredAt,
		ErrorMessage:    item.ErrorMessage,
		CreatedAt:       item.CreatedAt,
	}
}
