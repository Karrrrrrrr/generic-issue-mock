package paynda

import (
	"generic-mock/channel/paynda/biz"
	"generic-mock/channel/paynda/data"
	"generic-mock/channel/paynda/service"

	"github.com/samber/do/v2"
)

func RegisterProviders(injector do.Injector) {
	do.Provide(injector, data.NewPayndaRepository)
	do.Provide(injector, data.NewCardHolderRepository)
	do.Provide(injector, data.NewCardProductRepository)
	do.Provide(injector, data.NewCardRepository)
	do.Provide(injector, data.NewWalletRepository)
	do.Provide(injector, data.NewAccountRepository)
	do.Provide(injector, data.NewCardTransactionRepository)
	do.Provide(injector, data.NewAuthorizationRepository)
	do.Provide(injector, data.NewWebhookConfigRepository)
	do.Provide(injector, data.NewWebhookRecordRepository)
	do.Provide(injector, data.NewWebhookClient)
	do.Provide(injector, biz.NewPayndaOpenAPIUsecase)
	do.Provide(injector, biz.NewPayndaUIUsecase)
	do.Provide(injector, service.NewPayndaOpenAPIService)
	do.Provide(injector, service.NewPayndaUIService)
}
