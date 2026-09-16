package model

import (
	"generic-mock/enums"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/plugin/soft_delete"
)

// 根据渠道选择是int还是string(uuid)
type ID int64

type BaseModel struct {
	ID        ID
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt
}

type Card struct {
	BaseModel
	CardBin          string
	CardNumber       string
	Cvv              string
	Status           enums.CardStatus
	VirtualAccountID ID // 如果为0, 表示普通卡, 不为0 表示虚拟账户卡 共享余额
	BalanceID        ID // 如果为是虚拟账户, 指向虚拟账户的钱包id, 减少一次查询
	CardHolderID     ID //  持卡人id 可能为0
	FormType         enums.CardFormType

	// ref
	//CardHolderInline *CardHolder `gorm:"-"` // 对于不需要持卡人的渠道, 直接存json 不做关联
	VirtualAccount *VirtualAccount
	Wallet         *Wallet
	CardHolder     *CardHolder
	VirtualCard    *VirtualCard
	PhysicalCard   *PhysicalCard
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
}

type VirtualAccount struct {
	BaseModel
	WalletID ID
	Wallet   *Wallet
}

type Account struct {
	BaseModel
}

//type Transaction struct {
//
//}

type CardTransaction struct {
	BaseModel
	OriginCardTransactionID ID
	AuthorizationID         ID
	CardID                  ID
	Status                  enums.CardTransactionStatus
	Type                    enums.CardTransactionType
	Currency                enums.Currency
	TxAmount                decimal.Decimal

	Authorization         *Authorization
	OriginCardTransaction *CardTransaction
	CardTransactions      []*CardTransaction
}

type Authorization struct {
	BaseModel

	CardTransactions []*CardTransaction
}

type CardHolder struct {
	BaseModel
	Name   string
	Shared bool // 对于一些渠道, 不需要开卡传入持卡人id, 这种的给每一张卡单独分配持卡人, 而不是共用的, 为false, 支持共享的设置为true
}

type WebhookConfig struct {
	BaseModel
}

type WebhookRecord struct {
	BaseModel
}
