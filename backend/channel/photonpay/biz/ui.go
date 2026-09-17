package biz

import (
	"generic-mock/model"

	"github.com/samber/do"
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

type PhotonPayUIUsecase struct {
	transaction             PhotonPayTransaction
	cardHolderRepo          CardHolderRepository
	cardRepo                CardRepository
	cardProductRepo         CardProductRepository
	authorizationRepo       AuthorizationRepository
	cardTransactionRepo     CardTransactionRepository
	webhookRepo             WebhookConfigRepository
	authorizationConfigRepo AuthorizationConfigRepository
	webhookRecordRepo       WebhookRecordRepository
	webhookClient           WebhookClient
	accountRepo             AccountRepository
	walletRepo              WalletRepository
	virtualAccountRepo      VirtualAccountRepository
}

func NewPhotonPayUIUsecase(injector *do.Injector) (*PhotonPayUIUsecase, error) {
	return &PhotonPayUIUsecase{
		transaction:             do.MustInvoke[PhotonPayTransaction](injector),
		cardHolderRepo:          do.MustInvoke[CardHolderRepository](injector),
		cardRepo:                do.MustInvoke[CardRepository](injector),
		cardProductRepo:         do.MustInvoke[CardProductRepository](injector),
		authorizationRepo:       do.MustInvoke[AuthorizationRepository](injector),
		cardTransactionRepo:     do.MustInvoke[CardTransactionRepository](injector),
		webhookRepo:             do.MustInvoke[WebhookConfigRepository](injector),
		authorizationConfigRepo: do.MustInvoke[AuthorizationConfigRepository](injector),
		webhookRecordRepo:       do.MustInvoke[WebhookRecordRepository](injector),
		webhookClient:           do.MustInvoke[WebhookClient](injector),
		accountRepo:             do.MustInvoke[AccountRepository](injector),
		walletRepo:              do.MustInvoke[WalletRepository](injector),
		virtualAccountRepo:      do.MustInvoke[VirtualAccountRepository](injector),
	}, nil
}
