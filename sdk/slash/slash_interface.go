package slash

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"time"

	"resty.dev/v3"
)

type CardPostRequest struct {
	Name          string // 卡片持有人姓名
	CardProductID string // 卡挂在 slash某个产品下
	RequestID     string // Headers: X-Idempotency-Key
	CardID        string // 卡片id
}

type GetCardByIDResponse struct {
	Status      CardStatusEnum
	Pan         string
	Cvv         string
	ExpiryMonth string // 2 位如 12
	ExpiryYear  string // 4 位,如 2020,限制了成功一定是 4 位,否则报错
}

type card struct {
	ID                 string             `json:"id"`
	AccountID          string             `json:"accountId"`
	VirtualAccountID   string             `json:"virtualAccountId"`
	Last4              string             `json:"last4"`
	Name               string             `json:"name"`
	ExpiryMonth        string             `json:"expiryMonth"`
	ExpiryYear         string             `json:"expiryYear"`
	Status             string             `json:"status"` // active, paused, inactive, closed
	IsPhysical         bool               `json:"isPhysical"`
	IsSingleUse        bool               `json:"isSingleUse"`
	Pan                string             `json:"pan"`
	Cvv                string             `json:"cvv"`
	CardGroupID        string             `json:"cardGroupId"`
	CreatedAt          time.Time          `json:"createdAt"`
	SpendingConstraint SpendingConstraint `json:"spendingConstraint"`
	UserData           UserData           `json:"userData"`
	CardProductID      string             `json:"cardProductId"`
}

// CardPostRequest 开卡请求参数
type cardPostRequest struct {
	Name               string              `json:"name"`                         // 卡片持有人姓名
	UserData           any                 `json:"userData"`                     // 客户自定义参数exceed 4kb
	SpendingConstraint *SpendingConstraint `json:"spendingConstraint,omitempty"` // 卡控制参数
	Type               string              `json:"type"`                         // default: "virtual"
	IsSingleUse        bool                `json:"isSingleUse"`                  // 单笔消费 false
	AccountID          string              `json:"accountId"`                    // 我方在 slash 的账户
	VirtualAccountID   string              `json:"virtualAccountId"`             // 卡挂在 slash某个账户下
	CardGroupID        string              `json:"cardGroupId,omitempty"`        // 卡挂在 slash某个卡组下
	CardProductID      string              `json:"cardProductId,omitempty"`      // 卡挂在 slash某个产品下
}

type SlashSDKConf struct {
	Url              string
	VaultUrl         string
	ApiKey           string
	AccountID        string
	VirtualAccountID string
	CardGroupID      string
	LegalEntity      string
}

type slashSDK struct {
	Debug bool

	// slash 配置
	SlashSDKConf *SlashSDKConf

	// sdk组件
	Resty               *resty.Client
	RequestMiddleware   []resty.RequestMiddleware
	ResponseMiddlewares []resty.ResponseMiddleware
}

// Option SDK选项函数类型
type SlashOption func(t *slashSDK)

// WithRequestMiddlewares 设置请求中间件
func WithSlashRequestMiddlewares(middlewares ...resty.RequestMiddleware) SlashOption {
	return func(sh *slashSDK) {
		sh.RequestMiddleware = append(sh.RequestMiddleware, resty.PrepareRequestMiddleware)
		sh.RequestMiddleware = append(sh.RequestMiddleware, middlewares...)
	}
}

// WithResponseMiddlewares 设置响应中间件
func WithSlashResponseMiddlewares(middlewares ...resty.ResponseMiddleware) SlashOption {
	return func(sh *slashSDK) {
		sh.ResponseMiddlewares = append(sh.ResponseMiddlewares, resty.AutoParseResponseMiddleware)
		sh.ResponseMiddlewares = append(sh.ResponseMiddlewares, middlewares...)
	}
}

// slash 外面调用接口
type SlashSDKInterface interface {
	CardPostPre(params *CardPostRequest) (json.RawMessage, error)
	CardPost(ctx context.Context, params *CardPostRequest) (cardID string, err error)
	GetCardByID(ctx context.Context, cardID string) (*GetCardByIDResponse, error)
	CardFrozen(ctx context.Context, cardID string) error
	CardUnfrozen(ctx context.Context, cardID string) error
	GetVirtualAccount() string
}

