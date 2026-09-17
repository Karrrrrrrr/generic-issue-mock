package model

import (
	"generic-mock/enums"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/plugin/soft_delete"
)

type ID = int64

type BaseModel struct {
	ID        ID                    `gorm:"column:id;type:bigint;primaryKey;autoIncrement;not null"`
	CreatedAt time.Time             `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
	UpdatedAt time.Time             `gorm:"column:updated_at;type:timestamptz;not null;autoUpdateTime"`
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;softDelete;not null;default:0"`
}

type Card struct {
	BaseModel
	Channel                enums.Channel         `gorm:"column:channel;type:varchar;not null;default:''"`
	AccountID              ID                    `gorm:"column:account_id;type:bigint;not null;default:0"`
	CardProductID          ID                    `gorm:"column:card_product_id;type:bigint;not null;default:0"`
	CardBin                string                `gorm:"column:card_bin;type:varchar;not null;default:''"`
	CardNumber             string                `gorm:"column:card_number;type:varchar;not null;default:'';uniqueIndex"`
	Cvv                    string                `gorm:"column:cvv;type:varchar;not null;default:''"`
	ExpireAt               time.Time             `gorm:"column:expire_at;type:timestamptz;not null"`
	Status                 enums.CardStatus      `gorm:"column:status;type:varchar;not null;default:''"`
	VirtualAccountID       *ID                   `gorm:"column:virtual_account_id;type:bigint;default:null"`   // nil 表示普通卡，非 nil 表示虚拟账户卡，共享余额。
	WalletID               ID                    `gorm:"column:wallet_id;type:bigint;not null;default:0"`      // 虚拟账户卡指向虚拟账户的钱包 ID，减少一次查询。
	CardHolderID           ID                    `gorm:"column:card_holder_id;type:bigint;not null;default:0"` // 持卡人 ID，允许为空。
	FormType               enums.CardFormType    `gorm:"column:form_type;type:varchar;not null;default:''"`
	RequestID              string                `gorm:"column:request_id;type:varchar;not null;default:'';uniqueIndex"`
	LastOperationRequestID string                `gorm:"column:last_operation_request_id;type:varchar;not null;default:'';uniqueIndex"`
	LastOperationType      enums.OperationType   `gorm:"column:last_operation_type;type:varchar;not null;default:''"`
	LastOperationStatus    enums.OperationStatus `gorm:"column:last_operation_status;type:varchar;not null;default:''"`
	CardCurrency           enums.Currency        `gorm:"column:card_currency;type:varchar;not null;default:''"`
	CardScheme             enums.CardScheme      `gorm:"column:card_scheme;type:varchar;not null;default:''"`
	CardType               enums.CardType        `gorm:"column:card_type;type:varchar;not null;default:''"`
	RawRequest             []byte                `gorm:"column:raw_request;type:jsonb;not null;default:'{}'"`

	// ref
	//CardHolderInline *CardHolder `gorm:"-"` // 对于不需要持卡人的渠道, 直接存json 不做关联
	VirtualAccount *VirtualAccount `gorm:"foreignKey:VirtualAccountID;references:ID;->"`
	Wallet         *Wallet         `gorm:"foreignKey:WalletID;references:ID;->"`
	CardHolder     *CardHolder     `gorm:"foreignKey:CardHolderID;references:ID;->"`
	CardProduct    *CardProduct    `gorm:"foreignKey:CardProductID;references:ID;->"`
	VirtualCard    *VirtualCard    `gorm:"foreignKey:CardID;references:ID;->"`
	PhysicalCard   *PhysicalCard   `gorm:"foreignKey:CardID;references:ID;->"`
}

