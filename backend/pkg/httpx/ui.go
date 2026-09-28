package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type BindUIRequest[Req any, Resp any] struct {
	Service             ServiceFunc[Req, Resp]
	SuccessEncoder      SuccessEncoder[Resp]
	BindingErrorEncoder ErrorEncoder
	ErrorEncoder        ErrorEncoder
}

func BindUI[Req any, Resp any](input BindUIRequest[Req, Resp]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var request Req
		var err error
		if ctx.Request.Method == http.MethodGet {
			err = ctx.ShouldBindQuery(&request)
		} else {
			err = ctx.ShouldBindJSON(&request)
		}
		if err != nil {
			status, response := input.BindingErrorEncoder(err)
			ctx.JSON(status, response)
			return
		}
		response, err := input.Service(ctx.Request.Context(), &request)
		if err != nil {
			status, response := input.ErrorEncoder(err)
			ctx.JSON(status, response)
			return
		}
		ctx.JSON(http.StatusOK, input.SuccessEncoder(response))
	}
}
