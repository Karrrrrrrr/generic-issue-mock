package service

import (
	"context"

	"generic-mock/channel/slash/biz"
	channelenums "generic-mock/channel/slash/enums"
	"generic-mock/enums"
	sharedbiz "generic-mock/shared/biz"
	sharederrors "generic-mock/shared/errors"
)

type sharedUIWebhookAdapter struct{ webhookUsecase *biz.SlashWebhookUsecase }

var _ sharedbiz.UIWebhookEventCatalog = (*sharedUIWebhookAdapter)(nil)
var _ sharedbiz.UIWebhookReplayer = (*sharedUIWebhookAdapter)(nil)

func (adapter *sharedUIWebhookAdapter) ListEvents(ctx context.Context, req *sharedbiz.UIWebhookEventsRequest) ([]string, error) {
	if req == nil || req.Channel != enums.Channel_Slash {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	events := channelenums.WebhookEvents()
	result := make([]string, 0, len(events))
	for _, event := range events {
		result = append(result, string(event))
	}
	return result, nil
}

func (adapter *sharedUIWebhookAdapter) Replay(ctx context.Context, req *sharedbiz.UIWebhookReplayRequest) error {
	if req == nil || req.Record == nil || req.Record.Channel != enums.Channel_Slash {
		return sharederrors.ErrInvalidUIRequest
	}
	_, err := adapter.webhookUsecase.ReplayRecord(ctx, &biz.ReplayWebhookRecordRequest{
		AccountID: req.Record.AccountID,
		ID:        req.Record.ID,
	})
	return err
}