// NewSlashSDK 初始化SlashPaySDK
func NewSlashSDK(debug bool, conf SlashSDKConf, options ...SlashOption) SlashSDKInterface {
	sh := &slashSDK{
		Debug:        debug,
		SlashSDKConf: &conf,
		Resty:        resty.New(),
	}

	for _, option := range options {
		option(sh)
	}

	if len(sh.RequestMiddleware) > 0 {
		sh.Resty.SetRequestMiddlewares(sh.RequestMiddleware...)
	}
	if len(sh.ResponseMiddlewares) > 0 {
		sh.Resty.SetResponseMiddlewares(sh.ResponseMiddlewares...)
	}
	return sh
}

/** 对外 接口层 实现 **/
func (sh *slashSDK) CardPostPre(params *CardPostRequest) (json.RawMessage, error) {
	cardPostParams, err := sh.cardPostParams(params)
	if err != nil {
		return nil, err
	}

	marshal, err := json.Marshal(cardPostParams)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(marshal), nil
}

// PostCard 开卡接口
//
// 由于 cardPost有参数的缺失，返回值只能返回 ID，这个 id 是 salsh 的 card的id
func (sh *slashSDK) CardPost(ctx context.Context, params *CardPostRequest) (string, error) {
	cardPostParams, err := sh.cardPostParams(params)
	if err != nil {
		return "", err
	}

	result, err := sh.cardPost(ctx, params.RequestID, cardPostParams)
	if err != nil {
		return "", err
	}

	if result.ID == "" {
		return "", fmt.Errorf("slash open card failed, empty card id returned")
	}

	return result.ID, nil
}

// GetCardByID 获取卡信息
func (sh *slashSDK) GetCardByID(ctx context.Context, cardID string) (*GetCardByIDResponse, error) {
	if cardID == "" {
		return nil, errors.New("slash get card by id cardID is required")
	}

	result, err := sh.getCardByID(ctx, cardID, true, true)
	if err != nil {
		return nil, err
	}

	// 枚举转换
	status := cardStatusToCardStatusEnum(result.Status)
	// 拦截未知的枚举
	if status == CardStatusEnumUnknow {
		return nil, errors.New("slash getCardById but unknow status, please check")
	}

	if slices.Contains([]CardStatusEnum{CardStatusEnumActive, CardStatusEnumPaused}, status) &&
		len(result.ExpiryYear) != 4 {
		return nil, errors.New("slash getCardByID but return expire return not len 4")
	}

	return &GetCardByIDResponse{
		Status:      status,
		Pan:         result.Pan,
		Cvv:         result.Cvv,
		ExpiryMonth: result.ExpiryMonth,
		ExpiryYear:  result.ExpiryYear,
	}, nil
}

func (sh *slashSDK) CardFrozen(ctx context.Context, cardID string) error {
	// Slash 未提供状态更新任务 ID；上层需通过 GetCardByID 轮询确认 paused 是否生效。
	if cardID == "" {
		return ErrEmptyCardID
	}

	_, err := sh.restyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodPatch,
		Path:   "/card/" + cardID,
		Params: &UpdateCardRequest{Status: string(CardStatusEnumPaused)},
	})
	return err
}

func (sh *slashSDK) CardUnfrozen(ctx context.Context, cardID string) error {
	// 与冻结一样，PATCH 成功仅代表请求提交，active 状态仍由上层查询确认。
	if cardID == "" {
		return ErrEmptyCardID
	}

	_, err := sh.restyRequest(ctx, &RestyRequestOptions{
		Method: http.MethodPatch,
		Path:   "/card/" + cardID,
		Params: &UpdateCardRequest{Status: string(CardStatusEnumActive)},
	})
	return err
}

// GetVirtualAccount 获取当前虚拟账户
func (sh *slashSDK) GetVirtualAccount() string {
	return sh.SlashSDKConf.VirtualAccountID
}

/** 私有 httpClient 实现 **/

