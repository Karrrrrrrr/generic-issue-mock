package model

import (
	"generic-mock/enums"

	"github.com/shopspring/decimal"
)

type WalletTransfer struct {
	BaseModel
	AccountID      ID                       `gorm:"column:account_id;type:bigint;not null;default:0;uniqueIndex:idx_wallet_transfers_request"`
	Channel        enums.Channel            `gorm:"column:channel;type:varchar;not null;default:'';uniqueIndex:idx_wallet_transfers_request"`
	RequestID      string                   `gorm:"column:request_id;type:varchar;not null;default:'';uniqueIndex:idx_wallet_transfers_request,where:request_id <> ''"`
	Kind           enums.WalletTransferKind `gorm:"column:kind;type:varchar;not null;default:''"`
	CardID         *ID                      `gorm:"column:card_id;type:bigint;default:null"`
	SourceWalletID ID                       `gorm:"column:source_wallet_id;type:bigint;not null;default:0"`
	TargetWalletID ID                       `gorm:"column:target_wallet_id;type:bigint;not null;default:0"`
	Currency       enums.Currency           `gorm:"column:currency;type:varchar;not null;default:''"`
	Amount         decimal.Decimal          `gorm:"column:amount;type:numeric;not null;default:0"`
	SourceBefore   decimal.Decimal          `gorm:"column:source_before;type:numeric;not null;default:0"`
	SourceAfter    decimal.Decimal          `gorm:"column:source_after;type:numeric;not null;default:0"`
	TargetBefore   decimal.Decimal          `gorm:"column:target_before;type:numeric;not null;default:0"`
	TargetAfter    decimal.Decimal          `gorm:"column:target_after;type:numeric;not null;default:0"`
	Account        *Account                 `gorm:"foreignKey:AccountID,Channel;references:ID,Channel;->"`
	Card           *Card                    `gorm:"foreignKey:CardID;references:ID;->"`
}
