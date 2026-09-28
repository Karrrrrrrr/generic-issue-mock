package http

import (
	"net/http"

	"generic-mock/channel/paynda/service"
	"generic-mock/pkg/httpx"
	sharedhttp "generic-mock/shared/http"

	"github.com/gin-gonic/gin"
	kratosErrors "github.com/go-kratos/kratos/v2/errors"
)

const (
	payndaSuccessCode    = 200
	payndaSuccessMessage = "success"
)

type response[T any] struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
	Success bool   `json:"success"`
}

func bind[Req any, Resp any](fn httpx.ServiceFunc[Req, Resp]) gin.HandlerFunc {
	return httpx.Bind(
		fn,
		func(data *Resp) any {
			return response[Resp]{
				Code:    payndaSuccessCode,
				Message: payndaSuccessMessage,
				Data:    *data,
				Success: true,
			}
		},
		failure,
		failure,
	)
}

func bindUI[Req any, Resp any](fn httpx.ServiceFunc[Req, Resp]) gin.HandlerFunc {
	return httpx.BindUI(httpx.BindUIRequest[Req, Resp]{
		Service:             fn,
		SuccessEncoder:      func(data *Resp) any { return data },
		BindingErrorEncoder: uiBindingFailure,
		ErrorEncoder:        uiFailure,
	})
}

func failure(err error) (int, any) {
	app := kratosErrors.FromError(err)
	status := int(app.Code)
	if status == 0 {
		status = http.StatusInternalServerError
	}
	return status, response[any]{
		Code:    int64(status),
		Message: app.Message,
		Success: false,
	}
}

func uiFailure(err error) (int, any) {
	app := kratosErrors.FromError(err)
	status := int(app.Code)
	if status == 0 {
		status = http.StatusInternalServerError
	}
	return status, struct {
		Reason  string `json:"reason"`
		Message string `json:"message"`
	}{
		Reason:  app.Reason,
		Message: app.Message,
	}
}

type RegisterRequest struct {
	Router  *gin.RouterGroup
	OpenAPI *service.PayndaOpenAPIService
	UI      *service.PayndaUIService
}

func Register(req RegisterRequest) {
	router := req.Router
	openapi := req.OpenAPI
	openAPIRoutes := router.Group("/openapi")
	{
		openAPIRoutes.GET("/merchant/wallets", bind(openapi.ListMerchantWallets))
		openAPIRoutes.GET("/balanceAccounts", bind(openapi.ListBalanceAccounts))
		openAPIRoutes.POST("/balanceAccounts", bind(openapi.CreateBalanceAccount))
		openAPIRoutes.GET("/balanceAccounts/:balanceAccountId", bind(openapi.GetBalanceAccount))
		openAPIRoutes.PUT("/balanceAccounts/:balanceAccountId", bind(openapi.UpdateBalanceAccount))
		openAPIRoutes.DELETE("/balanceAccounts/:balanceAccountId", bind(openapi.DeleteBalanceAccount))
		openAPIRoutes.GET("/balanceAccounts/:balanceAccountId/wallets", bind(openapi.ListBalanceAccountWallets))
		openAPIRoutes.GET("/balanceAccounts/:balanceAccountId/cardBins", bind(openapi.ListCardBins))
		openAPIRoutes.GET("/balanceAccounts/:balanceAccountId/cardholders", bind(openapi.ListCardHolders))
		openAPIRoutes.POST("/balanceAccounts/:balanceAccountId/cardholders", bind(openapi.CreateCardHolder))
		openAPIRoutes.GET("/balanceAccounts/:balanceAccountId/cardholders/:cardholderId", bind(openapi.GetCardHolder))
		openAPIRoutes.PUT("/balanceAccounts/:balanceAccountId/cardholders/:cardholderId", bind(openapi.UpdateCardHolder))
		openAPIRoutes.DELETE("/balanceAccounts/:balanceAccountId/cardholders/:cardholderId", bind(openapi.DeleteCardholder))
		openAPIRoutes.GET("/balanceAccounts/:balanceAccountId/cardholders/:cardholderId/wallets", bind(openapi.ListCardholderWallets))
		openAPIRoutes.POST("/balanceAccounts/:balanceAccountId/cardholderWalletUpdates", bind(openapi.CardholderWallet))
		openAPIRoutes.GET("/balanceAccounts/:balanceAccountId/cards", bind(openapi.ListCards))
		openAPIRoutes.POST("/balanceAccounts/:balanceAccountId/cards", bind(openapi.CreateCard))
		openAPIRoutes.GET("/balanceAccounts/:balanceAccountId/cards/:cardId", bind(openapi.GetCard))
		openAPIRoutes.GET("/balanceAccounts/:balanceAccountId/cards/:cardId/balance", bind(openapi.GetCardBalance))
		openAPIRoutes.GET("/balanceAccounts/:balanceAccountId/cards/:cardId/controls", bind(openapi.GetCardControls))
		openAPIRoutes.PATCH("/balanceAccounts/:balanceAccountId/cards/:cardId/controls", bind(openapi.UpdateCardControls))
		openAPIRoutes.GET("/balanceAccounts/:balanceAccountId/cards/:cardId/sensitiveInfo", bind(openapi.GetCardSensitive))
		openAPIRoutes.PATCH("/balanceAccounts/:balanceAccountId/cards/:cardId/status/frozen", bind(openapi.FreezeCard))
		openAPIRoutes.PATCH("/balanceAccounts/:balanceAccountId/cards/:cardId/status/release", bind(openapi.ReleaseCard))
		openAPIRoutes.PATCH("/balanceAccounts/:balanceAccountId/cards/:cardId/status/unfrozen", bind(openapi.UnfreezeCard))
		openAPIRoutes.POST("/balanceAccounts/:balanceAccountId/cardBalanceTransfers", bind(openapi.TransferCardBalance))
		openAPIRoutes.GET("/balanceAccounts/:balanceAccountId/cardBalanceUpdates", bind(openapi.ListCardBalanceUpdates))
		openAPIRoutes.POST("/balanceAccounts/:balanceAccountId/cardBalanceUpdates", bind(openapi.UpdateCardBalance))
		openAPIRoutes.GET("/balanceAccounts/:balanceAccountId/transactions", bind(openapi.ListCardTransactions))
		openAPIRoutes.GET("/balanceAccounts/:balanceAccountId/transactions/:transactionId", bind(openapi.GetCardTransaction))
		openAPIRoutes.POST("/balanceAccountWalletTransfers", bind(openapi.TransferBalanceAccountWallet))
		openAPIRoutes.GET("/requestResults", bind(openapi.RequestResult))
	}

	uiRoutes := router.Group("/ui")
	sharedhttp.Register(sharedhttp.RegisterRequest{
		Router:                uiRoutes,
		Service:               req.UI.Shared,
		EnableVirtualAccounts: false,
		EnableWebhooks:        true,
	})
}

func uiBindingFailure(err error) (int, any) {
	_, response := uiFailure(err)
	return http.StatusBadRequest, response
}
