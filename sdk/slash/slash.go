package slash

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
	"resty.dev/v3"
)

const (
	DEFAULT_RETRY_TIMES = 10 // 开卡后查询卡信息重试次数
	CARD_INACTIVE_ERR   = "card inactive"
)

// Conf 配置接口
type Conf interface {
	GetApiKey() string
	GetAccount() string
	GetVirtualAccount() string
	GetCardGroup() string
	GetUrl() string
	GetVaultUrl() string
	GetDebug() bool
	GetRestriction() string
	GetCountryList() string
	GetRetryTimes() int32
	GetLegalEntity() string
}

// SlashSDK SDK结构体
type SlashSDK struct {
	resty               *resty.Client
	conf                Conf
	requestMiddleware   []resty.RequestMiddleware
	responseMiddlewares []resty.ResponseMiddleware
}

// Option SDK选项函数类型
type Option func(t *SlashSDK)

// WithRequestMiddlewares 设置请求中间件
func WithRequestMiddlewares(middlewares ...resty.RequestMiddleware) Option {
	return func(sh *SlashSDK) {
		sh.requestMiddleware = append(sh.requestMiddleware, resty.PrepareRequestMiddleware)
		sh.requestMiddleware = append(sh.requestMiddleware, middlewares...)
	}
}

// WithResponseMiddlewares 设置响应中间件
func WithResponseMiddlewares(middlewares ...resty.ResponseMiddleware) Option {
	return func(sh *SlashSDK) {
		sh.responseMiddlewares = append(sh.responseMiddlewares, resty.AutoParseResponseMiddleware)
		sh.responseMiddlewares = append(sh.responseMiddlewares, middlewares...)
	}
}

// New 初始化SlashPaySDK
func New(conf Conf, options ...Option) *SlashSDK {
	sh := &SlashSDK{
		conf:                conf,
		resty:               resty.New(),
		requestMiddleware:   []resty.RequestMiddleware{},
		responseMiddlewares: []resty.ResponseMiddleware{},
	}

	for _, option := range options {
		option(sh)
	}

	if len(sh.requestMiddleware) > 0 {
		sh.resty.SetRequestMiddlewares(sh.requestMiddleware...)
	}
	if len(sh.responseMiddlewares) > 0 {
		sh.resty.SetResponseMiddlewares(sh.responseMiddlewares...)
	}
	return sh
}

// GenerateHeaders
func GenerateHeaders(apiKey, idempotencyKey, legalEntity string) map[string]string {

	headers := map[string]string{
		"X-API-Key":         apiKey,
		"X-Idempotency-Key": idempotencyKey,
	}
	if legalEntity != "" { // fluxo需要legal-entity
		headers["x-legal-entity"] = legalEntity
	}
	return headers
}

