package service

import (
	"context"

	"generic-mock/channel/photonpay/biz"
	channelenums "generic-mock/channel/photonpay/enums"
	"generic-mock/enums"
	sharedbiz "generic-mock/shared/biz"
	sharederrors "generic-mock/shared/errors"
	sharedservice "generic-mock/shared/service"

	"github.com/samber/do/v2"
)

type SharedUIService struct{ *sharedservice.Service }

func NewSharedUIService(injector do.Injector) (*SharedUIService, error) {
	factory := do.MustInvoke[*sharedservice.Factory](injector)
	uc := do.MustInvoke[*biz.PhotonPayUIUsecase](injector)
	adapter := &sharedUIWebhookAdapter{uc: uc}
	createVirtualAccount := true
	management, err := factory.New(&sharedservice.NewRequest{
		Channel:                               enums.Channel_PhotonPay,
		Notificator:                           uc,
		WebhookEventCatalog:                   adapter,
		WebhookReplayer:                       adapter,
		CreateVirtualAccountOnAccountCreation: &createVirtualAccount,
	})
	if err != nil {
		return nil, err
	}
	return &SharedUIService{Service: management}, nil
}

type sharedUIWebhookAdapter struct{ uc *biz.PhotonPayUIUsecase }

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
	_, err := adapter.uc.ReplayWebhookRecord(ctx, req.Record.ID)
	return err
}
