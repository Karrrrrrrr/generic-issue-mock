package uqpay

import (
	"time"

	"github.com/shopspring/decimal"
)

// GetAccessTokenResp 获取access_token接口返回
type GetAccessTokenResp struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	AuthToken string `json:"auth_token"`
	ExpiredAt int64  `json:"expired_at"`
}

type CreateCardholderResp struct {
	Code             string `json:"code"` // 错误码(三方成功时不会给code，失败情况会给字符串型code)
	Message          string `json:"message"`
	Type             string `json:"type"`
	CardholderId     string `json:"cardholder_id"`
	CardholderStatus string `json:"cardholder_status"`
}

// 由于三方接口失败时只给了code，不给status,这里对失败情况补充一下FAILED状态
func (u *CreateCardholderResp) checkFailCode() {
	if len(u.Code) > 0 {
		u.CardholderStatus = THIRD_PARTY_UP_CARD_HOLDER_STATUS_ENUM_FAILED
	}
}

type UpdateCardholderResp struct {
	Code             string `json:"code"` // 错误码(三方成功时不会给code，失败情况会给字符串型code)
	Message          string `json:"message"`
	Type             string `json:"type"`
	CardholderId     string `json:"cardholder_id"`
	CardholderStatus string `json:"cardholder_status"`
}

// 由于三方接口失败时只给了code，不给status,这里对失败情况补充一下FAILED状态
func (u *UpdateCardholderResp) checkFailCode() {
	if len(u.Code) > 0 {
		u.CardholderStatus = THIRD_PARTY_UP_CARD_HOLDER_STATUS_ENUM_FAILED
	}
}

type CardholderDetail struct {
	CardholderId     string           `json:"cardholder_id"`
	NumberOfCards    int              `json:"number_of_cards"`
	Email            string           `json:"email"`
	FirstName        string           `json:"first_name"`
	LastName         string           `json:"last_name"`
	DateOfBirth      string           `json:"date_of_birth"`
	CountryCode      string           `json:"country_code"`
	PhoneNumber      string           `json:"phone_number"`
	DocumentType     string           `json:"document_type,omitempty"`
	Document         string           `json:"document,omitempty"`
	DeliveryAddress  *DeliveryAddress `json:"delivery_address"`
	CreateTime       string           `json:"create_time"`
	CardholderStatus string           `json:"cardholder_status"`
	ReviewStatus     string           `json:"review_status"`
}

