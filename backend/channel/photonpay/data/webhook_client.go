package data

import (
	"bytes"
	"context"
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"generic-mock/channel/photonpay/biz"

	"github.com/samber/do"
)

type webhookClient struct {
	client     *http.Client
	privateKey string
}

func NewWebhookClient(_ *do.Injector) (biz.WebhookClient, error) {
	return &webhookClient{
		client:     &http.Client{Timeout: 10 * time.Second},
		privateKey: os.Getenv("PHOTONPAY_WEBHOOK_PRIVATE_KEY"),
	}, nil
}

func (c *webhookClient) Deliver(ctx context.Context, req *biz.PhotonPayWebhookDeliveryRequest) (*biz.WebhookDeliveryResult, error) {
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, req.TargetURL, bytes.NewReader(req.Payload))
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("X-PD-NOTIFICATION-CATAGORY", req.NotifyCategory)
	httpRequest.Header.Set("X-PD-NOTIFICATION-TYPE", req.NotifyType)
	httpRequest.Header.Set("X-PD-PUBLISHED-AT", req.PublishedAt)
	if c.privateKey != "" {
		signature, err := photonPayWebhookSignature(req.Payload, c.privateKey)
		if err != nil {
			return nil, err
		}
		httpRequest.Header.Set("X-PD-SIGN", signature)
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
	result := &biz.WebhookDeliveryResult{
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

func photonPayWebhookSignature(payload []byte, privateKeyPEM string) (string, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", errors.New("invalid PhotonPay webhook private key")
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}
	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return "", errors.New("PhotonPay webhook private key is not RSA")
	}
	hash := md5.Sum(payload)
	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, crypto.MD5, hash[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

var _ biz.WebhookClient = (*webhookClient)(nil)
