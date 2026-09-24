package traffic

import (
	"encoding/json"

	"github.com/go-kratos/kratos/v2/log"
	"resty.dev/v3"
)

// PhotonTrafficLog 记录请求和响应
// NOTE:中间件的错误都忽略,避免事务回滚
func PhotonTrafficLog(logger log.Logger) func(c *resty.Client, r *resty.Response) error {
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
				"response", PhotonHandleRespBody(r),
				"dealer", "photonPay",
				"latency", r.Duration().String(), // 本次消耗的时间
			)

		return nil
	}
}

func PhotonHandleRespBody(r *resty.Response) map[string]any {
	body := struct {
		Code string          `json:"code"` // 返回码  0000:成功 其他：失败
		Msg  string          `json:"msg"`  // 返回信息
		Data json.RawMessage `json:"data"` // 使用 RawMessage 处理不同格式的数据
	}{}
	err := json.Unmarshal(r.Bytes(), &body)
	if err != nil {
		return map[string]any{
			"body": string(r.Bytes()),
		}
	}
	// 如果不成功直接返回
	if r.IsError() {
		return map[string]any{
			"body": body,
		}
	}

	result := make(map[string]any, 30)
	err = json.Unmarshal(body.Data, &result)
	if err != nil {
		return map[string]any{
			"body": body,
		}
	}

	if sensitiveInfo, ok := result["sensitiveInfo"]; ok {
		sensitive := sensitiveInfo.(map[string]any)
		// 卡号脱敏
		if v, ok := sensitive["cardNo"]; ok {
			cardNo := v.(string)
			if len(cardNo) != 16 {
				cardNo = "??????******????"
			}

			cardNo = cardNo[0:6] + "******" + cardNo[12:16]
			sensitive["cardNo"] = cardNo
		}

		// cvv 脱敏
		if _, ok := sensitive["cvv"]; ok {
			sensitive["cvv"] = "***"
		}

		// 过期时间脱敏
		if _, ok := sensitive["expirationDate"]; ok {
			sensitive["expirationDate"] = "**/**"
		}

		result["sensitiveInfo"] = sensitive
	}
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
	if _, ok := result["expirationDate"]; ok {
		result["expDate"] = "**/**"
	}
	data, _ := json.Marshal(result)
	body.Data = data
	rsp, _ := json.Marshal(body)
	return map[string]any{
		"body": string(rsp),
	}
}