type CardProduct struct {
	BaseModel
	AccountID      ID            `gorm:"column:account_id;type:bigint;not null;default:0;uniqueIndex:idx_card_products_account_channel_prefix"`
	Channel        enums.Channel `gorm:"column:channel;type:varchar;not null;default:'';uniqueIndex:idx_card_products_account_channel_prefix"`
	Prefix         string        `gorm:"column:prefix;type:varchar;not null;default:'';uniqueIndex:idx_card_products_account_channel_prefix"`
	NextCardNumber int64         `gorm:"column:next_card_number;type:bigint;not null;default:0"`
	IsDefault      bool          `gorm:"column:is_default;type:boolean;not null;default:false"`

	Cards []*Card `gorm:"foreignKey:CardProductID;references:ID;->"`
}

type VirtualCard struct {
	BaseModel
	AccountID ID            `gorm:"column:account_id;type:bigint;not null;default:0"`
	Channel   enums.Channel `gorm:"column:channel;type:varchar;not null;default:''"`
	CardID    ID            `gorm:"column:card_id;type:bigint;not null;default:0"`
}

type PhysicalCard struct {
	BaseModel
	AccountID ID            `gorm:"column:account_id;type:bigint;not null;default:0"`
	Channel   enums.Channel `gorm:"column:channel;type:varchar;not null;default:''"`
	CardID    ID            `gorm:"column:card_id;type:bigint;not null;default:0"`
}

type Wallet struct {
	BaseModel
	AccountID  ID               `gorm:"column:account_id;type:bigint;not null;default:0"`
	Channel    enums.Channel    `gorm:"column:channel;type:varchar;not null;default:''"`
	Amount     decimal.Decimal  `gorm:"column:amount;type:numeric;not null;default:0"`
	PendingIn  decimal.Decimal  `gorm:"column:pending_in;type:numeric;not null;default:0"`
	PendingOut decimal.Decimal  `gorm:"column:pending_out;type:numeric;not null;default:0"`
	In         decimal.Decimal  `gorm:"column:in;type:numeric;not null;default:0"`
	Out        decimal.Decimal  `gorm:"column:out;type:numeric;not null;default:0"`
	Type       enums.WalletType `gorm:"column:type;type:varchar;not null;default:''"`
	Currency   enums.Currency   `gorm:"column:currency;type:varchar;not null;default:''"`
}

type VirtualAccount struct {
	BaseModel
	AccountID ID            `gorm:"column:account_id;type:bigint;not null;default:0"`
	Channel   enums.Channel `gorm:"column:channel;type:varchar;not null;default:''"`
	WalletID  ID            `gorm:"column:wallet_id;type:bigint;not null;default:0"`
	Wallet    *Wallet       `gorm:"foreignKey:WalletID;references:ID;->"`
	Name      string        `gorm:"column:name;type:varchar;not null;default:''"`
}

type Account struct {
	BaseModel
	Channel  enums.Channel `gorm:"column:channel;type:varchar;not null;default:''"`
	Name     string        `gorm:"column:name;type:varchar;not null;default:''"`
	WalletID ID            `gorm:"column:wallet_id;type:bigint;not null;default:0"`
	Wallet   *Wallet       `gorm:"foreignKey:WalletID;references:ID;->"`
}

