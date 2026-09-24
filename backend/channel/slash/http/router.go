package http

import (
	"net/http"

	"generic-mock/channel/slash/service"
	"generic-mock/pkg/httpx"

	"github.com/gin-gonic/gin"
	kratosErrors "github.com/go-kratos/kratos/v2/errors"
)

func Register(router *gin.RouterGroup, service *service.SlashUIService, openAPIService *service.SlashOpenAPIService) {
	openAPIRoutes := router.Group("")
	{
		openAPIRoutes.GET("/legal-entity", bind(openAPIService.ListLegalEntities))
		openAPIRoutes.GET("/account", bind(openAPIService.ListAccounts))
		openAPIRoutes.GET("/account/:id", bind(openAPIService.GetAccount))
		openAPIRoutes.GET("/account/:id/authorization-webhook", bind(openAPIService.AuthWebhook))
		openAPIRoutes.PUT("/account/:id/authorization-webhook", bind(openAPIService.AuthWebhook))
		openAPIRoutes.GET("/account/:id/balance", bind(openAPIService.ListAccountBalances))
		openAPIRoutes.GET("/virtual-account", bind(openAPIService.ListVirtualAccounts))
		openAPIRoutes.POST("/virtual-account", bind(openAPIService.CreateVirtualAccount))
		openAPIRoutes.GET("/virtual-account/:id", bind(openAPIService.GetVirtualAccount))
		openAPIRoutes.PATCH("/virtual-account/:id", bind(openAPIService.UpdateVirtualAccount))
		openAPIRoutes.GET("/card-product", bind(openAPIService.ListCardProducts))
		openAPIRoutes.GET("/card", bind(openAPIService.ListCards))
		openAPIRoutes.POST("/card", bind(openAPIService.CreateCard))
		openAPIRoutes.GET("/card/:id", bind(openAPIService.GetCard))
		openAPIRoutes.PATCH("/card/:id", bind(openAPIService.UpdateCard))
		openAPIRoutes.GET("/card/:id/modifier", bind(openAPIService.GetCardModifiers))
		openAPIRoutes.PUT("/card/:id/modifier", bind(openAPIService.SetCardModifier))
		openAPIRoutes.PUT("/card/:id/spending-constraint", bind(openAPIService.SetCardSpendingConstraint))
		openAPIRoutes.PATCH("/card/:id/spending-constraint", bind(openAPIService.SetCardSpendingConstraint))
		openAPIRoutes.GET("/card/:id/utilization", bind(openAPIService.GetCardUtilization))
		openAPIRoutes.GET("/card-group", bind(openAPIService.ListCardGroups))
		openAPIRoutes.POST("/card-group", bind(openAPIService.CardGroup))
		openAPIRoutes.GET("/card-group/:id", bind(openAPIService.CardGroup))
		openAPIRoutes.PATCH("/card-group/:id", bind(openAPIService.CardGroup))
		openAPIRoutes.PATCH("/card-group/:id/spending-constraint", bind(openAPIService.SetGroupSpendingConstraint))
		openAPIRoutes.GET("/card-group/:id/utilization", bind(openAPIService.GetGroupUtilization))
		openAPIRoutes.GET("/transaction", bind(openAPIService.ListTransactions))
		openAPIRoutes.GET("/transaction/:id", bind(openAPIService.GetTransaction))
		openAPIRoutes.GET("/transaction/:id/fee-details", bind(openAPIService.GetTransactionFees))
		openAPIRoutes.GET("/transaction/aggregation", bind(openAPIService.GetTransactionAggregations))
		openAPIRoutes.POST("/transfer/virtual-account", bind(openAPIService.TransferVirtualAccount))
		openAPIRoutes.GET("/merchant", bind(openAPIService.ListMerchants))
		openAPIRoutes.GET("/merchant/:id", bind(openAPIService.GetMerchant))
		openAPIRoutes.GET("/merchant-category", bind(openAPIService.ListMerchantCategories))
		openAPIRoutes.GET("/webhook", bind(openAPIService.ListOpenAPIWebhooks))
		openAPIRoutes.POST("/webhook", bind(openAPIService.Webhook))
		openAPIRoutes.PATCH("/webhook/:id", bind(openAPIService.Webhook))
	}

	uiRoutes := router.Group("/ui")
	{
		uiRoutes.GET("/accounts", bind(service.ListAccounts))
		uiRoutes.POST("/accounts", bind(service.CreateAccount))
		uiRoutes.PUT("/accounts/:id", bind(service.UpdateAccount))
		uiRoutes.GET("/funds", bind(service.ListFunds))
		uiRoutes.POST("/funds/transfer", bind(service.MoveFunds))
		uiRoutes.GET("/managed-virtual-accounts", bind(service.ListManagedVirtualAccounts))
		uiRoutes.POST("/managed-virtual-accounts", bind(service.CreateManagedVirtualAccount))
		uiRoutes.GET("/virtual-accounts", bind(service.ListVirtualAccounts))
		uiRoutes.GET("/card-products", bind(service.ListCardProducts))
		uiRoutes.GET("/cardholders", bind(service.ListCardHolders))
		uiRoutes.GET("/cards", bind(service.ListCards))
		uiRoutes.GET("/cards/:id", bind(service.GetCard))
		uiRoutes.PUT("/cards/:id/status", bind(service.UpdateCardStatus))
		uiRoutes.GET("/authorization-config", bind(service.GetAuthorizationConfig))
		uiRoutes.PUT("/authorization-config", bind(service.UpdateAuthorizationConfig))
		uiRoutes.GET("/authorizations", bind(service.ListAuthorizations))
		uiRoutes.GET("/authorizations/:id", bind(service.GetAuthorization))
		uiRoutes.POST("/authorizations/:id/clear", bind(service.ClearAuthorization))
		uiRoutes.GET("/authorization-balances", bind(service.ListAuthorizationBalances))
		uiRoutes.GET("/transactions", bind(service.ListTransactions))
		uiRoutes.GET("/transactions/:id", bind(service.GetTransaction))
		uiRoutes.POST("/transactions/:id/refund", bind(service.RefundTransaction))
		uiRoutes.POST("/transactions/:id/reverse", bind(service.ReverseTransaction))
		uiRoutes.POST("/simulate/authorizations", bind(service.SimulateAuthorization))
		uiRoutes.POST("/simulate/refunds", bind(service.SimulateRefund))
		uiRoutes.GET("/webhooks", bind(service.ListWebhooks))
		uiRoutes.POST("/webhooks", bind(service.CreateWebhook))
		uiRoutes.PUT("/webhooks/:id", bind(service.UpdateWebhook))
		uiRoutes.DELETE("/webhooks/:id", bind(service.DeleteWebhook))
		uiRoutes.GET("/webhooks/events", bind(service.ListWebhookEvents))
		uiRoutes.GET("/webhook-records", bind(service.ListWebhookRecords))
		uiRoutes.POST("/webhook-records/:id/replay", bind(service.ReplayWebhookRecord))
	}
}

func bind[Req any, Resp any](fn httpx.ServiceFunc[Req, Resp]) gin.HandlerFunc {
	return httpx.Bind(
		fn,
		func(data *Resp) any {
			return data
		},
		failure,
		failure,
	)
}

type errorResponse struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

func failure(err error) (int, any) {
	appError := kratosErrors.FromError(err)
	status := int(appError.Code)
	if status == 0 {
		status = http.StatusInternalServerError
	}
	return status, errorResponse{
		Reason:  appError.Reason,
		Message: appError.Message,
	}
}
