package photonpay

import (
	"generic-mock/channel/photonpay/biz"
	"generic-mock/channel/photonpay/data"
	"generic-mock/channel/photonpay/service"

	"github.com/samber/do"
)

func RegisterProviders(injector *do.Injector) {
	do.Provide(injector, data.NewRepository)
	do.Provide(injector, data.NewTransaction)
	do.Provide(injector, data.NewCardHolderRepository)
	do.Provide(injector, data.NewCardRepository)
	do.Provide(injector, data.NewCardProductRepository)
	do.Provide(injector, data.NewAuthorizationRepository)
	do.Provide(injector, data.NewCardTransactionRepository)
	do.Provide(injector, data.NewAccountRepository)
	do.Provide(injector, data.NewWebhookConfigRepository)
	do.Provide(injector, data.NewWebhookRecordRepository)
	do.Provide(injector, data.NewWebhookClient)
	do.Provide(injector, biz.NewPhotonPayOpenAPIUsecase)
	do.Provide(injector, biz.NewPhotonPayUIUsecase)
	do.Provide(injector, service.NewPhotonPayOpenAPIService)
	do.Provide(injector, service.NewPhotonPayUIService)
}
