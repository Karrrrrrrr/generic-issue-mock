package http

import (
	"net/http"

	"generic-mock/channel/slash/service"
	"generic-mock/pkg/httpx"

	"github.com/gin-gonic/gin"
	kratosErrors "github.com/go-kratos/kratos/v2/errors"
)

func Register(router *gin.RouterGroup, service *service.Service) {
	router.GET("/card-product", bind(service.ListCardProducts))
	router.GET("/ui/card-products", bind(service.ListCardProducts))
	router.GET("/ui/cardholders", bind(service.ListCardHolders))
	router.POST("/ui/cardholders", bind(service.CreateCardHolder))
	router.GET("/ui/cards", bind(service.ListCards))
	router.GET("/ui/cards/:id", bind(service.GetCard))
	router.POST("/ui/cards", bind(service.CreateCard))
	router.PUT("/ui/cards/:id/status", bind(service.UpdateCardStatus))
	router.POST("/ui/simulate/authorizations", bind(service.SimulateAuthorization))
	router.GET("/ui/authorizations", bind(service.ListAuthorizations))
	router.GET("/ui/authorizations/:id", bind(service.GetAuthorization))
	router.GET("/ui/transactions", bind(service.ListTransactions))
	router.GET("/ui/transactions/:id", bind(service.GetTransaction))
	router.POST("/ui/transactions/:id/clear", bind(service.ClearTransaction))
	router.POST("/ui/transactions/:id/reverse", bind(service.ReverseTransaction))
	router.POST("/ui/transactions/:id/refund", bind(service.RefundTransaction))
}

func bind[Req any, Resp any](fn httpx.ServiceFunc[Req, Resp]) gin.HandlerFunc {
	return httpx.Bind(fn, func(data *Resp) any { return data }, failure, failure)
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
