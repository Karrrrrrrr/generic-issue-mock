package biz

import (
	sharedbiz "generic-mock/shared/biz"

	"github.com/samber/do"
)

type PhotonPayOpenAPIUsecase struct {
	simulator           sharedbiz.CardTransactionSimulator
	transaction         PhotonPayTransaction
	cardHolderRepo      CardHolderRepository
	cardRepo            CardRepository
	cardProductRepo     CardProductRepository
	cardTransactionRepo CardTransactionRepository
	virtualAccountRepo  VirtualAccountRepository
}

func NewPhotonPayOpenAPIUsecase(injector *do.Injector) (*PhotonPayOpenAPIUsecase, error) {
	return &PhotonPayOpenAPIUsecase{
		simulator:           do.MustInvoke[sharedbiz.CardTransactionSimulator](injector),
		transaction:         do.MustInvoke[PhotonPayTransaction](injector),
		cardHolderRepo:      do.MustInvoke[CardHolderRepository](injector),
		cardRepo:            do.MustInvoke[CardRepository](injector),
		cardProductRepo:     do.MustInvoke[CardProductRepository](injector),
		cardTransactionRepo: do.MustInvoke[CardTransactionRepository](injector),
		virtualAccountRepo:  do.MustInvoke[VirtualAccountRepository](injector),
	}, nil
}
