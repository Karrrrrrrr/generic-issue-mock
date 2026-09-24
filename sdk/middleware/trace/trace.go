package trace

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"resty.dev/v3"
)

type trackingKey struct{}

// RequestTracking 请求追踪
// 配合 ResponseTracking 使用
func RequestTracking() func(c *resty.Client, r *resty.Request) error {
	return func(c *resty.Client, r *resty.Request) error {
		spanName := fmt.Sprintf(`%s %s%s`, r.Method, r.RawRequest.Host, r.RawRequest.URL.Path)
		ctx, span := otel.Tracer("").Start(r.Context(), spanName)
		r.SetContext(context.WithValue(ctx, trackingKey{}, span))
		return nil
	}
}

// ResponseTracking 响应追踪
// 配合 RequestTracking 使用
func ResponseTracking() func(c *resty.Client, r *resty.Response) error {
	return func(c *resty.Client, r *resty.Response) error {
		span, _ := r.Request.Context().Value(trackingKey{}).(trace.Span)
		if span == nil {
			return nil
		}
		defer span.End()

		// 如果是get请求的话记录一下query参数
		if r.Request.Method == http.MethodGet && len(r.Request.QueryParams) > 0 {
			_, querySpan := otel.Tracer("").Start(r.Request.Context(), "http.request.query")
			for k, v := range r.Request.QueryParams {
				querySpan.SetAttributes(attribute.StringSlice(k, v))
			}
			defer querySpan.End()
		}

		return nil
	}
}
