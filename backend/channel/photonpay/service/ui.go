package service

import (
	"generic-mock/channel/photonpay/biz"
	"generic-mock/enums"
	sharedservice "generic-mock/shared/service"

	"github.com/samber/do/v2"
)

type PhotonPayUIService struct {
	Shared  *sharedservice.Service
	usecase *biz.PhotonPayUIUsecase
}

func NewPhotonPayUIService(injector do.Injector) (*PhotonPayUIService, error) {
	factory := do.MustInvoke[*sharedservice.Factory](injector)
	uc := do.MustInvoke[*biz.PhotonPayUIUsecase](injector)
	notificator := do.MustInvoke[*biz.PhotonPayWebhookNotificator](injector)
	authorizationRequester := do.MustInvoke[*biz.PhotonPayAuthorizationRequester](injector)

	adapter := &sharedUIWebhookAdapter{
		notificator: notificator,
	}
	createVirtualAccount := true
	management, err := factory.New(&sharedservice.NewRequest{
		Channel:                               enums.Channel_PhotonPay,
		Notificator:                           notificator,
		AuthorizationRequester:                authorizationRequester,
		WebhookEventCatalog:                   adapter,
		WebhookReplayer:                       adapter,
		CreateVirtualAccountOnAccountCreation: &createVirtualAccount,
	})
	if err != nil {
		return nil, err
	}

	return &PhotonPayUIService{
		usecase: uc,
		Shared:  management,
	}, nil
}
