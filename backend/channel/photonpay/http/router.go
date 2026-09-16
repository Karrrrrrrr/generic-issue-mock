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
	router.GET("/ui/cardholders", bindUI(uiService.ListCardHolders))
	router.POST("/ui/cardholders", bindUI(uiService.CreateCardHolder))
	router.GET("/ui/cards", bindUI(uiService.ListCards))
	router.POST("/ui/cards", bindUI(uiService.CreateCard))
	router.PUT("/ui/cards/:id/status", bindUI(uiService.UpdateCardStatus))
	router.POST("/ui/simulate/authorizations", bindUI(uiService.SimulateAuthorization))
	router.GET("/ui/transactions", bindUI(uiService.ListTransactions))
	router.POST("/ui/transactions/:id/clear", bindUI(uiService.ClearTransaction))
	router.POST("/ui/transactions/:id/reverse", bindUI(uiService.ReverseTransaction))
	router.POST("/ui/transactions/:id/refund", bindUI(uiService.RefundTransaction))

	router.POST("/oauth2/token/accessToken", bind(openAPIService.AccessToken))
	router.GET("/wallet/openApi/v4/account/single", bind(openAPIService.AccountSingle))
	router.POST("/vcc/openApi/v4/addCardholder", bind(openAPIService.CreateCardHolder))
	router.POST("/vcc/openApi/v4/editCardholder", bind(openAPIService.EditCardHolder))
	router.GET("/vcc/openApi/v4/pagingVccCardholder", bind(openAPIService.ListCardHolders))
	router.GET("/vcc/openApi/v4/getCardBin", bind(openAPIService.CardBins))
	router.POST("/vcc/openApi/v4/openCard", bind(openAPIService.OpenCard))
	router.GET("/vcc/openApi/v4/getRequestResult", bind(openAPIService.RequestResult))
	router.GET("/vcc/openApi/v4/getCardDetail", bind(openAPIService.CardDetail))
	router.GET("/vcc/openApi/v4/getCvv", bind(openAPIService.CardCVV))
	router.POST("/vcc/openApi/v4/freezeCard", bind(openAPIService.FreezeCard))
	router.POST("/vcc/openApi/v4/cancelCard", bind(openAPIService.CancelCard))
	router.GET("/vcc/openApi/v4/pagingVccTradeOrder", bind(openAPIService.ListTrades))
	router.POST("/file/apiUpload/:businessKey", bind(openAPIService.Upload))
}
