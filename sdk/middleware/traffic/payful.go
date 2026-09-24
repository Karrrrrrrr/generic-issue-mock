package traffic

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/wenzhenxi/gorsa"
	"resty.dev/v3"
)

// PayfulTrafficLog 记录请求和响应
// NOTE:中间件的错误都忽略,避免事务回滚
func PayfulTrafficLog(logger log.Logger, publicKey string) func(c *resty.Client, r *resty.Response) error {
	return func(c *resty.Client, r *resty.Response) error {
		req := map[string]any{
			"method": r.Request.Method,
			"url":    r.Request.URL,
		}

		if r.Request.RawRequest.ContentLength > 0 {
			reqBody := make(map[string]any, 10)
			body, _ := r.Request.RawRequest.GetBody()
			json.NewDecoder(body).Decode(&reqBody)

			if data, ok := reqBody["data"].(string); ok {
				args := make(map[string]any, 30)
				json.Unmarshal([]byte(data), &args)

				if _, ok := args["expdate"]; ok {
					args["expdate"] = "*"
				}

				req["body"] = args
			}
		}

		log.WithContext(r.Request.Context(), logger).
			Log(log.LevelInfo,
				"request", req,
				"response", PayfulHandleRespBody(r, publicKey),
				"dealer", "payful",
				"latency", r.Duration().String(), // 本次消耗的时间
			)

		return nil
	}
}

func PayfulHandleRespBody(r *resty.Response, publicKey string) map[string]any {
	// 这里对于图片或者非json 的报文，跳过body 的打印，避免 下载文件时打印大量二进制数据
	if shouldSummarizePayfulResp(r) {
		return payfulRespSummary(r)
	}

	body := struct {
		Success       bool   `json:"success"`
		ErrorMsg      string `json:"errorMsg"`
		Result        string `json:"result"`
		UserNo        string `json:"userNo"`
		CertificateId string `json:"certificateId"`
	}{}
	err := json.Unmarshal(r.Bytes(), &body)
	if err != nil {
		return payfulRespSummary(r)
	}
	// 如果不成功直接返回
	if r.IsError() {
		return map[string]any{
			"body": body,
		}
	}
	// 5. 判断 result 是加密字符串还是直接 JSON
	decryptedJSON := string(body.Result)
	// 尝试解析为字符串（加密情况），如果失败则直接使用（明文 JSON 情况）
	var resultStr string
	if err := json.Unmarshal([]byte(body.Result), &resultStr); err == nil {
		// result 是字符串，需要 RSA 解密
		decryptedJSON, err = DecodeRSA(resultStr, publicKey)
		if err != nil {
			return map[string]any{
				"body": body,
			}
		}
	} else {
		return map[string]any{
			"body": body,
		}
	}

	result := map[string]any{}
	err = json.Unmarshal([]byte(decryptedJSON), &result)
	if err != nil {
		return map[string]any{
			"body": body,
		}
	}
	// 修复：使用索引遍历并增加类型断言
	if v, ok := result["List"]; ok {
		if list, ok := v.([]any); ok {
			for i := range list {
				if m, ok := list[i].(map[string]any); ok {
					list[i] = desensitization(m)
				}
			}
		}
	}
	result = desensitization(result)
	data, _ := json.Marshal(result)
	body.Result = string(data)
	rsp, _ := json.Marshal(body)
	return map[string]any{
		"body": string(rsp),
	}
}

// 判断报文 是否是图片，或者非json
func shouldSummarizePayfulResp(r *resty.Response) bool {
	contentType := strings.ToLower(strings.TrimSpace(r.Header().Get("Content-Type")))
	if contentType == "" {
		return false
	}
	if strings.HasPrefix(contentType, "image/") {
		return true
	}
	return !strings.Contains(contentType, "json")
}

func payfulRespSummary(r *resty.Response) map[string]any {
	return map[string]any{
		"content_type": r.Header().Get("Content-Type"),
		"size":         len(r.Bytes()),
		"status":       r.Status(),
	}
}

func desensitization(result map[string]any) map[string]any {
	// 卡号脱敏
	if v, ok := result["cardNo"]; ok {
		cardNo := v.(string)
		if len(cardNo) != 16 {
			cardNo = "??????******????"
		}
		cardNo = cardNo[0:6] + "******" + cardNo[12:16]
		result["cardNo"] = cardNo
	}

	// cvv 脱敏
	if _, ok := result["cardVerifyNo"]; ok {
		result["cardVerifyNo"] = "***"
	}

	// 过期时间脱敏
	if _, ok := result["cardExpiryDate"]; ok {
		result["cardExpiryDate"] = "**/**"
	}
	return result
}

func DecodeRSA(encryptData string, publicKey string) (originData string, err error) {
	if encryptData == "" {
		return "", nil
	}

	encryptByte, err := hex.DecodeString(encryptData)
	if err != nil {
		return
	}

	originHex, err := gorsa.PublicDecrypt(base64.StdEncoding.EncodeToString(encryptByte), publicKey)
	if err != nil {
		return
	}

	originByte, err := hex.DecodeString(originHex)
	if err != nil {
		return
	}

	originDataByte, err := base64.StdEncoding.DecodeString(string(originByte))
	if err != nil {
		return
	}

	originData = string(originDataByte)
	return
}

func PayfulGlobalAccountTrafficLog(logger log.Logger, publicKey string) func(c *resty.Client, r *resty.Response) error {
	return func(c *resty.Client, r *resty.Response) error {
		req := map[string]any{
			"method": r.Request.Method,
			"url":    r.Request.URL,
		}

		if r.Request.RawRequest.ContentLength > 0 {
			reqBody := make(map[string]any, 10)
			body, _ := r.Request.RawRequest.GetBody()
			json.NewDecoder(body).Decode(&reqBody)

			if data, ok := reqBody["data"].(string); ok {
				args := make(map[string]any, 30)
				json.Unmarshal([]byte(data), &args)

				if _, ok := args["expdate"]; ok {
					args["expdate"] = "*"
				}

				req["body"] = args
			}
		}

		log.WithContext(r.Request.Context(), logger).
			Log(log.LevelInfo,
				"request", req,
				"response", PayfulHandleRespBody(r, publicKey),
				"dealer", "payful",
				"latency", r.Duration().String(), // 本次消耗的时间
			)

		return nil
	}
}
