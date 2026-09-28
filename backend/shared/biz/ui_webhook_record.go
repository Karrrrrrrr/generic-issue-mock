package biz

import (
	"context"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"
	sharederrors "generic-mock/shared/errors"

	"go.uber.org/zap"
)

type ListUIWebhookRecordsRequest struct {
	UIPageRequest
	UITimeRange
	ID        *model.ID
	AccountID *model.ID
	Event     *string
	Status    *enums.WebhookDeliveryStatus
}

func (req *ListUIWebhookRecordsRequest) Validate() error {
	if req == nil || !validUIIDs([]*model.ID{req.ID, req.AccountID}) || !validUIOptionalText(req.Event) ||
		(req.Status != nil && *req.Status != enums.WebhookDeliveryStatus_Pending && *req.Status != enums.WebhookDeliveryStatus_Succeeded && *req.Status != enums.WebhookDeliveryStatus_Failed) {
		return sharederrors.ErrInvalidUIRequest
	}
	if err := req.UITimeRange.Validate(); err != nil {
		return err
	}
	return req.UIPageRequest.Validate()
}

type GetUIWebhookRecordRequest struct {
	ID        model.ID
	AccountID model.ID
}

func (req *GetUIWebhookRecordRequest) Validate() error {
	if req == nil || req.ID <= 0 || req.AccountID <= 0 {
		return sharederrors.ErrInvalidUIRequest
	}
	return nil
}

type ReplayUIWebhookRecordRequest struct {
	ID        model.ID
	AccountID model.ID
}

func (req *ReplayUIWebhookRecordRequest) Validate() error {
	if req == nil || req.ID <= 0 || req.AccountID <= 0 {
		return sharederrors.ErrInvalidUIRequest
	}
	return nil
}

func (uc *ui) ListWebhookRecords(ctx context.Context, req *ListUIWebhookRecordsRequest) ([]*model.WebhookRecord, int64, error) {
	if err := req.Validate(); err != nil {
		return nil, 0, err
	}
	filters := WebhookRecordFilters{
		Channel:     uc.channel,
		AccountIDs:  types.PointerSlice(req.AccountID),
		IDs:         types.PointerSlice(req.ID),
		Events:      types.PointerSlice(req.Event),
		Statuses:    types.PointerSlice(req.Status),
		CreatedFrom: req.CreatedFrom,
		CreatedTo:   req.CreatedTo,
	}
	items, err := uc.webhookRecordRepo.List(ctx, &WebhookRecordListRequest{
		WebhookRecordFilters: filters,
		Offset:               req.Offset,
		Limit:                req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list shared UI webhook records", "error", err)
		return nil, 0, sharederrors.ErrDatabaseOperation
	}
	total, err := uc.webhookRecordRepo.Count(ctx, &WebhookRecordCountRequest{WebhookRecordFilters: filters})
	if err != nil {
		zap.S().Errorw("count shared UI webhook records", "error", err)
		return nil, 0, sharederrors.ErrDatabaseOperation
	}
	return items, total, nil
}

func (uc *ui) GetWebhookRecord(ctx context.Context, req *GetUIWebhookRecordRequest) (*model.WebhookRecord, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	exists, err := uc.webhookRecordRepo.Exist(ctx, &WebhookRecordExistRequest{
		ID:        req.ID,
		AccountID: req.AccountID,
		Channel:   uc.channel,
	})
	if err != nil {
		zap.S().Errorw("check shared UI webhook record", "error", err)
		return nil, sharederrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, sharederrors.ErrUIWebhookNotFound
	}
	item, err := uc.webhookRecordRepo.Find(ctx, &WebhookRecordFindRequest{
		ID:        req.ID,
		AccountID: req.AccountID,
		Channel:   uc.channel,
	})
	if err != nil {
		zap.S().Errorw("find shared UI webhook record", "error", err)
		return nil, sharederrors.ErrDatabaseOperation
	}
	return item, nil
}

func (uc *ui) ReplayWebhookRecord(ctx context.Context, req *ReplayUIWebhookRecordRequest) (*model.WebhookRecord, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if uc.webhookReplayer == nil {
		return nil, sharederrors.ErrUIUnsupported
	}
	record, err := uc.GetWebhookRecord(ctx, &GetUIWebhookRecordRequest{
		ID:        req.ID,
		AccountID: req.AccountID,
	})
	if err != nil {
		return nil, err
	}
	if err := uc.webhookReplayer.Replay(ctx, &UIWebhookReplayRequest{Record: record}); err != nil {
		return nil, err
	}
	return uc.GetWebhookRecord(ctx, &GetUIWebhookRecordRequest{
		ID:        req.ID,
		AccountID: req.AccountID,
	})
}
