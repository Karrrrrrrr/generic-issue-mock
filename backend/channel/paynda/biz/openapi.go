package biz

import (
	"github.com/samber/do"
)

type PayndaOpenAPIUsecase struct {
	transaction               PayndaTransaction
	cardHolderRepository      PayndaCardHolderRepository
	cardProductRepository     PayndaCardProductRepository
	cardRepository            PayndaCardRepository
	walletRepository          PayndaWalletRepository
	accountRepository         PayndaAccountRepository
	cardTransactionRepository PayndaCardTransactionRepository
	authorizationRepository   PayndaAuthorizationRepository
}

func NewPayndaOpenAPIUsecase(injector *do.Injector) (*PayndaOpenAPIUsecase, error) {
	return &PayndaOpenAPIUsecase{
		transaction:               do.MustInvoke[PayndaTransaction](injector),
		cardHolderRepository:      do.MustInvoke[PayndaCardHolderRepository](injector),
		cardProductRepository:     do.MustInvoke[PayndaCardProductRepository](injector),
		cardRepository:            do.MustInvoke[PayndaCardRepository](injector),
		walletRepository:          do.MustInvoke[PayndaWalletRepository](injector),
		accountRepository:         do.MustInvoke[PayndaAccountRepository](injector),
		cardTransactionRepository: do.MustInvoke[PayndaCardTransactionRepository](injector),
		authorizationRepository:   do.MustInvoke[PayndaAuthorizationRepository](injector),
	}, nil
}
