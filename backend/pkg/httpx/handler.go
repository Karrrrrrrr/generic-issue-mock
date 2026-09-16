package httpx

import (
	"context"

	"github.com/gin-gonic/gin"
)

type ServiceFunc[Req any, Resp any] func(context.Context, *Req) (*Resp, error)

type SuccessEncoder[Resp any] func(*Resp) any

type ErrorEncoder func(error) (int, any)

func Bind[Req any, Resp any](
	fn ServiceFunc[Req, Resp],
	successEncoder SuccessEncoder[Resp],
	bindingErrorEncoder ErrorEncoder,
	errorEncoder ErrorEncoder,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var request Req
		if err := ctx.ShouldBind(&request); err != nil {
			status, response := bindingErrorEncoder(err)
			ctx.JSON(status, response)
			return
		}
		if err := ctx.ShouldBindHeader(&request); err != nil {
			status, response := bindingErrorEncoder(err)
			ctx.JSON(status, response)
			return
		}

		response, err := fn(ctx.Request.Context(), &request)
		if err != nil {
			status, errorResponse := errorEncoder(err)
			ctx.JSON(status, errorResponse)
			return
		}

		ctx.JSON(200, successEncoder(response))
	}
}
