package httpx

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ServiceFunc[Req any, Resp any] func(context.Context, *Req) (*Resp, error)

type Error interface {
	error
	HTTPStatus() int
}

func Bind[Req any, Resp any](fn ServiceFunc[Req, Resp]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var request Req
		if err := ctx.ShouldBind(&request); err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse{
				Code:    "4000",
				Message: err.Error(),
			})
			return
		}

		response, err := fn(ctx.Request.Context(), &request)
		if err != nil {
			status := http.StatusInternalServerError
			if appError, ok := err.(Error); ok {
				status = appError.HTTPStatus()
			}
			ctx.JSON(status, errorResponse{
				Code:    "5000",
				Message: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, response)
	}
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"msg"`
	Data    any    `json:"data"`
}
