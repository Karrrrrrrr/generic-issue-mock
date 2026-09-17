package biz

import (
	"context"

	photon "generic-mock/channel/photonpay/enums"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"

	"go.uber.org/zap"
)

type WebhookListRequest struct {
	AccountID *model.ID
}

type UICreateWebhookRequest struct {
	AccountID model.ID
	Event     photon.WebhookEvent
	TargetURL string
	Enabled   bool
}

type UIUpdateWebhookRequest struct {
	ID        model.ID
	TargetURL string
	Enabled   bool
}

func (u *PhotonPayUIUsecase) CreateWebhook(ctx context.Context, req *UICreateWebhookRequest) (*model.WebhookConfig, error) {
	if !req.Event.Valid() {
		return nil, ErrInvalidOperation
	}
	exists, err := u.accountRepo.Exist(ctx, req.AccountID)
	if err != nil {
		zap.S().Errorw("check photonpay webhook account", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}

	account, err := u.accountRepo.Find(ctx, req.AccountID)
	if err != nil {
		zap.S().Errorw("find photonpay webhook account", "error", err)
		return nil, ErrDatabaseOperation
	}
	item := &model.WebhookConfig{
		Account:   account,
		AccountID: req.AccountID,
		Channel:   enums.Channel_PhotonPay,
		Event:     string(req.Event),
		TargetURL: req.TargetURL,
		Enabled:   req.Enabled,
	}
	if err := u.webhookRepo.Create(ctx, item); err != nil {
		zap.S().Errorw("create photonpay UI webhook", "error", err)
		return nil, ErrDatabaseOperation
	}
	return item, nil
}

func (u *PhotonPayUIUsecase) ListWebhooks(
	ctx context.Context,
	req *WebhookListRequest,
) ([]*model.WebhookConfig, error) {
	items, err := u.webhookRepo.List(ctx, &WebhookConfigListRequest{
		AccountIDs: types.PointerSlice(req.AccountID),
	})
	if err != nil {
		zap.S().Errorw("list photonpay UI webhooks", "error", err)
		return nil, ErrDatabaseOperation
	}
	result := make([]*model.WebhookConfig, 0, len(items))
	for _, item := range items {
		if photon.WebhookEvent(item.Event).Valid() {
			result = append(result, item)
		}
	}
	return result, nil
}

func (u *PhotonPayUIUsecase) UpdateWebhook(ctx context.Context, req *UIUpdateWebhookRequest) (*model.WebhookConfig, error) {
	exists, err := u.webhookRepo.ExistByID(ctx, req.ID)
	if err != nil {
		zap.S().Errorw("check photonpay UI webhook", "error", err)
		return nil, ErrDatabaseOperation
	}
	if !exists {
		return nil, ErrResourceNotFound
	}
	item, err := u.webhookRepo.FindByID(ctx, req.ID)
	if err != nil {
		zap.S().Errorw("find photonpay UI webhook", "error", err)
		return nil, ErrDatabaseOperation
	}
	item.TargetURL, item.Enabled = req.TargetURL, req.Enabled
	if err := u.webhookRepo.Save(ctx, item); err != nil {
		zap.S().Errorw("update photonpay UI webhook", "error", err)
		return nil, ErrDatabaseOperation
	}
	return item, nil
}

func (u *PhotonPayUIUsecase) DeleteWebhook(ctx context.Context, id model.ID) error {
	exists, err := u.webhookRepo.ExistByID(ctx, id)
	if err != nil {
		zap.S().Errorw("check photonpay UI webhook", "error", err)
		return ErrDatabaseOperation
	}
	if !exists {
		return ErrResourceNotFound
	}
	if err := u.webhookRepo.Delete(ctx, id); err != nil {
		zap.S().Errorw("delete photonpay UI webhook", "error", err)
		return ErrDatabaseOperation
	}
	return nil
}
