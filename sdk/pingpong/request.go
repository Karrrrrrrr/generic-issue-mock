package pingpong

import (
	"net/url"
	"strconv"
)

// QueryAccountsBalancesRequest 查询账户余额的筛选及分页参数。
type QueryAccountsBalancesRequest struct {
	AccountType      string   `json:"account_type,omitempty"`       // 账户类型筛选条件。
	CurrencyList     []string `json:"currency_list,omitempty"`      // 币种代码列表。
	SubaccountIDList []string `json:"subaccount_id_list,omitempty"` // 子账户 ID 列表。
	PageNo           int      `json:"page_no"`                      // 页码。
	PageSize         int      `json:"page_size"`                    // 每页记录数。
}

// Amount 表示带币种的金额。
type Amount struct {
	Amount   float64 `json:"amount"`             // 金额数值。
	Currency string  `json:"currency,omitempty"` // ISO 4217 币种代码；v3 开卡限额需填写。
}

// CreateCardRequest 创建卡片所需的产品、预算及限额信息。
type CreateCardRequest struct {
	RequestID           string  `json:"request_id"`                      // 唯一请求 ID，用于追踪请求。
	CardProductCode     string  `json:"card_product_code"`               // 卡产品代码，来自卡产品查询接口。
	CardCurrency        string  `json:"card_currency"`                   // 发卡币种，需与预算币种一致。
	BudgetID            string  `json:"budget_id"`                       // 卡片关联的预算账户 ID。
	CardholderID        string  `json:"cardholder_id,omitempty"`         // 持卡人 ID；启用 3DS 时需要提供。
	PerTransactionLimit *Amount `json:"per_transaction_limit,omitempty"` // 单笔交易限额。
	DailyLimit          *Amount `json:"daily_limit,omitempty"`           // 每日交易限额。
	WeeklyLimit         *Amount `json:"weekly_limit,omitempty"`          // 每周交易限额。
	MonthlyLimit        *Amount `json:"monthly_limit,omitempty"`         // 每月交易限额。
	LifetimeLimit       *Amount `json:"lifetime_limit,omitempty"`        // 卡片生命周期内的交易限额。
	CouponApplied       bool    `json:"coupon_applied"`                  // 是否使用奖励余额抵扣开卡费用。
	Remark              string  `json:"remark,omitempty"`                // 卡片备注。
}

// CardFundingRequest 卡片充值或转出的请求；重试时须保持 unique_order_id 不变。
type CardFundingRequest struct {
	CardID        string  `json:"card_id"`         // 目标卡片 ID。
	Action        string  `json:"action"`          // 资金操作类型：top_up（充值）或 withdraw（转出）。
	Amount        float64 `json:"amount"`          // 操作金额。
	UniqueOrderID string  `json:"unique_order_id"` // 业务订单唯一 ID，用于幂等。
}

// QueryCardFundingOrdersRequest 查询卡片充值及转出订单的筛选参数。
type QueryCardFundingOrdersRequest struct {
	PageNo        int    // 页码。
	PageSize      int    // 每页记录数。
	StartDate     string // 订单开始日期。
	EndDate       string // 订单结束日期。
	Status        string // 订单状态筛选条件。
	UniqueOrderID string // 业务订单唯一 ID。
	CardID        string // 卡片 ID。
}

// values 将非空订单筛选参数转换为 URL 查询参数。
func (r *QueryCardFundingOrdersRequest) values() url.Values {
	v := url.Values{}
	if r == nil {
		return v
	}
	if r.PageNo > 0 {
		v.Set("page_no", strconv.Itoa(r.PageNo))
	}
	if r.PageSize > 0 {
		v.Set("page_size", strconv.Itoa(r.PageSize))
	}
	if r.StartDate != "" {
		v.Set("start_date", r.StartDate)
	}
	if r.EndDate != "" {
		v.Set("end_date", r.EndDate)
	}
	if r.Status != "" {
		v.Set("status", r.Status)
	}
	if r.UniqueOrderID != "" {
		v.Set("unique_order_id", r.UniqueOrderID)
	}
	if r.CardID != "" {
		v.Set("card_id", r.CardID)
	}
	return v
}

// CardActionRequest 冻结、解冻、销卡或更新卡片备注的请求。
type CardActionRequest struct {
	CardID string `json:"card_id"`          // 待操作的卡片 ID。
	Action string `json:"action"`           // 操作类型：freeze、unfreeze、close 或 update_remark。
	Remark string `json:"remark,omitempty"` // 卡片备注；close 和 update_remark 操作需提供。
}

