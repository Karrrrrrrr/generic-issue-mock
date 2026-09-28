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
		uiRoutes.GET("/accounts", bindUI(service.ListAccounts))
		uiRoutes.POST("/accounts", bindUI(service.CreateAccount))
		uiRoutes.PUT("/accounts/:id", bindUI(service.UpdateAccount))
		uiRoutes.GET("/funds", bindUI(service.ListFunds))
		uiRoutes.POST("/funds/transfer", bindUI(service.MoveFunds))
		uiRoutes.GET("/managed-virtual-accounts", bindUI(service.ListManagedVirtualAccounts))
		uiRoutes.POST("/managed-virtual-accounts", bindUI(service.CreateManagedVirtualAccount))
		uiRoutes.GET("/virtual-accounts", bindUI(service.ListVirtualAccounts))
		uiRoutes.GET("/card-products", bindUI(service.ListCardProducts))
		uiRoutes.GET("/cardholders", bindUI(service.ListCardHolders))
		uiRoutes.GET("/cards", bindUI(service.ListCards))
		uiRoutes.GET("/cards/:id", bindUI(service.GetCard))
		uiRoutes.PUT("/cards/:id/status", bindUI(service.UpdateCardStatus))
		uiRoutes.GET("/authorization-config", bindUI(service.GetAuthorizationConfig))
		uiRoutes.PUT("/authorization-config", bindUI(service.UpdateAuthorizationConfig))
		uiRoutes.GET("/authorizations", bindUI(service.ListAuthorizations))
		uiRoutes.GET("/authorizations/:id", bindUI(service.GetAuthorization))
		uiRoutes.POST("/authorizations/:id/clear", bindUI(service.ClearAuthorization))
		uiRoutes.GET("/authorizations/:id/detail", bindUI(service.GetAuthorizationDetail))
		uiRoutes.POST("/authorizations/:id/reverse", bindUI(service.ReverseAuthorization))
		uiRoutes.POST("/authorizations/:id/refund", bindUI(service.RefundAuthorization))
		uiRoutes.GET("/authorization-balances", bindUI(service.ListAuthorizationBalances))
		uiRoutes.GET("/transactions", bindUI(service.ListTransactions))
		uiRoutes.GET("/transactions/:id", bindUI(service.GetTransaction))
		uiRoutes.POST("/transactions/:id/refund", bindUI(service.RefundTransaction))
		uiRoutes.POST("/transactions/:id/clear", bindUI(service.ClearTransaction))
		uiRoutes.POST("/transactions/:id/reverse", bindUI(service.ReverseTransaction))
		uiRoutes.POST("/simulate/authorizations", bindUI(service.SimulateAuthorization))
		uiRoutes.POST("/simulate/refunds", bindUI(service.SimulateRefund))
		uiRoutes.GET("/webhooks", bindUI(service.ListWebhooks))
		uiRoutes.POST("/webhooks", bindUI(service.CreateWebhook))
		uiRoutes.PUT("/webhooks/:id", bindUI(service.UpdateWebhook))
		uiRoutes.DELETE("/webhooks/:id", bindUI(service.DeleteWebhook))
		uiRoutes.GET("/webhooks/events", bindUI(service.ListWebhookEvents))
		uiRoutes.GET("/webhook-records", bindUI(service.ListWebhookRecords))
		uiRoutes.POST("/webhook-records/:id/replay", bindUI(service.ReplayWebhookRecord))
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

func bindUI[Req any, Resp any](fn httpx.ServiceFunc[Req, Resp]) gin.HandlerFunc {
	return httpx.Bind(
		fn,
		func(data *Resp) any { return data },
		uiBindingFailure,
		failure,
	)
}

func uiBindingFailure(err error) (int, any) {
	_, response := failure(err)
	return http.StatusBadRequest, response
}
