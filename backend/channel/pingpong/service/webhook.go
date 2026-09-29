package service

import (
	"context"

	"generic-mock/channel/pingpong/biz"
	ping "generic-mock/channel/pingpong/enums"
	pingerrors "generic-mock/channel/pingpong/errors"
	"generic-mock/model"
)

type DispatchWebhookRequest struct {
	AccountID model.ID          `json:"account_id" binding:"required,gt=0"`
	SourceID  model.ID          `json:"source_id" binding:"required,gt=0"`
	Event     ping.WebhookEvent `json:"event" binding:"required"`
}

func (s *PingPongUIService) DispatchWebhook(ctx context.Context, req *DispatchWebhookRequest) (*Empty, error) {
	if req == nil {
		return nil, pingerrors.ErrInvalid
	}
	if err := s.webhook.Dispatch(ctx, &biz.LoadWebhookSourceRequest{
		AccountID: req.AccountID,
		SourceID:  req.SourceID,
		Event:     req.Event,
	}); err != nil {
		return nil, err
	}
	return &Empty{}, nil
}
