package http

import (
	"generic-mock/channel/photonpay/service"
	"generic-mock/pkg/httpx"

	"github.com/gin-gonic/gin"
)

func Register(router *gin.RouterGroup, service *service.Service) {
	router.POST("/oauth2/token/accessToken", httpx.Bind(service.AccessToken))
	router.GET("/wallet/openApi/v4/account/single", httpx.Bind(service.AccountSingle))
	router.POST("/vcc/openApi/v4/addCardholder", httpx.Bind(service.CreateCardHolder))
	router.POST("/vcc/openApi/v4/editCardholder", httpx.Bind(service.EditCardHolder))
	router.GET("/vcc/openApi/v4/pagingVccCardholder", httpx.Bind(service.ListCardHolders))
	router.GET("/vcc/openApi/v4/getCardBin", httpx.Bind(service.CardBins))
	router.POST("/vcc/openApi/v4/openCard", httpx.Bind(service.OpenCard))
	router.GET("/vcc/openApi/v4/getRequestResult", httpx.Bind(service.RequestResult))
	router.GET("/vcc/openApi/v4/getCardDetail", httpx.Bind(service.CardDetail))
	router.GET("/vcc/openApi/v4/getCvv", httpx.Bind(service.CardCVV))
	router.POST("/vcc/openApi/v4/freezeCard", httpx.Bind(service.FreezeCard))
	router.POST("/vcc/openApi/v4/cancelCard", httpx.Bind(service.CancelCard))
	router.GET("/vcc/openApi/v4/pagingVccTradeOrder", httpx.Bind(service.ListTrades))
	router.POST("/file/apiUpload/:businessKey", httpx.Bind(service.Upload))
}
