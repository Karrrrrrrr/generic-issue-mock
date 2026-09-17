package biz

import (
	"context"

	paynda "generic-mock/channel/paynda/enums"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

type PayndaListWebhooksRequest struct {
	AccountID *model.ID
}

type PayndaUICreateWebhookRequest struct {
	AccountID model.ID
	Event     paynda.WebhookEvent
	TargetURL string
	Enabled   bool
}

type PayndaUIUpdateWebhookRequest struct {
	ID        model.ID
	TargetURL string
	Enabled   bool
}

func (u *PayndaUIUsecase) CreateWebhook(ctx context.Context, req *PayndaUICreateWebhookRequest) (*model.WebhookConfig, error) {
	if !req.Event.Valid() {
		return nil, ErrInvalidOperation
	}
	exists, err := u.accountRepository.ExistByID(ctx, req.AccountID)
	if err != nil {
		zap.S().Errorw("check paynda webhook account", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}
	account, err := u.accountRepository.FindByID(ctx, req.AccountID)
	if err != nil {
		zap.S().Errorw("find paynda webhook account", "error", err)
		return nil, ErrDatabaseOperation
	}
	item := &model.WebhookConfig{
		Account:   account,
		Channel:   enums.Channel_Paynda,
		AccountID: req.AccountID,
		Event:     string(req.Event),
		TargetURL: req.TargetURL,
		Enabled:   req.Enabled,
	}
	if err := u.webhookConfigRepository.Create(ctx, item); err != nil {
		zap.S().Errorw("create paynda UI webhook", "error", err)
		return nil, ErrDatabaseOperation
	}
	return item, nil
}

func (u *PayndaUIUsecase) ListWebhooks(ctx context.Context, req *PayndaListWebhooksRequest) ([]*model.WebhookConfig, error) {
	items, err := u.webhookConfigRepository.ListByAccountIDs(ctx, &WebhookConfigListByAccountIDsRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
	})
	if err != nil {
		zap.S().Errorw("list paynda UI webhooks", "error", err)
		return nil, ErrDatabaseOperation
	}
	result := make([]*model.WebhookConfig, 0, len(items))
	for _, item := range items {
		if paynda.WebhookEvent(item.Event).Valid() {
			result = append(result, item)
		}
	}
	return result, nil
}

func (u *PayndaUIUsecase) UpdateWebhook(ctx context.Context, req *PayndaUIUpdateWebhookRequest) (*model.WebhookConfig, error) {
	exists, err := u.webhookConfigRepository.ExistByID(ctx, req.ID)
	if err != nil {
		zap.S().Errorw("check paynda UI webhook", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}
	item, err := u.webhookConfigRepository.FindByID(ctx, req.ID)
	if err != nil {
		zap.S().Errorw("find paynda UI webhook", "error", err)
		return nil, ErrDatabaseOperation
	}
	item.TargetURL, item.Enabled = req.TargetURL, req.Enabled
	if err := u.webhookConfigRepository.Save(ctx, item); err != nil {
		zap.S().Errorw("update paynda UI webhook", "error", err)
		return nil, ErrDatabaseOperation
	}
	return item, nil
}

func (u *PayndaUIUsecase) DeleteWebhook(ctx context.Context, id model.ID) error {
	exists, err := u.webhookConfigRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check paynda UI webhook", "error", err)
		return ErrDatabaseOperation
	}
	if !exists {
		return ErrResourceNotFound
	}
	if err := u.webhookConfigRepository.Delete(ctx, id); err != nil {
		zap.S().Errorw("delete paynda UI webhook", "error", err)
		return ErrDatabaseOperation
	}
	return nil
}
