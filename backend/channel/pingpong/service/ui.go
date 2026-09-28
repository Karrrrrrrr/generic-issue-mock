package service

import (
	"generic-mock/enums"
	sharedbiz "generic-mock/shared/biz"
	sharedservice "generic-mock/shared/service"

	"github.com/samber/do/v2"
)

type PingPongUIService struct {
	Shared *sharedservice.Service
}

func NewUIService(injector do.Injector) (*PingPongUIService, error) {
	factory := do.MustInvoke[*sharedservice.Factory](injector)
	s := &PingPongUIService{}
	management, err := factory.New(&sharedservice.NewRequest{
		Channel:     enums.Channel_PingPong,
		Notificator: sharedbiz.NoopNotificator{},
	})
	if err != nil {
		return nil, err
	}
	s.Shared = management
	return s, nil
}
