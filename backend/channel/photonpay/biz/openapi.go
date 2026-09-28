package biz

import (
	sharedbiz "generic-mock/shared/biz"

	"github.com/samber/do/v2"
)

type PhotonPayOpenAPIUsecase struct {
	transaction         sharedbiz.Transaction
	cardHolderRepo      CardHolderRepository
	cardRepo            CardRepository
	cardProductRepo     CardProductRepository
	cardTransactionRepo CardTransactionRepository
	virtualAccountRepo  VirtualAccountRepository
}

func NewPhotonPayOpenAPIUsecase(injector do.Injector) (*PhotonPayOpenAPIUsecase, error) {
	return &PhotonPayOpenAPIUsecase{
		transaction:         do.MustInvoke[sharedbiz.Transaction](injector),
		cardHolderRepo:      do.MustInvoke[CardHolderRepository](injector),
		cardRepo:            do.MustInvoke[CardRepository](injector),
		cardProductRepo:     do.MustInvoke[CardProductRepository](injector),
		cardTransactionRepo: do.MustInvoke[CardTransactionRepository](injector),
		virtualAccountRepo:  do.MustInvoke[VirtualAccountRepository](injector),
	}, nil
}
