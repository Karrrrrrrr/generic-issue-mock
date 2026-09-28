package data

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"generic-mock/channel/slash/biz"

	"github.com/samber/do/v2"
)

type webhookClient struct {
	client *http.Client
}

func NewWebhookClient(_ do.Injector) (biz.SlashWebhookClient, error) {
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
	result := &biz.SlashWebhookDeliveryResult{
		RequestHeaders: requestHeaders,
	}
	response, err := c.client.Do(httpRequest)
	if err != nil {
		return result, err
	}
	defer response.Body.Close()
	result.StatusCode = response.StatusCode
	responseHeaders, err := json.Marshal(response.Header)
	if err != nil {
		return result, err
	}
	result.ResponseHeaders = responseHeaders
	body, err := io.ReadAll(response.Body)
	result.ResponseBody = string(body)
	if err != nil {
		return result, err
	}
	return result, nil
}

var _ biz.SlashWebhookClient = (*webhookClient)(nil)