func (sh *slashSDK) cardPostParams(params *CardPostRequest) (*cardPostRequest, error) {
	if params.Name == "" {
		return nil, errors.New("slash open card name is required")
	}
	return &cardPostRequest{
		Name: params.Name,
		UserData: struct {
			CardID    string `json:"cardId"`
			RequestOD string `json:"requestId"`
		}{
			CardID:    params.CardID,
			RequestOD: params.RequestID,
		},
		Type:             "virtual", // 固定虚拟卡
		AccountID:        sh.SlashSDKConf.AccountID,
		VirtualAccountID: sh.SlashSDKConf.VirtualAccountID,
		CardGroupID:      sh.SlashSDKConf.CardGroupID,
		CardProductID:    params.CardProductID,
	}, nil
}

// cardPost 开卡接口
//
// 文档地址: https://docs.slash.com/api-reference/card-post
//
// 特别注意1: 开卡的时候，幂等请求需要在 Headers 里设置 X-Idempotency-Key 字段没有生效
//
// 特别注意2: 返回参数有缺失
//
//	{
//	   "last4": ""
//	   "expiryYear": "",
//	   "expiryMonth": "",
//	   "expiryYear": "",
//	}
func (sh *slashSDK) cardPost(ctx context.Context, requestID string, params *cardPostRequest) (*card, error) {
	const path = "/card"
	var c *card

	_, err := sh.restyRequest(ctx, &RestyRequestOptions{
		Method:    http.MethodPost,
		RequestID: requestID,
		Path:      path,
		Params:    params,
		Result:    &c,
	})
	if err != nil {
		return nil, err
	}
	return c, nil
}

// getCardByID 获取卡信息
//
// 文档地址: https://docs.slash.com/api-reference/card-get-by-id
func (sh *slashSDK) getCardByID(ctx context.Context, cardID string, includePan bool, includeCvv bool) (*card, error) {
	var (
		result  *card
		apipath = "/card/" + cardID
	)

	if cardID == "" {
		return nil, ErrEmptyCardID
	}

	params := map[string]bool{
		"include_pan": includePan,
		"include_cvv": includeCvv,
	}

	_, err := sh.restyRequest(ctx,
		&RestyRequestOptions{Method: http.MethodGet,
			Url:    sh.SlashSDKConf.VaultUrl,
			Path:   apipath,
			Params: params,
			Result: &result})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// restyRequest 统一请求方法
func (sh *slashSDK) restyRequest(ctx context.Context, options *RestyRequestOptions) (*resty.Response, error) {
	// 生成请求头
	headers := GenerateHeaders(sh.SlashSDKConf.ApiKey, options.RequestID, sh.SlashSDKConf.LegalEntity)
	apiUrl := options.Url
	if apiUrl == "" {
		apiUrl = sh.SlashSDKConf.Url
	}

	apiurl, err := url.JoinPath(apiUrl, options.Path)
	if err != nil {
		return nil, fmt.Errorf("构建URL失败: %w", err)
	}

	// 创建请求
	req := sh.Resty.R().
		SetContext(ctx).
		SetURL(apiurl).
		SetHeaders(headers).
		SetMethod(options.Method).
		SetDebug(sh.Debug).
		SetResult(&options.Result)

	if options.Method != http.MethodGet {
		req.SetRetryCount(0)
	}
	// 根据请求方法和body类型设置请求参数
	if options.Params != nil {
		switch options.Method {
		case http.MethodGet:
			// 对于GET请求，将body作为查询参数
			switch p := options.Params.(type) {
			case map[string]any:
				for k, v := range p {
					req.SetQueryParam(k, fmt.Sprintf("%v", v))
				}
			default:
				jsonData, err := json.Marshal(options.Params)
				if err != nil {
					return nil, fmt.Errorf("序列化请求参数失败: %w", err)
				}

				var queryParams map[string]any
				if err := json.Unmarshal(jsonData, &queryParams); err != nil {
					return nil, fmt.Errorf("解析请求参数失败: %w", err)
				}

				for k, v := range queryParams {
					if v != nil {
						req.SetQueryParam(k, fmt.Sprintf("%v", v))
					}
				}
			}
		default:
			if options != nil && options.RequestID != "" {
				req.SetQueryParam("requestId", options.RequestID)
			}
			if options != nil && options.Params != nil {
				req.SetBody(options.Params)
			}

		}
	}

	res, err := req.Send()
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}

	if res.IsError() {
		return nil, fmt.Errorf("请求失败: %s", res.String())
	}

	return res, nil
}
