package http

import (
	pingerrors "generic-mock/channel/pingpong/errors"
	"generic-mock/channel/pingpong/service"
	"generic-mock/pkg/httpx"
	sharedhttp "generic-mock/shared/http"

	"github.com/gin-gonic/gin"
	kratoserrors "github.com/go-kratos/kratos/v2/errors"
)

type RegisterRequest struct {
	Router  *gin.RouterGroup
	OpenAPI *service.PingPongOpenAPIService
	UI      *service.PingPongUIService
}

func Register(req RegisterRequest) {
	router := req.Router
	openAPIRoutes := router.Group("")
	{
		api := req.OpenAPI
		openAPIRoutes.GET("/v2/token/get", bind(api.GetAccessToken))
		openAPIRoutes.GET("/api/issuing/v3/card-products", bind(api.ListCardProducts))
		openAPIRoutes.POST("/api/issuing/v3/budgets", bind(api.CreateBudget))
		openAPIRoutes.GET("/api/issuing/v3/budget/balance", bind(api.ListBudgetBalances))
		openAPIRoutes.POST("/api/issuing/v3/budgets/funding", bind(api.FundBudget))
		openAPIRoutes.GET("/api/issuing/v3/funding/orders", bind(api.GetBudgetFundingOrder))
		openAPIRoutes.POST("/api/issuing/v3/cards/apply", bind(api.CreateCard))
		openAPIRoutes.GET("/api/issuing/v3/cards/details", bind(api.GetCardDetails))
		openAPIRoutes.GET("/api/issuing/v3/card/balance", bind(api.GetCardBalance))
		openAPIRoutes.POST("/api/issuing/v3/cards/actions", bind(api.ApplyCardAction))
		openAPIRoutes.POST("/api/issuing/v3/cards/funding/actions", bind(api.FundCard))
		openAPIRoutes.GET("/api/issuing/v3/card/funding/orders", bind(api.ListCardFundingOrders))
		openAPIRoutes.POST("/api/reporting/v3/account-balance", bind(api.RejectUnsupportedOperation))
		openAPIRoutes.GET("/api/issuing/v3/transactions", bind(api.RejectUnsupportedOperation))
		openAPIRoutes.GET("/api/issuing/v4/account/transactions", bind(api.RejectUnsupportedOperation))
		openAPIRoutes.GET("/api/issuing/v3/cards/3ds/details", bind(api.RejectUnsupportedOperation))
	}
	uiRoutes := router.Group("/ui")
	sharedhttp.Register(sharedhttp.RegisterRequest{
		Router:                uiRoutes,
		Service:               req.UI.Shared,
		EnableVirtualAccounts: true,
		EnableWebhooks:        true,
	})
	uiRoutes.POST("/webhooks/dispatch", bindUI(req.UI.DispatchWebhook))
}

type envelope[Item any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *Item  `json:"data"`
}

type errorDetails struct {
	Reason string `json:"reason"`
}

type failureData struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details errorDetails `json:"details"`
}

func bind[Req any, Resp any](fn httpx.ServiceFunc[Req, Resp]) gin.HandlerFunc {
	return httpx.Bind(fn, func(data *Resp) any {
		return envelope[Resp]{
			Code:    0,
			Message: "success",
			Data:    data,
		}
	}, badRequest, failure)
}

func bindUI[Req any, Resp any](fn httpx.ServiceFunc[Req, Resp]) gin.HandlerFunc {
	return httpx.BindUI(httpx.BindUIRequest[Req, Resp]{
		Service:             fn,
		SuccessEncoder:      func(data *Resp) any { return data },
		BindingErrorEncoder: badRequest,
		ErrorEncoder:        failure,
	})
}

func badRequest(err error) (int, any) { return failure(pingerrors.ErrInvalid) }
func failure(err error) (int, any) {
	converted := kratoserrors.FromError(err)
	return int(converted.Code), failureData{
		Code:    converted.Reason,
		Message: converted.Message,
		Details: errorDetails{Reason: converted.Reason},
	}
}
