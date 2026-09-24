package payndapay

import (
	"errors"
	"regexp"
	"time"

	"github.com/shopspring/decimal"
)

// 错误定义
var (
	ErrEmptyBalanceAccountID = errors.New("balanceAccountID不能为空")
	ErrEmptyCardID           = errors.New("cardID不能为空")
	ErrEmptyCardholderID     = errors.New("cardholderID不能为空")
	ErrEmptyParams           = errors.New("参数不能为空")
	ErrEmptyResponse         = errors.New("响应数据为空")
	ErrSignatureVerify       = errors.New("签名验证失败")
	ErrNoSignature           = errors.New("响应头中没有签名信息")
	ErrEmptyNonce            = errors.New("nonce不能为空")
	ErrEmptyRequestID        = errors.New("requestID不能为空")

	// api error
	ErrRepeatedRequest   = errors.New("重复请求")    //{"code":500,"message":"repeated request","success":false}
	ErrNotFound          = errors.New("未找到")     //{"code":500,"message":"not found","success":false}
	ErrInsufficientFunds = errors.New("余额不足")    //{"code":500,"message":"Insufficient funds","success":false}
	ErrServerInternal    = errors.New("服务器内部错误") //{"code":500,"message":"error.server.internal","success":false}
	ErrInvalidParameter  = errors.New("无效的参数")   //{"code":500,"message":"invalid parameter","success":false}

)

func NewParamValidateError(msg string) error {
	return errors.New("参数验证失败:" + msg)
}

const (
	defaultCurrency = "USD"
)

var defaultPageQuery = map[string]any{
	"current":  1,
	"pageSize": 100,
}
var defaultPageQueryALL = map[string]any{
	"current":  1,
	"pageSize": 100,
}

type CardPaginateRequest struct {
	Current  int `json:"current"`
	PageSize int `json:"pageSize"`
}

type CardStatusUpdateRequest struct {
	CardID    string `json:"-"`
	RequestID string `json:"-"`
}

func (r *CardStatusUpdateRequest) Validate() error {
	if r == nil {
		return ErrEmptyParams
	}
	if r.CardID == "" {
		return ErrEmptyCardID
	}
	if r.RequestID == "" {
		return ErrEmptyRequestID
	}
	return nil
}

type BalanceAccountWalletTransferRequest struct {
	Nonce            string       `json:"-"`
	RequestID        string       `json:"-"`
	BalanceAccountID string       `json:"balanceAccountId"` // sdk根据环境获取
	Type             TransferType `json:"type"`
	Currency         string       `json:"currency"`
	Amount           string       `json:"amount"`
}

func (r *BalanceAccountWalletTransferRequest) Validate() error {
	if r.RequestID == "" {
		return ErrEmptyRequestID
	}
	if r.Type == "" {
		return errors.New("type不能为空")
	}
	if !r.Type.Valid() {
		return errors.New("type错误,不支持的类型:" + string(r.Type))
	}
	if r.Currency == "" {
		return errors.New("currency不能为空")
	}
	amount, err := decimal.NewFromString(r.Amount)
	if err != nil {
		return errors.New("金额格式错误")
	}
	if amount.LessThanOrEqual(decimal.NewFromFloat(0)) {
		return errors.New("amount必须大于0")
	}

	return nil

}

type RestyRequestOptions struct {
	Method    string
	Path      string
	Nonce     string
	RequestID string
	Params    any
	Result    any
}
type CardHolderRequest struct {
	Nonce               string `json:"-"`
	CardholderID        string `json:"-"`
	FirstName           string `json:"firstName,omitempty"`
	LastName            string `json:"lastName,omitempty"`
	MobilePrefix        string `json:"mobilePrefix,omitempty"`
	Mobile              string `json:"mobile,omitempty"`
	Email               string `json:"email,omitempty"`
	BillingAddressLine1 string `json:"billingAddressLine1,omitempty"`
	BillingAddressLine2 string `json:"billingAddressLine2,omitempty"`
	BillingCity         string `json:"billingCity,omitempty"`
	BillingCountryCode  string `json:"billingCountryCode,omitempty"`
	BillingPostalCode   string `json:"billingPostalCode,omitempty"`
	BillingState        string `json:"billingState,omitempty"`
	UnlimitedBalance    bool   `json:"unlimitedBalance,omitempty"` // default false. Unlimited balance option. true: unlimited. false: limited, using cardholder's wallet.
}

