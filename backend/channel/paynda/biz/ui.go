package biz

import (
	"generic-mock/model"
	sharedbiz "generic-mock/shared/biz"

	"github.com/samber/do/v2"
)

type PayndaResourceRequest struct {
	AccountID model.ID
	ID        model.ID
}

type PayndaListRequest struct {
	AccountID *model.ID
	Offset    int
	Limit     int
}

type PayndaUIUsecase struct {
	simulator                 sharedbiz.CardTransactionSimulator
	transaction               sharedbiz.Transaction
	cardRepository            PayndaCardRepository
	cardHolderRepository      PayndaCardHolderRepository
	cardProductRepository     PayndaCardProductRepository
	walletRepository          PayndaWalletRepository
	accountRepository         PayndaAccountRepository
	cardTransactionRepository PayndaCardTransactionRepository
	authorizationRepository   PayndaAuthorizationRepository
	webhookConfigRepository   PayndaWebhookConfigRepository
	webhookRecordRepository   PayndaWebhookRecordRepository
	webhookClient             PayndaWebhookClient
}

func NewPayndaUIUsecase(injector do.Injector) (*PayndaUIUsecase, error) {
	return &PayndaUIUsecase{
		simulator:                 do.MustInvoke[sharedbiz.CardTransactionSimulator](injector),
		transaction:               do.MustInvoke[sharedbiz.Transaction](injector),
		cardRepository:            do.MustInvoke[PayndaCardRepository](injector),
		cardHolderRepository:      do.MustInvoke[PayndaCardHolderRepository](injector),
		cardProductRepository:     do.MustInvoke[PayndaCardProductRepository](injector),
		walletRepository:          do.MustInvoke[PayndaWalletRepository](injector),
		accountRepository:         do.MustInvoke[PayndaAccountRepository](injector),
		cardTransactionRepository: do.MustInvoke[PayndaCardTransactionRepository](injector),
		authorizationRepository:   do.MustInvoke[PayndaAuthorizationRepository](injector),
		webhookConfigRepository:   do.MustInvoke[PayndaWebhookConfigRepository](injector),
		webhookRecordRepository:   do.MustInvoke[PayndaWebhookRecordRepository](injector),
		webhookClient:             do.MustInvoke[PayndaWebhookClient](injector),
	}, nil
}
