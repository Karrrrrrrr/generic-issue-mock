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
	notificator := do.MustInvoke[*biz.PayndaWebhookNotificator](injector)
	s := &PayndaUIService{}
	adapter := &sharedUIWebhookAdapter{notificator: notificator}
	management, err := factory.New(&sharedservice.NewRequest{
		Channel:             enums.Channel_Paynda,
		Notificator:         notificator,
		WebhookEventCatalog: adapter,
		WebhookReplayer:     adapter,
	})
	if err != nil {
		return nil, err
	}
	s.Shared = management
	return s, nil
}
