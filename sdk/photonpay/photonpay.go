package photonpay

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sdk/crypto"

	"github.com/go-kratos/kratos/v2/errors"

	"resty.dev/v3"
)

// GenerateAuthorization 生成授权请求参数
func GenerateAuthorization(appID, appSecret string) string {
	return "basic " + base64.StdEncoding.EncodeToString([]byte(appID+"/"+appSecret))
}

// Conf 配置接口
type Conf interface {
	GetAppId() string
	GetAppSecret() string
	GetThirdPartyPublicKey() string
	GetPublicKey() string
	GetPrivateKey() string
	GetBalanceAccount() string
	GetUrl() string
	GetDebug() bool
}

// PhotonPaySDK SDK结构体
type PhotonPaySDK struct {
	resty               *resty.Client
	conf                Conf
	requestMiddleware   []resty.RequestMiddleware
	responseMiddlewares []resty.ResponseMiddleware
}

// Option SDK选项函数类型
type Option func(t *PhotonPaySDK)

// WithRequestMiddlewares 设置请求中间件
func WithRequestMiddlewares(middlewares ...resty.RequestMiddleware) Option {
	return func(pp *PhotonPaySDK) {
		pp.requestMiddleware = append(pp.requestMiddleware, resty.PrepareRequestMiddleware)
		pp.requestMiddleware = append(pp.requestMiddleware, middlewares...)
	}
}

// WithResponseMiddlewares 设置响应中间件
func WithResponseMiddlewares(middlewares ...resty.ResponseMiddleware) Option {
	return func(pp *PhotonPaySDK) {
		pp.responseMiddlewares = append(pp.responseMiddlewares, resty.AutoParseResponseMiddleware)
		pp.responseMiddlewares = append(pp.responseMiddlewares, middlewares...)
	}
}

// New 初始化PhotonPaySDK
func New(conf Conf, options ...Option) *PhotonPaySDK {
	pp := &PhotonPaySDK{
		conf:                conf,
		resty:               resty.New(),
		requestMiddleware:   []resty.RequestMiddleware{},
		responseMiddlewares: []resty.ResponseMiddleware{},
	}

	for _, o := range options {
		o(pp)
	}

	if len(pp.requestMiddleware) > 0 {
		pp.resty.SetRequestMiddlewares(pp.requestMiddleware...)
	}
	if len(pp.responseMiddlewares) > 0 {
		pp.resty.SetResponseMiddlewares(pp.responseMiddlewares...)
	}

	return pp
}

