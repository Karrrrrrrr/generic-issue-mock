package service

import (
	"generic-mock/channel/slash/biz"
	"generic-mock/enums"
	sharedservice "generic-mock/shared/service"

	"github.com/samber/do/v2"
)

type SlashUIService struct {
	Shared         *sharedservice.Service
	usecase        *biz.SlashUIUsecase
	webhookUsecase *biz.SlashWebhookUsecase
}

func NewSlashUIService(injector do.Injector) (*SlashUIService, error) {
	factory := do.MustInvoke[*sharedservice.Factory](injector)
	s := &SlashUIService{
		usecase:        do.MustInvoke[*biz.SlashUIUsecase](injector),
		webhookUsecase: do.MustInvoke[*biz.SlashWebhookUsecase](injector),
	}
	authorizationRequester := do.MustInvoke[*biz.SlashAuthorizationRequester](injector)
	adapter := &sharedUIWebhookAdapter{webhookUsecase: s.webhookUsecase}
	management, err := factory.New(&sharedservice.NewRequest{
		Channel:                enums.Channel_Slash,
		Notificator:            s.webhookUsecase,
		AuthorizationRequester: authorizationRequester,
		WebhookEventCatalog:    adapter,
		WebhookReplayer:        adapter,
	})
	if err != nil {
		return nil, err
	}
	s.Shared = management
	return s, nil
}
