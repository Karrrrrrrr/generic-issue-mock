package model

import (
	"generic-mock/enums"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/plugin/soft_delete"
)

type ID = string

type BaseModel struct {
	ID        ID                    `gorm:"type:uuid;default:uuidv7();primaryKey"`
	CreatedAt time.Time             `gorm:"not null"`
	UpdatedAt time.Time             `gorm:"not null"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete"`
}

type Card struct {
	BaseModel
	Channel                enums.Channel
	CardProductID          ID
	CardBin                string
	CardNumber             string `gorm:"uniqueIndex"`
	Cvv                    string
	ExpireTime             string
	Status                 enums.CardStatus
	VirtualAccountID       *ID // nil 表示普通卡，非 nil 表示虚拟账户卡，共享余额。
	BalanceID              *ID // 虚拟账户卡指向虚拟账户的钱包 ID，减少一次查询。
	CardHolderID           ID  // 持卡人 ID，允许为空。
	FormType               enums.CardFormType
	RequestID              string `gorm:"uniqueIndex"`
	LastOperationRequestID string `gorm:"uniqueIndex"`
	LastOperationType      enums.OperationType
	LastOperationStatus    enums.OperationStatus
	CardCurrency           enums.Currency
	CardScheme             string
	CardType               enums.CardType
	RawRequest             []byte `gorm:"type:jsonb"`

	// ref
	//CardHolderInline *CardHolder `gorm:"-"` // 对于不需要持卡人的渠道, 直接存json 不做关联
	VirtualAccount *VirtualAccount
	Wallet         *Wallet `gorm:"foreignKey:BalanceID"`
	CardHolder     *CardHolder
	CardProduct    *CardProduct
	VirtualCard    *VirtualCard
	PhysicalCard   *PhysicalCard
}

type CardProduct struct {
	BaseModel
	Channel   enums.Channel `gorm:"uniqueIndex:idx_card_products_channel_prefix"`
	Prefix    string        `gorm:"uniqueIndex:idx_card_products_channel_prefix"`
	IsDefault bool

	Cards []*Card `gorm:"foreignKey:CardProductID"`
}

type VirtualCard struct {
	BaseModel
	CardID ID
}

type PhysicalCard struct {
	BaseModel
	CardID ID
}

type Wallet struct {
	BaseModel
	Amount     decimal.Decimal
	PendingIn  decimal.Decimal
	PendingOut decimal.Decimal
	In         decimal.Decimal
	Out        decimal.Decimal
	Type       enums.WalletType
	Currency   enums.Currency
}

type VirtualAccount struct {
	BaseModel
	WalletID ID
	Wallet   *Wallet
	Name     string
}

type Account struct {
	BaseModel
	Name     string
	WalletID ID
	Wallet   *Wallet
}

type CardTransaction struct {
	BaseModel
	Channel                 enums.Channel
	OriginCardTransactionID ID
	AuthorizationID         ID
	CardID                  ID
	Status                  enums.CardTransactionStatus
	Type                    enums.CardTransactionType
	Currency                enums.Currency
	TxAmount                decimal.Decimal
	TxCurrency              enums.Currency
	RequestID               string
	MerchantName            string
	MerchantCountry         string
	MerchantMCC             string
	AuthorizationCode       string
	OccurredAt              time.Time
	SettledAt               *time.Time
	RawPayload              []byte `gorm:"type:jsonb"`

	Authorization         *Authorization
	OriginCardTransaction *CardTransaction
	CardTransactions      []*CardTransaction `gorm:"foreignKey:OriginCardTransactionID"`
}

type Authorization struct {
	BaseModel
	Channel enums.Channel

	CardID                ID
	OriginAuthorizationID ID
	Currency              enums.Currency
	Amount                decimal.Decimal
	MerchantName          string
	MerchantCountry       string
	MerchantMCC           string
	AuthorizationCode     string
	Status                enums.CardTransactionStatus
	OccurredAt            time.Time
	RawPayload            []byte `gorm:"type:jsonb"`

	CardTransactions []*CardTransaction
}

type CardHolder struct {
	BaseModel
	Channel                enums.Channel
	FirstName              string
	LastName               string
	Email                  string
	Mobile                 string
	MobilePrefix           string
	DateOfBirth            string
	NationalityCountryCode string
	ResidentialAddress     string
	ResidentialCity        string
	ResidentialCountryCode string
	ResidentialPostalCode  string
	ResidentialState       string
	CertType               string
	CertCountryCode        string
	CertID                 string
	Portrait               string
	ReverseSide            string
	Status                 enums.CardHolderStatus
	ReviewStatus           enums.CardHolderReviewStatus
	Shared                 bool // 对于一些渠道, 不需要开卡传入持卡人id, 这种的给每一张卡单独分配持卡人, 而不是共用的, 为false, 支持共享的设置为true
}

type WebhookConfig struct {
	BaseModel
	Channel   enums.Channel
	Event     string
	TargetURL string
	Enabled   bool
}

type WebhookRecord struct {
	BaseModel
	WebhookConfigID ID
	Event           string
	Payload         []byte `gorm:"type:jsonb"`
	ResponseBody    string
	StatusCode      int
	DeliveredAt     *time.Time
	ErrorMessage    string
}
