package http

import (
	"generic-mock/channel/photonpay/service"
	sharedhttp "generic-mock/shared/http"

	"github.com/gin-gonic/gin"
)

type RegisterRequest struct {
	Router  *gin.RouterGroup
	OpenAPI *service.PhotonPayOpenAPIService
	UI      *service.PhotonPayUIService
}

func Register(req RegisterRequest) {
	router := req.Router
	uiService := req.UI
	openAPIService := req.OpenAPI
	openAPIRoutes := router.Group("")
	{
		openAPIRoutes.POST("/oauth2/token/accessToken", bind(openAPIService.AccessToken))
		openAPIRoutes.GET("/wallet/openApi/v4/account/history", bind(openAPIService.AccountHistory))
		openAPIRoutes.GET("/wallet/openApi/v4/account/single", bind(openAPIService.AccountSingle))
		openAPIRoutes.POST("/vcc/openApi/v4/addCardholder", bind(openAPIService.CreateCardHolder))
		openAPIRoutes.POST("/vcc/openApi/v4/cancelCard", bind(openAPIService.CancelCard))
		openAPIRoutes.POST("/vcc/openApi/v4/editCardBillingAddress", bind(openAPIService.EditCardBillingAddress))
		openAPIRoutes.POST("/vcc/openApi/v4/editCardholder", bind(openAPIService.EditCardHolder))
		openAPIRoutes.POST("/vcc/openApi/v4/freezeCard", bind(openAPIService.FreezeCard))
		openAPIRoutes.GET("/vcc/openApi/v4/getCardBin", bind(openAPIService.CardBins))
		openAPIRoutes.GET("/vcc/openApi/v4/getCardDetail", bind(openAPIService.CardDetail))
		openAPIRoutes.GET("/vcc/openApi/v4/getCvv", bind(openAPIService.CardCVV))
		openAPIRoutes.GET("/vcc/openApi/v4/getRequestResult", bind(openAPIService.RequestResult))
		openAPIRoutes.POST("/vcc/openApi/v4/openCard", bind(openAPIService.OpenCard))
		openAPIRoutes.GET("/vcc/openApi/v4/pagingIssuingHistory", bind(openAPIService.IssuingHistory))
		openAPIRoutes.GET("/vcc/openApi/v4/pagingRechargeCardFundsDetail", bind(openAPIService.CardFundsHistory))
		openAPIRoutes.GET("/vcc/openApi/v4/pagingShareCardTxnLimitDetail", bind(openAPIService.CardFundsHistory))
		openAPIRoutes.GET("/vcc/openApi/v4/pagingVccCard", bind(openAPIService.ListCards))
		openAPIRoutes.GET("/vcc/openApi/v4/pagingVccCardholder", bind(openAPIService.ListCardHolders))
		openAPIRoutes.GET("/vcc/openApi/v4/pagingVccTradeOrder", bind(openAPIService.ListTrades))
		openAPIRoutes.GET("/vcc/openApi/v4/preRecharge", bind(openAPIService.PreRecharge))
		openAPIRoutes.POST("/vcc/openApi/v4/recharge", bind(openAPIService.Recharge))
		openAPIRoutes.POST("/vcc/openApi/v4/rechargeReturn", bind(openAPIService.Recharge))
		openAPIRoutes.POST("/vcc/openApi/v4/updateCard", bind(openAPIService.UpdateCard))
		openAPIRoutes.GET("/exchange-center/open/api/v1/webhook/notification", bind(openAPIService.WebhookNotifications))
		openAPIRoutes.PUT("/exchange-center/open/api/v1/webhook/notification", bind(openAPIService.WebhookNotifications))
		openAPIRoutes.DELETE("/exchange-center/open/api/v1/webhook/notification", bind(openAPIService.WebhookNotifications))
		openAPIRoutes.POST("/file/apiUpload/:businessKey", bind(openAPIService.Upload))
	}

	uiRoutes := router.Group("/ui")
	sharedhttp.Register(sharedhttp.RegisterRequest{
		Router:                uiRoutes,
		Service:               req.UI.Shared,
		EnableVirtualAccounts: true,
		EnableWebhooks:        true,
	})
	uiRoutes.GET("/authorization-config", bindUI(uiService.GetAuthorizationConfig))
	uiRoutes.POST("/authorization-config", bindUI(uiService.UpdateAuthorizationConfig))
}
