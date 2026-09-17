package http

import (
	"net/http"

	"generic-mock/channel/slash/service"
	"generic-mock/pkg/httpx"

	"github.com/gin-gonic/gin"
	kratosErrors "github.com/go-kratos/kratos/v2/errors"
)

func Register(router *gin.RouterGroup, service *service.SlashUIService, openAPIService *service.SlashOpenAPIService) {
	router.GET("/ui/funds", bind(service.ListFunds))
	router.POST("/ui/funds/transfer", bind(service.MoveFunds))
	router.GET("/ui/managed-virtual-accounts", bind(service.ListManagedVirtualAccounts))
	router.POST("/ui/managed-virtual-accounts", bind(service.CreateManagedVirtualAccount))
	router.GET("/ui/authorization-balances", bind(service.ListAuthorizationBalances))
	router.POST("/ui/authorizations/:id/clear", bind(service.ClearAuthorization))
	router.GET("/ui/accounts", bind(service.ListAccounts))
	router.POST("/ui/accounts", bind(service.CreateAccount))
	router.PUT("/ui/accounts/:id", bind(service.UpdateAccount))
	router.GET("/ui/authorization-config", bind(service.GetAuthorizationConfig))
	router.PUT("/ui/authorization-config", bind(service.UpdateAuthorizationConfig))
	router.GET("/card", bind(openAPIService.ListCards))
	router.POST("/card", bind(openAPIService.CreateCard))
	router.GET("/card/:id", bind(openAPIService.GetCard))
	router.PATCH("/card/:id", bind(openAPIService.UpdateCard))
	router.GET("/card-product", bind(openAPIService.ListCardProducts))
	router.GET("/virtual-account", bind(openAPIService.ListVirtualAccounts))
	router.POST("/transfer/virtual-account", bind(openAPIService.TransferVirtualAccount))
	router.GET("/transaction", bind(openAPIService.ListTransactions))
	router.GET("/transaction/:id", bind(openAPIService.GetTransaction))

	router.GET("/ui/card-products", bind(service.ListCardProducts))
	router.GET("/ui/virtual-accounts", bind(service.ListVirtualAccounts))
	router.GET("/ui/cardholders", bind(service.ListCardHolders))
	router.GET("/ui/cards", bind(service.ListCards))
	router.GET("/ui/cards/:id", bind(service.GetCard))
	router.PUT("/ui/cards/:id/status", bind(service.UpdateCardStatus))
	router.POST("/ui/simulate/authorizations", bind(service.SimulateAuthorization))
	router.POST("/ui/simulate/refunds", bind(service.SimulateRefund))
	router.GET("/ui/authorizations", bind(service.ListAuthorizations))
	router.GET("/ui/authorizations/:id", bind(service.GetAuthorization))
	router.GET("/ui/transactions", bind(service.ListTransactions))
	router.GET("/ui/transactions/:id", bind(service.GetTransaction))
	router.POST("/ui/transactions/:id/reverse", bind(service.ReverseTransaction))
	router.POST("/ui/transactions/:id/refund", bind(service.RefundTransaction))
	router.GET("/ui/webhooks", bind(service.ListWebhooks))
	router.GET("/ui/webhooks/events", bind(service.ListWebhookEvents))
	router.POST("/ui/webhooks", bind(service.CreateWebhook))
	router.PUT("/ui/webhooks/:id", bind(service.UpdateWebhook))
	router.DELETE("/ui/webhooks/:id", bind(service.DeleteWebhook))
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
