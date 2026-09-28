package data

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"generic-mock/channel/pingpong/biz"

	"github.com/samber/do/v2"
)

type webhookClient struct{ client *http.Client }

var _ biz.WebhookClient = (*webhookClient)(nil)

func NewWebhookClient(do.Injector) (biz.WebhookClient, error) {
	return &webhookClient{client: &http.Client{
		Timeout:       10 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}, nil
}

func (client *webhookClient) Send(ctx context.Context, req *biz.SendWebhookRequest) (*biz.SendWebhookResult, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, req.TargetURL, bytes.NewReader(req.Payload))
	if err != nil {
		return nil, err
	}
	request.Header = req.Headers.Clone()
	response, err := client.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := &biz.SendWebhookResult{StatusCode: response.StatusCode, Headers: response.Header.Clone()}
	body, err := io.ReadAll(io.LimitReader(response.Body, (1<<20)+1))
	if len(body) > 1<<20 {
		result.Body = string(body[:1<<20])
		return result, fmt.Errorf("webhook response exceeds 1 MiB")
	}
	result.Body = string(body)
	return result, err
}