// QueryCardTransactionsRequest 查询卡交易的筛选及分页参数。
type QueryCardTransactionsRequest struct {
	CardID           string // 卡片 ID。
	Type             string // 交易类型。
	PageNo           int    // 页码。
	PageSize         int    // 每页记录数。
	StartTime        string // 交易时间范围起点。
	EndTime          string // 交易时间范围终点。
	StartPostingDate string // 入账日期范围起点。
	EndPostingDate   string // 入账日期范围终点。
	StartCreatedDate string // 创建日期范围起点。
	EndCreatedDate   string // 创建日期范围终点。
	ClearType        string // 清算类型筛选条件。
	Status           string // 交易状态筛选条件。
}

// values 将非空交易筛选参数转换为 URL 查询参数。
func (r *QueryCardTransactionsRequest) values() url.Values {
	v := url.Values{}
	if r == nil {
		return v
	}
	for key, value := range map[string]string{
		"card_id": r.CardID, "type": r.Type, "start_time": r.StartTime,
		"end_time": r.EndTime, "start_posting_date": r.StartPostingDate,
		"end_posting_date": r.EndPostingDate, "start_created_date": r.StartCreatedDate,
		"end_created_date": r.EndCreatedDate, "clear_type": r.ClearType, "status": r.Status,
	} {
		if value != "" {
			v.Set(key, value)
		}
	}
	if r.PageNo > 0 {
		v.Set("page_no", strconv.Itoa(r.PageNo))
	}
	if r.PageSize > 0 {
		v.Set("page_size", strconv.Itoa(r.PageSize))
	}
	return v
}

// CreateBudgetAccountRequest 创建预算账户的请求。
type CreateBudgetAccountRequest struct {
	BudgetName string `json:"budget_name"` // 预算账户名称。
}

// BudgetFundingRequest 预算账户充值或划转请求；重试时应复用相同的订单 ID。
type BudgetFundingRequest struct {
	UniqueOrderID  string  `json:"unique_order_id,omitempty"`  // 业务订单唯一 ID；top_up 时必填，重试复用，最多 36 字符。
	BudgetID       string  `json:"budget_id"`                  // 来源预算账户 ID。
	Action         string  `json:"action"`                     // 资金操作：top_up 或 transfer。
	Amount         float64 `json:"amount"`                     // 操作金额。
	Currency       string  `json:"currency"`                   // 充值币种或划出币种。
	TargetBudgetID string  `json:"target_budget_id,omitempty"` // transfer 时的目标预算账户 ID。
	TargetCurrency string  `json:"target_currency,omitempty"`  // transfer 时目标账户的币种。
}

// QueryBudgetFundingOrderRequest 预算账户资金订单的查询参数。
type QueryBudgetFundingOrderRequest struct {
	OrderID string // 预算资金订单 ID。
	Action  string // 操作类型：top_up 或 transfer。
}

// values 转换预算账户资金订单查询参数。
func (r *QueryBudgetFundingOrderRequest) values() url.Values {
	return url.Values{"order_id": {r.OrderID}, "action": {r.Action}}
}

// QueryAccountTransactionsRequest 查询预算账户或卡片的已入账交易。
type QueryAccountTransactionsRequest struct {
	PageNo           int    // 页码，从 1 开始。
	PageSize         int    // 每页条数，最大 100。
	BudgetID         string // 预算账户 ID；与 CardID 至少填写一个。
	CardID           string // 卡片 ID；与 BudgetID 至少填写一个。
	PostingStartTime string // 入账起始时间，ISO 8601 格式。
	PostingEndTime   string // 入账结束时间，ISO 8601 格式；区间不超过 31 天。
	TransactionType  string // 账户交易类型筛选条件。
	Direction        string // 资金方向：CREDIT 或 DEBIT。
}

// values 将非空账户交易筛选参数转换为 URL 查询参数。
func (r *QueryAccountTransactionsRequest) values() url.Values {
	v := url.Values{}
	for key, value := range map[string]string{
		"budget_id": r.BudgetID, "card_id": r.CardID,
		"posting_start_time": r.PostingStartTime, "posting_end_time": r.PostingEndTime,
		"transaction_type": r.TransactionType, "direction": r.Direction,
	} {
		if value != "" {
			v.Set(key, value)
		}
	}
	if r.PageNo > 0 {
		v.Set("page_no", strconv.Itoa(r.PageNo))
	}
	if r.PageSize > 0 {
		v.Set("page_size", strconv.Itoa(r.PageSize))
	}
	return v
}
