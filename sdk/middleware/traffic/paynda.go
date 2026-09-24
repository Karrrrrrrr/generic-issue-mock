package traffic

import (
	"encoding/json"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"resty.dev/v3"
)

// PayndaTrafficLog 记录请求和响应
// NOTE:中间件的错误都忽略,避免事务回滚
func PayndaTrafficLog(logger log.Logger) func(c *resty.Client, r *resty.Response) error {
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
				"response", payndaHandleRespBody(r),
				"dealer", "paynda",
				"latency", r.Duration().String(), // 本次消耗的时间
			)

		return nil
	}
}

func payndaHandleRespBody(r *resty.Response) map[string]any {
	// Parse the original response body to extract structure
	var body struct {
		Code    json.Number     `json:"code"` // Use json.Number for large numbers
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"` // Keep as RawMessage to parse separately
		Success bool            `json:"success"`
	}
	if r.IsError() {
		return map[string]any{
			"body": body,
		}
	}

	err := json.Unmarshal(r.Bytes(), &body)
	if err != nil {
		return map[string]any{
			"body": string(r.Bytes()),
		}
	}

	// Parse the data field separately with UseNumber
	decoder := json.NewDecoder(strings.NewReader(string(body.Data)))
	decoder.UseNumber()

	var result map[string]any
	err = decoder.Decode(&result)
	if err != nil {
		return map[string]any{
			"body": body,
		}
	}

	// Process sensitive information with proper type checking
	if sensitiveInfo, ok := result["sensitiveInfo"]; ok {
		if sensitive, ok := sensitiveInfo.(map[string]any); ok {
			// Card number masking
			if v, ok := sensitive["cardNo"]; ok {
				if cardNo, ok := v.(string); ok {
					maskedCardNo := maskCardNumber(cardNo)
					sensitive["cardNo"] = maskedCardNo
				}
			}

			// CVV masking
			if _, ok := sensitive["cvv"]; ok {
				sensitive["cvv"] = "***"
			}

			// Expiration date masking
			if _, ok := sensitive["expirationDate"]; ok {
				sensitive["expirationDate"] = "**/**"
			}

			result["sensitiveInfo"] = sensitive
		}
	}

	if v, ok := result["cardNo"]; ok {
		if cardNo, ok := v.(string); ok {
			maskedCardNo := maskCardNumber(cardNo)
			result["cardNo"] = maskedCardNo
		}
	}

	if _, ok := result["cvv"]; ok {
		result["cvv"] = "***"
	}

	if _, ok := result["expirationDate"]; ok {
		result["expirationDate"] = "**/**"
	}

	data, err := json.Marshal(result)
	if err != nil {
		originalBody := map[string]any{
			"code":    body.Code,
			"message": body.Message,
			"data":    string(body.Data),
			"success": body.Success,
		}
		return map[string]any{
			"body": originalBody,
		}
	}

	body.Data = data
	rsp, err := json.Marshal(body)
	if err != nil {
		originalBody := map[string]any{
			"code":    body.Code,
			"message": body.Message,
			"data":    string(body.Data),
			"success": body.Success,
		}
		return map[string]any{
			"body": originalBody,
		}
	}

	return map[string]any{
		"body": string(rsp),
	}
}

func maskCardNumber(cardNo string) string {
	if len(cardNo) < 16 {
		return "??????******????"
	}
	return cardNo[0:6] + "******" + cardNo[len(cardNo)-4:]
}
