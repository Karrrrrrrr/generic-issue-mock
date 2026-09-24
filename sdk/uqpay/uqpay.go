package uqpay

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"resty.dev/v3"
)

const defaultUpCardLimitAmount = 250000.0

// Conf 配置接口
type Conf interface {
	GetUrl() string
	GetDebug() bool
	GetApiKey() string
	GetClientId() string
	GetCardLimitAmount() float64
}

// UqPaySDK SDK结构体
type UqPaySDK struct {
	resty               *resty.Client
	conf                Conf
	requestMiddleware   []resty.RequestMiddleware
	responseMiddlewares []resty.ResponseMiddleware
}

// Option SDK选项函数类型
type Option func(t *UqPaySDK)

// WithRequestMiddlewares 设置请求中间件
func WithRequestMiddlewares(middlewares ...resty.RequestMiddleware) Option {
	return func(pp *UqPaySDK) {
		pp.requestMiddleware = append(pp.requestMiddleware, resty.PrepareRequestMiddleware)
		pp.requestMiddleware = append(pp.requestMiddleware, middlewares...)
	}
}

// WithResponseMiddlewares 设置响应中间件
func WithResponseMiddlewares(middlewares ...resty.ResponseMiddleware) Option {
	return func(pp *UqPaySDK) {
		pp.responseMiddlewares = append(pp.responseMiddlewares, resty.AutoParseResponseMiddleware)
		pp.responseMiddlewares = append(pp.responseMiddlewares, middlewares...)
	}
}

// New 初始化PhotonPaySDK
func New(conf Conf, options ...Option) *UqPaySDK {
	pp := &UqPaySDK{
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

// GetAccessToken 获取access_token
// 文档地址: https://docs.uqpay.com/reference/access-token
func (up *UqPaySDK) GetAccessToken(ctx context.Context) (*GetAccessTokenResp, error) {
	resp, err := up.resty.R().
		SetContext(ctx).
		SetDebug(up.conf.GetDebug()).
		SetHeader("x-api-key", up.conf.GetApiKey()).
		SetHeader("x-client-id", up.conf.GetClientId()).
		Post(up.conf.GetUrl() + "/api/v1/connect/token")
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, errors.New(resp.String())
	}

	result := &GetAccessTokenResp{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}
	if result.Code >= 400 {
		return nil, errors.New(result.Message)
	}

	return result, nil
}

// CreateCardholder 创建持卡人
// 文档地址: https://docs.uqpay.com/reference/create-cardholder
func (up *UqPaySDK) CreateCardholder(ctx context.Context, token, idempotencyKey string, req *CreateCardholderReq) (*CreateCardholderResp, error) {
	resp, err := up.resty.R().
		SetContext(ctx).
		SetDebug(up.conf.GetDebug()).
		SetHeader("x-auth-token", token).
		SetHeader("x-idempotency-key", idempotencyKey).
		SetBody(req).
		Post(up.conf.GetUrl() + "/api/v1/issuing/cardholders")
	if err != nil {
		return nil, err
	}

	result := &CreateCardholderResp{}
	if !resp.IsSuccess() {
		if err := json.Unmarshal([]byte(resp.String()), &result); err == nil {
			result.checkFailCode()
			return result, nil
		}
		return nil, errors.New(resp.String())
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}
	result.checkFailCode()
	return result, nil
}

// UpdateCardholder 更新持卡人
// 文档地址: https://docs.uqpay.com/reference/update-cardholder
func (up *UqPaySDK) UpdateCardholder(ctx context.Context, token, idempotencyKey string, req *UpdateCardholderReq) (*UpdateCardholderResp, error) {
	resp, err := up.resty.R().
		SetContext(ctx).
		SetDebug(up.conf.GetDebug()).
		SetHeader("x-auth-token", token).
		SetHeader("x-idempotency-key", idempotencyKey).
		SetBody(req).
		Post(up.conf.GetUrl() + "/api/v1/issuing/cardholders/" + req.CardHolderID)
	if err != nil {
		return nil, err
	}

	result := &UpdateCardholderResp{}
	if !resp.IsSuccess() {
		if err := json.Unmarshal([]byte(resp.String()), &result); err == nil {
			result.checkFailCode()
			return result, nil
		}
		return nil, errors.New(resp.String())
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}
	result.checkFailCode()
	return result, nil
}

// GetCardholder 查询持卡人
// 文档地址: https://docs.uqpay.com/reference/update-cardholder
func (up *UqPaySDK) GetCardholder(ctx context.Context, token, cardHolderID string) (*CardholderDetail, error) {
	resp, err := up.resty.R().
		SetContext(ctx).
		SetDebug(up.conf.GetDebug()).
		SetHeader("x-auth-token", token).
		Get(up.conf.GetUrl() + "/api/v1/issuing/cardholders/" + cardHolderID)
	if err != nil {
		return nil, err
	}

	result := &CardholderDetail{}
	if !resp.IsSuccess() {
		if err := json.Unmarshal([]byte(resp.String()), &result); err == nil {
			return result, nil
		}
		return nil, errors.New(resp.String())
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}
	return result, nil
}

// ListCardProducts 获取product列表
// 文档地址: https://docs.uqpay.com/reference/list-card-products
func (up *UqPaySDK) ListCardProducts(ctx context.Context, token string, pageSize, pageNumber string) (*ListCardProductsResp, error) {
	resp, err := up.resty.R().
		SetContext(ctx).
		SetDebug(up.conf.GetDebug()).
		SetHeader("x-auth-token", token).
		SetHeader("x-idempotency-key", uuid.New().String()).
		SetQueryParam("page_size", pageSize).
		SetQueryParam("page_number", pageNumber).
		Get(up.conf.GetUrl() + "/api/v1/issuing/products")
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, errors.New(resp.String())
	}

	result := &ListCardProductsResp{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}
	if result.Code >= 400 {
		return nil, errors.New(result.Message)
	}
	return result, nil
}

