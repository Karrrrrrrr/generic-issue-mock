package biz

import (
	"github.com/samber/do/v2"
)

type PayndaWebhookNotificator struct {
	cardRepository            PayndaCardRepository
	cardHolderRepository      PayndaCardHolderRepository
	walletRepository          PayndaWalletRepository
	accountRepository         PayndaAccountRepository
	cardTransactionRepository PayndaCardTransactionRepository
	authorizationRepository   PayndaAuthorizationRepository
	webhookConfigRepository   PayndaWebhookConfigRepository
	webhookRecordRepository   PayndaWebhookRecordRepository
	webhookClient             PayndaWebhookClient
}

func NewPayndaWebhookNotificator(injector do.Injector) (*PayndaWebhookNotificator, error) {
	return &PayndaWebhookNotificator{
		cardRepository:            do.MustInvoke[PayndaCardRepository](injector),
		cardHolderRepository:      do.MustInvoke[PayndaCardHolderRepository](injector),
		walletRepository:          do.MustInvoke[PayndaWalletRepository](injector),
		accountRepository:         do.MustInvoke[PayndaAccountRepository](injector),
		cardTransactionRepository: do.MustInvoke[PayndaCardTransactionRepository](injector),
		authorizationRepository:   do.MustInvoke[PayndaAuthorizationRepository](injector),
		webhookConfigRepository:   do.MustInvoke[PayndaWebhookConfigRepository](injector),
		webhookRecordRepository:   do.MustInvoke[PayndaWebhookRecordRepository](injector),
		webhookClient:             do.MustInvoke[PayndaWebhookClient](injector),
	}, nil
}