// Validate 验证CardHolderRequest的必填字段和格式
func (r *CardHolderRequest) Validate() error {
	if r.Nonce == "" {
		return ErrEmptyNonce
	}
	// if r.CardholderID == "" {
	// 	return errors.New("cardholderId不能为空")
	// }
	if r.FirstName == "" {
		return errors.New("firstName不能为空")
	}
	if r.LastName == "" {
		return errors.New("lastName不能为空")
	}

	// 验证Email格式
	if r.Email != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(r.Email) {
			return errors.New("email格式不正确")
		}
	}

	// 验证手机号
	if r.Mobile != "" && r.MobilePrefix == "" {
		return errors.New("mobilePrefix不能为空")
	}

	// 验证国家代码
	// if r.BillingCountryCode != "" {
	// 	if len(r.BillingCountryCode) != 2 {
	// 		return errors.New("billingCountryCode错误")
	// 	}
	// }

	return nil
}

func NewDefaultCardHolderRequest() *CardHolderRequest {
	return &CardHolderRequest{
		BillingCountryCode: "USA",
		UnlimitedBalance:   true,
	}
}

type CardHolderWalletRequest struct {
	Nonce        string               `json:"-"`
	CardholderID string               `json:"cardholderId,omitempty"`
	Type         BalanceOperationType `json:"type,omitempty"` //  INC: Increase DEC: Decrease [Enum: INC(1) , DEC(2)
	Currency     string               `json:"currency,omitempty"`
	Amount       string               `json:"amount,omitempty"`
}

// Validate 验证CardHolderWalletRequest的必填字段和格式
func (r *CardHolderWalletRequest) Validate() error {
	if r.Nonce == "" {
		return ErrEmptyNonce
	}
	if r.CardholderID == "" {
		return ErrEmptyCardholderID
	}

	if r.Type == "" {
		return errors.New("type不能为空")
	}

	if !r.Type.Valid() {
		return errors.New("type错误,不支持的类型:" + string(r.Type))
	}
	if r.Currency == "" {
		return errors.New("currency不能为空")
	}
	amount, err := decimal.NewFromString(r.Amount)
	if err != nil {
		return errors.New("金额格式错误")
	}
	if amount.LessThanOrEqual(decimal.NewFromFloat(0)) {
		return errors.New("amount必须大于0")
	}

	return nil
}

// Card represents a card
type CardCreateRequest struct {
	Nonce                      string `json:"-"`
	RequestID                  string `json:"-"`
	CardholderID               string `json:"cardholderId"`
	Currency                   string `json:"currency"`
	Amount                     string `json:"amount"`
	ExpirationDate             string `json:"expirationDate"`
	CardBinID                  string `json:"cardBinId"`                              // Card bin id
	SingleUse                  bool   `json:"singleUse,omitempty"`                    // default false
	TransactionCountLimitTotal int64  `json:"transactionCountLimitToTotal,omitempty"` // default 0
}

// Validate 验证CardCreateRequest的必填字段和格式
func (r *CardCreateRequest) Validate() error {
	if r.Nonce == "" {
		return ErrEmptyNonce
	}
	if r.RequestID == "" {
		return ErrEmptyRequestID
	}
	if r.CardholderID == "" {
		return ErrEmptyCardholderID
	}

	if r.Currency == "" {
		return errors.New("currency不能为空")
	}

	if r.Currency != "USD" {
		return errors.New("currency错误,只支持USD")
	}
	amount, err := decimal.NewFromString(r.Amount)
	if err != nil {
		return errors.New("金额格式错误")
	}
	if amount.LessThanOrEqual(decimal.NewFromFloat(0)) {
		return errors.New("amount必须大于0")
	}

	if r.ExpirationDate != "" {
		// 验证日期格式 MM/YY
		if _, err := time.Parse("01/06", r.ExpirationDate); err != nil {
			return errors.New("expirationDate格式不正确, 应为MM/YY")
		}
	}

	if r.CardBinID == "" {
		return errors.New("cardBinID不能为空")
	}

	return nil
}
func NewDefaultCardCreateRequest() *CardCreateRequest {
	return &CardCreateRequest{
		Currency:                   defaultCurrency,
		ExpirationDate:             genExpirationDate(), //从当前月份往后 6 个月到 3 年之间随机生成一个月份
		SingleUse:                  false,
		TransactionCountLimitTotal: 0,
	}

}

