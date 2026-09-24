package biz

import (
	"github.com/samber/do"
)

type SlashOpenAPIUsecase struct {
	accountRepository         SlashAccountRepository
	transaction               SlashTransaction
	cardHolderRepository      SlashCardHolderRepository
	cardRepository            SlashCardRepository
	cardProductRepository     SlashCardProductRepository
	cardTransactionRepository SlashCardTransactionRepository
	virtualAccountRepository  SlashVirtualAccountRepository
	walletRepository          SlashWalletRepository
}

func NewSlashOpenAPIUsecase(injector *do.Injector) (*SlashOpenAPIUsecase, error) {
	return &SlashOpenAPIUsecase{
		accountRepository:         do.MustInvoke[SlashAccountRepository](injector),
		transaction:               do.MustInvoke[SlashTransaction](injector),
		cardHolderRepository:      do.MustInvoke[SlashCardHolderRepository](injector),
		cardRepository:            do.MustInvoke[SlashCardRepository](injector),
		cardProductRepository:     do.MustInvoke[SlashCardProductRepository](injector),
		cardTransactionRepository: do.MustInvoke[SlashCardTransactionRepository](injector),
		virtualAccountRepository:  do.MustInvoke[SlashVirtualAccountRepository](injector),
		walletRepository:          do.MustInvoke[SlashWalletRepository](injector),
	}, nil
}
