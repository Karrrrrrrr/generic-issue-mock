package service

import (
	"context"
	"time"

	"generic-mock/model"
	"generic-mock/shared/biz"
	sharederrors "generic-mock/shared/errors"
)

type WebhookConfigData struct {
	ID          model.ID  `json:"id"`
	AccountID   model.ID  `json:"account_id"`
	AccountName string    `json:"account_name"`
	Event       string    `json:"event"`
	TargetURL   string    `json:"target_url"`
	CreatedAt   time.Time `json:"created_at"`
	Enabled     bool      `json:"enabled"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ListWebhooksRequest struct {
	PageRequest
	ID        *model.ID `form:"id" binding:"omitempty,gt=0"`
	AccountID *model.ID `form:"account_id" binding:"omitempty,gt=0"`
	Event     *string   `form:"event" binding:"omitempty,min=1"`
}

func (s *Service) ListWebhooks(ctx context.Context, req *ListWebhooksRequest) (*Page[WebhookConfigData], error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	page, err := req.PageRequest.toBizPage()
	if err != nil {
		return nil, err
	}
	items, total, err := s.uc.ListWebhooks(ctx, &biz.ListUIWebhooksRequest{
		UIPageRequest: page,
		ID:            req.ID,
		AccountID:     req.AccountID,
		Event:         req.Event,
	})
	if err != nil {
		return nil, err
	}
	result := &Page[WebhookConfigData]{
		Items: make([]WebhookConfigData, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		result.Items = append(result.Items, toWebhookConfigData(item))
	}
	return result, nil
}

func toWebhookConfigData(item *model.WebhookConfig) WebhookConfigData {
	return WebhookConfigData{
		ID:          item.ID,
		AccountID:   item.AccountID,
		AccountName: item.Account.GetName(),
		Event:       item.Event,
		TargetURL:   item.TargetURL,
		CreatedAt:   item.CreatedAt,
		Enabled:     item.Enabled,
		UpdatedAt:   item.UpdatedAt,
	}
}

type WebhookEventsData struct {
	Items []string `json:"items"`
}

func (s *Service) ListWebhookEvents(ctx context.Context, req *Empty) (*WebhookEventsData, error) {
	events, err := s.uc.ListWebhookEvents(ctx)
	if err != nil {
		return nil, err
	}
	return &WebhookEventsData{Items: append([]string{}, events...)}, nil
}

type CreateWebhookRequest struct {
	AccountID model.ID `json:"account_id" binding:"required,gt=0"`
	Event     string   `json:"event" binding:"required"`
	TargetURL string   `json:"target_url" binding:"required"`
	Enabled   *bool    `json:"enabled" binding:"required"`
}

func (s *Service) CreateWebhook(ctx context.Context, req *CreateWebhookRequest) (*WebhookConfigData, error) {
	if req == nil || req.Enabled == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	item, err := s.uc.CreateWebhook(ctx, &biz.CreateUIWebhookRequest{
		AccountID: req.AccountID,
		Event:     req.Event,
		TargetURL: req.TargetURL,
		Enabled:   *req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	result := toWebhookConfigData(item)
	return &result, nil
}

type UpdateWebhookRequest struct {
	ID        model.ID `uri:"id" binding:"required,gt=0"`
	AccountID model.ID `json:"account_id" binding:"required,gt=0"`
	TargetURL string   `json:"target_url" binding:"required"`
	Enabled   *bool    `json:"enabled" binding:"required"`
}

func (s *Service) UpdateWebhook(ctx context.Context, req *UpdateWebhookRequest) (*WebhookConfigData, error) {
	if req == nil || req.Enabled == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	item, err := s.uc.UpdateWebhook(ctx, &biz.UpdateUIWebhookRequest{
		ID:        req.ID,
		AccountID: req.AccountID,
		TargetURL: req.TargetURL,
		Enabled:   *req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	result := toWebhookConfigData(item)
	return &result, nil
}

type DeleteWebhookRequest struct {
	ID        model.ID `uri:"id" binding:"required,gt=0"`
	AccountID model.ID `form:"account_id" binding:"required,gt=0"`
}

func (s *Service) DeleteWebhook(ctx context.Context, req *DeleteWebhookRequest) (*Empty, error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	err := s.uc.DeleteWebhook(ctx, &biz.DeleteUIWebhookRequest{
		ID:        req.ID,
		AccountID: req.AccountID,
	})
	if err != nil {
		return nil, err
	}
	return &Empty{}, nil
}
