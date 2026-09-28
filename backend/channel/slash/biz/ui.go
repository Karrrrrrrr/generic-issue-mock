package biz

import (
	"github.com/samber/do/v2"
)

type SlashUIUsecase struct {
	authorizationConfigRepo SlashAuthorizationConfigRepository
	accountRepository       SlashAccountRepository
}

func NewSlashUIUsecase(injector do.Injector) (*SlashUIUsecase, error) {
	return &SlashUIUsecase{
		authorizationConfigRepo: do.MustInvoke[SlashAuthorizationConfigRepository](injector),
		accountRepository:       do.MustInvoke[SlashAccountRepository](injector),
	}, nil
}
