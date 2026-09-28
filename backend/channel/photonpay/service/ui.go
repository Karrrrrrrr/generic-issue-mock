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
	s := &PhotonPayUIService{usecase: uc}
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
	s.Shared = management
	return s, nil
}