type CardTransaction struct {
	BaseModel
	AccountID               ID                          `gorm:"column:account_id;type:bigint;not null;default:0"`
	Channel                 enums.Channel               `gorm:"column:channel;type:varchar;not null;default:''"`
	OriginCardTransactionID ID                          `gorm:"column:origin_card_transaction_id;type:bigint;not null;default:0"`
	AuthorizationID         ID                          `gorm:"column:authorization_id;type:bigint;not null;default:0"`
	CardID                  ID                          `gorm:"column:card_id;type:bigint;not null;default:0"`
	Status                  enums.CardTransactionStatus `gorm:"column:status;type:varchar;not null;default:''"`
	Type                    enums.CardTransactionType   `gorm:"column:type;type:varchar;not null;default:''"`
	Currency                enums.Currency              `gorm:"column:currency;type:varchar;not null;default:''"`
	TxAmount                decimal.Decimal             `gorm:"column:tx_amount;type:numeric;not null;default:0"`
	TxCurrency              enums.Currency              `gorm:"column:tx_currency;type:varchar;not null;default:''"`
	RequestID               string                      `gorm:"column:request_id;type:varchar;not null;default:''"`
	MerchantName            string                      `gorm:"column:merchant_name;type:varchar;not null;default:''"`
	MerchantCountry         string                      `gorm:"column:merchant_country;type:varchar;not null;default:''"`
	MerchantMCC             string                      `gorm:"column:merchant_mcc;type:varchar;not null;default:''"`
	AuthorizationCode       string                      `gorm:"column:authorization_code;type:varchar;not null;default:''"`
	RawPayload              []byte                      `gorm:"column:raw_payload;type:jsonb;not null;default:'{}'"`

	Authorization         *Authorization     `gorm:"foreignKey:AuthorizationID;references:ID;->"`
	OriginCardTransaction *CardTransaction   `gorm:"foreignKey:OriginCardTransactionID;references:ID;->"`
	CardTransactions      []*CardTransaction `gorm:"foreignKey:OriginCardTransactionID;references:ID;->"`
}

type Authorization struct {
	BaseModel
	AccountID ID            `gorm:"column:account_id;type:bigint;not null;default:0"`
	Channel   enums.Channel `gorm:"column:channel;type:varchar;not null;default:''"`

	CardID                ID                          `gorm:"column:card_id;type:bigint;not null;default:0"`
	OriginAuthorizationID ID                          `gorm:"column:origin_authorization_id;type:bigint;not null;default:0"`
	Currency              enums.Currency              `gorm:"column:currency;type:varchar;not null;default:''"`
	Amount                decimal.Decimal             `gorm:"column:amount;type:numeric;not null;default:0"`
	MerchantName          string                      `gorm:"column:merchant_name;type:varchar;not null;default:''"`
	MerchantCountry       string                      `gorm:"column:merchant_country;type:varchar;not null;default:''"`
	MerchantMCC           string                      `gorm:"column:merchant_mcc;type:varchar;not null;default:''"`
	AuthorizationCode     string                      `gorm:"column:authorization_code;type:varchar;not null;default:''"`
	Status                enums.CardTransactionStatus `gorm:"column:status;type:varchar;not null;default:''"`
	RawPayload            []byte                      `gorm:"column:raw_payload;type:jsonb;not null;default:'{}'"`

	CardTransactions []*CardTransaction `gorm:"foreignKey:AuthorizationID;references:ID;->"`
}

type CardHolder struct {
	BaseModel
	AccountID              ID                           `gorm:"column:account_id;type:bigint;not null;default:0"`
	Channel                enums.Channel                `gorm:"column:channel;type:varchar;not null;default:''"`
	FirstName              string                       `gorm:"column:first_name;type:varchar;not null;default:''"`
	LastName               string                       `gorm:"column:last_name;type:varchar;not null;default:''"`
	Email                  string                       `gorm:"column:email;type:varchar;not null;default:''"`
	Mobile                 string                       `gorm:"column:mobile;type:varchar;not null;default:''"`
	MobilePrefix           string                       `gorm:"column:mobile_prefix;type:varchar;not null;default:''"`
	DateOfBirth            *time.Time                   `gorm:"column:date_of_birth;type:timestamptz;default:null"`
	NationalityCountryCode string                       `gorm:"column:nationality_country_code;type:varchar;not null;default:''"`
	ResidentialAddress     string                       `gorm:"column:residential_address;type:varchar;not null;default:''"`
	ResidentialCity        string                       `gorm:"column:residential_city;type:varchar;not null;default:''"`
	ResidentialCountryCode string                       `gorm:"column:residential_country_code;type:varchar;not null;default:''"`
	ResidentialPostalCode  string                       `gorm:"column:residential_postal_code;type:varchar;not null;default:''"`
	ResidentialState       string                       `gorm:"column:residential_state;type:varchar;not null;default:''"`
	CertType               string                       `gorm:"column:cert_type;type:varchar;not null;default:''"`
	CertCountryCode        string                       `gorm:"column:cert_country_code;type:varchar;not null;default:''"`
	CertID                 string                       `gorm:"column:cert_id;type:varchar;not null;default:''"`
	Portrait               string                       `gorm:"column:portrait;type:varchar;not null;default:''"`
	ReverseSide            string                       `gorm:"column:reverse_side;type:varchar;not null;default:''"`
	Status                 enums.CardHolderStatus       `gorm:"column:status;type:varchar;not null;default:''"`
	ReviewStatus           enums.CardHolderReviewStatus `gorm:"column:review_status;type:varchar;not null;default:''"`
	Shared                 bool                         `gorm:"column:shared;type:boolean;not null;default:false"` // 对于一些渠道, 不需要开卡传入持卡人id, 这种的给每一张卡单独分配持卡人, 而不是共用的, 为false, 支持共享的设置为true
}

