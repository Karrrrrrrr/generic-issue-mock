package biz

import (
	"context"
	"net/http"
)

type SendWebhookRequest struct {
	TargetURL string
	Payload   []byte
	Headers   http.Header
}

type SendWebhookResult struct {
	StatusCode int
	Headers    http.Header
	Body       string
}

type WebhookClient interface {
	Send(context.Context, *SendWebhookRequest) (*SendWebhookResult, error)
}
