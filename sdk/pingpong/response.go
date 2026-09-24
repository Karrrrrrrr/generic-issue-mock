package pingpong

// 金额字段按接口定义分别保留 number 和 string，避免交易金额的隐式浮点转换。

// AccessTokenResponse 获取访问令牌的响应。
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"` // 业务接口使用的访问令牌。
	ExpiresIn   int    `json:"expires_in"`   // 令牌有效期，单位为秒。
}

// QueryAccountsBalancesResponse 账户余额分页查询结果。
type QueryAccountsBalancesResponse struct {
	PageNo   int              `json:"page_no"`   // 当前页码。
	PageSize int              `json:"page_size"` // 每页记录数。
	TotalNum int              `json:"total_num"` // 记录总数。
	ItemList []AccountBalance `json:"item_list"` // 账户余额列表。
}

// AccountBalance 子账户各币种的余额信息。
type AccountBalance struct {
	SubaccountID    string  `json:"subaccount_id"`    // 子账户 ID。
	AccountType     string  `json:"account_type"`     // 账户类型。
	Currency        string  `json:"currency"`         // 余额币种。
	Source          string  `json:"source"`           // 余额来源。
	Status          string  `json:"status"`           // 账户状态。
	AvailableAmount float64 `json:"available_amount"` // 可用金额。
	FrozenAmount    float64 `json:"frozen_amount"`    // 冻结金额。
	TotalAmount     float64 `json:"total_amount"`     // 总金额。
}

// QueryCardProductsResponse 可用卡产品查询结果。
type QueryCardProductsResponse struct {
	ProductList []CardProduct `json:"product_list"` // 卡产品列表。
}

// CardProduct 卡产品的发卡及计费信息。
type CardProduct struct {
	CardNetwork        string `json:"card_network"`         // 卡组织。
	CardProductCode    string `json:"card_product_code"`    // 卡产品代码。
	ValidMonth         int    `json:"valid_month"`          // 卡片有效月数。
	AllowWithdrawal    bool   `json:"allow_withdrawal"`     // 是否允许提现。
	CardApplicationFee string `json:"card_application_fee"` // 开卡费用。
	BillingCurrency    string `json:"billing_currency"`     // 账单币种。
	BinRange           string `json:"bin_range"`            // 卡 BIN 范围。
	Share              bool   `json:"share"`                // 是否为共享产品。
	NameCN             string `json:"name_cn"`              // 产品中文名称。
	NameEN             string `json:"name_en"`              // 产品英文名称。
}

// CreateCardResponse 创建卡片的响应。
type CreateCardResponse struct {
	CardID string `json:"card_id"` // 新建卡片的 ID。
}

// CardDetailsResponse 包含完整卡号与 CVC，调用方不得将该响应写入日志。
type CardDetailsResponse struct {
	CardID                 string  `json:"card_id"`                  // 卡片 ID。
	BudgetID               string  `json:"budget_id"`                // 关联预算 ID。
	CardStatus             string  `json:"card_status"`              // 卡片状态。
	CardType               string  `json:"card_type"`                // 卡片类型。
	CardNumber             string  `json:"card_number"`              // 完整卡号，属于敏感信息。
	CVC                    string  `json:"cvc"`                      // 卡片安全码，属于敏感信息。
	CardExpiryDate         string  `json:"card_expiry_date"`         // 卡片到期日期。
	WithdrawalAllowed      bool    `json:"withdrawal_allowed"`       // 是否允许提现。
	CancellationInProgress bool    `json:"cancellation_in_progress"` // 是否正在销卡。
	Cancelled              bool    `json:"cancelled"`                // 是否已销卡。
	BillingCurrency        string  `json:"billing_currency"`         // 账单币种。
	MaxTransactions        int     `json:"max_transactions"`         // 最大交易笔数。
	PerTransactionLimit    *Amount `json:"per_transaction_limit"`    // 单笔交易限额。
	DailyLimit             *Amount `json:"daily_limit"`              // 每日交易限额。
	WeeklyLimit            *Amount `json:"weekly_limit"`             // 每周交易限额。
	MonthlyLimit           *Amount `json:"monthly_limit"`            // 每月交易限额。
	LifetimeLimit          *Amount `json:"lifetime_limit"`           // 卡片生命周期内的交易限额。
	Remark                 string  `json:"remark"`                   // 卡片备注。
	CreatedAt              string  `json:"created_at"`               // 卡片创建时间。
}

// CardFundingResponse 卡片充值或转出的操作结果。
type CardFundingResponse struct {
	RecordID string  `json:"record_id"` // 资金操作记录 ID。
	Action   string  `json:"action"`    // 资金操作类型。
	Amount   float64 `json:"amount"`    // 操作金额。
	Currency string  `json:"currency"`  // 操作币种。
}

