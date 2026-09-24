package pingpong

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"tman/pkg/dealer/crypto"

	"resty.dev/v3"
)

type PingPongConf struct {
	Url         string
	AppId       string
	AppSecret   string
	PrivateKey  string
	SignVersion string
	Debug       bool
}

// Conf 提供网关地址、应用凭据和已在 PingPong 开发者中心绑定的签名私钥。
type Conf interface {
	GetDebug() bool
	// GetUrl 返回网关基础地址。
	GetUrl() string
	// GetAppId 返回应用 ID。
	GetAppId() string
	// GetAppSecret 返回应用密钥。
	GetAppSecret() string
	// GetPrivateKey 返回用于请求签名的 RSA 私钥。
	GetPrivateKey() string
	// GetSignVersion 返回非对称密钥版本号，由 PingPong 配置密钥后提供。
	GetSignVersion() string
}

// PingPongSDK 封装 PingPong 网关请求。
type PingPongSDK struct {
	resty               *resty.Client
	conf                Conf
	requestMiddleware   []resty.RequestMiddleware
	responseMiddlewares []resty.ResponseMiddleware
}

// Option 用于配置 PingPongSDK。
type Option func(*PingPongSDK)

// New 创建并配置 PingPongSDK。
func New(conf Conf, options ...Option) *PingPongSDK {
	pp := &PingPongSDK{
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

// WithRequestMiddlewares 设置请求中间件
func WithRequestMiddlewares(middlewares ...resty.RequestMiddleware) Option {
	return func(pp *PingPongSDK) {
		pp.requestMiddleware = append(pp.requestMiddleware, resty.PrepareRequestMiddleware)
		pp.requestMiddleware = append(pp.requestMiddleware, middlewares...)
	}
}

// WithResponseMiddlewares 设置响应中间件
func WithResponseMiddlewares(middlewares ...resty.ResponseMiddleware) Option {
	return func(p *PingPongSDK) {
		p.responseMiddlewares = append(p.responseMiddlewares, resty.AutoParseResponseMiddleware)
		p.responseMiddlewares = append(p.responseMiddlewares, middlewares...)
	}
}

// APIError 表示网关返回的 HTTP 或业务错误；不会在错误信息中打印敏感响应体。
type APIError struct {
	StatusCode int    // StatusCode 是 HTTP 状态码。
	Code       string // Code 是网关业务错误码。
	Message    string // Message 是网关错误信息。
	Reason     string // Reason 是网关返回的具体失败原因。
}

// Error 优先返回三方错误码和原因；没有三方错误码时保留 HTTP 状态码。
func (e *APIError) Error() string {
	if e.Code == "" {
		return fmt.Sprintf("pingpong: HTTP %d: %s", e.StatusCode, e.Message)
	}
	if e.Reason != "" {
		return fmt.Sprintf("code=%s, message=%s, reason=%s", e.Code, e.Message, e.Reason)
	}
	return fmt.Sprintf("code=%s, message=%s", e.Code, e.Message)
}

// request 构建并发送网关请求，处理签名、认证和业务响应。
func (p *PingPongSDK) request(ctx context.Context, method, path, token string, query url.Values, body, result any) error {

	headers := map[string]string{
		"Content-Type": "application/json",
	}
	if token != "" {
		headers["Authorization"] = token
	}
	apiUrl, err := url.JoinPath(p.conf.GetUrl(), path)
	if err != nil {
		return fmt.Errorf("构建URL失败: %w", err)
	}
	// 4. 使用 resty 发送 POST 请求
	client := resty.New()

	req := client.R().
		SetContext(ctx).
		SetMethod(method).
		EnableDebug().
		SetURL(apiUrl).
		SetRetryCount(0)
	if token != "" {
		req.SetHeader("Authorization", "Bearer "+token)
	}
	if len(query) != 0 {
		req.SetQueryParamsFromValues(query)
	}
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("pingpong: 序列化请求失败: %w", err)
		}
		signature, err := crypto.Sign(data, p.conf.GetPrivateKey())
		if err != nil {
			return err
		}
		req.SetHeader("Content-Type", "application/json").SetHeader("sign", signature).
			SetHeader("sign-version", p.conf.GetSignVersion()).SetBody(data)
	}
	resp, err := req.Send()
	if err != nil {
		return fmt.Errorf("pingpong: 发送请求失败: %w", err)
	}

	// 网关通常返回 code/message/data；部分文档示例直接展示 data 对象。
	var envelope struct {
		Code    json.RawMessage `json:"code"`    // 网关业务状态码，兼容字符串和数字。
		Message string          `json:"message"` // 网关错误信息。
		Msg     string          `json:"msg"`     // 兼容的网关错误信息字段。
		Details struct {
			Reason string `json:"reason"`
		} `json:"details"`
		Data json.RawMessage `json:"data"` // 封装的业务响应数据。
	}
	if err := json.Unmarshal(resp.Bytes(), &envelope); err != nil {
		if !resp.IsSuccess() {
			return &APIError{StatusCode: resp.StatusCode(), Message: http.StatusText(resp.StatusCode())}
		}
		return fmt.Errorf("pingpong: 解析响应失败: %w", err)
	}
	message := envelope.Message
	if message == "" {
		message = envelope.Msg
	}
	code := strings.Trim(string(envelope.Code), `"`)
	if !resp.IsSuccess() || (len(envelope.Code) != 0 && code != "0") {
		if message == "" {
			message = http.StatusText(resp.StatusCode())
		}
		return &APIError{StatusCode: resp.StatusCode(), Code: code, Message: message, Reason: envelope.Details.Reason}
	}
	if result == nil {
		return nil
	}
	data := resp.Bytes()
	if envelope.Code != nil {
		data = envelope.Data
	}
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	if err := json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("pingpong: 解析业务响应失败: %w", err)
	}
	return nil
}

