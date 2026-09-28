package biz

import (
	"generic-mock/model"
	sharedbiz "generic-mock/shared/biz"

	"github.com/samber/do/v2"
)

type ResourceRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type SlashOpenAPIUsecase struct {
	accountRepository         SlashAccountRepository
	transaction               sharedbiz.Transaction
	cardHolderRepository      SlashCardHolderRepository
	cardRepository            SlashCardRepository
	cardProductRepository     SlashCardProductRepository
	cardTransactionRepository SlashCardTransactionRepository
	virtualAccountRepository  SlashVirtualAccountRepository
	walletRepository          SlashWalletRepository
}

func NewSlashOpenAPIUsecase(injector do.Injector) (*SlashOpenAPIUsecase, error) {
	return &SlashOpenAPIUsecase{
		accountRepository:         do.MustInvoke[SlashAccountRepository](injector),
		transaction:               do.MustInvoke[sharedbiz.Transaction](injector),
		cardHolderRepository:      do.MustInvoke[SlashCardHolderRepository](injector),
		cardRepository:            do.MustInvoke[SlashCardRepository](injector),
		cardProductRepository:     do.MustInvoke[SlashCardProductRepository](injector),
		cardTransactionRepository: do.MustInvoke[SlashCardTransactionRepository](injector),
		virtualAccountRepository:  do.MustInvoke[SlashVirtualAccountRepository](injector),
		walletRepository:          do.MustInvoke[SlashWalletRepository](injector),
	}, nil
}