type CardBalanceUpdateRequest struct {
	Nonce     string `json:"-"`
	RequestID string `json:"-"`
	CardID    string `json:"cardId,omitempty"`
	Amount    string `json:"amount,omitempty"`
}

func (r *CardBalanceUpdateRequest) Validate() error {
	if r.Nonce == "" {
		return ErrEmptyNonce
	}
	if r.RequestID == "" {
		return ErrEmptyRequestID
	}
	if r.CardID == "" {
		return errors.New("cardID不能为空")
	}

	amount, err := decimal.NewFromString(r.Amount)
	if err != nil {
		return errors.New("金额格式错误")
	}
	if amount.LessThanOrEqual(decimal.NewFromFloat(0)) {
		return errors.New("amount必须大于0")
	}

	// if r.Amount != "" {
	// 	// 验证金额格式
	// 	amountRegex := regexp.MustCompile(`^[0-9]+(\.[0-9]{1,2})?$`)
	// 	if !amountRegex.MatchString(r.Amount) {
	// 		return errors.New("amount格式不正确,应为数字或小数点后最多两位")
	// 	}
	// }
	return nil
}

type CardControlRequest struct {
	Nonce            string            `json:"-"`
	CardID           string            `json:"-"`
	Period           CardControlPeriod `json:"period"`           // DAY: Daily transaction limit. MONTH: Monthly transaction limit. TOTAL: Cumulative transaction limit. ONCE: Single transaction limit.
	TransactionCount int64             `json:"transactionCount"` // Transaction count limit in the period. When the period is ONCE, this field is not used. When it is less or equal than 0, there are no restrictions.
	Amount           string            `json:"amount"`           // Transaction amount limit in the period. When it is less or equal than 0, there are no restrictions.
}

// Validate 验证CardControlRequest的必填字段和格式
func (r *CardControlRequest) Validate() error {
	if r.Nonce == "" {
		return ErrEmptyNonce
	}
	if r.CardID == "" {
		return errors.New("cardID不能为空")
	}
	if r.Period == "" {
		return errors.New("period不能为空")
	}

	if !r.Period.Valid() {
		return errors.New("period错误,不支持的值:" + string(r.Period))
	}
	if r.TransactionCount < 0 {
		return errors.New("transactionCount不能小于0")
	}
	amount, err := decimal.NewFromString(r.Amount)
	if err != nil {
		return errors.New("金额格式错误")
	}
	if amount.LessThanOrEqual(decimal.NewFromFloat(0)) {
		return errors.New("amount必须大于0")
	}

	return nil
}

// Card Top-up/Withdra
type CardBalanceTransferRequest struct {
	Nonce     string       `json:"-"`
	RequestID string       `json:"-"`
	CardID    string       `json:"cardId"`
	Amount    string       `json:"amount"`
	Type      TransferType `json:"type"` // IN: Top-up OUT: Withdraw
}

// Validate 验证CardBalanceTransferRequest的必填字段和格式
func (r *CardBalanceTransferRequest) Validate() error {
	if r.Nonce == "" {
		return ErrEmptyNonce
	}
	if r.RequestID == "" {
		return ErrEmptyRequestID
	}
	if r.CardID == "" {
		return errors.New("cardID不能为空")
	}

	amount, err := decimal.NewFromString(r.Amount)
	if err != nil {
		return errors.New("金额格式错误")
	}
	if amount.LessThanOrEqual(decimal.NewFromFloat(0)) {
		return errors.New("amount必须大于0")
	}

	if r.Type == "" {
		return errors.New("type不能为空")
	}
	if !r.Type.Valid() {
		return errors.New("type错误,不支持的类型:" + string(r.Type))
	}
	return nil
}

type CardTransactionsRequest struct {
	// Nonce                string `json:"-"`
	Current              int64  `json:"current"`
	PageSize             int64  `json:"pageSize"`
	BalanceAccountID     string `json:"balanceAccountId,omitempty"`
	CardholderID         string `json:"cardholderId,omitempty"`
	CardID               string `json:"cardId,omitempty"`
	TransactionTimeStart string `json:"transactionTimeStart,omitempty"` //2025-09-01 07:10:23
	TransactionTimeEnd   string `json:"transactionTimeEnd,omitempty"`
}

// Validate 验证CardTransactionsRequest的必填字段和格式
func (r *CardTransactionsRequest) Validate() error {
	return nil
}
