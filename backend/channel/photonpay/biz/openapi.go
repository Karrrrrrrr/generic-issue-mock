package biz

import (
	"github.com/samber/do"
)

type PhotonPayOpenAPIUsecase struct {
	transaction         PhotonPayTransaction
	cardHolderRepo      CardHolderRepository
	cardRepo            CardRepository
	cardProductRepo     CardProductRepository
	authorizationRepo   AuthorizationRepository
	cardTransactionRepo CardTransactionRepository
	virtualAccountRepo  VirtualAccountRepository
}

func NewPhotonPayOpenAPIUsecase(injector *do.Injector) (*PhotonPayOpenAPIUsecase, error) {
	return &PhotonPayOpenAPIUsecase{
		transaction:         do.MustInvoke[PhotonPayTransaction](injector),
		cardHolderRepo:      do.MustInvoke[CardHolderRepository](injector),
		cardRepo:            do.MustInvoke[CardRepository](injector),
		cardProductRepo:     do.MustInvoke[CardProductRepository](injector),
		authorizationRepo:   do.MustInvoke[AuthorizationRepository](injector),
		cardTransactionRepo: do.MustInvoke[CardTransactionRepository](injector),
		virtualAccountRepo:  do.MustInvoke[VirtualAccountRepository](injector),
	}, nil
}