// RestyRequest 创建并发送请求，处理响应
func (sh *SlashSDK) RestyRequest(ctx context.Context, options *RestyRequestOptions) (*resty.Response, error) {
	// 生成请求头
	headers := GenerateHeaders(sh.conf.GetApiKey(), options.RequestID, sh.conf.GetLegalEntity())
	apiUrl := options.Url
	if apiUrl == "" {
		apiUrl = sh.conf.GetUrl()
	}

	apiurl, err := url.JoinPath(apiUrl, options.Path)
	if err != nil {
		return nil, fmt.Errorf("构建URL失败: %w", err)
	}

	// 创建请求
	req := sh.resty.R().
		SetContext(ctx).
		SetURL(apiurl).
		SetHeaders(headers).
		SetAllowNonIdempotentRetry(false).
		SetMethod(options.Method).
		SetDebug(sh.conf.GetDebug()).
		SetResult(&options.Result)

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

// List legal entities
// path: /legal-entity
// method: GET
func (sh *SlashSDK) ListLegalEntity(ctx context.Context) ([]LegalEntity, error) {
	var (
		apipath = "/legal-entity"
		result  *LegalEntitiesResponse
	)
	_, err := sh.RestyRequest(ctx, &RestyRequestOptions{Method: http.MethodGet, Url: sh.conf.GetUrl(), Path: apipath, Result: &result})
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

// List accounts
// path: /account
// method: GET
func (sh *SlashSDK) ListAccounts(ctx context.Context) ([]Account, error) {

	var (
		apipath = "/account"
		result  *AccountsResponse
	)
	_, err := sh.RestyRequest(ctx, &RestyRequestOptions{Method: http.MethodGet, Url: sh.conf.GetUrl(), Path: apipath, Result: &result})
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

// Get account
// path: /account/{accountId}
// method: GET
func (sh *SlashSDK) GetAccount(ctx context.Context) (*Account, error) {
	var (
		accountId = sh.conf.GetAccount()
		result    *Account
	)
	apipath := fmt.Sprintf("/account/%s", accountId)
	_, err := sh.RestyRequest(ctx, &RestyRequestOptions{Method: http.MethodGet, Url: sh.conf.GetUrl(), Path: apipath, Result: &result})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// List account balances
// path: /account/{accountId}/balance
// method: GET
func (sh *SlashSDK) ListAccountBalances(ctx context.Context) ([]Balance, error) {
	var (
		accountId = sh.conf.GetAccount()
		result    *AccountBalancesResponse
	)

	apipath := fmt.Sprintf("/account/%s/balance", accountId)
	_, err := sh.RestyRequest(ctx, &RestyRequestOptions{Method: http.MethodGet, Url: sh.conf.GetUrl(), Path: apipath, Result: &result})
	if err != nil {
		return nil, err
	}
	return result.Balances, nil
}

// 创建VirtualAccount账户
// path: /virtual-account
// method: POST
func (sh *SlashSDK) CreateVirtualAccount(ctx context.Context, name string) (*VirtualAccountResponse, error) {
	var (
		apipath = "/virtual-account"
		result  *VirtualAccountResponse
	)

	if name == "" {
		return nil, errors.New("name is required")
	}
	req := CreateVirtualAccountRequest{
		AccountID: sh.conf.GetAccount(),
		Name:      name,
		CommissionDetails: CommissionDetails{
			Amount: Amount{
				AmountCents: 0,
			},
			Frequency: COMMISSION_FREQUENCY_MONTHLY,
			StartDate: time.Now(),
			Type:      COMMISSION_TYPE_FLATFEE,
		},
	}
	_, err := sh.RestyRequest(ctx, &RestyRequestOptions{Method: http.MethodPost, Url: sh.conf.GetUrl(), Path: apipath, Params: req, Result: &result})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 更新va账户
// path: /virtual-account/{virtualAccountId}
// method: PATCH
func (sh *SlashSDK) UpdateVirtualAccount(ctx context.Context, vaId string, name string) (*VirtualAccountResponse, error) {
	var (
		apipath = fmt.Sprintf("/virtual-account/%s", vaId)
		result  *VirtualAccountResponse
	)

	req := UpdateVirtualAccountRequest{
		Action: "update",
		Name:   name,
		// CommissionDetails: &CommissionDetails{
		// 	Amount: Amount{
		// 		AmountCents: 0,
		// 	},
		// 	Frequency: COMMISSION_FREQUENCY_MONTHLY,
		// 	StartDate: time.Now(),
		// 	Type:      COMMISSION_TYPE_FLATFEE,
		// },
	}
	_, err := sh.RestyRequest(ctx, &RestyRequestOptions{Method: http.MethodPatch, Url: sh.conf.GetUrl(), Path: apipath, Params: req, Result: &result})
	if err != nil {
		return nil, err
	}
	return result, nil

}

// 获取va账户详情
// path: /virtual-account/{virtualAccountId}
// method: GET
func (sh *SlashSDK) GetVirtualAccount(ctx context.Context) (*VirtualAccount, error) {
	var result *VirtualAccount

	va := sh.conf.GetVirtualAccount()
	apipath := fmt.Sprintf("/virtual-account/%s", va)

	_, err := sh.RestyRequest(ctx, &RestyRequestOptions{Method: http.MethodGet, Url: sh.conf.GetUrl(), Path: apipath, Result: &result})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 查询va账户
// path: /virtual-account
// method: GET
func (sh *SlashSDK) ListVirtualAccounts(ctx context.Context) ([]VirtualAccount, error) {
	var (
		apipath = "/virtual-account"
		result  *ListVirtualAccountResponse
	)

	_, err := sh.RestyRequest(ctx, &RestyRequestOptions{Method: http.MethodGet, Url: sh.conf.GetUrl(), Path: apipath, Result: &result})

	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

// 资金账户钱包转账
// path:  /transfer/virtual-account
// method: POST
func (sh *SlashSDK) VirtualAccountTransfer(ctx context.Context, param *TransferRequest) (*TransferResponse, error) {
	var (
		apipath = "/transfer/virtual-account"
		result  *TransferResponse
	)
	if err := param.Validate(); err != nil {
		return nil, err
	}

	// param.Source = sh.conf.GetVirtualAccount()
	_, err := sh.RestyRequest(ctx, &RestyRequestOptions{
		Method:    http.MethodPost,
		Url:       sh.conf.GetUrl(),
		RequestID: param.RequestID,
		Path:      apipath,
		Params:    param,
		Result:    &result})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 获取卡BIN信息
// path:  /card-product
// method: GET
func (sh *SlashSDK) ListCardProducts(ctx context.Context) ([]CardProduct, error) {
	var (
		result  *CardProductsResponse
		apipath = "/card-product"
	)
	_, err := sh.RestyRequest(ctx, &RestyRequestOptions{Method: http.MethodGet, Url: sh.conf.GetUrl(), Path: apipath, Result: &result})

	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

// 新增卡
// path： /card
// method： POST
func (sh *SlashSDK) createCard(ctx context.Context, in *CreateCardReq) (*Card, error) {
	var (
		apipath = "/card"
		result  *Card
	)

	if in == nil {
		return nil, ErrEmptyParams
	}
	// 验证请求参数
	if err := in.Validate(); err != nil {
		return nil, err
	}
	balanceAccountID := sh.conf.GetAccount()
	VirtualAccountID := sh.conf.GetVirtualAccount()
	cardGroupID := sh.conf.GetCardGroup()

	aountryRule := &CountryRule{
		Countries:   []string{},
		Restriction: RESTRICTION_BLACKLIST,
	}
	// 根据配置设置国家限制
	restriction := sh.conf.GetRestriction()
	if restriction != "" {
		aountryRule.Restriction = restriction

		countryList := sh.conf.GetCountryList()
		countries := strings.Split(countryList, ",")
		for i, v := range countries {
			countries[i] = strings.ToUpper(strings.TrimSpace(v))
		}
		aountryRule.Countries = countries
	}

	request := CreateCardRequest{
		AccountID:     balanceAccountID,
		CardGroupID:   cardGroupID,
		CardProductID: in.CardBinID,
		IsSingleUse:   false,
		Name:          in.Name,
		SpendingConstraint: &SpendingConstraint{
			CountryRule: aountryRule,
		},
		Type:             CARD_TYPE_VIRTUAL,
		UserData:         UserData{CardID: in.CardID, RequestID: in.RequestID},
		VirtualAccountID: VirtualAccountID,
	}

	_, err := sh.RestyRequest(ctx, &RestyRequestOptions{Method: http.MethodPost, Url: sh.conf.GetUrl(), Path: apipath, Params: request, Result: &result})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (sh *SlashSDK) CreateCard(ctx context.Context, in *CreateCardReq) (*Card, error) {
	r, err := sh.createCard(ctx, in)
	if err != nil {
		return nil, err
	}

	retryTimes := sh.conf.GetRetryTimes()
	if retryTimes == 0 {
		retryTimes = DEFAULT_RETRY_TIMES
	}
	retryOptions := []retry.Option{
		retry.Attempts(uint(retryTimes)),
		retry.Delay(time.Second), // 查询间隔
		retry.DelayType(retry.FixedDelay),
		retry.LastErrorOnly(true),
		retry.Context(ctx),
	}

	// 卡信息查询
	qryFunc := func() (*Card, error) {
		card, getErr := sh.GetCard(ctx, r.ID)
		if getErr != nil {
			return nil, getErr
		}
		if card.Status == CARD_STATUS_INACTIVE {
			return nil, errors.New(CARD_INACTIVE_ERR)
		}
		return card, nil
	}

	card, err := retry.DoWithData(qryFunc, retryOptions...)
	if err != nil && err.Error() == CARD_INACTIVE_ERR {
		return r, nil
	}

	return card, err
}

// 获取卡信息(带CVVPAN)
// path： /card/{cardId}
// method： GET
func (sh *SlashSDK) GetCard(ctx context.Context, cardID string) (*Card, error) {
	var result *Card

	if cardID == "" {
		return nil, ErrEmptyCardID
	}
	apipath := fmt.Sprintf("/card/%s", cardID)

	params := map[string]string{
		"include_pan": "true",
		"include_cvv": "true",
	}

	_, err := sh.RestyRequest(ctx,
		&RestyRequestOptions{Method: http.MethodGet,
			Url:    sh.conf.GetVaultUrl(),
			Path:   apipath,
			Params: params,
			Result: &result})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 获取卡信息（不带CVV、PAN）
// path： /card/{cardId}
// method： GET
func (sh *SlashSDK) QueryCard(ctx context.Context, cardID string) (*Card, error) {
	var result *Card

	if cardID == "" {
		return nil, ErrEmptyCardID
	}
	apipath := fmt.Sprintf("/card/%s", cardID)

	_, err := sh.RestyRequest(ctx,
		&RestyRequestOptions{Method: http.MethodGet,
			Url:    sh.conf.GetUrl(),
			Path:   apipath,
			Result: &result})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 卡列表
// path： /card
// method： GET
func (sh *SlashSDK) ListCards(ctx context.Context, in *QueryCardsParams) (*ListCardsResponse, error) {
	var (
		result  *ListCardsResponse
		apipath = "/card"
	)

	_, err := sh.RestyRequest(ctx,
		&RestyRequestOptions{
			Method: http.MethodGet,
			Url:    sh.conf.GetUrl(),
			Path:   apipath,
			Params: in.BuildQuery(),
			Result: &result}) // TODO: 分批次查询
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 银行卡冻结
// path： /card/{cardID}
// method： PATCH
func (sh *SlashSDK) FrozenCard(ctx context.Context, cardID string) error {
	if cardID == "" {
		return ErrEmptyCardID
	}
	apipath := fmt.Sprintf("/card/%s", cardID)
	cardReq := UpdateCardRequest{
		Status: CARD_STATUS_PAUSED,
	}

	_, err := sh.RestyRequest(ctx, &RestyRequestOptions{Method: http.MethodPatch, Url: sh.conf.GetUrl(), Path: apipath, Params: &cardReq})
	return err
}

// 银行卡解冻
// path： /card/{cardID}
// method： PATCH
func (sh *SlashSDK) UnfrozenCard(ctx context.Context, cardID string) error {
	if cardID == "" {
		return ErrEmptyCardID
	}
	apipath := fmt.Sprintf("/card/%s", cardID)
	cardReq := UpdateCardRequest{
		Status: CARD_STATUS_ACTIVE,
	}

	_, err := sh.RestyRequest(ctx, &RestyRequestOptions{Method: http.MethodPatch, Url: sh.conf.GetUrl(), Path: apipath, Params: &cardReq})
	return err
}

// closed 银行卡
// path： /card/{cardID}
// method： PATCH
func (sh *SlashSDK) ReleaseCard(ctx context.Context, cardID string) error {
	if cardID == "" {
		return ErrEmptyCardID
	}
	apipath := fmt.Sprintf("/card/%s", cardID)
	cardReq := UpdateCardRequest{
		Status: CARD_STATUS_CLOSED, //active, paused, inactive, closed
	}

	_, err := sh.RestyRequest(ctx, &RestyRequestOptions{Method: http.MethodPatch, Url: sh.conf.GetUrl(), Path: apipath, Params: &cardReq})
	return err
}

// List transactions
// path: /transaction
// method: GET
func (sh *SlashSDK) ListTransactions(ctx context.Context, params QueryTransactionsParams) (*TransactionsResponse, error) {
	var (
		result  *TransactionsResponse
		apipath = "/transaction"
	)
	params.AccountId = sh.conf.GetAccount()

	_, err := sh.RestyRequest(ctx,
		&RestyRequestOptions{Method: http.MethodGet, Url: sh.conf.GetUrl(), Path: apipath, Params: params, Result: &result})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Get transaction
// path: /transaction/{transactionId}
func (sh *SlashSDK) GetTransaction(ctx context.Context, transactionId string) (*Transaction, error) {
	var result *Transaction

	apipath := fmt.Sprintf("/transaction/%s", transactionId)
	_, err := sh.RestyRequest(ctx,
		&RestyRequestOptions{Method: http.MethodGet, Url: sh.conf.GetUrl(), Path: apipath, Result: &result})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, ErrEmptyResult
	}
	return result, nil
}
