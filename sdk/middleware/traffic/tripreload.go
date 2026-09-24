package traffic

import (
	"encoding/json"
	"sdk/crypto"

	"github.com/go-kratos/kratos/v2/log"
	"resty.dev/v3"
)

// TripreloadTrafficLog 路淘记录请求和响应
// NOTE:中间件的错误都忽略,避免事务回滚
func TripreloadTrafficLog(conf crypto.TripreloadCrypto, logger log.Logger) func(c *resty.Client, r *resty.Response) error {
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
				b, _ := crypto.TripreloadDecrypt(conf, data)
				args := make(map[string]any, 30)
				json.Unmarshal(b, &args)

				if _, ok := args["expdate"]; ok {
					args["expdate"] = "*"
				}

				req["body"] = args
			}
		}

		log.WithContext(r.Request.Context(), logger).
			Log(log.LevelInfo,
				"request", req,
				"response", tripreloadHandleRespBody(conf, r),
				"dealer", "tripreload",
				"latency", r.Duration().String(), // 本次消耗的时间
			)

		return nil
	}
}

func tripreloadHandleRespBody(conf crypto.TripreloadCrypto, r *resty.Response) map[string]any {
	body := struct {
		Code int             `json:"code"` // 返回码  1:成功 其他：失败
		Msg  string          `json:"msg"`  // 返回信息
		Data json.RawMessage `json:"data"` // 使用 RawMessage 处理不同格式的数据
	}{}
	json.Unmarshal(r.Bytes(), &body)

	// 如果不成功直接返回
	if body.Code != 1 {
		return map[string]any{
			"body": body,
		}
	}

	var data string
	json.Unmarshal(body.Data, &data)
	b, _ := crypto.TripreloadDecrypt(conf, data)

	result := make(map[string]any, 30)
	json.Unmarshal(b, &result)

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
	if _, ok := result["cvv"]; ok {
		result["cvv"] = "***"
	}

	// 过期时间脱敏
	if _, ok := result["expDate"]; ok {
		result["expDate"] = "**/**"
	}

	return map[string]any{
		"body": result,
	}
}