// GetAccessToken 使用 APP ID 和 APP Key 获取访问令牌；调用方应根据 expires_in 统一缓存、刷新令牌。
func (p *PingPongSDK) GetAccessToken(ctx context.Context) (*AccessTokenResponse, error) {
	query := url.Values{"app_id": {p.conf.GetAppId()}, "app_secret": {p.conf.GetAppSecret()}}
	var result AccessTokenResponse
	if err := p.request(ctx, http.MethodGet, "/v2/token/get", "", query, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// QueryAccountsBalances 查询账户余额。大账户，真实的账户余额应该看所有预算账户的和
func (p *PingPongSDK) QueryAccountsBalances(ctx context.Context, token string, params *QueryAccountsBalancesRequest) (*QueryAccountsBalancesResponse, error) {
	var result QueryAccountsBalancesResponse
	if err := p.request(ctx, http.MethodPost, "/api/reporting/v3/account-balance", token, nil, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// QueryCardProducts 查询可用的卡产品。
func (p *PingPongSDK) QueryCardProducts(ctx context.Context, token string) (*QueryCardProductsResponse, error) {
	var result QueryCardProductsResponse
	if err := p.request(ctx, http.MethodGet, "/api/issuing/v3/card-products", token, nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateCard 申请创建卡片。
func (p *PingPongSDK) CreateCard(ctx context.Context, token string, params *CreateCardRequest) (*CreateCardResponse, error) {
	var result CreateCardResponse
	if err := p.request(ctx, http.MethodPost, "/api/issuing/v3/cards/apply", token, nil, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetCardDetails 根据卡片 ID 查询卡片详情。
func (p *PingPongSDK) GetCardDetails(ctx context.Context, token, cardID string) (*CardDetailsResponse, error) {
	var result CardDetailsResponse
	if err := p.request(ctx, http.MethodGet, "/api/issuing/v3/cards/details", token, url.Values{"card_id": {cardID}}, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CardFunding 执行卡片充值或转出操作。
func (p *PingPongSDK) CardFunding(ctx context.Context, token string, params *CardFundingRequest) (*CardFundingResponse, error) {
	var result CardFundingResponse
	if err := p.request(ctx, http.MethodPost, "/api/issuing/v3/cards/funding/actions", token, nil, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// QueryCardFundingOrders 查询卡片充值订单。
func (p *PingPongSDK) QueryCardFundingOrders(ctx context.Context, token string, params *QueryCardFundingOrdersRequest) (*QueryCardFundingOrdersResponse, error) {
	var result QueryCardFundingOrdersResponse
	if err := p.request(ctx, http.MethodGet, "/api/issuing/v3/card/funding/orders", token, params.values(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// QueryDedicatedCardBalance 根据卡片 ID 查询专用卡余额。
func (p *PingPongSDK) QueryDedicatedCardBalance(ctx context.Context, token, cardID string) (*DedicatedCardBalanceResponse, error) {
	var result DedicatedCardBalanceResponse
	if err := p.request(ctx, http.MethodGet, "/api/issuing/v3/card/balance", token, url.Values{"card_id": {cardID}}, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CardAction 支持 freeze、unfreeze、close 和 update_remark（close/update_remark 需提供 remark）。
func (p *PingPongSDK) CardAction(ctx context.Context, token string, params *CardActionRequest) error {
	return p.request(ctx, http.MethodPost, "/api/issuing/v3/cards/actions", token, nil, params, nil)
}

// QueryCardTransactions 查询卡片交易记录。
func (p *PingPongSDK) QueryCardTransactions(ctx context.Context, token string, params *QueryCardTransactionsRequest) (*QueryCardTransactionsResponse, error) {
	var result QueryCardTransactionsResponse
	if err := p.request(ctx, http.MethodGet, "/api/issuing/v3/transactions", token, params.values(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Query3DSDetails 根据卡片 ID 查询 3DS 详情。
func (p *PingPongSDK) Query3DSDetails(ctx context.Context, token, cardID string) (*ThreeDSDetailsResponse, error) {
	var result ThreeDSDetailsResponse
	if err := p.request(ctx, http.MethodGet, "/api/issuing/v3/cards/3ds/details", token, url.Values{"card_id": {cardID}}, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateBudgetAccount 创建预算账户。
func (p *PingPongSDK) CreateBudgetAccount(ctx context.Context, token string, params *CreateBudgetAccountRequest) (*CreateBudgetAccountResponse, error) {
	var result CreateBudgetAccountResponse
	if err := p.request(ctx, http.MethodPost, "/api/issuing/v3/budgets", token, nil, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BudgetFunding 对预算账户充值或在预算账户间划转资金。
func (p *PingPongSDK) BudgetFunding(ctx context.Context, token string, params *BudgetFundingRequest) (*BudgetFundingResponse, error) {
	var result BudgetFundingResponse
	if err := p.request(ctx, http.MethodPost, "/api/issuing/v3/budgets/funding", token, nil, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// QueryBudgetFundingOrder 按订单 ID 和操作类型查询预算账户资金订单状态。
func (p *PingPongSDK) QueryBudgetFundingOrder(ctx context.Context, token string, params *QueryBudgetFundingOrderRequest) (*QueryBudgetFundingOrderResponse, error) {
	var result QueryBudgetFundingOrderResponse
	if err := p.request(ctx, http.MethodGet, "/api/issuing/v3/funding/orders", token, params.values(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// QueryBudgetAccountBalance 查询指定预算账户余额；budgetID 为空时查询全部预算账户。
func (p *PingPongSDK) QueryBudgetAccountBalance(ctx context.Context, token, budgetID string) (*QueryBudgetAccountBalanceResponse, error) {
	query := url.Values{}
	if budgetID != "" {
		query.Set("budget_id", budgetID)
	}
	var result QueryBudgetAccountBalanceResponse
	if err := p.request(ctx, http.MethodGet, "/api/issuing/v3/budget/balance", token, query, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// QueryAccountTransactions 查询预算账户或卡片的已入账交易，最长查询 31 天。
func (p *PingPongSDK) QueryAccountTransactions(ctx context.Context, token string, params *QueryAccountTransactionsRequest) (*QueryAccountTransactionsResponse, error) {
	var result QueryAccountTransactionsResponse
	if err := p.request(ctx, http.MethodGet, "/api/issuing/v4/account/transactions", token, params.values(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
