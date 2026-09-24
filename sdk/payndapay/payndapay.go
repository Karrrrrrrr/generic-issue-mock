package payndapay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"sdk/crypto"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
	"resty.dev/v3"
)

// const RETRYCOUNT = 3

// GenerateHeaders 生成带签名的请求头
func GenerateHeaders(appID, appSecret, nonce, path string) map[string]string {

	// nonce := fmt.Sprint(core.Global.GenSnowflakeID())
	if nonce == "" {
		nonce = genNonce()
	}
	signParams := &crypto.SignatureParams{
		AppID:     appID,
		AppSecret: appSecret,
		Path:      path,
		Nonce:     nonce,
		Timestamp: fmt.Sprint(time.Now().Unix()),
	}
	signature := crypto.GenerateSignature(signParams)

	return map[string]string{
		"appId":     appID,
		"timestamp": signParams.Timestamp,
		"nonce":     nonce,
		"sign":      signature,
	}
}

// Conf 配置接口
type Conf interface {
	GetAppId() string
	GetAppSecret() string
	GetBalanceAccount() string
	GetUrl() string
	GetDebug() bool
	GetDisabledBins() string
}

// PayndaPaySDK SDK结构体
type PayndaPaySDK struct {
	resty               *resty.Client
	conf                Conf
	requestMiddleware   []resty.RequestMiddleware
	responseMiddlewares []resty.ResponseMiddleware
}

// Option SDK选项函数类型
type Option func(t *PayndaPaySDK)

// WithRequestMiddlewares 设置请求中间件
func WithRequestMiddlewares(middlewares ...resty.RequestMiddleware) Option {
	return func(pp *PayndaPaySDK) {
		pp.requestMiddleware = append(pp.requestMiddleware, resty.PrepareRequestMiddleware)
		pp.requestMiddleware = append(pp.requestMiddleware, middlewares...)
	}
}

// WithResponseMiddlewares 设置响应中间件
func WithResponseMiddlewares(middlewares ...resty.ResponseMiddleware) Option {
	return func(pp *PayndaPaySDK) {
		pp.responseMiddlewares = append(pp.responseMiddlewares, resty.AutoParseResponseMiddleware)
		pp.responseMiddlewares = append(pp.responseMiddlewares, middlewares...)
	}
}

