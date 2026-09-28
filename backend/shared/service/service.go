package service

import (
	"generic-mock/enums"
	"generic-mock/shared/biz"
	sharederrors "generic-mock/shared/errors"

	"github.com/samber/do/v2"
)

type NewRequest struct {
	WebhookEventCatalog                   biz.UIWebhookEventCatalog
	WebhookReplayer                       biz.UIWebhookReplayer
	Channel                               enums.Channel
	Notificator                           biz.Notificator
	CreateVirtualAccountOnAccountCreation *bool
}

type Factory struct {
	usecaseFactory *biz.UIFactory
}

type Service struct {
	uc biz.UI
}

func NewFactory(injector do.Injector) (*Factory, error) {
	return &Factory{usecaseFactory: do.MustInvoke[*biz.UIFactory](injector)}, nil
}

func (factory *Factory) New(req *NewRequest) (*Service, error) {
	if req == nil {
		return nil, sharederrors.ErrInvalidUIConfiguration
	}
	uc, err := factory.usecaseFactory.New(&biz.NewUIRequest{
		WebhookEventCatalog:                   req.WebhookEventCatalog,
		WebhookReplayer:                       req.WebhookReplayer,
		Channel:                               req.Channel,
		Notificator:                           req.Notificator,
		CreateVirtualAccountOnAccountCreation: req.CreateVirtualAccountOnAccountCreation,
	})
	if err != nil {
		return nil, err
	}
	return &Service{uc: uc}, nil
}

type Empty struct{}