type WebhookConfig struct {
	BaseModel
	Channel   enums.Channel `gorm:"column:channel;type:varchar;not null;default:''"`
	AccountID ID            `gorm:"column:account_id;type:bigint;not null;default:0"`
	Event     string        `gorm:"column:event;type:varchar;not null;default:''"`
	TargetURL string        `gorm:"column:target_url;type:varchar;not null;default:''"`
	Enabled   bool          `gorm:"column:enabled;type:boolean;not null;default:false"`
}

// AuthorizationConfig configures the synchronous authorization callback for
// one channel account. It is intentionally separate from asynchronous event
// webhook subscriptions.
type AuthorizationConfig struct {
	BaseModel
	AccountID     ID            `gorm:"column:account_id;type:bigint;not null;default:0;uniqueIndex:idx_authorization_configs_account_channel"`
	Channel       enums.Channel `gorm:"column:channel;type:varchar;not null;default:'';uniqueIndex:idx_authorization_configs_account_channel"`
	TargetURL     string        `gorm:"column:target_url;type:varchar;not null;default:''"`
	Enabled       bool          `gorm:"column:enabled;type:boolean;not null;default:false"`
	TimeoutMillis int           `gorm:"column:timeout_millis;type:integer;not null;default:0"`
}

type WebhookRecord struct {
	BaseModel
	WebhookConfigID ID                          `gorm:"column:webhook_config_id;type:bigint;not null;default:0"`
	AccountID       ID                          `gorm:"column:account_id;type:bigint;not null;default:0"`
	Channel         enums.Channel               `gorm:"column:channel;type:varchar;not null;default:''"`
	Event           string                      `gorm:"column:event;type:varchar;not null;default:''"`
	TargetURL       string                      `gorm:"column:target_url;type:varchar;not null;default:''"`
	SourceID        string                      `gorm:"column:source_id;type:varchar;not null;default:''"`
	Payload         []byte                      `gorm:"column:payload;type:jsonb;not null;default:'{}'"`
	RequestHeaders  []byte                      `gorm:"column:request_headers;type:jsonb;not null;default:'{}'"`
	ResponseBody    string                      `gorm:"column:response_body;type:varchar;not null;default:''"`
	ResponseHeaders []byte                      `gorm:"column:response_headers;type:jsonb;not null;default:'{}'"`
	StatusCode      int                         `gorm:"column:status_code;type:integer;not null;default:0"`
	Status          enums.WebhookDeliveryStatus `gorm:"column:status;type:varchar;not null;default:''"`
	AttemptCount    int                         `gorm:"column:attempt_count;type:integer;not null;default:0"`
	DeliveredAt     *time.Time                  `gorm:"column:delivered_at;type:timestamptz;default:null"`
	ErrorMessage    string                      `gorm:"column:error_message;type:varchar;not null;default:''"`
}
