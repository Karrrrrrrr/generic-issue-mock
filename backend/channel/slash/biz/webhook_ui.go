package biz

import (
	"context"

	slash "generic-mock/channel/slash/enums"
	slasherrors "generic-mock/channel/slash/errors"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

type ListWebhooksRequest struct {
	AccountID *model.ID
}

type CreateWebhookRequest struct {
	AccountID model.ID
	Event     slash.WebhookEvent
	TargetURL string
	Enabled   bool
}

type UpdateWebhookRequest struct {
	ID        model.ID
	TargetURL string
	Enabled   bool
}

func (u *SlashUIUsecase) CreateWebhook(ctx context.Context, req *CreateWebhookRequest) (*model.WebhookConfig, error) {
	if !req.Event.Valid() {
		return nil, slasherrors.ErrInvalidOperation
	}
	exists, err := u.accountRepository.Exist(ctx, req.AccountID)
	if err != nil {
		zap.S().Errorw("check slash webhook account", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, slasherrors.ErrResourceNotFound
	}

	account, err := u.accountRepository.Find(ctx, req.AccountID)
	if err != nil {
		zap.S().Errorw("find slash webhook account", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	item := &model.WebhookConfig{
		Account:   account,
		AccountID: req.AccountID,
		Channel:   enums.Channel_Slash,
		Event:     string(req.Event),
		TargetURL: req.TargetURL,
		Enabled:   req.Enabled,
	}
	if err := u.webhookConfigRepository.Create(ctx, item); err != nil {
		zap.S().Errorw("create slash webhook", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	return item, nil
}

func (u *SlashUIUsecase) ListWebhooks(
	ctx context.Context,
	req *ListWebhooksRequest,
) ([]*model.WebhookConfig, error) {
	items, err := u.webhookConfigRepository.List(ctx, &WebhookConfigListRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
	})
	if err != nil {
		zap.S().Errorw("list slash webhooks", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	return items, nil
}

func (u *SlashUIUsecase) UpdateWebhook(ctx context.Context, req *UpdateWebhookRequest) (*model.WebhookConfig, error) {
	exists, err := u.webhookConfigRepository.ExistByID(ctx, req.ID)
	if err != nil {
		zap.S().Errorw("check slash webhook", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	if !exists {
		return nil, slasherrors.ErrResourceNotFound
	}
	item, err := u.webhookConfigRepository.FindByID(ctx, req.ID)
	if err != nil {
		zap.S().Errorw("find slash webhook", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	item.TargetURL = req.TargetURL
	item.Enabled = req.Enabled
	if err := u.webhookConfigRepository.Save(ctx, item); err != nil {
		zap.S().Errorw("update slash webhook", "error", err)
		return nil, slasherrors.ErrDatabaseOperation
	}
	return item, nil
}

func (u *SlashUIUsecase) DeleteWebhook(ctx context.Context, id model.ID) error {
	exists, err := u.webhookConfigRepository.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check slash webhook", "error", err)
		return slasherrors.ErrDatabaseOperation
	}
	if !exists {
		return slasherrors.ErrResourceNotFound
	}
	if err := u.webhookConfigRepository.Delete(ctx, id); err != nil {
		zap.S().Errorw("delete slash webhook", "error", err)
		return slasherrors.ErrDatabaseOperation
	}
	return nil
}
