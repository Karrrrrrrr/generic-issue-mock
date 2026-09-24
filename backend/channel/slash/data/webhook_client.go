package data

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"generic-mock/channel/slash/biz"

	"github.com/samber/do"
)

type webhookClient struct {
	client *http.Client
}

func NewWebhookClient(_ *do.Injector) (biz.SlashWebhookClient, error) {
	return &webhookClient{
		client: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *webhookClient) Deliver(ctx context.Context, req *biz.SlashWebhookDeliveryRequest) (*biz.SlashWebhookDeliveryResult, error) {
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, req.TargetURL, bytes.NewReader(req.Payload))
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	if len(req.RequestHeaders) != 0 {
		if err := json.Unmarshal(req.RequestHeaders, &httpRequest.Header); err != nil {
			return nil, err
		}
	}
	requestHeaders, err := json.Marshal(httpRequest.Header)
	if err != nil {
		return nil, err
	}
	response, err := c.client.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	responseHeaders, err := json.Marshal(response.Header)
	if err != nil {
		return nil, err
	}
	return &biz.SlashWebhookDeliveryResult{
		StatusCode:      response.StatusCode,
		ResponseBody:    string(responseBody),
		RequestHeaders:  requestHeaders,
		ResponseHeaders: responseHeaders,
	}, nil
}

var _ biz.SlashWebhookClient = (*webhookClient)(nil)
