package shared

import (
	"generic-mock/shared/biz"
	"generic-mock/shared/data"

	"github.com/samber/do"
)

func RegisterProviders(injector *do.Injector) {
	do.Provide(injector, data.NewRepository)
	do.Provide(injector, data.NewTransaction)
	do.Provide(injector, data.NewAccountRepository)
	do.Provide(injector, data.NewCardRepository)
	do.Provide(injector, data.NewCardHolderRepository)
	do.Provide(injector, data.NewCardProductRepository)
	do.Provide(injector, data.NewVirtualAccountRepository)
	do.Provide(injector, data.NewWalletRepository)
	do.Provide(injector, biz.NewCardIssuer)
}
