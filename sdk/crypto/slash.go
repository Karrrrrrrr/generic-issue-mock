package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
	"tman/errors"
)

const (
	slashSignaturePrefix     = "v1="
	slashSigningSecretPrefix = "whsec_"
)

type ValidateSlashSignatureReq struct {
	WebhookID       string
	Timestamp       string
	SignatureHeader string
	Payload         string
	SigningSecret   string
}

// 校验timestamp
func validateSlashWebhookTimestamp(timestamp string) error {
	// Parse the timestamp
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return errors.NewBadRequest("invalid x-webhook-timestamp")
	}

	// Convert to time.Time (assuming timestamp is in seconds)
	timestampTime := time.Unix(ts, 0)

	// Get current time
	now := time.Now()

	// Calculate the difference
	diff := now.Sub(timestampTime)
	if diff < 0 {
		diff = -diff
	}

	// Check if within 2 minutes (120 seconds)
	if diff > 2*time.Minute {
		return errors.NewBadRequest("timestamp is not within acceptable window")
	}

	return nil
}

// ValidateSlashWebhookSignature 验签
func ValidateSlashWebhookSignature(req *ValidateSlashSignatureReq) (string, error) {
	// 校验timestamp
	err := validateSlashWebhookTimestamp(req.Timestamp)
	if err != nil {
		return "", err
	}

	// 去除signatureHeader中的"v1="前缀
	signature := strings.TrimPrefix(req.SignatureHeader, slashSignaturePrefix)

	// 去除signingSecret中的"whsec_"前缀
	secret := strings.TrimPrefix(req.SigningSecret, slashSigningSecretPrefix)
	decodedSecret, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", errors.ErrServer
	}

	// 组装待签名内容: webhookId.timestamp.requestBody
	payloadString := fmt.Sprintf("%s.%s.%s", req.WebhookID, req.Timestamp, req.Payload)
	// 生成 HMAC SHA256 签名
	h := hmac.New(sha256.New, decodedSecret)
	h.Write([]byte(payloadString))
	expectedSignature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// 验签
	if signature != expectedSignature {
		return "", errors.NewBadRequest("signature err")
	}

	return expectedSignature, nil
}
