package traffic

import (
	"encoding/json"

	"github.com/go-kratos/kratos/v2/log"
	"resty.dev/v3"
)

// SlashTrafficLog 记录请求和响应
// NOTE:中间件的错误都忽略,避免事务回滚
func SlashTrafficLog(logger log.Logger) func(c *resty.Client, r *resty.Response) error {
	return func(c *resty.Client, r *resty.Response) error {
		req := map[string]any{
			"method": r.Request.Method,
			"url":    r.Request.URL,
		}

		if r.Request.RawRequest.ContentLength > 0 {
			body, _ := r.Request.RawRequest.GetBody()
			args := make(map[string]any, 10)
			json.NewDecoder(body).Decode(&args)

			req["body"] = args
		}

		log.WithContext(r.Request.Context(), logger).
			Log(log.LevelInfo,
				"request", req,
				"response", slashHandleRespBody(r),
				"dealer", "slash",
				"latency", r.Duration().String(), // 本次消耗的时间
			)

		return nil
	}
}

func slashHandleRespBody(r *resty.Response) map[string]any {
	body := make(map[string]any, 30)
	err := json.Unmarshal(r.Bytes(), &body)
	if err != nil {
		return map[string]any{
			"body": string(r.Bytes()),
		}
	}
	// 如果不成功直接返回
	if r.IsError() {
		return body
	}

	if v, ok := body["pan"]; ok && v != nil {
		cardNo := v.(string)
		if len(cardNo) != 16 {
			cardNo = "??????******????"
		}

		cardNo = cardNo[0:6] + "******" + cardNo[12:16]
		body["pan"] = cardNo
	}

	// cvv 脱敏
	if _, ok := body["cvv"]; ok {
		body["cvv"] = "***"
	}

	// 过期时间脱敏
	if _, ok := body["expiryYear"]; ok {
		body["expiryYear"] = "**"
		body["expiryMonth"] = "**"
	}
	return map[string]any{
		"body": body,
	}
}
