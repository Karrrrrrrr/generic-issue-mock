package service

import (
	"generic-mock/enums"
	sharedbiz "generic-mock/shared/biz"
	sharedservice "generic-mock/shared/service"

	"github.com/samber/do/v2"
)

type SharedUIService struct{ *sharedservice.Service }

func NewSharedUIService(injector do.Injector) (*SharedUIService, error) {
	factory := do.MustInvoke[*sharedservice.Factory](injector)

	management, err := factory.New(&sharedservice.NewRequest{
		Channel:     enums.Channel_PingPong,
		Notificator: sharedbiz.NoopNotificator{},
	})
	if err != nil {
		return nil, err
	}
	return &SharedUIService{Service: management}, nil
}
