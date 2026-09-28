package service

import (
	"generic-mock/channel/paynda/biz"
	"generic-mock/enums"
	sharedservice "generic-mock/shared/service"

	"github.com/samber/do/v2"
)

type PayndaUIService struct {
	Shared *sharedservice.Service
}

func NewPayndaUIService(injector do.Injector) (*PayndaUIService, error) {
	factory := do.MustInvoke[*sharedservice.Factory](injector)
	uc := do.MustInvoke[*biz.PayndaUIUsecase](injector)
	s := &PayndaUIService{}
	adapter := &sharedUIWebhookAdapter{uc: uc}
	management, err := factory.New(&sharedservice.NewRequest{
		Channel:             enums.Channel_Paynda,
		Notificator:         uc,
		WebhookEventCatalog: adapter,
		WebhookReplayer:     adapter,
	})
	if err != nil {
		return nil, err
	}
	s.Shared = management
	return s, nil
}
