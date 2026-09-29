package biz

import (
	sharedbiz "generic-mock/shared/biz"

	"github.com/samber/do/v2"
)

type PingPongOpenAPIUsecase struct {
	tx                        sharedbiz.Transaction
	accountRepo               PingPongAccountRepository
	virtualAccountRepo        PingPongVirtualAccountRepository
	cardRepo                  PingPongCardRepository
	sharedCardRepo            sharedbiz.CardRepo
	sharedCardTransactionRepo sharedbiz.CardTransactionRepo
	cardProductRepo           PingPongProductRepository
	sharedWalletRepo          sharedbiz.WalletRepo
	walletTransferRepo        PingPongTransferRepository
}

type PingPongUIUsecase struct{}

func NewOpenAPIUsecase(injector do.Injector) (*PingPongOpenAPIUsecase, error) {
	return &PingPongOpenAPIUsecase{
		tx:                        do.MustInvoke[sharedbiz.Transaction](injector),
		accountRepo:               do.MustInvoke[PingPongAccountRepository](injector),
		virtualAccountRepo:        do.MustInvoke[PingPongVirtualAccountRepository](injector),
		cardRepo:                  do.MustInvoke[PingPongCardRepository](injector),
		sharedCardRepo:            do.MustInvoke[sharedbiz.CardRepo](injector),
		sharedCardTransactionRepo: do.MustInvoke[sharedbiz.CardTransactionRepo](injector),
		cardProductRepo:           do.MustInvoke[PingPongProductRepository](injector),
		sharedWalletRepo:          do.MustInvoke[sharedbiz.WalletRepo](injector),
		walletTransferRepo:        do.MustInvoke[PingPongTransferRepository](injector),
	}, nil
}

func NewUIUsecase(injector do.Injector) (*PingPongUIUsecase, error) {
	return &PingPongUIUsecase{}, nil
}