// RestyRequest 创建并发送请求，处理响应
func (pp *PhotonPaySDK) RestyRequest(ctx context.Context, options *RestyRequestOptions) error {

	headers := map[string]string{
		"Content-Type": "application/json",
	}
	if options.Token != "" {
		headers["X-PD-TOKEN"] = options.Token
	}
	if options.Authorization != "" {
		headers["Authorization"] = options.Authorization
	}
	apiUrl, err := url.JoinPath(pp.conf.GetUrl(), options.Path)
	if err != nil {
		return fmt.Errorf("构建URL失败: %w", err)
	}
	// 创建请求
	req := pp.resty.R().
		SetContext(ctx).
		SetURL(apiUrl).
		SetDebug(pp.conf.GetDebug()).
		SetAllowNonIdempotentRetry(false).
		SetMethod(options.Method)

	if options.Method != http.MethodGet {
		req.SetRetryCount(0)
	}
	// 根据请求方法和body类型设置请求参数
	switch options.Method {
	case http.MethodGet:
		// 对于GET请求，将body作为查询参数
		jsonData, err := json.Marshal(options.Params)
		if err != nil {
			return fmt.Errorf("序列化请求参数失败: %w", err)
		}
		var queryParams map[string]any
		if err := json.Unmarshal(jsonData, &queryParams); err != nil {
			return fmt.Errorf("解析请求参数失败: %w", err)
		}
		for k, v := range queryParams {
			if v != nil {
				req.SetQueryParam(k, fmt.Sprintf("%v", v))
			}
		}
	case http.MethodDelete:
		// 对于GET请求，将body作为查询参数
		jsonData, err := json.Marshal(options.Params)
		if err != nil {
			return fmt.Errorf("序列化请求参数失败: %w", err)
		}
		var queryParams map[string]any
		if err := json.Unmarshal(jsonData, &queryParams); err != nil {
			return fmt.Errorf("解析请求参数失败: %w", err)
		}
		for k, v := range queryParams {
			if v != nil {
				req.SetQueryParam(k, fmt.Sprintf("%v", v))
			}
		}
	default:
		if options != nil && options.Params != nil {
			// 序列化请求参数为 JSON 字节
			json_data, err := json.Marshal(options.Params)
			if err != nil {
				return fmt.Errorf("签名序列化参数失败: %w", err)
			}
			req.SetBody(json_data)

			// 生成签名
			signature, err := crypto.PhotonpaySignature(json_data, pp.conf.GetPrivateKey())
			if err != nil {
				return fmt.Errorf("生成签名失败: %w", err)
			}
			headers["X-PD-SIGN"] = signature
		}
	}

	client, err := req.
		SetHeaders(headers).
		Send()
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}

	resp := struct {
		Code string          `json:"code"` // 返回码  0000:成功 其他：失败
		Msg  string          `json:"msg"`  // 返回信息
		Data json.RawMessage `json:"data"` // 使用 RawMessage 处理不同格式的数据
	}{}

	if err := json.NewDecoder(client.Body).Decode(&resp); err != nil {
		return err
	}
	// // 检查响应状态
	// if client.IsError() {
	// 	return fmt.Errorf("%s", client.String())
	// }

	if resp.Code != "0000" {
		return errors.New(0, resp.Code, resp.Msg)
	}

	if options.Result != nil {
		if err := json.Unmarshal(resp.Data, &options.Result); err != nil {
			return err
		}
	}

	return nil
}

