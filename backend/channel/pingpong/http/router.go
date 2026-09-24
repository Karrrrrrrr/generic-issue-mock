package http

import (
	pingerrors "generic-mock/channel/pingpong/errors"
	"generic-mock/channel/pingpong/service"
	"generic-mock/pkg/httpx"

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
	{
		api := req.OpenAPI
		router.GET("/v2/token/get", bind(api.Token))
		router.GET("/api/issuing/v3/card-products", bind(api.Products))
		router.POST("/api/issuing/v3/budgets", bind(api.CreateBudget))
		router.GET("/api/issuing/v3/budget/balance", bind(api.BudgetBalances))
		router.POST("/api/issuing/v3/budgets/funding", bind(api.BudgetFunding))
		router.GET("/api/issuing/v3/funding/orders", bind(api.BudgetOrder))
		router.POST("/api/issuing/v3/cards/apply", bind(api.CreateCard))
		router.GET("/api/issuing/v3/cards/details", bind(api.CardDetails))
		router.GET("/api/issuing/v3/card/balance", bind(api.CardBalance))
		router.POST("/api/issuing/v3/cards/actions", bind(api.CardAction))
		router.POST("/api/issuing/v3/cards/funding/actions", bind(api.CardFunding))
		router.GET("/api/issuing/v3/card/funding/orders", bind(api.CardOrders))
		router.POST("/api/reporting/v3/account-balance", bind(api.Unsupported))
		router.GET("/api/issuing/v3/transactions", bind(api.Unsupported))
		router.GET("/api/issuing/v4/account/transactions", bind(api.Unsupported))
		router.GET("/api/issuing/v3/cards/3ds/details", bind(api.Unsupported))
	}
	{
		ui := router.Group("/ui")
		ui.GET("/accounts", bindUI(req.UI.Accounts))
		ui.POST("/accounts", bindUI(req.UI.CreateAccount))
		ui.POST("/accounts/:id/balance", bindUI(req.UI.AdjustAccount))
		ui.GET("/virtual-accounts", bindUI(req.UI.VirtualAccounts))
		ui.POST("/virtual-accounts", bindUI(req.UI.CreateVirtualAccount))
		ui.POST("/virtual-accounts/:id/fund", bindUI(req.UI.FundVirtualAccount))
		ui.GET("/products", bindUI(req.UI.Products))
		ui.GET("/cards", bindUI(req.UI.Cards))
		ui.PUT("/cards/:id/status", bindUI(req.UI.ChangeCard))
		ui.POST("/cards/:id/fund", bindUI(req.UI.FundCard))
		ui.GET("/transfers", bindUI(req.UI.Transfers))
		ui.GET("/authorizations", bindUI(req.UI.Authorizations))
		ui.POST("/simulate/authorizations", bindUI(req.UI.SimulateAuthorization))
		ui.POST("/authorizations/:id/stages", bindUI(req.UI.Stage))
	}
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
	return httpx.Bind(fn, func(data *Resp) any { return data }, badRequest, failure)
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
