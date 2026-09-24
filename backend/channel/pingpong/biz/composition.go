package biz

import "github.com/samber/do"

type PingPongOpenAPIUsecase struct {
	tx                 PingPongTransaction
	accountRepo        PingPongAccountRepository
	virtualAccountRepo PingPongVirtualAccountRepository
	cardRepo           PingPongCardRepository
	cardProductRepo    PingPongProductRepository
	walletRepo         PingPongWalletRepository
	walletTransferRepo PingPongTransferRepository
}

type PingPongUIUsecase struct {
	tx                  PingPongTransaction
	accountRepo         PingPongAccountRepository
	virtualAccountRepo  PingPongVirtualAccountRepository
	cardRepo            PingPongCardRepository
	cardProductRepo     PingPongProductRepository
	walletRepo          PingPongWalletRepository
	walletTransferRepo  PingPongTransferRepository
	authorizationRepo   PingPongAuthorizationRepository
	cardTransactionRepo PingPongCardTransactionRepository
}

func NewOpenAPIUsecase(injector *do.Injector) (*PingPongOpenAPIUsecase, error) {
	return &PingPongOpenAPIUsecase{
		tx:                 do.MustInvoke[PingPongTransaction](injector),
		accountRepo:        do.MustInvoke[PingPongAccountRepository](injector),
		virtualAccountRepo: do.MustInvoke[PingPongVirtualAccountRepository](injector),
		cardRepo:           do.MustInvoke[PingPongCardRepository](injector),
		cardProductRepo:    do.MustInvoke[PingPongProductRepository](injector),
		walletRepo:         do.MustInvoke[PingPongWalletRepository](injector),
		walletTransferRepo: do.MustInvoke[PingPongTransferRepository](injector),
	}, nil
}

func NewUIUsecase(injector *do.Injector) (*PingPongUIUsecase, error) {
	return &PingPongUIUsecase{
		tx:                  do.MustInvoke[PingPongTransaction](injector),
		accountRepo:         do.MustInvoke[PingPongAccountRepository](injector),
		virtualAccountRepo:  do.MustInvoke[PingPongVirtualAccountRepository](injector),
		cardRepo:            do.MustInvoke[PingPongCardRepository](injector),
		cardProductRepo:     do.MustInvoke[PingPongProductRepository](injector),
		walletRepo:          do.MustInvoke[PingPongWalletRepository](injector),
		walletTransferRepo:  do.MustInvoke[PingPongTransferRepository](injector),
		authorizationRepo:   do.MustInvoke[PingPongAuthorizationRepository](injector),
		cardTransactionRepo: do.MustInvoke[PingPongCardTransactionRepository](injector),
	}, nil
}
