package biz

import (
	"generic-mock/model"
	sharedbiz "generic-mock/shared/biz"

	"github.com/samber/do"
)

type ResourceRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type SlashUIUsecase struct {
	simulator                 sharedbiz.CardTransactionSimulator
	transaction               SlashTransaction
	cardHolderRepository      SlashCardHolderRepository
	cardRepository            SlashCardRepository
	cardProductRepository     SlashCardProductRepository
	authorizationRepository   SlashAuthorizationRepository
	cardTransactionRepository SlashCardTransactionRepository
	webhookConfigRepository   SlashWebhookConfigRepository
	authorizationConfigRepo   SlashAuthorizationConfigRepository
	virtualAccountRepository  SlashVirtualAccountRepository
	accountRepository         SlashAccountRepository
	walletRepository          SlashWalletRepository
}

func NewSlashUIUsecase(injector *do.Injector) (*SlashUIUsecase, error) {
	return &SlashUIUsecase{
		simulator:                 do.MustInvoke[sharedbiz.CardTransactionSimulator](injector),
		transaction:               do.MustInvoke[SlashTransaction](injector),
		cardHolderRepository:      do.MustInvoke[SlashCardHolderRepository](injector),
		cardRepository:            do.MustInvoke[SlashCardRepository](injector),
		cardProductRepository:     do.MustInvoke[SlashCardProductRepository](injector),
		authorizationRepository:   do.MustInvoke[SlashAuthorizationRepository](injector),
		cardTransactionRepository: do.MustInvoke[SlashCardTransactionRepository](injector),
		webhookConfigRepository:   do.MustInvoke[SlashWebhookConfigRepository](injector),
		authorizationConfigRepo:   do.MustInvoke[SlashAuthorizationConfigRepository](injector),
		virtualAccountRepository:  do.MustInvoke[SlashVirtualAccountRepository](injector),
		accountRepository:         do.MustInvoke[SlashAccountRepository](injector),
		walletRepository:          do.MustInvoke[SlashWalletRepository](injector),
	}, nil
}
