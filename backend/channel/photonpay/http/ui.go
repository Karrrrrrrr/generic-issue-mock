package http

import (
	"net/http"

	"generic-mock/pkg/httpx"

	"github.com/gin-gonic/gin"
	kratosErrors "github.com/go-kratos/kratos/v2/errors"
)

func bindUI[Req any, Resp any](fn httpx.ServiceFunc[Req, Resp]) gin.HandlerFunc {
	return httpx.Bind(fn, func(data *Resp) any { return data }, uiFailure, uiFailure)
}

type uiErrorResponse struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

func uiFailure(err error) (int, any) {
	appError := kratosErrors.FromError(err)
	status := int(appError.Code)
	if status == 0 {
		status = http.StatusInternalServerError
	}

	return status, uiErrorResponse{
		Reason:  appError.Reason,
		Message: appError.Message,
	}
}
