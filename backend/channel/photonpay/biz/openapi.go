package biz

import (
	"generic-mock/model"
	sharedbiz "generic-mock/shared/biz"

	"github.com/samber/do/v2"
)

type ListRequest struct {
	AccountID *model.ID
	Offset    int
	Limit     int
}

type ResourceRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type PhotonPayOpenAPIUsecase struct {
	transaction         sharedbiz.Transaction
	cardHolderRepo      CardHolderRepository
	cardRepo            CardRepository
	cardProductRepo     CardProductRepository
	cardTransactionRepo sharedbiz.CardTransactionRepo
	virtualAccountRepo  VirtualAccountRepository
}

func NewPhotonPayOpenAPIUsecase(injector do.Injector) (*PhotonPayOpenAPIUsecase, error) {
	return &PhotonPayOpenAPIUsecase{
		transaction:         do.MustInvoke[sharedbiz.Transaction](injector),
		cardHolderRepo:      do.MustInvoke[CardHolderRepository](injector),
		cardRepo:            do.MustInvoke[CardRepository](injector),
		cardProductRepo:     do.MustInvoke[CardProductRepository](injector),
		cardTransactionRepo: do.MustInvoke[sharedbiz.CardTransactionRepo](injector),
		virtualAccountRepo:  do.MustInvoke[VirtualAccountRepository](injector),
	}, nil
}
