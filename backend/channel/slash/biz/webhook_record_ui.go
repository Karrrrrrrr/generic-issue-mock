package biz

import (
	"context"
	"encoding/json"
	"time"

	slash "generic-mock/channel/slash/enums"
	slasherrors "generic-mock/channel/slash/errors"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"github.com/samber/do"
	"go.uber.org/zap"
)

type DispatchWebhookRequest struct {
	AccountID model.ID
	Event     slash.WebhookEvent
	EntityID  string
	EventID   string
}

type ListWebhookRecordsRequest struct {
	AccountID   *model.ID
	Event       *slash.WebhookEvent
	Status      *enums.WebhookDeliveryStatus
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	Offset      int
	Limit       int
}

type ReplayWebhookRecordRequest struct {
	AccountID model.ID
	ID        model.ID
}

type SlashWebhookUsecase struct {
	webhookConfigRepository SlashWebhookConfigRepository
	webhookRecordRepository SlashWebhookRecordRepository
	webhookClient           SlashWebhookClient
}

func NewSlashWebhookUsecase(injector *do.Injector) (*SlashWebhookUsecase, error) {
	return &SlashWebhookUsecase{
		webhookConfigRepository: do.MustInvoke[SlashWebhookConfigRepository](injector),
		webhookRecordRepository: do.MustInvoke[SlashWebhookRecordRepository](injector),
		webhookClient:           do.MustInvoke[SlashWebhookClient](injector),
	}, nil
}

