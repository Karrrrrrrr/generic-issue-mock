package httpx

import (
	"context"
	"encoding/json"
	"io"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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
		defer func() {
			if ctx.Request.MultipartForm != nil {
				_ = ctx.Request.MultipartForm.RemoveAll()
			}
		}()
		var request Req
		if err := bindRequest(&bindRequestInput{
			Context: ctx,
			Request: &request,
		}); err != nil {
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

type bindRequestInput struct {
	Context *gin.Context
	Request any
}

func bindRequest(input *bindRequestInput) error {
	ctx := input.Context
	request := input.Request
	if ctx.ContentType() == binding.MIMEJSON && ctx.Request.Body != nil {
		if err := json.NewDecoder(ctx.Request.Body).Decode(request); err != nil && err != io.EOF {
			return err
		}
	} else if ctx.ContentType() == binding.MIMEMultipartPOSTForm {
		if err := ctx.Request.ParseMultipartForm(32 << 20); err != nil {
			return err
		}
	} else if err := ctx.Request.ParseForm(); err != nil {
		return err
	}
	if err := binding.MapFormWithTag(request, ctx.Request.URL.Query(), "form"); err != nil {
		return err
	}
	if err := binding.MapFormWithTag(request, ctx.Request.PostForm, "form"); err != nil {
		return err
	}
	params := make(map[string][]string, len(ctx.Params))
	for _, param := range ctx.Params {
		params[param.Key] = []string{param.Value}
	}
	if err := binding.MapFormWithTag(request, params, "uri"); err != nil {
		return err
	}
	headers := make(map[string][]string)
	collectHeaders(&collectHeadersInput{
		Context:     ctx,
		RequestType: reflect.TypeOf(request),
		Headers:     headers,
	})
	if err := binding.MapFormWithTag(request, headers, "header"); err != nil {
		return err
	}
	if ctx.ContentType() == binding.MIMEMultipartPOSTForm {
		return binding.FormMultipart.Bind(ctx.Request, request)
	}
	return binding.Validator.ValidateStruct(request)
}

type collectHeadersInput struct {
	Context     *gin.Context
	RequestType reflect.Type
	Headers     map[string][]string
}

func collectHeaders(input *collectHeadersInput) {
	ctx := input.Context
	requestType := input.RequestType
	headers := input.Headers
	if requestType.Kind() == reflect.Pointer {
		requestType = requestType.Elem()
	}
	if requestType.Kind() != reflect.Struct {
		return
	}
	for index := 0; index < requestType.NumField(); index++ {
		field := requestType.Field(index)
		if field.Anonymous {
			collectHeaders(&collectHeadersInput{
				Context:     ctx,
				RequestType: field.Type,
				Headers:     headers,
			})
		}
		if tag := field.Tag.Get("header"); tag != "" && tag != "-" {
			if values := ctx.Request.Header.Values(tag); len(values) > 0 {
				headers[tag] = values
			}
		}
	}
}