// QueryCardFundingOrdersResponse 卡片资金订单的分页查询结果。
type QueryCardFundingOrdersResponse struct {
	TotalNum int                `json:"total_num"` // 订单总数。
	PageNo   int                `json:"page_no"`   // 当前页码。
	PageSize int                `json:"page_size"` // 每页记录数。
	List     []CardFundingOrder `json:"list"`      // 资金订单列表。
}

// CardFundingOrder 卡片充值或转出订单。
type CardFundingOrder struct {
	RecordID      string  `json:"record_id"`       // 资金操作记录 ID。
	UniqueOrderID string  `json:"unique_order_id"` // 业务订单唯一 ID。
	Created       string  `json:"created"`         // 订单创建时间。
	Amount        float64 `json:"amount"`          // 订单金额。
	Currency      string  `json:"currency"`        // 订单币种。
	CardNumber    string  `json:"card_number"`     // 卡号，调用方应避免记录在日志中。
	CardID        string  `json:"card_id"`         // 卡片 ID。
	Remark        string  `json:"remark"`          // 订单备注。
	Status        string  `json:"status"`          // 订单状态。
	Action        string  `json:"action"`          // 资金操作类型。
}

// DedicatedCardBalanceResponse 指定卡片的余额查询结果。
type DedicatedCardBalanceResponse struct {
	CardNumber       string  `json:"card_number"`       // 卡号，调用方应避免记录在日志中。
	AvailableBalance float64 `json:"available_balance"` // 卡片可用余额。
	Currency         string  `json:"currency"`          // 余额币种。
}

// QueryCardTransactionsResponse 卡交易的分页查询结果。
type QueryCardTransactionsResponse struct {
	TotalNum        int               `json:"total_num"`        // 交易总数。
	PageNo          int               `json:"page_no"`          // 当前页码。
	PageSize        int               `json:"page_size"`        // 每页记录数。
	TransactionList []CardTransaction `json:"transaction_list"` // 卡交易列表。
}

// CardTransaction 卡片交易记录。
type CardTransaction struct {
	AuthorizationID     string          `json:"authorization_id"`     // 授权记录 ID。
	CardNumber          string          `json:"card_number"`          // 卡号，调用方应避免记录在日志中。
	CardID              string          `json:"card_id"`              // 卡片 ID。
	BudgetID            string          `json:"budget_id"`            // 关联预算 ID。
	TransactionDate     string          `json:"transaction_date"`     // 交易时间。
	BillingAmount       string          `json:"billing_amount"`       // 账单金额。
	BillingCurrency     string          `json:"billing_currency"`     // 账单币种。
	TransactionAmount   string          `json:"transaction_amount"`   // 原交易金额。
	TransactionCurrency string          `json:"transaction_currency"` // 原交易币种。
	MerchantName        string          `json:"merchant_name"`        // 商户名称。
	MerchantCountry     string          `json:"merchant_country"`     // 商户国家或地区。
	MCC                 string          `json:"mcc"`                  // 商户类别码。
	Remark              string          `json:"remark"`               // 交易备注。
	Type                string          `json:"type"`                 // 交易类型。
	Status              string          `json:"status"`               // 交易状态。
	FailReason          string          `json:"fail_reason"`          // 交易失败原因。
	ApproveCode         string          `json:"approve_code"`         // 交易授权码。
	Fees                TransactionFees `json:"fees"`                 // 交易费用明细。
}

// TransactionFees 卡交易产生的费用。
type TransactionFees struct {
	RateFee            string `json:"rate_fee"`             // 费率费用金额。
	RateFeeCurrency    string `json:"rate_fee_currency"`    // 费率费用币种。
	ThreeDSFee         string `json:"threeds_fee"`          // 3DS 验证费用金额。
	ThreeDSFeeCurrency string `json:"threeds_fee_currency"` // 3DS 验证费用币种。
}

// ThreeDSDetailsResponse 含个人信息及安全问题答案，调用方不得记录响应内容。
type ThreeDSDetailsResponse struct {
	CardholderID     string `json:"cardholder_id"`     // 持卡人 ID。
	BudgetID         string `json:"budget_id"`         // 关联预算 ID。
	FirstName        string `json:"first_name"`        // 持卡人名。
	LastName         string `json:"last_name"`         // 持卡人姓。
	DateOfBirth      string `json:"date_of_birth"`     // 持卡人出生日期。
	CallPrefix       string `json:"call_prefix"`       // 电话区号。
	Mobile           string `json:"mobile"`            // 手机号码。
	PostCode         string `json:"post_code"`         // 邮政编码。
	CountryCode      string `json:"country_code"`      // 国家或地区代码。
	State            string `json:"state"`             // 州或省份。
	City             string `json:"city"`              // 城市。
	AddressLine      string `json:"address_line"`      // 地址详情。
	Email            string `json:"email"`             // 电子邮箱。
	SecurityIndex    string `json:"security_index"`    // 安全问题索引。
	SecurityQuestion string `json:"security_question"` // 安全问题。
	SecurityAnswer   string `json:"security_answer"`   // 安全问题答案，属于敏感信息。
}

