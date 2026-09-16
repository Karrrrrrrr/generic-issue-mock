package http

import (
	"net/http"

	"generic-mock/channel/paynda/service"
	"generic-mock/pkg/httpx"

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
	return httpx.Bind(
		fn,
		func(data *Resp) any {
			return data
		},
		uiFailure,
		uiFailure,
	)
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

func Register(router *gin.RouterGroup, openapi *service.PayndaOpenAPIService, ui *service.PayndaUIService) {
	router.GET("/ui/cardholders", bindUI(ui.ListCardHolders))
	router.POST("/ui/cardholders", bindUI(ui.CreateCardHolder))
	router.GET("/ui/cards", bindUI(ui.ListCards))
	router.POST("/ui/cards", bindUI(ui.CreateCard))
	router.PUT("/ui/cards/:id/status", bindUI(ui.UpdateCardStatus))
	router.POST("/ui/simulate/authorizations", bindUI(ui.SimulateAuthorization))
	router.POST("/ui/simulate/refunds", bindUI(ui.SimulateRefund))
	router.GET("/ui/authorizations", bindUI(ui.ListAuthorizations))
	router.GET("/ui/transactions", bindUI(ui.ListTransactions))
	router.POST("/ui/transactions/:id/clear", bindUI(ui.ClearTransaction))
	router.POST("/ui/transactions/:id/reverse", bindUI(ui.ReverseTransaction))
	router.POST("/ui/transactions/:id/refund", bindUI(ui.RefundTransaction))
	router.GET("/openapi/merchant/wallets", bind(openapi.ListMerchantWallets))
	router.POST("/openapi/balanceAccountWalletTransfers", bind(openapi.TransferBalanceAccountWallet))
	router.POST("/openapi/balanceAccounts/:balanceAccountId/cardholders", bind(openapi.CreateCardHolder))
	router.GET("/openapi/balanceAccounts/:balanceAccountId/cardholders", bind(openapi.ListCardHolders))
	router.GET("/openapi/balanceAccounts/:balanceAccountId/cardholders/:cardholderId", bind(openapi.GetCardHolder))
	router.PUT("/openapi/balanceAccounts/:balanceAccountId/cardholders/:cardholderId", bind(openapi.UpdateCardHolder))
	router.GET("/openapi/balanceAccounts/:balanceAccountId/cardBins", bind(openapi.ListCardBins))
	router.GET("/openapi/balanceAccounts/:balanceAccountId/wallets", bind(openapi.ListBalanceAccountWallets))
	router.POST("/openapi/balanceAccounts/:balanceAccountId/cards", bind(openapi.CreateCard))
	router.GET("/openapi/balanceAccounts/:balanceAccountId/cards", bind(openapi.ListCards))
	router.GET("/openapi/balanceAccounts/:balanceAccountId/cards/:cardId", bind(openapi.GetCard))
	router.GET("/openapi/balanceAccounts/:balanceAccountId/cards/:cardId/sensitiveInfo", bind(openapi.GetCardSensitive))
	router.GET("/openapi/balanceAccounts/:balanceAccountId/cards/:cardId/balance", bind(openapi.GetCardBalance))
	router.PATCH("/openapi/balanceAccounts/:balanceAccountId/cards/:cardId/status/frozen", bind(openapi.FreezeCard))
	router.PATCH("/openapi/balanceAccounts/:balanceAccountId/cards/:cardId/status/unfrozen", bind(openapi.UnfreezeCard))
	router.PATCH("/openapi/balanceAccounts/:balanceAccountId/cards/:cardId/status/release", bind(openapi.ReleaseCard))
	router.POST("/openapi/balanceAccounts/:balanceAccountId/cardBalanceTransfers", bind(openapi.TransferCardBalance))
	router.GET("/openapi/balanceAccounts/:balanceAccountId/cardBalanceUpdates", bind(openapi.ListCardBalanceUpdates))
	router.GET("/openapi/balanceAccounts/:balanceAccountId/transactions", bind(openapi.ListCardTransactions))
	router.GET("/openapi/balanceAccounts/:balanceAccountId/transactions/:transactionId", bind(openapi.GetCardTransaction))
	router.GET("/openapi/requestResults", bind(openapi.RequestResult))
}