// CreateCard 创建卡片
// 文档地址: https://docs.uqpay.com/reference/create-card
func (up *UqPaySDK) CreateCard(ctx context.Context, token, idempotencyKey string, req *CreateCardReq) (*CreateCardResp, error) {
	resp, err := up.resty.R().
		SetContext(ctx).
		SetDebug(up.conf.GetDebug()).
		SetHeader("x-auth-token", token).
		SetHeader("x-idempotency-key", idempotencyKey).
		SetBody(req).
		Post(up.conf.GetUrl() + "/api/v1/issuing/cards")
	if err != nil {
		return nil, err
	}

	result := &CreateCardResp{}
	if !resp.IsSuccess() {
		if err := json.Unmarshal([]byte(resp.String()), &result); err == nil {
			result.checkFailCode()
			return result, nil
		}
		return nil, errors.New(resp.String())
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}
	result.checkFailCode()
	return result, nil
}

// UpdateCard 更新卡片
// 文档地址:https://docs.uqpay.com/reference/update-card
func (up *UqPaySDK) UpdateCard(ctx context.Context, token, idempotencyKey string, req *UpdateCardReq) (*UpdateCardResp, error) {
	resp, err := up.resty.R().
		SetContext(ctx).
		SetDebug(up.conf.GetDebug()).
		SetHeader("x-auth-token", token).
		SetHeader("x-idempotency-key", idempotencyKey).
		SetBody(req).
		Post(up.conf.GetUrl() + "/api/v1/issuing/cards/" + req.CardID)
	if err != nil {
		return nil, err
	}
	result := &UpdateCardResp{}
	if !resp.IsSuccess() {
		if err := json.Unmarshal([]byte(resp.String()), &result); err == nil {
			result.checkFailCode()
			return result, nil
		}
		return nil, errors.New(resp.String())
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}
	result.checkFailCode()
	return result, nil
}