// 获取令牌
// path: /oauth2/token/accessToken
// method: POST
func (pp *PhotonPaySDK) GetAccessToken(ctx context.Context) (*AccessTokenResponse, error) {
	apiPath := "/oauth2/token/accessToken"
	result := &AccessTokenResponse{}
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method:        http.MethodPost,
		Path:          apiPath,
		Authorization: GenerateAuthorization(pp.conf.GetAppId(), pp.conf.GetAppSecret()),
		Params:        nil,
		Result:        &result,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 实时金额查询
// path: /wallet/openApi/v4/account/single
// method: GET
func (pp *PhotonPaySDK) GetAccountSingle(ctx context.Context, token string, param *AccountSingleRequest) (*AccountSingleResponse, error) {
	apiPath := "/wallet/openApi/v4/account/single"
	if err := param.Validate(); err != nil {
		return nil, err
	}
	var result *AccountSingleResponse
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodGet,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 光子易账户账单查询
// path: /wallet/openApi/v4/account/history
// method: GET
func (pp *PhotonPaySDK) AccountHistory(ctx context.Context, token string, param *AccountHistoryRequest) (*AccountHistoryResponse, error) {
	apiPath := "/wallet/openApi/v4/account/history"
	if err := param.Validate(); err != nil {
		return nil, err
	}
	var result *AccountHistoryResponse
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodGet,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 添加用卡人
// path: /vcc/openApi/v4/addCardholder
// method: POST
func (pp *PhotonPaySDK) AddCardholder(ctx context.Context, token string, param *AddCardholderRequest) (*AddCardholderResponse, error) {
	apiPath := "/vcc/openApi/v4/addCardholder"
	if err := param.Validate(); err != nil {
		return nil, err
	}

	var result *AddCardholderResponse
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodPost,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 更新用卡人
// path: /vcc/openApi/v4/editCardholder
// method: POST
// 1、modify - 待修改：姓、名、生日不能修改，其他的都可以
// 2、approved - 审核通过：只有手机、邮箱、地址可以修改，其他不能修改
func (pp *PhotonPaySDK) EditCardholder(ctx context.Context, token string, param *EditCardholderRequest) (*EditCardholderResponse, error) {
	apiPath := "/vcc/openApi/v4/editCardholder"
	if err := param.Validate(); err != nil {
		return nil, err
	}
	var result *EditCardholderResponse
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodPost,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 用卡人查询
// path: /vcc/openApi/v4/pagingVccCardholder
// method: GET
func (pp *PhotonPaySDK) PagingVccCardholder(ctx context.Context, token string, param *PagingVccCardholderRequest) ([]*PagingVccCardholderResponse, error) {
	apiPath := "/vcc/openApi/v4/pagingVccCardholder"
	if err := param.Validate(); err != nil {
		return nil, err
	}
	var result []*PagingVccCardholderResponse
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodGet,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 卡bin查询
// path: /vcc/openApi/v4/getCardBin
// method: GET
func (pp *PhotonPaySDK) GetCardBin(ctx context.Context, token string, param *CardBinRequest) ([]*CardBinResponse, error) {
	apiPath := "/vcc/openApi/v4/getCardBin"
	// if err := param.Validate(); err != nil {
	// 	return nil, err
	// }
	var result []*CardBinResponse
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodGet,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 单卡开卡
// path: /vcc/openApi/v4/openCard
// method: POST
func (pp *PhotonPaySDK) OpenCard(ctx context.Context, token string, param *OpenCardRequest) (*OpenCardResponse, error) {
	apiPath := "/vcc/openApi/v4/openCard"
	if err := param.Validate(); err != nil {
		return nil, err
	}
	var result *OpenCardResponse
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodPost,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 请求结果查询
// path: /vcc/openApi/v4/getRequestResult
// method: GET
func (pp *PhotonPaySDK) GetRequestResult(ctx context.Context, token string, param *GetRequestResultRequest) (*OpenCardResponse, error) {
	apiPath := "/vcc/openApi/v4/getRequestResult"
	if err := param.Validate(); err != nil {
		return nil, err
	}
	var result *OpenCardResponse
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodGet,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 卡信息查询
// path: /vcc/openApi/v4/getCardDetail
// method: GET
func (pp *PhotonPaySDK) GetCardDetail(ctx context.Context, token string, param *GetCardDetailRequest) (*GetCardDetailResponse, error) {
	apiPath := "/vcc/openApi/v4/getCardDetail"
	if err := param.Validate(); err != nil {
		return nil, err
	}
	var result *GetCardDetailResponse
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodGet,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 卡列表
// path: /vcc/openApi/v4/pagingVccCard
// method: GET
func (pp *PhotonPaySDK) PagingVccCard(ctx context.Context, token string, param *PagingVccCardRequest) ([]*CardDetailResponse, error) {
	apiPath := "/vcc/openApi/v4/pagingVccCard"
	if err := param.Validate(); err != nil {
		return nil, err
	}
	var result []*CardDetailResponse
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodGet,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CVV查询
// path: /vcc/openApi/v4/getCvv
// method: GET
func (pp *PhotonPaySDK) GetCvv(ctx context.Context, token string, param *GetCvvRequest) (*GetCvvResponse, error) {
	apiPath := "/vcc/openApi/v4/getCvv"
	if err := param.Validate(); err != nil {
		return nil, err
	}
	var result *GetCvvResponse
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodGet,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 卡更新
// path: /vcc/openApi/v4/updateCard
// method: POST
func (pp *PhotonPaySDK) UpdateCard(ctx context.Context, token string, param *UpdateCardRequest) (*OpenCardResponse, error) {
	apiPath := "/vcc/openApi/v4/updateCard"
	// if err := param.Validate(); err != nil {
	// 	return nil, err
	// }
	var result *OpenCardResponse
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodPost,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 更新账单地址
// path: /vcc/openApi/v4/editCardBillingAddress
// method: POST
func (pp *PhotonPaySDK) EditCardBillingAddress(ctx context.Context, token string, param *EditCardBillingAddressRequest) error {
	apiPath := "/vcc/openApi/v4/editCardBillingAddress"
	if err := param.Validate(); err != nil {
		return err
	}
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodPost,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: nil,
	})
	if err != nil {
		return err
	}
	return nil
}

// 冻结
// path: /vcc/openApi/v4/freezeCard
// method: POST
func (pp *PhotonPaySDK) FreezeCard(ctx context.Context, token string, param *FreezeCardRequest) error {
	apiPath := "/vcc/openApi/v4/freezeCard"
	if err := param.Validate(); err != nil {
		return err
	}
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodPost,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: nil,
	})
	if err != nil {
		return err
	}
	return nil
}

// 销卡
// path: /vcc/openApi/v4/cancelCard
// method: POST
func (pp *PhotonPaySDK) CancelCard(ctx context.Context, token string, param *CancelCardRequest) error {
	apiPath := "/vcc/openApi/v4/cancelCard"
	if err := param.Validate(); err != nil {
		return err
	}
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodPost,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: nil,
	})
	if err != nil {
		return err
	}
	return nil
}

// 换汇询价
// path: /vcc/openApi/v4/preRecharge
// method: GET
func (pp *PhotonPaySDK) PreRecharge(ctx context.Context, token string, param *PreRechargeRequest) (*PreRechargeResponse, error) {
	apiPath := "/vcc/openApi/v4/preRecharge"
	if err := param.Validate(); err != nil {
		return nil, err
	}
	var result *PreRechargeResponse
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodGet,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 转入下单
// path: /vcc/openApi/v4/recharge
// method: POST
func (pp *PhotonPaySDK) Recharge(ctx context.Context, token string, param *RechargeRequest) (*RechargeResponse, error) {
	apiPath := "/vcc/openApi/v4/recharge"
	if err := param.Validate(); err != nil {
		return nil, err
	}
	var result *RechargeResponse
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodPost,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 卡金额退还
// path: /vcc/openApi/v4/rechargeReturn
// method: POST
func (pp *PhotonPaySDK) RechargeReturn(ctx context.Context, token string, param *RechargeReturnRequest) (*RechargeReturnResponse, error) {
	apiPath := "/vcc/openApi/v4/rechargeReturn"
	if err := param.Validate(); err != nil {
		return nil, err
	}
	var result *RechargeReturnResponse
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodPost,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 卡历史明细
// path: /vcc/openApi/v4/pagingIssuingHistory
// method: GET
func (pp *PhotonPaySDK) PagingIssuingHistory(ctx context.Context, token string, param *PagingIssuingHistoryRequest) ([]*PagingIssuingHistoryResponse, error) {
	apiPath := "/vcc/openApi/v4/pagingIssuingHistory"
	if err := param.Validate(); err != nil {
		return nil, err
	}
	var result []*PagingIssuingHistoryResponse
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodGet,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 常规卡资金明细
// path: /vcc/openApi/v4/pagingRechargeCardFundsDetail
// method: GET
func (pp *PhotonPaySDK) PagingRechargeCardFundsDetail(ctx context.Context, token string, param *PagingRechargeCardFundsDetailRequest) ([]*PagingRechargeCardFundsDetailResponse, error) {
	apiPath := "/vcc/openApi/v4/pagingRechargeCardFundsDetail"
	if err := param.Validate(); err != nil {
		return nil, err
	}
	var result []*PagingRechargeCardFundsDetailResponse
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodGet,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 共享卡交易额度明细
// path: /vcc/openApi/v4/pagingShareCardTxnLimitDetail
// method: GET
func (pp *PhotonPaySDK) PagingShareCardTxnLimitDetail(ctx context.Context, token string, param *PagingShareCardTxnLimitDetailRequest) ([]*PagingShareCardTxnLimitDetailResponse, error) {
	apiPath := "/vcc/openApi/v4/pagingShareCardTxnLimitDetail"
	if err := param.Validate(); err != nil {
		return nil, err
	}
	var result []*PagingShareCardTxnLimitDetailResponse
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodGet,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 交易明细
// path: /vcc/openApi/v4/pagingVccTradeOrder
// method: GET
func (pp *PhotonPaySDK) PagingVccTradeOrder(ctx context.Context, token string, param *PagingVccTradeOrderRequest) ([]*PagingVccTradeOrderResponse, error) {
	apiPath := "/vcc/openApi/v4/pagingVccTradeOrder"
	if err := param.Validate(); err != nil {
		return nil, err
	}
	var result []*PagingVccTradeOrderResponse
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodGet,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 交易模拟
// path: /vcc/open/v2/sandBoxTransaction
// method: POST
func (pp *PhotonPaySDK) SandBoxTransaction(ctx context.Context, token string, param *SandBoxTransactionRequest) error {
	apiPath := "/vcc/open/v2/sandBoxTransaction"
	if err := param.Validate(); err != nil {
		return err
	}
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodPost,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: nil,
	})
	if err != nil {
		return err
	}
	return nil
}

// ApiUploadFile 文件上传
func (pp *PhotonPaySDK) ApiUploadFile(ctx context.Context, token string, param *ApiUploadFileRequest) (string, error) {
	apiPath := fmt.Sprintf("/file/apiUpload/%s", param.BusinessKey)

	headers := map[string]string{
		"X-PD-TOKEN": token,
	}

	var resp struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data string `json:"data,omitempty"`
	}
	apiUrl, err := url.JoinPath(pp.conf.GetUrl(), apiPath)
	if err != nil {
		return "", fmt.Errorf("构建URL失败: %w", err)
	}
	// 使用resty的文件上传功能
	client, err := pp.resty.R().
		SetContext(ctx).
		SetDebug(pp.conf.GetDebug()).
		SetHeaders(headers).
		SetFileReader("file", param.FileName, param.FileReader).
		SetResult(&resp).
		Post(apiUrl)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	// 检查HTTP状态码错误
	if client.IsError() {
		// 尝试解析响应体中的错误信息
		if parseErr := json.NewDecoder(client.Body).Decode(&resp); parseErr == nil {
			// 成功解析错误响应
			return "", errors.New(0, resp.Code, resp.Msg)
		}
		// 无法解析错误响应，返回原始错误信息
		return "", errors.New(0, "HTTP_ERROR", client.Status())
	}

	if resp.Code != "0000" {
		return "", errors.New(0, resp.Code, resp.Msg)
	}

	return resp.Data, nil
}

// 查询 Webhook 订阅通知
// path: /exchange-center/open/api/v1/webhook/notification
// method: GET
func (pp *PhotonPaySDK) GetWebhookNotification(ctx context.Context, token string) (*GetWebhookNotificationResponse, error) {
	apiPath := "/exchange-center/open/api/v1/webhook/notification"
	result := &GetWebhookNotificationResponse{}
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodGet,
		Path:   apiPath,
		Token:  token,
		Params: nil,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 订阅 Webhook 通知
// path: /exchange-center/open/api/v1/webhook/notification
// method: PUT
func (pp *PhotonPaySDK) SetWebhookNotification(ctx context.Context, token string, param *WebhookNotificationRequest) error {
	apiPath := "/exchange-center/open/api/v1/webhook/notification"
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodPut,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: nil,
	})
	if err != nil {
		return err
	}
	return nil
}

// 取消订阅 Webhook 通知
// path: /exchange-center/open/api/v1/webhook/notification
// method: DELETE
func (pp *PhotonPaySDK) DelWebhookNotification(ctx context.Context, token string, param *WebhookNotificationRequest) error {
	apiPath := "/exchange-center/open/api/v1/webhook/notification"
	err := pp.RestyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodDelete,
		Path:   apiPath,
		Token:  token,
		Params: param,
		Result: nil,
	})
	if err != nil {
		return err
	}
	return nil
}
