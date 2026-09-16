package slash

import (
	"generic-mock/channel/slash/biz"
	"generic-mock/channel/slash/data"
	"generic-mock/channel/slash/service"

	"github.com/samber/do"
)

func RegisterProviders(injector *do.Injector) {
	do.Provide(injector, data.NewRepository)
	do.ProvideNamed(injector, "slash.transaction", data.NewTransaction)
	do.ProvideNamed(injector, "slash.card-holder-repository", data.NewCardHolderRepository)
	do.ProvideNamed(injector, "slash.card-repository", data.NewCardRepository)
	do.ProvideNamed(injector, "slash.card-product-repository", data.NewCardProductRepository)
	do.ProvideNamed(injector, "slash.authorization-repository", data.NewAuthorizationRepository)
	do.ProvideNamed(injector, "slash.card-transaction-repository", data.NewCardTransactionRepository)
	do.ProvideNamed(injector, "slash.usecase", biz.NewUsecase)
	do.ProvideNamed(injector, "slash.openapi-usecase", biz.NewOpenAPIUsecase)
	do.ProvideNamed(injector, "slash.service", service.NewService)
	do.ProvideNamed(injector, "slash.openapi-service", service.NewOpenAPIService)
}
