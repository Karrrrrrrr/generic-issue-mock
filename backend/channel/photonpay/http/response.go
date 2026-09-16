package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	kratosErrors "github.com/go-kratos/kratos/v2/errors"

	"generic-mock/channel/photonpay/enums"
	"generic-mock/pkg/httpx"
)

const successMessage = "success"

type response[T any] struct {
	Code enums.ResponseCode `json:"code"`
	Msg  string             `json:"msg"`
	Data T                  `json:"data"`
}

func bind[Req any, Resp any](fn httpx.ServiceFunc[Req, Resp]) gin.HandlerFunc {
	return httpx.Bind(fn, success[Resp], bindingFailure, failure)
}

func success[T any](data *T) any {
	return response[T]{
		Code: enums.ResponseCode_Success,
		Msg:  successMessage,
		Data: *data,
	}
}

func failure(err error) (int, any) {
	appError := kratosErrors.FromError(err)
	status := int(appError.Code)
	if status == 0 {
		status = http.StatusInternalServerError
	}

	code := photonResponseCode(status, appError.Reason)

	return status, response[any]{
		Code: code,
		Msg:  appError.Message,
		Data: nil,
	}
}

func photonResponseCode(status int, reason string) enums.ResponseCode {
	code := enums.ResponseCode(reason)
	if code != "" {
		return code
	}
	if status == http.StatusNotFound {
		return enums.ResponseCode_NotFound
	}
	if status == http.StatusBadRequest {
		return enums.ResponseCode_BadInput
	}

	return enums.ResponseCode_InternalError
}

func bindingFailure(err error) (int, any) {
	return http.StatusBadRequest, response[any]{
		Code: enums.ResponseCode_BadInput,
		Msg:  err.Error(),
		Data: nil,
	}
}
