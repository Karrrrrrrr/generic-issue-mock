package uqpay

import (
	"strconv"

	enum "tman/enums"

	"github.com/shopspring/decimal"
)

type CreateCardholderReq struct {
	Email           string           `json:"email"`
	FirstName       string           `json:"first_name"`
	LastName        string           `json:"last_name"`
	DateOfBirth     string           `json:"date_of_birth"`
	CountryCode     string           `json:"country_code"`
	PhoneNumber     string           `json:"phone_number"`
	DocumentType    *string          `json:"document_type,omitempty"`
	Document        *string          `json:"document,omitempty"`
	DeliveryAddress *DeliveryAddress `json:"delivery_address"`
}

type DeliveryAddress struct {
	City       string `json:"city"`
	Country    string `json:"country"`
	Line1      string `json:"line1"`
	State      string `json:"state,omitempty"`
	Line2      string `json:"line2,omitempty"`
	PostalCode string `json:"postal_code"`
}

type UpdateCardholderReq struct {
	CardHolderID    string           `json:"cardholder_id"`
	CountryCode     *string          `json:"country_code,omitempty"`
	Email           *string          `json:"email,omitempty"`
	PhoneNumber     *string          `json:"phone_number,omitempty"`
	DocumentType    *string          `json:"document_type,omitempty"`
	Document        *string          `json:"document,omitempty"`
	DateOfBirth     *string          `json:"date_of_birth,omitempty"`
	DeliveryAddress *DeliveryAddress `json:"delivery_address,omitempty"`
}
type CreateCardReq struct {
	CardCurrency     string             `json:"card_currency"`
	CardLimit        *float64           `json:"card_limit,omitempty"`
	CardholderID     string             `json:"cardholder_id"`
	CardProductID    string             `json:"card_product_id"`
	SpendingControls []*SpendingControl `json:"spending_controls"`
	RiskControls     *RiskControl       `json:"risk_controls,omitempty"`
}
type SpendingControl struct {
	Amount   string `json:"amount"`
	Interval string `json:"interval"`
}
type RiskControl struct {
	Allow3dsTransactions string   `json:"allow_3ds_transactions"`
	AllowedMcc           []string `json:"allowed_mcc"`
	BlockedMcc           []string `json:"blocked_mcc"`
}
type UpdateCardReq struct {
	CardID             string
	CardLimit          *float64           `json:"card_limit,omitempty"`
	NoPinPaymentAmount *float64           `json:"no_pin_payment_amount,omitempty"`
	SpendingControls   []*SpendingControl `json:"spending_controls,omitempty"`
	RiskControls       *RiskControl       `json:"risk_controls,omitempty"`
}

type UpdateCardStatusReq struct {
	CardID       string  `json:"card_id"`
	CardStatus   string  `json:"card_status"`
	UpdateReason *string `json:"update_reason"`
}

type AssignCardReq struct {
	CardHolderID string            `json:"cardholder_id,omitempty"`
	CardNumber   string            `json:"card_number,omitempty"`
	CardCurrency enum.CurrencyEnum `json:"card_currency,omitempty"`
	CardMode     string            `json:"card_mode,omitempty"`
}

type ActivateCardReq struct {
	ActivationCode     string           `json:"activation_code"`       // 卡片激活码
	CardID             string           `json:"card_id"`               // 卡片的唯一标识符(UUID)
	NoPinPaymentAmount *decimal.Decimal `json:"no_pin_payment_amount"` // 免密额度,默认200USD
	Pin                string           `json:"pin"`                   // 6位密码
}

type ResetCardPinReq struct {
	CardID string `json:"card_id"` // 卡片的唯一标识符(UUID)
	Pin    string `json:"pin"`     // 6位密码
}

type ListTransactionReq struct {
	CardID     *string `json:"card_id,omitempty"`
	PageSize   int     `json:"page_size"`
	PageNumber int     `json:"page_number"`
	StartTime  *string `json:"start_time,omitempty"`
	EndTime    *string `json:"end_time,omitempty"`
}

// SimulateAuthorizationReq 三方授权接口参数
type SimulateAuthorizationReq struct {
	CardID               string          `json:"card_id"`
	TransactionAmount    decimal.Decimal `json:"transaction_amount"`
	TransactionCurrency  string          `json:"transaction_currency"`
	MerchantName         string          `json:"merchant_name"`
	MerchantCategoryCode string          `json:"merchant_category_code"`
}

type SimulateReversalReq struct {
	TransactionID string `json:"transaction_id"`
}

func (r *ListTransactionReq) toQueryParams() map[string]string {
	params := make(map[string]string)
	params["page_number"] = strconv.Itoa(r.PageNumber)
	params["page_size"] = strconv.Itoa(r.PageSize)

	if r.CardID != nil {
		params["card_id"] = *r.CardID
	}
	if r.StartTime != nil {
		params["start_time"] = *r.StartTime
	}
	if r.EndTime != nil {
		params["end_time"] = *r.EndTime
	}

	return params
}

type RetrieveBalanceReq struct {
	Currency string `json:"currency"` // 货币
}

type RetrieveIssuingBalanceReq struct {
	Currency string `json:"currency"` // 货币
}