// GetCardInfo 获取card信息
// 文档地址: https://docs.uqpay.com/reference/retrieve-card
func (up *UqPaySDK) GetCardInfo(ctx context.Context, token string, cardID string) (*GetCardInfo, error) {
	resp, err := up.resty.R().
		SetContext(ctx).
		SetDebug(up.conf.GetDebug()).
		SetHeader("x-auth-token", token).
		Get(up.conf.GetUrl() + "/api/v1/issuing/cards/" + cardID)
	if err != nil {
		return nil, err
	}

	result := &GetCardInfo{}
	if !resp.IsSuccess() {
		if err := json.Unmarshal([]byte(resp.String()), &result); err == nil {
			result.checkFailCode()
			return result, nil
		}
		return nil, errors.New(resp.String())
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}
	result.checkFailCode()
	return result, nil
}

// GetCardPrivateInfo 获取card的私密信息
// 文档地址: https://docs.uqpay.com/reference/retrieve-card-secure
func (up *UqPaySDK) GetCardPrivateInfo(ctx context.Context, token string, cardID string) (*CardPrivateInfo, error) {
	resp, err := up.resty.R().
		SetContext(ctx).
		SetDebug(up.conf.GetDebug()).
		SetHeader("x-auth-token", token).
		Get(up.conf.GetUrl() + "/api/v1/issuing/cards/" + cardID + "/secure")
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, errors.New(resp.String())
	}

	result := &CardPrivateInfo{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}
	if len(result.Code) > 0 {
		return nil, errors.New(result.Message)
	}
	return result, nil
}

// UpdateCardStatus 更新卡片状态
// 文档地址: https://docs.uqpay.com/reference/update-card-status
func (up *UqPaySDK) UpdateCardStatus(ctx context.Context, token, idempotencyKey string, req *UpdateCardStatusReq) (*UpdateCardStatusResp, error) {
	resp, err := up.resty.R().
		SetContext(ctx).
		SetDebug(up.conf.GetDebug()).
		SetHeader("x-auth-token", token).
		SetHeader("x-idempotency-key", idempotencyKey).
		SetBody(req).
		Post(up.conf.GetUrl() + "/api/v1/issuing/cards/" + req.CardID + "/status")
	if err != nil {
		return nil, err
	}
	result := &UpdateCardStatusResp{}
	if !resp.IsSuccess() {
		if err := json.Unmarshal([]byte(resp.String()), &result); err == nil {
			result.checkFailCode()
			return result, nil
		}
		return nil, errors.New(resp.String())
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}
	result.checkFailCode()
	return result, nil
}

// AssignCard 开实体卡
// 文档地址: https://docs.uqpay.com/reference/assign-card
func (up *UqPaySDK) AssignCard(ctx context.Context, token, idempotencyKey string, req *AssignCardReq) (*AssignCardResp, error) {
	resp, err := up.resty.R().
		SetContext(ctx).
		SetDebug(up.conf.GetDebug()).
		SetHeader("x-auth-token", token).
		SetHeader("x-idempotency-key", idempotencyKey).
		SetBody(req).
		Post(up.conf.GetUrl() + "/api/v1/issuing/cards/assign")
	if err != nil {
		return nil, err
	}
	result := &AssignCardResp{}
	if !resp.IsSuccess() {
		if err := json.Unmarshal([]byte(resp.String()), &result); err == nil {
			result.checkFailCode()
			return result, nil
		}
		return nil, errors.New(resp.String())
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}
	result.checkFailCode()
	return result, nil
}

