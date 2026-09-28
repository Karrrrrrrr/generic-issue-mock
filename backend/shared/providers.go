package shared

import (
	"generic-mock/shared/biz"
	"generic-mock/shared/data"

	"github.com/samber/do/v2"
)

func RegisterProviders(injector do.Injector) {
	do.Provide(injector, data.NewRepository)
	do.Provide(injector, data.NewTransaction)
	do.Provide(injector, data.NewAccountRepository)
	do.Provide(injector, data.NewCardRepository)
	do.Provide(injector, data.NewCardHolderRepository)
	do.Provide(injector, data.NewCardProductRepository)
	do.Provide(injector, data.NewVirtualAccountRepository)
	do.Provide(injector, data.NewWalletRepository)
	do.Provide(injector, data.NewAuthorizationRepository)
	do.Provide(injector, data.NewCardTransactionRepository)
	do.Provide(injector, data.NewWalletTransferRepository)
	do.Provide(injector, data.NewWebhookConfigRepository)
	do.Provide(injector, data.NewWebhookRecordRepository)
	do.Provide(injector, biz.NewCardIssuer)
	do.Provide(injector, biz.NewBalanceChanger)
	do.Provide(injector, biz.NewCardTransactionSimulator)
}
