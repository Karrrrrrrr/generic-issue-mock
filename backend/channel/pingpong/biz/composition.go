package biz

import (
	sharedbiz "generic-mock/shared/biz"

	"github.com/samber/do/v2"
)

type PingPongOpenAPIUsecase struct {
	tx                 sharedbiz.Transaction
	accountRepo        PingPongAccountRepository
	virtualAccountRepo PingPongVirtualAccountRepository
	cardRepo           PingPongCardRepository
	cardProductRepo    PingPongProductRepository
	walletRepo         PingPongWalletRepository
	walletTransferRepo PingPongTransferRepository
}

type PingPongUIUsecase struct{}

func NewOpenAPIUsecase(injector do.Injector) (*PingPongOpenAPIUsecase, error) {
	return &PingPongOpenAPIUsecase{
		tx:                 do.MustInvoke[sharedbiz.Transaction](injector),
		accountRepo:        do.MustInvoke[PingPongAccountRepository](injector),
		virtualAccountRepo: do.MustInvoke[PingPongVirtualAccountRepository](injector),
		cardRepo:           do.MustInvoke[PingPongCardRepository](injector),
		cardProductRepo:    do.MustInvoke[PingPongProductRepository](injector),
		walletRepo:         do.MustInvoke[PingPongWalletRepository](injector),
		walletTransferRepo: do.MustInvoke[PingPongTransferRepository](injector),
	}, nil
}

func NewUIUsecase(injector do.Injector) (*PingPongUIUsecase, error) {
	return &PingPongUIUsecase{}, nil
}
