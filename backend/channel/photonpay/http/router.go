package http

import (
	"generic-mock/channel/photonpay/service"

	"github.com/gin-gonic/gin"
)

func Register(router *gin.RouterGroup, service *service.Service) {
	router.POST("/oauth2/token/accessToken", bind(service.AccessToken))
	router.GET("/wallet/openApi/v4/account/single", bind(service.AccountSingle))
	router.POST("/vcc/openApi/v4/addCardholder", bind(service.CreateCardHolder))
	router.POST("/vcc/openApi/v4/editCardholder", bind(service.EditCardHolder))
	router.GET("/vcc/openApi/v4/pagingVccCardholder", bind(service.ListCardHolders))
	router.GET("/vcc/openApi/v4/getCardBin", bind(service.CardBins))
	router.POST("/vcc/openApi/v4/openCard", bind(service.OpenCard))
	router.GET("/vcc/openApi/v4/getRequestResult", bind(service.RequestResult))
	router.GET("/vcc/openApi/v4/getCardDetail", bind(service.CardDetail))
	router.GET("/vcc/openApi/v4/getCvv", bind(service.CardCVV))
	router.POST("/vcc/openApi/v4/freezeCard", bind(service.FreezeCard))
	router.POST("/vcc/openApi/v4/cancelCard", bind(service.CancelCard))
	router.GET("/vcc/openApi/v4/pagingVccTradeOrder", bind(service.ListTrades))
	router.POST("/file/apiUpload/:businessKey", bind(service.Upload))
}
