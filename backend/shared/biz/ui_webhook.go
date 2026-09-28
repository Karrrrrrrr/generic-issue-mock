package biz

import (
	"context"
	"net/url"
	"strings"

	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"
	sharederrors "generic-mock/shared/errors"

	"go.uber.org/zap"
)

type UIWebhookEventsRequest struct {
	Channel enums.Channel
}

type UIWebhookReplayRequest struct {
	Record *model.WebhookRecord
}

type UIWebhookEventCatalog interface {
	ListEvents(context.Context, *UIWebhookEventsRequest) ([]string, error)
}

type UIWebhookReplayer interface {
	Replay(context.Context, *UIWebhookReplayRequest) error
}

type UIWebhooks interface {
	ListWebhookEvents(context.Context) ([]string, error)
	ListWebhooks(context.Context, *ListUIWebhooksRequest) ([]*model.WebhookConfig, int64, error)
	CreateWebhook(context.Context, *CreateUIWebhookRequest) (*model.WebhookConfig, error)
	UpdateWebhook(context.Context, *UpdateUIWebhookRequest) (*model.WebhookConfig, error)
	DeleteWebhook(context.Context, *DeleteUIWebhookRequest) error
	ListWebhookRecords(context.Context, *ListUIWebhookRecordsRequest) ([]*model.WebhookRecord, int64, error)
	GetWebhookRecord(context.Context, *GetUIWebhookRecordRequest) (*model.WebhookRecord, error)
	ReplayWebhookRecord(context.Context, *ReplayUIWebhookRecordRequest) (*model.WebhookRecord, error)
}

type ListUIWebhooksRequest struct {
	UIPageRequest
	ID        *model.ID
	AccountID *model.ID
	Event     *string
}

func (req *ListUIWebhooksRequest) Validate() error {
	if req == nil || !validUIIDs([]*model.ID{req.ID, req.AccountID}) || !validUIOptionalText(req.Event) {
		return sharederrors.ErrInvalidUIRequest
	}
	return req.UIPageRequest.Validate()
}

type CreateUIWebhookRequest struct {
	AccountID model.ID
	Event     string
	TargetURL string
	Enabled   bool
}

func validUIWebhookURL(value string) bool {
	target, err := url.ParseRequestURI(value)
	return err == nil && (target.Scheme == "http" || target.Scheme == "https") && target.Host != "" && target.User == nil
}

func (req *CreateUIWebhookRequest) Validate() error {
	if req == nil || req.AccountID <= 0 || strings.TrimSpace(req.Event) == "" || !validUIWebhookURL(req.TargetURL) {
		return sharederrors.ErrInvalidUIRequest
	}
	return nil
}

type UpdateUIWebhookRequest struct {
	ID        model.ID
	AccountID model.ID
	TargetURL string
	Enabled   bool
}

func (req *UpdateUIWebhookRequest) Validate() error {
	if req == nil || req.ID <= 0 || req.AccountID <= 0 || !validUIWebhookURL(req.TargetURL) {
		return sharederrors.ErrInvalidUIRequest
	}
	return nil
}

type DeleteUIWebhookRequest struct {
	ID        model.ID
	AccountID model.ID
}

func (req *DeleteUIWebhookRequest) Validate() error {
	if req == nil || req.ID <= 0 || req.AccountID <= 0 {
		return sharederrors.ErrInvalidUIRequest
	}
	return nil
}

func (uc *ui) ListWebhookEvents(ctx context.Context) ([]string, error) {
	if uc.webhookEventCatalog == nil {
		return nil, sharederrors.ErrUIUnsupported
	}
	return uc.webhookEventCatalog.ListEvents(ctx, &UIWebhookEventsRequest{Channel: uc.channel})
}