// CreateBudgetAccountResponse 新建预算账户的结果。
type CreateBudgetAccountResponse struct {
	BudgetID string `json:"budget_id"` // 新建预算账户 ID。
}

// BudgetFundingResponse 预算账户资金操作结果。
type BudgetFundingResponse struct {
	RecordID string `json:"record_id"` // 资金操作记录 ID。
}

// QueryBudgetFundingOrderResponse 预算账户资金订单状态。
type QueryBudgetFundingOrderResponse struct {
	OrderID string `json:"order_id"` // 资金订单 ID。
	Action  string `json:"action"`   // 操作类型：top_up 或 transfer。
	Status  string `json:"status"`   // 订单状态：SUCCESS、FAIL 或 PROCESSING。
}

// QueryBudgetAccountBalanceResponse 预算账户余额查询结果。
type QueryBudgetAccountBalanceResponse struct {
	BalanceList []BudgetAccountBalance `json:"balance_list"` // 各币种预算账户余额。
}

// BudgetAccountBalance 单个预算账户的币种余额。
type BudgetAccountBalance struct {
	Balance    float64 `json:"balance"`     // 账户余额。
	Currency   string  `json:"currency"`    // 余额币种。
	BudgetID   string  `json:"budget_id"`   // 预算账户 ID。
	BudgetName string  `json:"budget_name"` // 预算账户名称。
}

// QueryAccountTransactionsResponse 已入账账户交易分页结果。
type QueryAccountTransactionsResponse struct {
	TotalNum        int                  `json:"total_num"`        // 匹配的交易总数。
	PageNo          int                  `json:"page_no"`          // 当前页码。
	PageSize        int                  `json:"page_size"`        // 本页返回的记录数。
	TransactionList []AccountTransaction `json:"transaction_list"` // 已入账交易列表。
}

// AccountTransaction 预算账户或卡片的已入账余额变动记录。
type AccountTransaction struct {
	TransactionID          string                     `json:"transaction_id"`          // 交易 ID。
	TransactionType        string                     `json:"transaction_type"`        // 交易类型。
	TransactionDescription string                     `json:"transaction_description"` // 交易说明。
	AccountType            string                     `json:"account_type"`            // 账户类型。
	BudgetID               string                     `json:"budget_id"`               // 预算账户 ID。
	BudgetName             string                     `json:"budget_name"`             // 预算账户名称。
	CardID                 string                     `json:"card_id"`                 // 卡片 ID。
	IsOTA                  bool                       `json:"is_ota"`                  // 是否为 OTA 卡。
	CardNumber             string                     `json:"card_number"`             // 脱敏卡号，避免写入业务日志。
	CardRemark             string                     `json:"card_remark"`             // 卡片备注。
	TransactionTime        string                     `json:"transaction_time"`        // 交易时间。
	PostingTime            string                     `json:"posting_time"`            // 入账时间。
	TransactionAmount      float64                    `json:"transaction_amount"`      // 原交易金额。
	TransactionCurrency    string                     `json:"transaction_currency"`    // 原交易币种。
	BillingAmount          float64                    `json:"billing_amount"`          // 记账金额。
	BillingCurrency        string                     `json:"billing_currency"`        // 记账币种。
	Direction              string                     `json:"direction"`               // 余额变动方向：CREDIT 或 DEBIT。
	BalancePaid            float64                    `json:"balance_paid"`            // 已支付金额。
	OutstandingAmount      float64                    `json:"outstanding_amount"`      // 未结清金额。
	BeforeBalance          float64                    `json:"before_balance"`          // 交易前余额。
	AfterBalance           float64                    `json:"after_balance"`           // 交易后余额。
	ReconciliationInfo     *AccountReconciliationInfo `json:"reconciliation_info"`     // 对账业务信息。
}

// AccountReconciliationInfo 账户交易关联的业务对账字段。
type AccountReconciliationInfo struct {
	ServiceType  string `json:"service_type"`   // 业务类型。
	CustomField1 string `json:"custom_field_1"` // 自定义字段 1。
}
