package pingpong

import (
	"context"
	"encoding/json"
	"errors"

	kratosErr "github.com/go-kratos/kratos/v2/errors"
)

// PingPongSDKInterface 是服务 biz 层调用 PingPong 的接口；业务请求的 token 由调用方统一管理。
type PingPongSDKInterface interface {
	// CreateCardPre 创建卡片参数处理
	CreateCardPre(params *CreateCardRequest) (json.RawMessage, error)
	// GetAccessToken 获取访问令牌；调用方应缓存令牌，避免重复获取使现有令牌失效。
	GetAccessToken(ctx context.Context) (*AccessTokenResponse, error)
	// QueryAccountsBalances 查询账户余额。
	QueryAccountsBalances(ctx context.Context, token string, params *QueryAccountsBalancesRequest) (*QueryAccountsBalancesResponse, error)
	// QueryCardProducts 查询可用的卡产品。
	QueryCardProducts(ctx context.Context, token string) (*QueryCardProductsResponse, error)
	// CreateCard 申请创建卡片。
	CreateCard(ctx context.Context, token string, params *CreateCardRequest) (*CreateCardResponse, error)
	// GetCardDetails 根据卡片 ID 查询卡片详情。
	GetCardDetails(ctx context.Context, token, cardID string) (*CardDetailsResponse, error)
	// CardFunding 执行卡片充值或转出。
	CardFunding(ctx context.Context, token string, params *CardFundingRequest) (*CardFundingResponse, error)
	// QueryCardFundingOrders 查询卡片资金订单。
	QueryCardFundingOrders(ctx context.Context, token string, params *QueryCardFundingOrdersRequest) (*QueryCardFundingOrdersResponse, error)
	// QueryDedicatedCardBalance 根据卡片 ID 查询卡片余额。
	QueryDedicatedCardBalance(ctx context.Context, token, cardID string) (*DedicatedCardBalanceResponse, error)
	// CardAction 执行卡片冻结、解冻或销卡操作。
	CardAction(ctx context.Context, token string, params *CardActionRequest) error
	// QueryCardTransactions 查询卡片交易记录。
	QueryCardTransactions(ctx context.Context, token string, params *QueryCardTransactionsRequest) (*QueryCardTransactionsResponse, error)
	// Query3DSDetails 根据卡片 ID 查询 3DS 详情。
	Query3DSDetails(ctx context.Context, token, cardID string) (*ThreeDSDetailsResponse, error)
	// CreateBudgetAccount 创建预算账户。
	CreateBudgetAccount(ctx context.Context, token string, params *CreateBudgetAccountRequest) (*CreateBudgetAccountResponse, error)
	// BudgetFunding 对预算账户充值或划转资金。
	BudgetFunding(ctx context.Context, token string, params *BudgetFundingRequest) (*BudgetFundingResponse, error)
	// QueryBudgetFundingOrder 查询预算账户资金订单状态。
	QueryBudgetFundingOrder(ctx context.Context, token string, params *QueryBudgetFundingOrderRequest) (*QueryBudgetFundingOrderResponse, error)
	// QueryBudgetAccountBalance 查询指定或全部预算账户余额。
	QueryBudgetAccountBalance(ctx context.Context, token, budgetID string) (*QueryBudgetAccountBalanceResponse, error)
	// QueryAccountTransactions 查询预算账户或卡片的已入账交易。
	QueryAccountTransactions(ctx context.Context, token string, params *QueryAccountTransactionsRequest) (*QueryAccountTransactionsResponse, error)
}

type pingPongSDKInterface struct {
	sdk *PingPongSDK
}

func NewPingPongSDK(sdk *PingPongSDK) PingPongSDKInterface {
	return &pingPongSDKInterface{
		sdk: sdk,
	}
}

func (ge *pingPongSDKInterface) CreateCardPre(params *CreateCardRequest) (json.RawMessage, error) {
	cardPostParams, err := ge.cardPostParams(params)
	if err != nil {
		return nil, err
	}

	marshal, err := json.Marshal(cardPostParams)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(marshal), nil
}

func (ge *pingPongSDKInterface) cardPostParams(params *CreateCardRequest) (*CreateCardRequest, error) {
	return &CreateCardRequest{
		RequestID:           params.RequestID,           // 唯一请求 ID，用于追踪请求。
		CardProductCode:     params.CardProductCode,     // 卡产品代码，来自卡产品查询接口。
		CardCurrency:        params.CardCurrency,        // 发卡币种，需与预算币种一致。
		BudgetID:            params.BudgetID,            // 卡片关联的预算账户 ID。
		CardholderID:        params.CardholderID,        // 持卡人 ID；启用 3DS 时需要提供。
		PerTransactionLimit: params.PerTransactionLimit, // 单笔交易限额。
		DailyLimit:          params.DailyLimit,          // 每日交易限额。
		WeeklyLimit:         params.WeeklyLimit,         // 每周交易限额。
		MonthlyLimit:        params.MonthlyLimit,        // 每月交易限额。
		LifetimeLimit:       params.LifetimeLimit,       // 卡片生命周期内的交易限额。
		CouponApplied:       params.CouponApplied,       // 是否使用奖励余额抵扣开卡费用。
		Remark:              params.Remark,              // 卡片备注。
	}, nil
}

