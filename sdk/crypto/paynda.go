package crypto

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

// SignatureParams 签名参数
type SignatureParams struct {
	AppID     string
	AppSecret string
	Path      string
	Nonce     string
	Timestamp string
}

// GenerateSignature 生成API签名
// 根据文档: sign = md5(appId + appSecret + timestamp + nonce + path)
func GenerateSignature(params *SignatureParams) string {
	// 如果没有提供timestamp，则使用当前时间
	timestamp := params.Timestamp
	if timestamp == "" {
		timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	}
	data := params.AppID + params.AppSecret + timestamp + params.Nonce + params.Path

	// 计算MD5哈希
	hash := md5.Sum([]byte(data))

	// 转换为小写十六进制字符串
	return hex.EncodeToString(hash[:])
}

// 校验timestamp
func validatePayndaWebhookTimestamp(timestamp string) error {
	// Parse the timestamp
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid x-webhook-timestamp")
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
		return fmt.Errorf("timestamp is not within acceptable window")
	}

	return nil
}

// ValidatePayndaWebhookSignature 验签
func ValidatePayndaWebhookSignature(req *SignatureParams, signature string) error {
	// 校验timestamp
	err := validatePayndaWebhookTimestamp(req.Timestamp)
	if err != nil {
		return err
	}

	expectedSignature := GenerateSignature(req)

	// 验签
	if signature != expectedSignature {
		return fmt.Errorf("signature err")
	}

	return nil
}
