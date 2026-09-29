package photonpay

import (
	"generic-mock/channel/photonpay/biz"
	"generic-mock/channel/photonpay/data"
	"generic-mock/channel/photonpay/service"

	"github.com/samber/do/v2"
)

func RegisterProviders(injector do.Injector) {
	do.Provide(injector, data.NewRepository)
	do.Provide(injector, data.NewCardHolderRepository)
	do.Provide(injector, data.NewCardRepository)
	do.Provide(injector, data.NewCardProductRepository)
	do.Provide(injector, data.NewAuthorizationRepository)
	do.Provide(injector, data.NewCardTransactionRepository)
	do.Provide(injector, data.NewAccountRepository)
	do.Provide(injector, data.NewWalletRepository)
	do.Provide(injector, data.NewVirtualAccountRepository)
	do.Provide(injector, data.NewWebhookConfigRepository)
	do.Provide(injector, data.NewAuthorizationConfigRepository)
	do.Provide(injector, data.NewWebhookRecordRepository)
	do.Provide(injector, data.NewWebhookClient)
	do.Provide(injector, biz.NewPhotonPayOpenAPIUsecase)
	do.Provide(injector, biz.NewPhotonPayUIUsecase)
	do.Provide(injector, biz.NewPhotonPayWebhookNotificator)
	do.Provide(injector, service.NewPhotonPayOpenAPIService)
	do.Provide(injector, service.NewPhotonPayUIService)
}
