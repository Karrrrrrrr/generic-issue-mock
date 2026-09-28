package service

import (
	"generic-mock/channel/pingpong/biz"
	"generic-mock/enums"
	sharedservice "generic-mock/shared/service"

	"github.com/samber/do/v2"
)

type PingPongUIService struct {
	Shared  *sharedservice.Service
	webhook *biz.PingPongWebhookUsecase
}

func NewUIService(injector do.Injector) (*PingPongUIService, error) {
	factory := do.MustInvoke[*sharedservice.Factory](injector)
	s := &PingPongUIService{webhook: do.MustInvoke[*biz.PingPongWebhookUsecase](injector)}
	management, err := factory.New(&sharedservice.NewRequest{
		Channel:             enums.Channel_PingPong,
		Notificator:         s.webhook,
		WebhookEventCatalog: s.webhook,
		WebhookReplayer:     s.webhook,
	})
	if err != nil {
		return nil, err
	}
	s.Shared = management
	return s, nil
}
