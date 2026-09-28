package pingpong

import (
	"generic-mock/channel/pingpong/biz"
	"generic-mock/channel/pingpong/data"
	"generic-mock/channel/pingpong/service"

	"github.com/samber/do/v2"
)

func RegisterProviders(injector do.Injector) {
	do.Provide(injector, data.NewRepository)
	do.Provide(injector, data.NewAccountRepository)
	do.Provide(injector, data.NewVirtualAccountRepository)
	do.Provide(injector, data.NewCardRepository)
	do.Provide(injector, data.NewProductRepository)
	do.Provide(injector, data.NewWalletRepository)
	do.Provide(injector, data.NewTransferRepository)
	do.Provide(injector, data.NewAuthorizationRepository)
	do.Provide(injector, data.NewCardTransactionRepository)
	do.Provide(injector, biz.NewOpenAPIUsecase)
	do.Provide(injector, biz.NewUIUsecase)
	do.Provide(injector, service.NewOpenAPIService)
	do.Provide(injector, service.NewUIService)
	do.Provide(injector, data.NewWebhookClient)
	do.Provide(injector, biz.NewWebhookUsecase)
}
