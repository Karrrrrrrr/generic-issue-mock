package service

import (
	"context"

	"generic-mock/channel/paynda/biz"
	channelenums "generic-mock/channel/paynda/enums"
	"generic-mock/enums"
	sharedbiz "generic-mock/shared/biz"
	sharederrors "generic-mock/shared/errors"
)

type sharedUIWebhookAdapter struct{ uc *biz.PayndaUIUsecase }

var _ sharedbiz.UIWebhookEventCatalog = (*sharedUIWebhookAdapter)(nil)
var _ sharedbiz.UIWebhookReplayer = (*sharedUIWebhookAdapter)(nil)

func (adapter *sharedUIWebhookAdapter) ListEvents(ctx context.Context, req *sharedbiz.UIWebhookEventsRequest) ([]string, error) {
	if req == nil || req.Channel != enums.Channel_Paynda {
		return nil, sharederrors.ErrInvalidUIRequest
	}
	webhookTypes := channelenums.WebhookTypes()
	result := make([]string, 0, len(webhookTypes))
	for _, webhookType := range webhookTypes {
		result = append(result, string(webhookType))
	}
	return result, nil
}

func (adapter *sharedUIWebhookAdapter) Replay(ctx context.Context, req *sharedbiz.UIWebhookReplayRequest) error {
	if req == nil || req.Record == nil || req.Record.Channel != enums.Channel_Paynda {
		return sharederrors.ErrInvalidUIRequest
	}
	_, err := adapter.uc.ReplayWebhookRecord(ctx, req.Record.ID)
	return err
}
