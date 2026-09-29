package biz

import (
	sharedbiz "generic-mock/shared/biz"

	"github.com/samber/do/v2"
)

type PhotonPayUIUsecase struct {
	authorizationConfigRepo AuthorizationConfigRepository
	accountRepo             AccountRepository
}

type PhotonPayWebhookNotificator struct {
	sharedbiz.NoopNotificator
	cardRepo            CardRepository
	cardTransactionRepo CardTransactionRepository
	webhookRepo         WebhookConfigRepository
	webhookRecordRepo   WebhookRecordRepository
	webhookClient       WebhookClient
}

func NewPhotonPayUIUsecase(injector do.Injector) (*PhotonPayUIUsecase, error) {
	return &PhotonPayUIUsecase{
		authorizationConfigRepo: do.MustInvoke[AuthorizationConfigRepository](injector),
		accountRepo:             do.MustInvoke[AccountRepository](injector),
	}, nil
}

func NewPhotonPayWebhookNotificator(injector do.Injector) (*PhotonPayWebhookNotificator, error) {
	return &PhotonPayWebhookNotificator{
		cardRepo:            do.MustInvoke[CardRepository](injector),
		cardTransactionRepo: do.MustInvoke[CardTransactionRepository](injector),
		webhookRepo:         do.MustInvoke[WebhookConfigRepository](injector),
		webhookRecordRepo:   do.MustInvoke[WebhookRecordRepository](injector),
		webhookClient:       do.MustInvoke[WebhookClient](injector),
	}, nil
}
