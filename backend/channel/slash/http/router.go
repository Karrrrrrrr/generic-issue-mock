package http

import (
	"net/http"

	"generic-mock/channel/slash/service"
	"generic-mock/pkg/httpx"
	sharedhttp "generic-mock/shared/http"

	"github.com/gin-gonic/gin"
	kratosErrors "github.com/go-kratos/kratos/v2/errors"
)

type RegisterRequest struct {
	Router  *gin.RouterGroup
	OpenAPI *service.SlashOpenAPIService
	UI      *service.SlashUIService
}

func Register(req RegisterRequest) {
	router := req.Router
	service := req.UI
	openAPIService := req.OpenAPI
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
	sharedhttp.Register(sharedhttp.RegisterRequest{
		Router:                uiRoutes,
		Service:               req.UI.Shared,
		EnableVirtualAccounts: true,
		EnableWebhooks:        true,
	})
	uiRoutes.GET("/authorization-config", bindUI(service.GetAuthorizationConfig))
	uiRoutes.POST("/authorization-config", bindUI(service.UpdateAuthorizationConfig))
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
	return httpx.BindUI(httpx.BindUIRequest[Req, Resp]{
		Service:             fn,
		SuccessEncoder:      func(data *Resp) any { return data },
		BindingErrorEncoder: uiBindingFailure,
		ErrorEncoder:        failure,
	})
}

func uiBindingFailure(err error) (int, any) {
	_, response := failure(err)
	return http.StatusBadRequest, response
}
