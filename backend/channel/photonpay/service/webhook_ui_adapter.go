package service

import (
	"context"

	"generic-mock/channel/photonpay/biz"
	channelenums "generic-mock/channel/photonpay/enums"
	"generic-mock/enums"
	sharedbiz "generic-mock/shared/biz"
	sharederrors "generic-mock/shared/errors"
)

type sharedUIWebhookAdapter struct {
	notificator *biz.PhotonPayWebhookNotificator
}

var _ sharedbiz.UIWebhookEventCatalog = (*sharedUIWebhookAdapter)(nil)
var _ sharedbiz.UIWebhookReplayer = (*sharedUIWebhookAdapter)(nil)

func (adapter *sharedUIWebhookAdapter) ListEvents(ctx context.Context, req *sharedbiz.UIWebhookEventsRequest) ([]string, error) {
	if req == nil || req.Channel != enums.Channel_PhotonPay {
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
	if req == nil || req.Record == nil || req.Record.Channel != enums.Channel_PhotonPay {
		return sharederrors.ErrInvalidUIRequest
	}
	_, err := adapter.notificator.ReplayWebhookRecord(ctx, req.Record.ID)
	return err
}
