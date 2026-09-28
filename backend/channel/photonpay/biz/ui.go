package biz

import (
	"github.com/samber/do/v2"
)

type PhotonPayUIUsecase struct {
	cardRepo                CardRepository
	cardTransactionRepo     CardTransactionRepository
	webhookRepo             WebhookConfigRepository
	authorizationConfigRepo AuthorizationConfigRepository
	webhookRecordRepo       WebhookRecordRepository
	webhookClient           WebhookClient
	accountRepo             AccountRepository
}

func NewPhotonPayUIUsecase(injector do.Injector) (*PhotonPayUIUsecase, error) {
	return &PhotonPayUIUsecase{
		cardRepo:                do.MustInvoke[CardRepository](injector),
		cardTransactionRepo:     do.MustInvoke[CardTransactionRepository](injector),
		webhookRepo:             do.MustInvoke[WebhookConfigRepository](injector),
		authorizationConfigRepo: do.MustInvoke[AuthorizationConfigRepository](injector),
		webhookRecordRepo:       do.MustInvoke[WebhookRecordRepository](injector),
		webhookClient:           do.MustInvoke[WebhookClient](injector),
		accountRepo:             do.MustInvoke[AccountRepository](injector),
	}, nil
}
