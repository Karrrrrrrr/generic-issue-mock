package service

import (
	"context"
	"time"

	common "generic-mock/enums"

	"generic-mock/channel/slash/biz"
	slash "generic-mock/channel/slash/enums"
	"generic-mock/channel/slash/pkg/idconv"
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
	ID        string `uri:"id" binding:"required"`
	AccountID string `json:"account_id" binding:"required"`
}

type WebhookRecordData struct {
	ID              string                       `json:"id"`
	AccountID       string                       `json:"account_id"`
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
	accountID, err := idconv.FromOptionalUUID(req.AccountID)
	if err != nil {
		return nil, err
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
	accountID, err := idconv.FromAccountUUID(req.AccountID)
	if err != nil {
		return nil, err
	}
	id, err := idconv.FromUUID(req.ID)
	if err != nil {
		return nil, err
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
		ID:              idconv.ToUUID(item.ID),
		AccountID:       idconv.ToUUID(item.AccountID),
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