// ActivateCard 激活实体卡
// 文档地址: https://docs.uqpay.com/reference/activate-card
func (up *UqPaySDK) ActivateCard(ctx context.Context, token, idempotencyKey string, req *ActivateCardReq) (*ActivateCardResp, error) {
	resp, err := up.resty.R().
		SetContext(ctx).
		SetDebug(up.conf.GetDebug()).
		SetHeader("x-auth-token", token).
		SetHeader("x-idempotency-key", idempotencyKey).
		SetBody(req).
		Post(up.conf.GetUrl() + "/api/v1/issuing/cards/activate")
	if err != nil {
		return nil, err
	}
	result := &ActivateCardResp{}
	if !resp.IsSuccess() {
		if err := json.Unmarshal([]byte(resp.String()), &result); err == nil {
			result.checkFailCode()
			return result, nil
		}
		return nil, errors.New(resp.String())
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}
	result.checkFailCode()
	return result, nil
}

// ResetCardPin 激活实体卡
// 文档地址: https://docs.uqpay.com/reference/reset-pin
func (up *UqPaySDK) ResetCardPin(ctx context.Context, token, idempotencyKey string, req *ResetCardPinReq) (*ResetCardPinResp, error) {
	resp, err := up.resty.R().
		SetContext(ctx).
		SetDebug(up.conf.GetDebug()).
		SetHeader("x-auth-token", token).
		SetHeader("x-idempotency-key", idempotencyKey).
		SetBody(req).
		Post(up.conf.GetUrl() + "/api/v1/issuing/cards/pin")
	if err != nil {
		return nil, err
	}
	result := &ResetCardPinResp{}
	if !resp.IsSuccess() {
		if err := json.Unmarshal([]byte(resp.String()), &result); err == nil {
			result.checkFailCode()
			return result, nil
		}
		return nil, errors.New(resp.String())
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}
	result.checkFailCode()
	return result, nil
}

// ListTransaction 获取交易列表
// 文档地址: https://docs.uqpay.com/reference/list-cards-transactions
func (up *UqPaySDK) ListTransaction(ctx context.Context, token string, req *ListTransactionReq) (*ListTransactionResp, error) {
	queryParams := req.toQueryParams()
	resp, err := up.resty.R().
		SetContext(ctx).
		SetDebug(up.conf.GetDebug()).
		SetHeader("x-auth-token", token).
		SetQueryParams(queryParams).
		Get(up.conf.GetUrl() + "/api/v1/issuing/transactions")
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, errors.New(resp.String())
	}

	result := &ListTransactionResp{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}
	if len(result.Code) > 0 {
		return nil, errors.New(result.Message)
	}
	return result, nil
}

// RetrieveBalance 取回余额
// 文档地址: https://docs.uqpay.com/reference/retrievebalance
func (up *UqPaySDK) RetrieveBalance(ctx context.Context, token, idempotencyKey string, req *RetrieveBalanceReq) (*RetrieveBalanceResp, error) {
	resp, err := up.resty.R().
		SetContext(ctx).
		SetDebug(up.conf.GetDebug()).
		SetHeader("x-auth-token", token).
		SetHeader("x-idempotency-key", idempotencyKey).
		SetBody(req).
		Get(up.conf.GetUrl() + "/api/v1/balances/" + req.Currency)
	if err != nil {
		return nil, err
	}

	result := &RetrieveBalanceResp{}
	if !resp.IsSuccess() {
		if err := json.Unmarshal([]byte(resp.String()), &result); err == nil {
			return result, nil
		}
		return nil, errors.New(resp.String())
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}
	return result, nil
}

// RetrieveIssuingBalance 取回余额
// 文档地址: https://docs.uqpay.com/reference/retrieve-issuing-balance
func (up *UqPaySDK) RetrieveIssuingBalance(ctx context.Context, token, idempotencyKey string, req *RetrieveIssuingBalanceReq) (*RetrieveIssuingBalanceResp, error) {
	resp, err := up.resty.R().
		SetContext(ctx).
		SetDebug(up.conf.GetDebug()).
		SetHeader("x-auth-token", token).
		SetHeader("x-idempotency-key", idempotencyKey).
		SetBody(req).
		Post(up.conf.GetUrl() + "/api/v1/issuing/balances")
	if err != nil {
		return nil, err
	}

	result := &RetrieveIssuingBalanceResp{}
	if !resp.IsSuccess() {
		if err := json.Unmarshal([]byte(resp.String()), &result); err == nil {
			return result, nil
		}
		return nil, errors.New(resp.String())
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}
	return result, nil
}