func (uc *ui) ListWebhooks(ctx context.Context, req *ListUIWebhooksRequest) ([]*model.WebhookConfig, int64, error) {
	if err := req.Validate(); err != nil {
		return nil, 0, err
	}
	filters := WebhookConfigFilters{
		Channel:    uc.channel,
		AccountIDs: types.PointerSlice(req.AccountID),
		IDs:        types.PointerSlice(req.ID),
		Events:     types.PointerSlice(req.Event),
	}
	items, err := uc.webhookConfigRepo.List(ctx, &WebhookConfigListRequest{
		WebhookConfigFilters: filters,
		Offset:               req.Offset,
		Limit:                req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list shared UI webhook configs", "error", err)
		return nil, 0, sharederrors.ErrDatabaseOperation
	}
	total, err := uc.webhookConfigRepo.Count(ctx, &WebhookConfigCountRequest{WebhookConfigFilters: filters})
	if err != nil {
		zap.S().Errorw("count shared UI webhook configs", "error", err)
		return nil, 0, sharederrors.ErrDatabaseOperation
	}
	return items, total, nil
}

func (uc *ui) CreateWebhook(ctx context.Context, req *CreateUIWebhookRequest) (*model.WebhookConfig, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	events, err := uc.ListWebhookEvents(ctx)
	if err != nil {
		return nil, err
	}
	validEvent := false
	for _, event := range events {
		if req.Event == event {
			validEvent = true
			break
		}
	}
	if !validEvent {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	item := &model.WebhookConfig{
		AccountID: req.AccountID,
		Channel:   uc.channel,
		Event:     req.Event,
		TargetURL: req.TargetURL,
		Enabled:   req.Enabled,
	}
	err = uc.tx.InTx(ctx, func(ctx context.Context) error {
		account, err := uc.lockUIAccount(ctx, req.AccountID)
		if err != nil {
			return err
		}
		item.Account = account
		if err := uc.webhookConfigRepo.Create(ctx, &WebhookConfigCreateRequest{WebhookConfig: item}); err != nil {
			zap.S().Errorw("create shared UI webhook config", "error", err)
			return sharederrors.ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (uc *ui) UpdateWebhook(ctx context.Context, req *UpdateUIWebhookRequest) (*model.WebhookConfig, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var item *model.WebhookConfig
	err := uc.tx.InTx(ctx, func(ctx context.Context) error {
		if _, err := uc.lockUIAccount(ctx, req.AccountID); err != nil {
			return err
		}
		exists, err := uc.webhookConfigRepo.Exist(ctx, &WebhookConfigExistRequest{
			ID:        req.ID,
			AccountID: req.AccountID,
			Channel:   uc.channel,
		})
		if err != nil {
			zap.S().Errorw("check shared UI webhook update", "error", err)
			return sharederrors.ErrDatabaseOperation
		}
		if !exists {
			return sharederrors.ErrUIWebhookNotFound
		}
		if err := uc.webhookConfigRepo.Update(ctx, &WebhookConfigUpdateRequest{
			ID:        req.ID,
			AccountID: req.AccountID,
			Channel:   uc.channel,
			TargetURL: req.TargetURL,
			Enabled:   req.Enabled,
		}); err != nil {
			zap.S().Errorw("update shared UI webhook", "error", err)
			return sharederrors.ErrDatabaseOperation
		}
		item, err = uc.webhookConfigRepo.Find(ctx, &WebhookConfigFindRequest{
			ID:        req.ID,
			AccountID: req.AccountID,
			Channel:   uc.channel,
		})
		if err != nil {
			zap.S().Errorw("find updated shared UI webhook", "error", err)
			return sharederrors.ErrDatabaseOperation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (uc *ui) DeleteWebhook(ctx context.Context, req *DeleteUIWebhookRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}
	return uc.tx.InTx(ctx, func(ctx context.Context) error {
		if _, err := uc.lockUIAccount(ctx, req.AccountID); err != nil {
			return err
		}
		exists, err := uc.webhookConfigRepo.Exist(ctx, &WebhookConfigExistRequest{
			ID:        req.ID,
			AccountID: req.AccountID,
			Channel:   uc.channel,
		})
		if err != nil {
			zap.S().Errorw("check shared UI webhook deletion", "error", err)
			return sharederrors.ErrDatabaseOperation
		}
		if !exists {
			return sharederrors.ErrUIWebhookNotFound
		}
		if err := uc.webhookConfigRepo.Delete(ctx, &WebhookConfigDeleteRequest{
			ID:        req.ID,
			AccountID: req.AccountID,
			Channel:   uc.channel,
		}); err != nil {
			zap.S().Errorw("delete shared UI webhook", "error", err)
			return sharederrors.ErrDatabaseOperation
		}
		return nil
	})
}
