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

	"github.com/samber/do"
)

type webhookClient struct {
	client    *http.Client
	appID     string
	appSecret string
}

func NewWebhookClient(_ *do.Injector) (biz.PayndaWebhookClient, error) {
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
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("X-VK-NOTIFICATION-CATEGORY", req.Category)
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
	if len(req.RequestHeaders) > 0 {
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
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	responseHeaders, err := json.Marshal(response.Header)
	if err != nil {
		return nil, err
	}

	return &biz.PayndaWebhookDeliveryResult{
		StatusCode:      response.StatusCode,
		ResponseBody:    string(body),
		RequestHeaders:  requestHeaders,
		ResponseHeaders: responseHeaders,
	}, nil
}

var _ biz.PayndaWebhookClient = (*webhookClient)(nil)