// GetAccessToken 获取访问令牌；调用方应缓存令牌，避免重复获取使现有令牌失效。
func (p *pingPongSDKInterface) GetAccessToken(ctx context.Context) (*AccessTokenResponse, error) {
	return p.sdk.GetAccessToken(ctx)
}

// QueryAccountsBalances 查询账户余额。
func (p *pingPongSDKInterface) QueryAccountsBalances(ctx context.Context, token string, params *QueryAccountsBalancesRequest) (*QueryAccountsBalancesResponse, error) {
	return p.sdk.QueryAccountsBalances(ctx, token, params)
}

// QueryCardProducts 查询可用的卡产品。
func (p *pingPongSDKInterface) QueryCardProducts(ctx context.Context, token string) (*QueryCardProductsResponse, error) {
	return p.sdk.QueryCardProducts(ctx, token)
}

// CreateCard 申请创建卡片。
func (p *pingPongSDKInterface) CreateCard(ctx context.Context, token string, params *CreateCardRequest) (*CreateCardResponse, error) {
	return p.sdk.CreateCard(ctx, token, params)
}

// GetCardDetails 根据卡片 ID 查询卡片详情。
func (p *pingPongSDKInterface) GetCardDetails(ctx context.Context, token, cardID string) (*CardDetailsResponse, error) {
	return p.sdk.GetCardDetails(ctx, token, cardID)
}

// CardFunding 执行卡片充值或转出。
func (p *pingPongSDKInterface) CardFunding(ctx context.Context, token string, params *CardFundingRequest) (*CardFundingResponse, error) {
	return p.sdk.CardFunding(ctx, token, params)
}

// QueryCardFundingOrders 查询卡片资金订单。
func (p *pingPongSDKInterface) QueryCardFundingOrders(ctx context.Context, token string, params *QueryCardFundingOrdersRequest) (*QueryCardFundingOrdersResponse, error) {
	return p.sdk.QueryCardFundingOrders(ctx, token, params)
}

// QueryDedicatedCardBalance 根据卡片 ID 查询卡片余额。
func (p *pingPongSDKInterface) QueryDedicatedCardBalance(ctx context.Context, token, cardID string) (*DedicatedCardBalanceResponse, error) {
	return p.sdk.QueryDedicatedCardBalance(ctx, token, cardID)
}

// CardAction 执行卡片冻结、解冻或销卡操作。
func (p *pingPongSDKInterface) CardAction(ctx context.Context, token string, params *CardActionRequest) error {
	return p.sdk.CardAction(ctx, token, params)
}

// QueryCardTransactions 查询卡片交易记录。
func (p *pingPongSDKInterface) QueryCardTransactions(ctx context.Context, token string, params *QueryCardTransactionsRequest) (*QueryCardTransactionsResponse, error) {
	return p.sdk.QueryCardTransactions(ctx, token, params)
}

// Query3DSDetails 根据卡片 ID 查询 3DS 详情。
func (p *pingPongSDKInterface) Query3DSDetails(ctx context.Context, token, cardID string) (*ThreeDSDetailsResponse, error) {
	return p.sdk.Query3DSDetails(ctx, token, cardID)
}

// CreateBudgetAccount 创建预算账户。
func (p *pingPongSDKInterface) CreateBudgetAccount(ctx context.Context, token string, params *CreateBudgetAccountRequest) (*CreateBudgetAccountResponse, error) {
	return p.sdk.CreateBudgetAccount(ctx, token, params)
}

// BudgetFunding 对预算账户充值或划转资金。
func (p *pingPongSDKInterface) BudgetFunding(ctx context.Context, token string, params *BudgetFundingRequest) (*BudgetFundingResponse, error) {
	return p.sdk.BudgetFunding(ctx, token, params)
}

// QueryBudgetFundingOrder 查询预算账户资金订单状态。
func (p *pingPongSDKInterface) QueryBudgetFundingOrder(ctx context.Context, token string, params *QueryBudgetFundingOrderRequest) (*QueryBudgetFundingOrderResponse, error) {
	return p.sdk.QueryBudgetFundingOrder(ctx, token, params)
}

// QueryBudgetAccountBalance 查询指定或全部预算账户余额。
func (p *pingPongSDKInterface) QueryBudgetAccountBalance(ctx context.Context, token, budgetID string) (*QueryBudgetAccountBalanceResponse, error) {
	return p.sdk.QueryBudgetAccountBalance(ctx, token, budgetID)
}

// QueryAccountTransactions 查询预算账户或卡片的已入账交易。
func (p *pingPongSDKInterface) QueryAccountTransactions(ctx context.Context, token string, params *QueryAccountTransactionsRequest) (*QueryAccountTransactionsResponse, error) {
	return p.sdk.QueryAccountTransactions(ctx, token, params)
}

// IsInvalidToken 判断是否为token失效错误（1002）
func IsInvalidToken(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code == "1002" || apiErr.Reason == "Invalid Token"
	}
	var ke *kratosErr.Error
	if errors.As(err, &ke) {
		return ke.Reason == "Invalid Token"
	}
	return false
}
