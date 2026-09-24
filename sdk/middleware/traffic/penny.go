package traffic

import (
	"encoding/json"

	"github.com/go-kratos/kratos/v2/log"

	"resty.dev/v3"
)

// PennyTrafficLog Penny记录请求和响应
// NOTE:中间件的错误都忽略,避免事务回滚
func PennyTrafficLog(logger log.Logger) func(c *resty.Client, r *resty.Response) error {
	return func(c *resty.Client, r *resty.Response) error {
		req := map[string]any{
			"method": r.Request.Method,
			"url":    r.Request.URL,
		}

		if r.Request.RawRequest.ContentLength > 0 {
			body, _ := r.Request.RawRequest.GetBody()
			args := make(map[string]any, 10)
			json.NewDecoder(body).Decode(&args)

			if _, ok := args["expiration_date"]; ok {
				args["expiration_date"] = "*"
			}

			req["body"] = args
		}

		log.WithContext(r.Request.Context(), logger).
			Log(log.LevelInfo,
				"request", req,
				"response", pennyHandleRespBody(r),
				"dealer", "penny",
				"latency", r.Duration().String(), // 本次消耗的时间
			)

		return nil
	}
}

func pennyHandleRespBody(r *resty.Response) map[string]any {
	body := make(map[string]any, 30)
	json.Unmarshal(r.Bytes(), &body)
	// 如果不成功的话直接return
	if r.IsError() {
		return map[string]any{"body": body}
	}

	// 卡号脱敏
	if accountNumber, ok := body["account_number"]; ok {
		cardNo, ok := accountNumber.(string)
		if ok {
			if len(cardNo) != 16 {
				cardNo = "??????******????"
			}

			cardNo = cardNo[0:6] + "******" + cardNo[12:16]
			body["account_number"] = cardNo
		}
	}

	// cvv 脱敏
	if _, ok := body["security_code"]; ok {
		body["security_code"] = "***"
	}

	// 过期时间脱敏
	if _, ok := body["expiration"]; ok {
		body["expiration"] = "**/**"
	}

	// 过期日期
	if _, ok := body["expiration_date"]; ok {
		body["expiration_date"] = "*"
	}

	return map[string]any{
		"body": body,
	}
}