type ListCardProductsResp struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Type       string `json:"type"`
	TotalPages int    `json:"total_pages"`
	TotalItems int    `json:"total_items"`
	Data       []struct {
		ProductID          string   `json:"product_id"`
		ModeType           string   `json:"mode_type"`
		CardBin            string   `json:"card_bin"`
		CardForm           []string `json:"card_form"`
		MaxCardQuota       int      `json:"max_card_quota"`
		CardScheme         string   `json:"card_scheme"`
		NoPinPaymentAmount []struct {
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"no_pin_payment_amount"`
		CardCurrency  []string `json:"card_currency"`
		CreateTime    string   `json:"create_time"`
		UpdateTime    string   `json:"update_time"`
		ProductStatus string   `json:"product_status"`
	} `json:"data"`
}

type CreateCardResp struct {
	Code        string `json:"code"` // 错误码(三方成功时不会给code，失败情况会给字符串型code)
	Message     string `json:"message"`
	Type        string `json:"type"`
	CardID      string `json:"card_id"`
	CardOrderID string `json:"card_order_id"`
	CreateTime  string `json:"create_time"`
	CardStatus  string `json:"card_status"`
	OrderStatus string `json:"order_status"`
}

// 由于三方接口失败时只给了code，不给status,这里对失败情况补充一下FAILED状态
func (u *CreateCardResp) checkFailCode() {
	if len(u.Code) > 0 {
		u.CardStatus = THIRD_PARTY_UP_CARD_STATUS_ENUM_FAILED
		u.OrderStatus = THIRD_PARTY_UP_ORDER_STATUS_ENUM_FAILED
	}
}

type UpdateCardResp struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Type        string `json:"type"`
	CardID      string `json:"card_id"`
	CardOrderID string `json:"card_order_id"`
	CardStatus  string `json:"card_status"`
	OrderStatus string `json:"order_status"`
}

// 由于三方接口失败时只给了code，不给status,这里对失败情况补充一下FAILED状态
func (u *UpdateCardResp) checkFailCode() {
	if len(u.Code) > 0 {
		u.CardStatus = THIRD_PARTY_UP_CARD_STATUS_ENUM_FAILED
	}
}

type GetCardInfo struct {
	Code               string             `json:"code"`
	Message            string             `json:"message"`
	Type               string             `json:"type"`
	CardId             string             `json:"card_id"`
	CardBin            string             `json:"card_bin"`
	CardScheme         string             `json:"card_scheme"`
	CardCurrency       string             `json:"card_currency"`
	CardNumber         string             `json:"card_number"`
	FormFactor         string             `json:"form_factor"`
	ModeType           string             `json:"mode_type"`
	CardProductID      string             `json:"card_product_id"`
	CardLimit          decimal.Decimal    `json:"card_limit"`
	AvailableBalance   string             `json:"available_balance"`
	Cardholder         *CardholderDetail  `json:"cardholder"`
	SpendingControls   []*SpendingControl `json:"spending_controls"`
	NoPinPaymentAmount string             `json:"no_pin_payment_amount"`
	RiskControls       *RiskControl       `json:"risk_controls"`
	Metadata           map[string]string  `json:"metadata"`
	CardStatus         string             `json:"card_status"`
	UpdateReason       string             `json:"update_reason"`
	ConsumedAmount     string             `json:"consumed_amount"`
}

// 由于三方接口失败时只给了code，不给status,这里对失败情况补充一下FAILED状态
func (u *GetCardInfo) checkFailCode() {
	if len(u.Code) > 0 {
		u.CardStatus = THIRD_PARTY_UP_CARD_STATUS_ENUM_FAILED
	}
}

type UpdateCardStatusResp struct {
	Code         string `json:"code"`
	Message      string `json:"message"`
	Type         string `json:"type"`
	CardID       string `json:"card_id"`
	CardOrderID  string `json:"card_order_id"`
	OrderStatus  string `json:"order_status"`
	UpdateReason string `json:"update_reason"`
}

// 由于三方接口失败时只给了code，不给status,这里对失败情况补充一下FAILED状态
func (u *UpdateCardStatusResp) checkFailCode() {
	if len(u.Code) > 0 {
		u.OrderStatus = THIRD_PARTY_UP_ORDER_STATUS_ENUM_FAILED
	}
}

type CardPrivateInfo struct {
	Code               string `json:"code"`
	Message            string `json:"message"`
	Type               string `json:"type"`
	Cvv                string `json:"cvv"`
	ExpireDate         string `json:"expire_date"`
	CardNumber         string `json:"card_number"`
	CardExpirationDate string `json:"card_expiration_date"`
}

type AssignCardResp struct {
	CardID      string     `json:"card_id,omitempty"`       // UUID
	CardOrderID string     `json:"card_order_id,omitempty"` // UUID
	CreateTime  *time.Time `json:"create_time,omitempty"`
	CardStatus  string     `json:"card_status,omitempty"`
	OrderStatus string     `json:"order_status,omitempty"`

	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// 由于三方接口失败时只给了code，不给status,这里对失败情况补充一下FAILED状态
func (u *AssignCardResp) checkFailCode() {
	if len(u.Code) > 0 {
		u.OrderStatus = THIRD_PARTY_UP_ORDER_STATUS_ENUM_FAILED
	}
}

type ActivateCardResp struct {
	RequestStatus string `json:"request_status,omitempty"`

	Code        string `json:"code,omitempty"`
	Message     string `json:"message,omitempty"`
	OrderStatus string `json:"order_status"`
}

func (u *ActivateCardResp) checkFailCode() {
	if len(u.Code) > 0 {
		u.OrderStatus = THIRD_PARTY_UP_ORDER_STATUS_ENUM_FAILED
	}
}

type ResetCardPinResp struct {
	RequestStatus string `json:"request_status,omitempty"`

	Code        string `json:"code,omitempty"`
	Message     string `json:"message,omitempty"`
	OrderStatus string `json:"order_status"`
}

func (u *ResetCardPinResp) checkFailCode() {
	if len(u.Code) > 0 {
		u.OrderStatus = THIRD_PARTY_UP_ORDER_STATUS_ENUM_FAILED
	}
}

type ListTransactionResp struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Type       string            `json:"type"`
	TotalPages uint32            `json:"total_pages"`
	TotalItems uint32            `json:"total_items"`
	Data       []TransactionInfo `json:"data"`
}

type TransactionInfo struct {
	CardID                 string       `json:"card_id"`
	CardNumber             string       `json:"card_number"`
	CardholderID           string       `json:"cardholder_id"`
	TransactionID          string       `json:"transaction_id"`
	ShortTransactionID     string       `json:"short_transaction_id"`
	OriginalTransactionID  string       `json:"original_transaction_id"`
	TransactionType        string       `json:"transaction_type"`
	TransactionFee         string       `json:"transaction_fee"`
	TransactionFeeCurrency string       `json:"transaction_fee_currency"`
	FeePassThrough         string       `json:"fee_pass_through"`
	CardAvailableBalance   string       `json:"card_available_balance"`
	AuthorizationCode      string       `json:"authorization_code"`
	BillingAmount          string       `json:"billing_amount"`
	BillingCurrency        string       `json:"billing_currency"`
	TransactionAmount      string       `json:"transaction_amount"`
	TransactionCurrency    string       `json:"transaction_currency"`
	TransactionTime        string       `json:"transaction_time"`
	PostedTime             string       `json:"posted_time"`
	MerchantData           MerchantData `json:"merchant_data"`
	Description            string       `json:"description"`
	TransactionStatus      string       `json:"transaction_status"`
	WalletType             string       `json:"wallet_type"`
}

type MerchantData struct {
	CategoryCode string `json:"category_code"`
	City         string `json:"city"`
	Country      string `json:"country"`
	Name         string `json:"name"`
}

type RetrieveBalanceResp struct {
	BalanceID        string `json:"balance_id"`
	Currency         string `json:"currency"`
	AvailableBalance string `json:"available_balance"`
	PrepaidBalance   string `json:"prepaid_balance"`
	MarginBalance    string `json:"margin_balance"`
	FrozenBalance    string `json:"frozen_balance"`
	CreateTime       string `json:"create_time"`
	LastTradeTime    string `json:"last_trade_time"`
	BalanceStatus    string `json:"balance_status"`
	Code             string `json:"code"`
	Message          string `json:"message"`
}

type RetrieveIssuingBalanceResp struct {
	BalanceID        string `json:"balance_id"`
	Currency         string `json:"currency"`
	AvailableBalance string `json:"available_balance"`
	MarginBalance    string `json:"margin_balance"`
	FrozenBalance    string `json:"frozen_balance"`
	CreateTime       string `json:"create_time"`
	LastTradeTime    string `json:"last_trade_time"`
	BalanceStatus    string `json:"balance_status"`
	Type             string `json:"type"`
	Code             string `json:"code"`
	Message          string `json:"message"`
}

type SimulateAuthorizationResp struct {
	CardId               string          `json:"card_id"`
	CardNumber           string          `json:"card_number"`
	CardholderId         string          `json:"cardholder_id"`
	TransactionId        string          `json:"transaction_id"`
	TransactionType      string          `json:"transaction_type"`
	CardAvailableBalance decimal.Decimal `json:"card_available_balance"`
	AuthorizationCode    string          `json:"authorization_code"`
	BillingAmount        decimal.Decimal `json:"billing_amount"`
	BillingCurrency      string          `json:"billing_currency"`
	TransactionAmount    decimal.Decimal `json:"transaction_amount"`
	TransactionCurrency  string          `json:"transaction_currency"`
	TransactionTime      time.Time       `json:"transaction_time"`
	PostedTime           time.Time       `json:"posted_time"`
	MerchantData         MerchantData    `json:"merchant_data"`
	FailureReason        string          `json:"failure_reason"`
	TransactionStatus    string          `json:"transaction_status"`

	Code        string `json:"code"`
	Message     string `json:"message"`
	OrderStatus string `json:"order_status"`
}

func (c *SimulateAuthorizationResp) checkFailCode() {
	if len(c.Code) > 0 {
		c.OrderStatus = THIRD_PARTY_UP_ORDER_STATUS_ENUM_FAILED
	}
}

type SimulateReversalResp struct {
	AuthorizationCode    string          `json:"authorization_code"`
	BillingAmount        decimal.Decimal `json:"billing_amount"`
	BillingCurrency      string          `json:"billing_currency"`
	CardAvailableBalance decimal.Decimal `json:"card_available_balance"`
	CardId               string          `json:"card_id"`
	CardNumber           string          `json:"card_number"`
	CardholderId         string          `json:"cardholder_id"`
	FailureReason        string          `json:"failure_reason"`
	MerchantData         MerchantData    `json:"merchant_data"`
	OriginTransactionId  string          `json:"origin_transaction_id"`
	PostedTime           time.Time       `json:"posted_time"`
	TransactionAmount    decimal.Decimal `json:"transaction_amount"`
	TransactionCurrency  string          `json:"transaction_currency"`
	TransactionId        string          `json:"transaction_id"` // uuid
	TransactionStatus    string          `json:"transaction_status"`
	TransactionTime      time.Time       `json:"transaction_time"`
	TransactionType      string          `json:"transaction_type"`

	Code        string `json:"code"`
	OrderStatus string `json:"order_status"`
	Message     string `json:"message"`
}

func (c *SimulateReversalResp) checkFailCode() {
	if len(c.Code) > 0 {
		c.OrderStatus = THIRD_PARTY_UP_ORDER_STATUS_ENUM_FAILED
	}
}

type PanToken struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Type      string    `json:"type"`
	Token     string    `json:"token"`
	ExpiresIn int       `json:"expires_in"`
	ExpiresAt time.Time `json:"expires_at"`
}

// CardOrder 卡片订单
type CardOrder struct {
	Code             string          `json:"code"`
	Message          string          `json:"message"`
	CardID           string          `json:"card_id"`
	CardOrderID      string          `json:"card_order_id"`
	OrderType        string          `json:"order_type"`
	Amount           decimal.Decimal `json:"amount"`
	CardCurrency     string          `json:"card_currency"`
	CreateTime       string          `json:"create_time"`
	UpdateUpdateTime string          `json:"update_time"`
	CompleteTime     string          `json:"complete_time"`
	OrderStatus      string          `json:"order_status"` // PENDING PROCESSING SUCCESS FAILED
}
