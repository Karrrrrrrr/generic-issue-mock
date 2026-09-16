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
	do.Provide(injector, data.NewCardTransactionRepository)
	do.Provide(injector, biz.NewUsecase)
	do.ProvideNamed(injector, "photonpay.ui-usecase", biz.NewUIUsecase)
	do.Provide(injector, service.NewService)
	do.ProvideNamed(injector, "photonpay.ui-service", service.NewUIService)
}