// SimulateAuthorization 授权
// 文档地址: https://docs.uqpay.com/reference/simulate-authorization
func (up *UqPaySDK) SimulateAuthorization(ctx context.Context, token, idempotencyKey string, req *SimulateAuthorizationReq) (*SimulateAuthorizationResp, error) {

	resp, err := up.resty.R().
		SetContext(ctx).
		SetDebug(up.conf.GetDebug()).
		SetHeader("x-auth-token", token).
		SetHeader("x-idempotency-key", idempotencyKey).
		SetBody(req).
		Post(up.conf.GetUrl() + "/api/v1/simulation/issuing/authorization")
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, errors.New(resp.String())
	}

	var result SimulateAuthorizationResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Code) > 0 {
		return nil, errors.New(result.Message)
	}
	result.checkFailCode()
	return &result, nil
}

// SimulateReversal 撤销
// 文档地址: https://docs.uqpay.com/reference/simulate-reversal
func (up *UqPaySDK) SimulateReversal(ctx context.Context, token, idempotencyKey string, req *SimulateReversalReq) (*SimulateReversalResp, error) {
	resp, err := up.resty.R().
		SetContext(ctx).
		SetDebug(up.conf.GetDebug()).
		SetHeader("x-auth-token", token).
		SetHeader("x-idempotency-key", idempotencyKey).
		SetBody(req).
		Post(up.conf.GetUrl() + "/api/v1/simulation/issuing/reversal")
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, errors.New(resp.String())
	}

	var result SimulateReversalResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Code) > 0 {
		return nil, errors.New(result.Message)
	}
	result.checkFailCode()
	return &result, nil
}

// CreatePanToken 创建PAN token
// 文档地址: https://docs.uqpay.com/reference/create-pan-token
func (up *UqPaySDK) CreatePanToken(ctx context.Context, token, cardID string) (*PanToken, error) {
	resp, err := up.resty.R().
		SetContext(ctx).
		SetDebug(up.conf.GetDebug()).
		SetHeader("x-auth-token", token).
		SetHeader("x-idempotency-key", uuid.New().String()).
		Post(up.conf.GetUrl() + "/api/v1/issuing/cards/" + cardID + "/token")
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, errors.New(resp.String())
	}

	result := &PanToken{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}
	if len(result.Code) > 0 {
		return nil, errors.New(result.Message)
	}
	return result, nil
}

// Retrieve Card Order 取回卡片订单
// 文档地址: https://docs.uqpay.com/reference/retrieve-card-order
// api: https://api-sandbox.uqpaytech.com/api/v1/issuing/cards/{id}/order
func (up *UqPaySDK) RetrieveCardOrder(ctx context.Context, token, orderID string) (*CardOrder, error) {
	resp, err := up.resty.R().
		SetContext(ctx).
		SetDebug(up.conf.GetDebug()).
		SetHeader("x-auth-token", token).
		Get(up.conf.GetUrl() + "/api/v1/issuing/cards/" + orderID + "/order")
	if err != nil {
		return nil, err
	}

	result := CardOrder{}

	if !resp.IsSuccess() {
		if err := json.Unmarshal([]byte(resp.String()), &result); err == nil {
			return nil, errors.New(result.Code + ":" + result.Message)
		}
		return nil, errors.New(resp.String())
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (up *UqPaySDK) GetDefaultCardLimitAmount(ctx context.Context) float64 {
	cardLimit := defaultUpCardLimitAmount
	if up.conf.GetCardLimitAmount() > 0 {
		cardLimit = up.conf.GetCardLimitAmount()
	}
	return cardLimit
}