// New 初始化PayndaPaySDK
func New(conf Conf, options ...Option) *PayndaPaySDK {
	pp := &PayndaPaySDK{
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

// processResponse 处理API响应

// map[string]interface {}
// [
// "card": map[string]interface {} ["createTime": *(*interface {})(0x140000744a0), "merchantId": *(*interface {})(0x140000744c0), "firstName": *(*interface {})(0x140000744e0), "mobilePrefix": *(*interface {})(0x14000074500), "id": *(*interface {})(0x140000745a8), "cardholderId": *(*interface {})(0x140000745c8), "creditLimitType": *(*interface {})(0x140000745e8), "updateTime": *(*interface {})(0x140000746b0), "status": *(*interface {})(0x140000746d0), "maskCardNo": *(*interface {})(0x140000746f0), "mobile": *(*interface {})(0x14000074710), "email": *(*interface {})(0x14000074730), "singleUse": *(*interface {})(0x14000074750), "balanceAccountId": *(*interface {})(0x140000747b8), "cardBin": *(*interface {})(0x140000747d8), "currency": *(*interface {})(0x140000747f8), "lastName": *(*interface {})(0x14000074818), ],
// "sensitiveInfo": map[string]interface {} ["id": *(*interface {})(0x1400021a258), "createTime": *(*interface {})(0x1400021a278), "updateTime": *(*interface {})(0x1400021a298), "cardId": *(*interface {})(0x1400021a2b8), "cvv": *(*interface {})(0x1400021a2d8), "expirationDate": *(*interface {})(0x1400021a2f8), "cardNo": *(*interface {})(0x1400021a318), ],
// "balance": map[string]interface {} ["id": *(*interface {})(0x1400021a378), "createTime": *(*interface {})(0x1400021a398), "updateTime": *(*interface {})(0x1400021a3b8), "amountUsed": *(*interface {})(0x1400021a3d8), "amountFrozen": *(*interface {})(0x1400021a3f8), "amount": *(*interface {})(0x1400021a418), "availableAmount": *(*interface {})(0x1400021a438), ],
// ]
func (pp *PayndaPaySDK) processResponse(r *resty.Response, result any) error {
	if r == nil {
		return errors.New("payndapay response is nil")
	}
	if r.IsError() {
		if r.Err != nil {
			return r.Err
		}
		return fmt.Errorf("payndapay response error: status=%d body=%s", r.StatusCode(), r.String())
	}

	rv := reflect.ValueOf(result)
	if result == nil || rv.Kind() != reflect.Pointer || rv.IsNil() {
		return errors.New("result receiver must be a non-nil pointer")
	}

	var body struct {
		Code    json.Number     `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
		Success bool            `json:"success"`
	}

	decoder := json.NewDecoder(bytes.NewReader(r.Bytes()))
	decoder.UseNumber()
	if err := decoder.Decode(&body); err != nil {
		return err
	}

	if !(body.Code.String() == "200" || body.Success) {
		return fmt.Errorf("payndapay business error: code=%s message=%s", body.Code.String(), body.Message)
	}

	dataBytes := bytes.TrimSpace(body.Data)
	if len(dataBytes) == 0 || bytes.Equal(dataBytes, []byte("null")) {
		return ErrEmptyResponse
	}

	if dataBytes[0] == '{' && rv.Elem().Kind() == reflect.Slice {
		var dataMap map[string]json.RawMessage
		decoder := json.NewDecoder(bytes.NewReader(body.Data))
		decoder.UseNumber()
		if err := decoder.Decode(&dataMap); err != nil {
			return fmt.Errorf("解析响应数据失败: %w", err)
		}

		if records, ok := dataMap["records"]; ok {
			decoder := json.NewDecoder(bytes.NewReader(records))
			decoder.UseNumber()
			if err := decoder.Decode(result); err != nil {
				return fmt.Errorf("解析records数据到目标结构失败: %w", err)
			}
			return nil
		}
	}

	switch dataBytes[0] {
	case '{', '[':
		decoder := json.NewDecoder(bytes.NewReader(body.Data))
		decoder.UseNumber()
		if err := decoder.Decode(result); err != nil {
			return fmt.Errorf("解析响应数据到目标结构失败: %w", err)
		}
	default:
		return fmt.Errorf("响应数据格式错误: 未知类型")
	}

	return nil
}

func (pp *PayndaPaySDK) processResponseV0(resp *APIResponse, result any) error {
	if resp == nil {
		return ErrEmptyResponse
	}
	if !(resp.Code == 200 || resp.Success) {
		return fmt.Errorf("payndapay result error: code=%d message=%s", resp.Code, resp.Message)
	}
	if resp.Data == nil {
		return ErrEmptyResponse
	}

	rv := reflect.ValueOf(result)
	if result == nil || rv.Kind() != reflect.Pointer || rv.IsNil() {
		return errors.New("result receiver must be a non-nil pointer")
	}

	marshalUnmarshal := func(src any, dst any) error {
		b, err := json.Marshal(src)
		if err != nil {
			return fmt.Errorf("序列化响应数据失败: %w", err)
		}
		if err := json.Unmarshal(b, dst); err != nil {
			return fmt.Errorf("解析响应数据失败: %w", err)
		}
		return nil
	}

	switch data := resp.Data.(type) {
	case []any:
		return marshalUnmarshal(data, result)
	case map[string]any:
		if records, ok := data["records"]; ok && rv.Elem().Kind() == reflect.Slice {
			return marshalUnmarshal(records, result)
		}
		return marshalUnmarshal(data, result)
	default:
		return fmt.Errorf("响应数据格式错误: %T", data)
	}
}

// verifyResponseSignature 验证响应签名
func (pp *PayndaPaySDK) verifyResponseSignature(res *resty.Response, path string) error {
	// 从响应头中获取签名相关信息
	appID := res.Header().Get("appId")
	timestamp := res.Header().Get("timestamp")
	nonce := res.Header().Get("nonce")
	responseSign := res.Header().Get("sign")

	// 如果响应头中没有签名信息，则跳过验证
	if responseSign == "" {
		return ErrNoSignature
	}

	// 使用相同的参数生成签名
	signParams := &crypto.SignatureParams{
		AppID:     appID,
		AppSecret: pp.conf.GetAppSecret(),
		Path:      path,
		Nonce:     nonce,
		Timestamp: timestamp,
	}

	calculatedSign := crypto.GenerateSignature(signParams)

	// 比较计算出的签名与响应中的签名
	if calculatedSign != strings.ToLower(responseSign) {
		return ErrSignatureVerify
	}

	return nil
}

// buildAPIPath 构建API路径
func (pp *PayndaPaySDK) buildAPIPath(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}

// RestyRequest 创建并发送请求，处理响应
func (pp *PayndaPaySDK) RestyRequest(ctx context.Context, options *RestyRequestOptions) (*resty.Response, error) {
	if options == nil {
		return nil, errors.New("payndapay request options is nil")
	}
	// 生成请求头
	headers := GenerateHeaders(pp.conf.GetAppId(), pp.conf.GetAppSecret(), options.Nonce, options.Path)

	apiurl, err := url.JoinPath(pp.conf.GetUrl(), options.Path)
	if err != nil {
		return nil, fmt.Errorf("构建URL失败: %w", err)
	}

	// 创建请求
	req := pp.resty.R().
		SetContext(ctx).
		SetURL(apiurl).
		SetHeaders(headers).
		SetMethod(options.Method).
		SetAllowNonIdempotentRetry(false)

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
			if options.RequestID != "" {
				req.SetQueryParam("requestId", options.RequestID)
			}
			if options.Params != nil {
				req.SetBody(options.Params)
			}

		}
	}
	// var resp APIResponse
	// req.SetResult(&resp)

	if pp.conf.GetDebug() {
		req.EnableDebug()
	}
	res, err := req.Send()
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}

	// 检查响应状态
	if res.IsError() {
		return nil, fmt.Errorf("请求失败: %s", res.String())
	}

	// 验证响应签名
	if err := pp.verifyResponseSignature(res, options.Path); err != nil {
		// 如果是签名不存在的错误，可以考虑忽略或记录日志，但不中断流程
		if err != ErrNoSignature {
			return nil, err
		}
		// 可以在这里添加日志记录
	}

	// 处理响应数据
	if options.Result != nil {
		if err := pp.processResponse(res, options.Result); err != nil {
			return nil, err
		}
	}

	return res, nil
}

// 查询商户钱包
// path: /openapi/merchant/wallets
// method: GET
func (pp *PayndaPaySDK) GetMerchantWallet(ctx context.Context) ([]*MerchantWallet, error) {
	apipath := "/openapi/merchant/wallets"
	var result []*MerchantWallet
	nonce := genNonce()
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodGet, apipath, nonce, "", defaultPageQuery, &result})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// 创建资金账户
// path: /openapi/balanceAccounts
// method: POST
func (pp *PayndaPaySDK) CreateBalanceAccount(ctx context.Context, name string, nonce string) (*BalanceAccount, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	apipath := "/openapi/balanceAccounts"
	var result *BalanceAccount
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodPost, apipath, nonce, "", map[string]any{"name": name}, &result})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 修改资金账户
// path: /openapi/balanceAccounts/{id}
// method: PUT
func (pp *PayndaPaySDK) UpdateBalanceAccount(ctx context.Context, name string, nonce string) error {
	if name == "" {
		return errors.New("name is required")
	}
	balanceAccountID := pp.conf.GetBalanceAccount()
	apipath := fmt.Sprintf("/openapi/balanceAccounts/%s", balanceAccountID)
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodPut, apipath, nonce, "", map[string]any{"name": name}, nil})
	if err != nil {
		return err
	}
	return nil
}

// 删除资金账户
// path: /openapi/balanceAccounts/{id}
// method: DELETE
func (pp *PayndaPaySDK) DeleteBalanceAccount(ctx context.Context, nonce string) error {
	balanceAccountID := pp.conf.GetBalanceAccount()
	apipath := fmt.Sprintf("/openapi/balanceAccounts/%s", balanceAccountID)
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodDelete, apipath, nonce, "", nil, nil})
	if err != nil {
		return err
	}
	return nil
}

// 获取资金账户详情
// path: /openapi/balanceAccounts/{id}
// method: GET
func (pp *PayndaPaySDK) GetBalanceAccount(ctx context.Context) (*BalanceAccount, error) {
	balanceAccountID := pp.conf.GetBalanceAccount()
	apipath := fmt.Sprintf("/openapi/balanceAccounts/%s", balanceAccountID)
	var result BalanceAccount
	nonce := genNonce()
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodGet, apipath, nonce, "", nil, &result})

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// 查询资金账户
// path: /openapi/balanceAccounts
// method: GET
func (pp *PayndaPaySDK) GetBalanceAccounts(ctx context.Context) ([]*BalanceAccount, error) {
	apipath := "/openapi/balanceAccounts"
	var result []*BalanceAccount
	nonce := genNonce()
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodGet, apipath, nonce, "", defaultPageQuery, &result})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// 获取资金账户钱包
// path: /openapi/balanceAccounts/{balanceAccountId}/wallets
// method: GET
func (pp *PayndaPaySDK) GetBalanceAccountWallets(ctx context.Context) ([]*BalanceAccountWallet, error) {
	balanceAccountID := pp.conf.GetBalanceAccount()
	apipath := fmt.Sprintf("/openapi/balanceAccounts/%s/wallets", balanceAccountID)
	var result []*BalanceAccountWallet
	nonce := genNonce()
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodGet, apipath, nonce, "", defaultPageQuery, &result})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 资金账户钱包转账
// path:  /openapi/balanceAccountWalletTransfers
// method: POST
func (pp *PayndaPaySDK) BalanceAccountWalletTransfer(ctx context.Context, param *BalanceAccountWalletTransferRequest) (*BalanceAccountWalletTransfer, error) {

	if err := param.Validate(); err != nil {
		return nil, err
	}
	apipath := pp.buildAPIPath("/openapi/balanceAccountWalletTransfers")

	param.BalanceAccountID = pp.conf.GetBalanceAccount()
	var result *BalanceAccountWalletTransfer
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodPost, apipath, param.Nonce, param.RequestID, param, &result})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 获取卡BIN信息
// path:  /openapi/balanceAccounts/{balanceAccountId}/cardBins
// method: GET
func (pp *PayndaPaySDK) GetCardBin(ctx context.Context, creditLimitType CreditLimitType) ([]*CardBin, error) {
	balanceAccountID := pp.conf.GetBalanceAccount()
	req := map[string]any{}
	if creditLimitType != "" {
		req["creditLimitType"] = creditLimitType
	}

	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/cardBins", balanceAccountID)

	var result []*CardBin
	nonce := genNonce()
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodGet, apipath, nonce, "", req, &result})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// 获取持卡人详情
// path： /openapi/balanceAccounts/{balanceAccountId}/cardholders/{id}
// method： GET
func (pp *PayndaPaySDK) GetCardholder(ctx context.Context, cardholderID string) (*Cardholder, error) {
	if cardholderID == "" {
		return nil, ErrEmptyCardholderID
	}
	balanceAccountID := pp.conf.GetBalanceAccount()
	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/cardholders/%s", balanceAccountID, cardholderID)
	nonce := genNonce()
	var result *Cardholder
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodGet, apipath, nonce, "", nil, &result})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 新增持卡人
// path： /openapi/balanceAccounts/{balanceAccountId}/cardholders
// method： POST
func (pp *PayndaPaySDK) CreateCardHolder(ctx context.Context, param *CardHolderRequest) (*Cardholder, error) {
	if param == nil {
		return nil, NewParamValidateError(ErrEmptyParams.Error())
	}

	// 验证请求参数
	if err := param.Validate(); err != nil {
		return nil, NewParamValidateError(err.Error())
	}
	balanceAccountID := pp.conf.GetBalanceAccount()

	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/cardholders", balanceAccountID)

	var result *Cardholder
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodPost, apipath, param.Nonce, "", param, &result})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 修改持卡人信息
// path： /openapi/balanceAccounts/{balanceAccountId}/cardholders/{id}
// method： PUT
func (pp *PayndaPaySDK) UpdateCardholder(ctx context.Context, param *CardHolderRequest) error {
	if param == nil {
		return ErrEmptyParams
	}

	if param.CardholderID == "" {
		return ErrEmptyCardholderID

	}
	// 验证请求参数
	if err := param.Validate(); err != nil {
		return err
	}
	balanceAccountID := pp.conf.GetBalanceAccount()

	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/cardholders/%s", balanceAccountID, param.CardholderID)

	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodPut, apipath, param.Nonce, "", param, nil})
	if err != nil {
		return err
	}

	return nil
}

// 获取持卡人列表
// path： /openapi/balanceAccounts/{balanceAccountId}/cardholders
// method： GET
func (pp *PayndaPaySDK) GetCardholders(ctx context.Context) ([]*Cardholder, error) {
	balanceAccountID := pp.conf.GetBalanceAccount()

	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/cardholders", balanceAccountID)
	nonce := genNonce()
	var result []*Cardholder
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodGet, apipath, nonce, "", defaultPageQuery, &result})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 删除持卡人
// path： /openapi/balanceAccounts/{balanceAccountId}/cardholders/{id}
// method： DELETE
func (pp *PayndaPaySDK) DeleteCardholder(ctx context.Context, cardholderID string) error {
	if cardholderID == "" {
		return ErrEmptyCardholderID
	}
	balanceAccountID := pp.conf.GetBalanceAccount()
	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/cardholders/%s", balanceAccountID, cardholderID)
	nonce := genNonce()
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodDelete, apipath, nonce, "", nil, nil})
	return err
}

// 查询持卡人钱包金额
// path： /openapi/balanceAccounts/{balanceAccountId}/cardholders/{cardholderID}/wallets
// method： GET
func (pp *PayndaPaySDK) GetCardholderWallet(ctx context.Context, cardholderID string) ([]*CardholderWallet, error) {
	if cardholderID == "" {
		return nil, ErrEmptyCardholderID
	}
	balanceAccountID := pp.conf.GetBalanceAccount()
	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/cardholders/%s/wallets", balanceAccountID, cardholderID)
	nonce := genNonce()
	var result []*CardholderWallet
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodGet, apipath, nonce, "", defaultPageQuery, &result})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 修改持卡人钱包金额
// path： /openapi/balanceAccounts/{balanceAccountId}/cardholderWalletUpdates
// method： POST
func (pp *PayndaPaySDK) UpdateCardholderWallet(ctx context.Context, param *CardHolderWalletRequest) (*CardholderWallet, error) {
	if param == nil {
		return nil, ErrEmptyParams
	}

	// 验证请求参数
	if err := param.Validate(); err != nil {
		return nil, err
	}
	balanceAccountID := pp.conf.GetBalanceAccount()
	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/cardholderWalletUpdates", balanceAccountID)
	nonce := genNonce()
	var result *CardholderWallet
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodPost, apipath, nonce, "", param, &result})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 新增卡
// path： /openapi/balanceAccounts/{balanceAccountId}/cards
// method： POST
func (pp *PayndaPaySDK) CreateCard(ctx context.Context, param *CardCreateRequest) (*CardDetail, error) {
	if param == nil {
		return nil, ErrEmptyParams
	}
	balanceAccountID := pp.conf.GetBalanceAccount()
	// 验证请求参数
	if err := param.Validate(); err != nil {
		return nil, err
	}

	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/cards", balanceAccountID)

	var result *CardDetail

	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodPost, apipath, param.Nonce, param.RequestID, param, &result})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 获取卡信息
// path： /openapi/balanceAccounts/{balanceAccountId}/cards/{id}
// method： GET
func (pp *PayndaPaySDK) GetCard(ctx context.Context, cardID string) (*Card, error) {
	if cardID == "" {
		return nil, ErrEmptyCardID
	}
	balanceAccountID := pp.conf.GetBalanceAccount()

	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/cards/%s", balanceAccountID, cardID)
	nonce := genNonce()
	var result *Card
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodGet, apipath, nonce, "", nil, &result})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 获取卡余额
// path： /openapi/balanceAccounts/{balanceAccountId}/cards/{id}/balance
// method： GET
func (pp *PayndaPaySDK) GetCardBalance(ctx context.Context, cardID string) (*CardBalance, error) {
	if cardID == "" {
		return nil, ErrEmptyCardID
	}
	balanceAccountID := pp.conf.GetBalanceAccount()

	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/cards/%s/balance", balanceAccountID, cardID)
	nonce := genNonce()
	var result *CardBalance
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodGet, apipath, nonce, "", nil, &result})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 卡列表
// path： /openapi/balanceAccounts/{balanceAccountId}/cards
// method： GET

func (pp *PayndaPaySDK) GetCards(ctx context.Context) ([]*Card, error) {
	balanceAccountID := pp.conf.GetBalanceAccount()
	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/cards", balanceAccountID)
	nonce := genNonce()
	var result []*Card
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodGet, apipath, nonce, "", defaultPageQueryALL, &result}) // TODO: 分批次查询
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 卡列表
// path： /openapi/balanceAccounts/{balanceAccountId}/cards
// method： GET

func (pp *PayndaPaySDK) ListCardByPaginate(ctx context.Context, param *CardPaginateRequest) (*CardData, error) {
	balanceAccountID := pp.conf.GetBalanceAccount()
	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/cards", balanceAccountID)
	nonce := genNonce()
	var result *CardData
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodGet, apipath, nonce, "", param, &result}) // TODO: 分批次查询
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 获取卡sensitive信息
// path： /openapi/balanceAccounts/{balanceAccountId}/cards/{id}/sensitiveInfo
// method： GET
func (pp *PayndaPaySDK) GetCardSensitive(ctx context.Context, cardID string) (*CardSensitive, error) {
	if cardID == "" {
		return nil, ErrEmptyCardID
	}
	balanceAccountID := pp.conf.GetBalanceAccount()

	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/cards/%s/sensitiveInfo", balanceAccountID, cardID)
	nonce := genNonce()
	var result *CardSensitive
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodGet, apipath, nonce, "", nil, &result})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 银行卡冻结
// path： /openapi/balanceAccounts/{balanceAccountId}/cards/{id}/status/frozen
// method： PATCH
func (pp *PayndaPaySDK) FrozenCard(ctx context.Context, cardID string) error {
	if cardID == "" {
		return ErrEmptyCardID
	}
	balanceAccountID := pp.conf.GetBalanceAccount()

	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/cards/%s/status/frozen", balanceAccountID, cardID)
	nonce := genNonce()
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodPatch, apipath, nonce, "", nil, nil})
	return err
}

// 银行卡解冻
// path： /openapi/balanceAccounts/{balanceAccountId}/cards/{id}/status/unfrozen
// method： PATCH
func (pp *PayndaPaySDK) UnfrozenCard(ctx context.Context, cardID string) error {
	if cardID == "" {
		return ErrEmptyCardID
	}
	balanceAccountID := pp.conf.GetBalanceAccount()

	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/cards/%s/status/unfrozen", balanceAccountID, cardID)
	nonce := genNonce()
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodPatch, apipath, nonce, "", nil, nil})
	return err
}

// release 银行卡
// path： /openapi/balanceAccounts/{balanceAccountId}/cards/{id}/status/release
// method： PATCH
func (pp *PayndaPaySDK) ReleaseCard(ctx context.Context, nonce string, cardID string) error {

	if cardID == "" {
		return ErrEmptyCardID
	}
	balanceAccountID := pp.conf.GetBalanceAccount()
	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/cards/%s/status/release", balanceAccountID, cardID)
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodPatch, apipath, nonce, "", nil, nil})
	return err
}

// 获取card controls
// path： /openapi/balanceAccounts/{balanceAccountId}/cards/{id}/controls
// method： GET
func (pp *PayndaPaySDK) GetCardControls(ctx context.Context, cardID string) ([]*CardControl, error) {
	if cardID == "" {
		return nil, ErrEmptyCardID
	}
	balanceAccountID := pp.conf.GetBalanceAccount()

	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/cards/%s/controls", balanceAccountID, cardID)
	nonce := genNonce()
	var result []*CardControl
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodGet, apipath, nonce, "", nil, &result})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 修改card controls
// path： /openapi/balanceAccounts/{balanceAccountId}/cards/{id}/controls
// method： PATCH
func (pp *PayndaPaySDK) UpdateCardControls(ctx context.Context, req *CardControlRequest) error {
	if req == nil {
		return ErrEmptyParams
	}

	// 验证请求参数
	if err := req.Validate(); err != nil {
		return err
	}
	balanceAccountID := pp.conf.GetBalanceAccount()

	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/cards/%s/controls", balanceAccountID, req.CardID)

	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodPatch, apipath, req.Nonce, "", req, nil})
	if err != nil {
		return err
	}

	return nil
}

// 查询card balance update
func (pp *PayndaPaySDK) QueryCardBalanceUpdate(ctx context.Context) ([]*CardBalanceUpdateHistory, error) {
	balanceAccountID := pp.conf.GetBalanceAccount()
	apipath := fmt.Sprintf("/openapi/balanceAccounts/%s/cardBalanceUpdates", balanceAccountID)
	nonce := genNonce()
	var result []*CardBalanceUpdateHistory
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodGet, apipath, nonce, "", defaultPageQueryALL, &result})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 更新card balance
// path： /openapi/balanceAccounts/{balanceAccountId}/cardBalanceUpdates
// method： POST
func (pp *PayndaPaySDK) UpdateCardBalance(ctx context.Context, req *CardBalanceUpdateRequest) (*CardBalance, error) {
	if req == nil {
		return nil, ErrEmptyParams
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	balanceAccountID := pp.conf.GetBalanceAccount()

	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/cardBalanceUpdates", balanceAccountID)

	var result *CardBalance
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodPost, apipath, req.Nonce, req.RequestID, req, &result})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 银行卡充值/提现
// path： /openapi/balanceAccounts/{balanceAccountId}/cardBalanceTransfers
// method： POST
func (pp *PayndaPaySDK) CardTransfer(ctx context.Context, req *CardBalanceTransferRequest) (*CardBalanceTransfer, error) {
	if req == nil {
		return nil, ErrEmptyParams
	}

	// 验证请求参数
	if err := req.Validate(); err != nil {
		return nil, err
	}
	balanceAccountID := pp.conf.GetBalanceAccount()

	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/cardBalanceTransfers", balanceAccountID)

	var result *CardBalanceTransfer
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodPost, apipath, req.Nonce, req.RequestID, req, &result})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 获取卡交易列表
// path： /openapi/balanceAccounts/{balanceAccountId}/transactions
// method： GET
func (pp *PayndaPaySDK) GetCardTransactions(ctx context.Context, req *CardTransactionsRequest) (*CardTransactionsData, error) {
	if req == nil {
		return nil, ErrEmptyParams
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	req.BalanceAccountID = pp.conf.GetBalanceAccount()

	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/transactions", req.BalanceAccountID)

	// 创建自定义请求
	var result *CardTransactionsData

	nonce := genNonce()
	// 使用RestyRequest发送请求并处理响应
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodGet, apipath, nonce, "", req, &result})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 获取卡交易详情
// path： /openapi/balanceAccounts/{balanceAccountId}/transactions/{id}
// method： GET
func (pp *PayndaPaySDK) QueryCardTransaction(ctx context.Context, thirdId string) (*CardTransaction, error) {
	if thirdId == "" {
		return nil, ErrEmptyParams
	}
	balanceAccountID := pp.conf.GetBalanceAccount()
	apipath := pp.buildAPIPath("/openapi/balanceAccounts/%s/transactions/%s", balanceAccountID, thirdId)

	// 创建自定义请求
	var result *CardTransaction
	nonce := genNonce()
	// 使用RestyRequest发送请求并处理响应
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodGet, apipath, nonce, "", nil, &result})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (pp *PayndaPaySDK) QueryRequestResult(ctx context.Context, requestID string) (*RequestResult, error) {
	if requestID == "" {
		return nil, ErrEmptyRequestID
	}

	apipath := pp.buildAPIPath("/openapi/requestResults")

	// 创建自定义请求
	var result *RequestResult

	nonce := genNonce()
	// 使用RestyRequest发送请求并处理响应
	_, err := pp.RestyRequest(ctx, &RestyRequestOptions{http.MethodGet, apipath, nonce, "",
		map[string]any{
			"requestId": requestID,
		}, &result})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (pp *PayndaPaySDK) QueryTransfer(ctx context.Context, requestID string) (*CardBalanceTransferReuslt, error) {

	if requestID == "" {
		return nil, ErrEmptyRequestID
	}

	r, err := pp.QueryRequestResult(ctx, requestID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return &CardBalanceTransferReuslt{
				Message:    "transfer record not found",
				Success:    false,
				CreateTime: time.Now().Format(time.DateTime),
			}, nil
		}

		return nil, err
	}

	var rsp *APIResponse
	err = json.Unmarshal([]byte(r.Result), &rsp)
	if err != nil {
		return nil, err
	}

	result := &CardBalanceTransferReuslt{
		Code:       rsp.Code,
		Message:    rsp.Message,
		Success:    rsp.Success,
		CreateTime: r.CreateTime,
	}

	if rsp.Success {
		record := CardBalanceTransferRecord{
			ID:         r.ID,
			RequestID:  r.RequestID,
			CreateTime: r.CreateTime,
			UpdateTime: r.UpdateTime,
		}

		err = pp.processResponseV0(rsp, &record)
		if err != nil {
			return nil, err
		}
		result.TransferRecord = &record
	}

	return result, nil
}

// 生成过期日期
// 过期日期 格式: MM/YY，从当前月份往后 6 个月到 3 年之间随机生成一个月份
func genExpirationDate() string {
	// Get current time
	now := time.Now()

	randMonths := 6 + rand.Intn(31)
	expirationTime := now.AddDate(0, randMonths, 0)

	return expirationTime.Format("01/06")
}

func genNonce() string {
	return uuid.NewV4().String()
}
