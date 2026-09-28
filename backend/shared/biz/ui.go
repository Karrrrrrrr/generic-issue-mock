package biz

import (
	"generic-mock/enums"
	sharederrors "generic-mock/shared/errors"

	"github.com/samber/do/v2"
)

type NewUIRequest struct {
	WebhookEventCatalog                   UIWebhookEventCatalog
	WebhookReplayer                       UIWebhookReplayer
	Channel                               enums.Channel
	Notificator                           CardTransactionNotificator
	CreateVirtualAccountOnAccountCreation *bool
}

func (req *NewUIRequest) Validate() error {
	if req == nil {
		return sharederrors.ErrInvalidUIConfiguration
	}
	switch req.Channel {
	case enums.Channel_Slash, enums.Channel_Paynda, enums.Channel_PhotonPay, enums.Channel_PingPong:
		return nil
	default:
		return sharederrors.ErrInvalidUIConfiguration
	}
}

type UI interface {
	UIAccounts
	UICardProducts
	UICardHolders
	UIWallets
	UIVirtualAccounts
	UICards
	UIAuthorizations
	UICardTransactions
	UIWalletTransfers
	UISimulation
	UIFunding
	UIWebhooks
}

type ui struct {
	webhookEventCatalog         UIWebhookEventCatalog
	webhookReplayer             UIWebhookReplayer
	webhookConfigRepo           WebhookConfigRepo
	webhookRecordRepo           WebhookRecordRepo
	channel                     enums.Channel
	notificator                 CardTransactionNotificator
	createAccountVirtualAccount bool
	accountRepo                 AccountRepo
	cardProductRepo             CardProductRepo
	cardHolderRepo              CardHolderRepo
	virtualAccountRepo          VirtualAccountRepo
	cardRepo                    CardRepo
	authorizationRepo           AuthorizationRepo
	cardTransactionRepo         CardTransactionRepo
	walletTransferRepo          WalletTransferRepo
	walletRepo                  WalletRepo
	tx                          Transaction
	balanceChanger              BalanceChanger
	simulator                   CardTransactionSimulator
}

var _ UI = (*ui)(nil)

type UIFactory struct {
	prototype ui
}

func NewUIFactory(injector do.Injector) (*UIFactory, error) {
	return &UIFactory{
		prototype: ui{
			webhookConfigRepo:   do.MustInvoke[WebhookConfigRepo](injector),
			webhookRecordRepo:   do.MustInvoke[WebhookRecordRepo](injector),
			accountRepo:         do.MustInvoke[AccountRepo](injector),
			cardProductRepo:     do.MustInvoke[CardProductRepo](injector),
			cardHolderRepo:      do.MustInvoke[CardHolderRepo](injector),
			virtualAccountRepo:  do.MustInvoke[VirtualAccountRepo](injector),
			cardRepo:            do.MustInvoke[CardRepo](injector),
			authorizationRepo:   do.MustInvoke[AuthorizationRepo](injector),
			cardTransactionRepo: do.MustInvoke[CardTransactionRepo](injector),
			walletTransferRepo:  do.MustInvoke[WalletTransferRepo](injector),
			walletRepo:          do.MustInvoke[WalletRepo](injector),
			tx:                  do.MustInvoke[Transaction](injector),
			balanceChanger:      do.MustInvoke[BalanceChanger](injector),
			simulator:           do.MustInvoke[CardTransactionSimulator](injector),
		},
	}, nil
}

func (factory *UIFactory) New(req *NewUIRequest) (UI, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	uc := factory.prototype
	uc.webhookEventCatalog = req.WebhookEventCatalog
	uc.webhookReplayer = req.WebhookReplayer
	uc.channel = req.Channel
	uc.notificator = req.Notificator
	uc.createAccountVirtualAccount = req.CreateVirtualAccountOnAccountCreation != nil && *req.CreateVirtualAccountOnAccountCreation
	return &uc, nil
}
