package biz

import (
	"context"
	"time"

	slash "generic-mock/channel/slash/enums"
	"generic-mock/enums"
	"generic-mock/model"

	"encoding/json"
	"github.com/samber/do"
	"go.uber.org/zap"
)

type DispatchWebhookRequest struct {
	AccountID model.ID
	Event     slash.WebhookEvent
	EntityID  string
	EventID   string
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
	result, deliveryErr := u.webhookClient.Deliver(ctx, &SlashWebhookDeliveryRequest{
		TargetURL: config.TargetURL,
		Payload:   payload,
	})
	if deliveryErr != nil {
		record.Status = enums.WebhookDeliveryStatus_Failed
		record.ErrorMessage = deliveryErr.Error()
		zap.S().Errorw("deliver slash webhook", "error", deliveryErr, "webhook_record_id", record.ID)
	} else {
		record.StatusCode = result.StatusCode
		record.ResponseBody = result.ResponseBody
		record.RequestHeaders = result.RequestHeaders
		record.ResponseHeaders = result.ResponseHeaders
		if result.StatusCode >= 200 && result.StatusCode < 300 {
			deliveredAt := time.Now().UTC()
			record.Status = enums.WebhookDeliveryStatus_Succeeded
			record.DeliveredAt = &deliveredAt
		} else {
			record.Status = enums.WebhookDeliveryStatus_Failed
			record.ErrorMessage = "unexpected Slash webhook response"
		}
	}
	if err := u.webhookRecordRepository.Save(ctx, record); err != nil {
		zap.S().Errorw("save slash webhook record", "error", err)
	}
}