func (u *SlashWebhookUsecase) ListRecords(ctx context.Context, req *ListWebhookRecordsRequest) ([]*model.WebhookRecord, int64, error) {
	if (req.CreatedFrom != nil && req.CreatedFrom.IsZero()) ||
		(req.CreatedTo != nil && req.CreatedTo.IsZero()) ||
		(req.CreatedFrom != nil && req.CreatedTo != nil && req.CreatedFrom.After(*req.CreatedTo)) {
		return nil, 0, slasherrors.ErrInvalidOperation
	}
	filters := WebhookRecordFilters{
		AccountIDs:  types.PointerSlice(req.AccountID),
		Statuses:    types.PointerSlice(req.Status),
		CreatedFrom: req.CreatedFrom,
		CreatedTo:   req.CreatedTo,
	}
	if req.Event != nil {
		filters.Events = []string{string(*req.Event)}
	}
	items, err := u.webhookRecordRepository.List(ctx, &WebhookRecordListRequest{
		WebhookRecordFilters: filters,
		Offset:               req.Offset,
		Limit:                req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list slash webhook records", "error", err)
		return nil, 0, slasherrors.ErrDatabaseOperation
	}
	total, err := u.webhookRecordRepository.Count(ctx, &WebhookRecordCountRequest{
		WebhookRecordFilters: filters,
	})
	if err != nil {
		zap.S().Errorw("count slash webhook records", "error", err)
		return nil, 0, slasherrors.ErrDatabaseOperation
	}
	return items, total, nil
}

func (u *SlashWebhookUsecase) ReplayRecord(ctx context.Context, req *ReplayWebhookRecordRequest) (*model.WebhookRecord, error) {
	exists, err := u.webhookRecordRepository.Exist(ctx, &WebhookRecordExistRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("check slash webhook replay record", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, slasherrors.ErrResourceNotFound
	}
	original, err := u.webhookRecordRepository.Find(ctx, &WebhookRecordFindRequest{
		AccountID: req.AccountID,
		ID:        req.ID,
	})
	if err != nil {
		zap.S().Errorw("find slash webhook replay record", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	replay := &model.WebhookRecord{
		Account:         original.Account,
		AccountID:       original.AccountID,
		Channel:         original.Channel,
		WebhookConfigID: original.WebhookConfigID,
		Event:           original.Event,
		TargetURL:       original.TargetURL,
		SourceID:        original.SourceID,
		Payload:         original.Payload,
		RequestHeaders:  original.RequestHeaders,
		Status:          enums.WebhookDeliveryStatus_Pending,
		AttemptCount:    original.AttemptCount + 1,
	}
	if err := u.webhookRecordRepository.Create(ctx, replay); err != nil {
		zap.S().Errorw("create slash webhook replay record", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	startedAt := time.Now()
	zap.S().Infow("webhook delivery started",
		"channel", replay.Channel,
		"account_id", replay.AccountID,
		"webhook_record_id", replay.ID,
		"event", replay.Event,
		"source_id", replay.SourceID,
		"attempt", replay.AttemptCount,
		"method", "POST",
		"url", replay.TargetURL,
		"request_body", string(replay.Payload),
	)
	result, deliveryErr := u.webhookClient.Deliver(ctx, &SlashWebhookDeliveryRequest{
		TargetURL:      replay.TargetURL,
		Payload:        replay.Payload,
		RequestHeaders: replay.RequestHeaders,
	})
	if result != nil {
		replay.StatusCode = result.StatusCode
		replay.ResponseBody = result.ResponseBody
		replay.RequestHeaders = result.RequestHeaders
		replay.ResponseHeaders = result.ResponseHeaders
	}
	if deliveryErr != nil {
		replay.Status = enums.WebhookDeliveryStatus_Failed
		replay.ErrorMessage = deliveryErr.Error()
	} else {
		if result.StatusCode >= 200 && result.StatusCode < 300 {
			deliveredAt := time.Now().UTC()
			replay.Status = enums.WebhookDeliveryStatus_Succeeded
			replay.DeliveredAt = &deliveredAt
		} else {
			replay.Status = enums.WebhookDeliveryStatus_Failed
			replay.ErrorMessage = "unexpected Slash webhook response"
		}
	}
	logFields := []any{
		"channel", replay.Channel,
		"account_id", replay.AccountID,
		"webhook_record_id", replay.ID,
		"event", replay.Event,
		"source_id", replay.SourceID,
		"attempt", replay.AttemptCount,
		"method", "POST",
		"url", replay.TargetURL,
		"status", replay.Status,
		"status_code", replay.StatusCode,
		"request_headers", string(replay.RequestHeaders),
		"response_headers", string(replay.ResponseHeaders),
		"response_body", replay.ResponseBody,
		"duration_ms", time.Since(startedAt).Milliseconds(),
		"error", replay.ErrorMessage,
	}
	if replay.Status == enums.WebhookDeliveryStatus_Failed {
		zap.S().Errorw("webhook delivery failed", logFields...)
	} else {
		zap.S().Infow("webhook delivery succeeded", logFields...)
	}
	if err := u.webhookRecordRepository.Save(ctx, replay); err != nil {
		zap.S().Errorw("save slash webhook replay record", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	return replay, nil
}

func (u *SlashWebhookUsecase) Dispatch(ctx context.Context, req *DispatchWebhookRequest) {
	payload, err := json.Marshal(struct {
		EntityID string             `json:"entityId"`
		Event    slash.WebhookEvent `json:"event"`
		EventID  string             `json:"eventId"`
	}{
		EntityID: req.EntityID,
		Event:    req.Event,
		EventID:  req.EventID,
	})
	if err != nil {
		zap.S().Errorw("marshal slash webhook payload", "error", err)
		return
	}
	configs, err := u.webhookConfigRepository.List(ctx, &WebhookConfigListRequest{
		AccountIDs: []model.ID{req.AccountID},
	})
	if err != nil {
		zap.S().Errorw("list slash webhook configs", "error", err)
		return
	}
	for _, config := range configs {
		if !config.Enabled || config.Event != string(req.Event) {
			continue
		}
		u.dispatchToConfig(ctx, config, req, payload)
	}
}

func (u *SlashWebhookUsecase) dispatchToConfig(
	ctx context.Context,
	config *model.WebhookConfig,
	req *DispatchWebhookRequest,
	payload []byte,
) {
	record := &model.WebhookRecord{
		WebhookConfigID: config.ID,
		AccountID:       config.AccountID,
		Channel:         enums.Channel_Slash,
		Event:           string(req.Event),
		TargetURL:       config.TargetURL,
		SourceID:        req.EntityID,
		Payload:         payload,
		Status:          enums.WebhookDeliveryStatus_Pending,
		AttemptCount:    1,
	}
	if err := u.webhookRecordRepository.Create(ctx, record); err != nil {
		zap.S().Errorw("create slash webhook record", "error", err)
		return
	}
	startedAt := time.Now()
	zap.S().Infow("webhook delivery started",
		"channel", record.Channel,
		"account_id", record.AccountID,
		"webhook_record_id", record.ID,
		"event", record.Event,
		"source_id", record.SourceID,
		"attempt", record.AttemptCount,
		"method", "POST",
		"url", record.TargetURL,
		"request_body", string(record.Payload),
	)
	result, deliveryErr := u.webhookClient.Deliver(ctx, &SlashWebhookDeliveryRequest{
		TargetURL: config.TargetURL,
		Payload:   payload,
	})
	if result != nil {
		record.StatusCode = result.StatusCode
		record.ResponseBody = result.ResponseBody
		record.RequestHeaders = result.RequestHeaders
		record.ResponseHeaders = result.ResponseHeaders
	}
	if deliveryErr != nil {
		record.Status = enums.WebhookDeliveryStatus_Failed
		record.ErrorMessage = deliveryErr.Error()
	} else {
		if result.StatusCode >= 200 && result.StatusCode < 300 {
			deliveredAt := time.Now().UTC()
			record.Status = enums.WebhookDeliveryStatus_Succeeded
			record.DeliveredAt = &deliveredAt
		} else {
			record.Status = enums.WebhookDeliveryStatus_Failed
			record.ErrorMessage = "unexpected Slash webhook response"
		}
	}
	logFields := []any{
		"channel", record.Channel,
		"account_id", record.AccountID,
		"webhook_record_id", record.ID,
		"event", record.Event,
		"source_id", record.SourceID,
		"attempt", record.AttemptCount,
		"method", "POST",
		"url", record.TargetURL,
		"status", record.Status,
		"status_code", record.StatusCode,
		"request_headers", string(record.RequestHeaders),
		"response_headers", string(record.ResponseHeaders),
		"response_body", record.ResponseBody,
		"duration_ms", time.Since(startedAt).Milliseconds(),
		"error", record.ErrorMessage,
	}
	if record.Status == enums.WebhookDeliveryStatus_Failed {
		zap.S().Errorw("webhook delivery failed", logFields...)
	} else {
		zap.S().Infow("webhook delivery succeeded", logFields...)
	}
	if err := u.webhookRecordRepository.Save(ctx, record); err != nil {
		zap.S().Errorw("save slash webhook record", "error", err)
	}
}
