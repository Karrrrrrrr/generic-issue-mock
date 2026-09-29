package data

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"generic-mock/channel/photonpay/biz"

	"github.com/samber/do/v2"
)

type authorizationClient struct {
	client     *http.Client
	privateKey string
}

func NewAuthorizationClient(_ do.Injector) (biz.AuthorizationClient, error) {
	return &authorizationClient{
		client:     &http.Client{Timeout: 10 * time.Second},
		privateKey: os.Getenv("PHOTONPAY_WEBHOOK_PRIVATE_KEY"),
	}, nil
}

func (c *authorizationClient) RequestAuthorization(
	ctx context.Context,
	req *biz.AuthorizationRequestDeliveryRequest,
) (*biz.AuthorizationRequestDeliveryResult, error) {
	if req.TimeoutMillis > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(
			ctx,
			time.Duration(req.TimeoutMillis)*time.Millisecond,
		)
		defer cancel()
	}

	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		req.TargetURL,
		bytes.NewReader(req.Payload),
	)
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	if c.privateKey != "" {
		signature, err := photonPayWebhookSignature(req.Payload, c.privateKey)
		if err != nil {
			return nil, err
		}
		httpRequest.Header.Set("X-PD-SIGN", signature)
	}

	requestHeaders, err := json.Marshal(httpRequest.Header)
	if err != nil {
		return nil, err
	}
	result := &biz.AuthorizationRequestDeliveryResult{
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

var _ biz.AuthorizationClient = (*authorizationClient)(nil)
