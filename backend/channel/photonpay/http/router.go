package http

import (
	"generic-mock/channel/photonpay/service"

	"github.com/gin-gonic/gin"
)

func Register(
	router *gin.RouterGroup,
	openAPIService *service.PhotonPayOpenAPIService,
	uiService *service.PhotonPayUIService,
) {
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
	{
		uiRoutes.GET("/accounts", bindUI(uiService.ListAccounts))
		uiRoutes.POST("/accounts", bindUI(uiService.CreateAccount))
		uiRoutes.PUT("/accounts/:id", bindUI(uiService.UpdateAccount))
		uiRoutes.GET("/funds", bindUI(uiService.ListFunds))
		uiRoutes.POST("/funds/transfer", bindUI(uiService.MoveFunds))
		uiRoutes.GET("/managed-virtual-accounts", bindUI(uiService.ListManagedVirtualAccounts))
		uiRoutes.POST("/managed-virtual-accounts", bindUI(uiService.CreateManagedVirtualAccount))
		uiRoutes.GET("/virtual-accounts", bindUI(uiService.ListVirtualAccounts))
		uiRoutes.POST("/virtual-accounts", bindUI(uiService.CreateVirtualAccount))
		uiRoutes.GET("/card-products", bindUI(uiService.ListCardProducts))
		uiRoutes.GET("/cardholders", bindUI(uiService.ListCardHolders))
		uiRoutes.GET("/cards", bindUI(uiService.ListCards))
		uiRoutes.POST("/cards/:id/fund", bindUI(uiService.FundCard))
		uiRoutes.PUT("/cards/:id/status", bindUI(uiService.UpdateCardStatus))
		uiRoutes.GET("/authorization-config", bindUI(uiService.GetAuthorizationConfig))
		uiRoutes.PUT("/authorization-config", bindUI(uiService.UpdateAuthorizationConfig))
		uiRoutes.GET("/authorizations", bindUI(uiService.ListAuthorizations))
		uiRoutes.POST("/authorizations/:id/clear", bindUI(uiService.ClearAuthorization))
		uiRoutes.GET("/authorizations/:id/detail", bindUI(uiService.GetAuthorizationDetail))
		uiRoutes.POST("/authorizations/:id/reverse", bindUI(uiService.ReverseAuthorization))
		uiRoutes.POST("/authorizations/:id/refund", bindUI(uiService.RefundAuthorization))
		uiRoutes.GET("/authorization-balances", bindUI(uiService.ListAuthorizationBalances))
		uiRoutes.GET("/transactions", bindUI(uiService.ListTransactions))
		uiRoutes.POST("/transactions/:id/refund", bindUI(uiService.RefundTransaction))
		uiRoutes.POST("/transactions/:id/clear", bindUI(uiService.ClearTransaction))
		uiRoutes.POST("/transactions/:id/reverse", bindUI(uiService.ReverseTransaction))
		uiRoutes.POST("/simulate/authorizations", bindUI(uiService.SimulateAuthorization))
		uiRoutes.POST("/simulate/refunds", bindUI(uiService.SimulateRefund))
		uiRoutes.GET("/webhooks", bindUI(uiService.ListWebhooks))
		uiRoutes.POST("/webhooks", bindUI(uiService.CreateWebhook))
		uiRoutes.PUT("/webhooks/:id", bindUI(uiService.UpdateWebhook))
		uiRoutes.DELETE("/webhooks/:id", bindUI(uiService.DeleteWebhook))
		uiRoutes.GET("/webhooks/events", bindUI(uiService.ListWebhookEvents))
		uiRoutes.GET("/webhook-records", bindUI(uiService.ListWebhookRecords))
		uiRoutes.POST("/webhook-records/:id/replay", bindUI(uiService.ReplayWebhookRecord))
	}
}
