package shared

import (
	"generic-mock/shared/biz"
	"generic-mock/shared/service"

	"github.com/samber/do/v2"
)

func RegisterUIProviders(injector do.Injector) {
	do.Provide(injector, biz.NewUIFactory)
	do.Provide(injector, service.NewFactory)
}
