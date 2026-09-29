package data

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"generic-mock/channel/paynda/biz"
	"generic-mock/pkg/randomx"

	"github.com/samber/do/v2"
)

type webhookClient struct {
	client    *http.Client
	appID     string
	appSecret string
}

func NewWebhookClient(_ do.Injector) (biz.PayndaWebhookClient, error) {
	return &webhookClient{
		client:    &http.Client{Timeout: 10 * time.Second},
		appID:     os.Getenv("PAYNDA_WEBHOOK_APP_ID"),
		appSecret: os.Getenv("PAYNDA_WEBHOOK_APP_SECRET"),
	}, nil
}

func (c *webhookClient) Deliver(ctx context.Context, req *biz.PayndaWebhookDeliveryRequest) (*biz.PayndaWebhookDeliveryResult, error) {
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, req.TargetURL, bytes.NewReader(req.Payload))
	if err != nil {
		return nil, err
	}
	if len(req.RequestHeaders) > 0 {
		var requestHeaders http.Header
		if err := json.Unmarshal(req.RequestHeaders, &requestHeaders); err != nil {
			return nil, err
		}
		for key, values := range requestHeaders {
			httpRequest.Header.Del(key)
			for _, value := range values {
				httpRequest.Header.Add(key, value)
			}
		}
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	if req.Category != "" {
		httpRequest.Header.Set("X-VK-NOTIFICATION-CATEGORY", req.Category)
	}
	if c.appID != "" && c.appSecret != "" {
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		nonce := randomx.Digits(16)
		signatureData := c.appID + c.appSecret + timestamp + nonce + httpRequest.URL.Path
		hash := md5.Sum([]byte(signatureData))
		httpRequest.Header.Set("appId", c.appID)
		httpRequest.Header.Set("timestamp", timestamp)
		httpRequest.Header.Set("nonce", nonce)
		httpRequest.Header.Set("sign", hex.EncodeToString(hash[:]))
	}
	requestHeaders, err := json.Marshal(httpRequest.Header)
	if err != nil {
		return nil, err
	}
	result := &biz.PayndaWebhookDeliveryResult{
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

var _ biz.PayndaWebhookClient = (*webhookClient)(nil)
