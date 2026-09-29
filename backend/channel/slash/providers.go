package slash

import (
	"generic-mock/channel/slash/biz"
	"generic-mock/channel/slash/data"
	"generic-mock/channel/slash/service"

	"github.com/samber/do/v2"
)

func RegisterProviders(injector do.Injector) {
	do.Provide(injector, data.NewRepository)
	do.Provide(injector, data.NewCardHolderRepository)
	do.Provide(injector, data.NewCardRepository)
	do.Provide(injector, data.NewCardProductRepository)
	do.Provide(injector, data.NewAuthorizationRepository)
	do.Provide(injector, data.NewCardTransactionRepository)
	do.Provide(injector, data.NewVirtualAccountRepository)
	do.Provide(injector, data.NewWalletRepository)
	do.Provide(injector, data.NewAccountRepository)
	do.Provide(injector, data.NewWebhookConfigRepository)
	do.Provide(injector, data.NewWebhookRecordRepository)
	do.Provide(injector, data.NewWebhookClient)
	do.Provide(injector, data.NewAuthorizationClient)
	do.Provide(injector, data.NewAuthorizationConfigRepository)
	do.Provide(injector, biz.NewSlashWebhookUsecase)
	do.Provide(injector, biz.NewSlashAuthorizationRequester)
	do.Provide(injector, biz.NewSlashUIUsecase)
	do.Provide(injector, biz.NewSlashOpenAPIUsecase)
	do.Provide(injector, service.NewSlashUIService)
	do.Provide(injector, service.NewSlashOpenAPIService)
}
