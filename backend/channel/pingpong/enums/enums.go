package enums

import common "generic-mock/enums"

type CardStatus string

const (
	CardActive    CardStatus = "ACTIVE"
	CardFrozen    CardStatus = "REVOKED"
	CardCancelled CardStatus = "CANCELLED"
	CardInactive  CardStatus = "INACTIVE"
)

func FromGenericCardStatus(status common.CardStatus) CardStatus {
	switch status {
	case common.CardStatus_Active:
		return CardActive
	case common.CardStatus_Frozen:
		return CardFrozen
	case common.CardStatus_Deleted:
		return CardCancelled
	default:
		return CardInactive
	}
}

func ToGenericCardStatus(status CardStatus) common.CardStatus {
	switch status {
	case CardActive:
		return common.CardStatus_Active
	case CardFrozen:
		return common.CardStatus_Frozen
	case CardCancelled:
		return common.CardStatus_Deleted
	default:
		return ""
	}
}

type CardAction string

const (
	Freeze       CardAction = "freeze"
	Unfreeze     CardAction = "unfreeze"
	Close        CardAction = "close"
	UpdateRemark CardAction = "update_remark"
)

type FundingAction string

const (
	TopUp    FundingAction = "top_up"
	Withdraw FundingAction = "withdraw"
	Transfer FundingAction = "transfer"
)

type FundingStatus string

const FundingSuccess FundingStatus = "SUCCESS"

type Network string

const Visa Network = "visa"

type CardType string
type NotificationStatus string

const NotificationPendingContract NotificationStatus = "contract_pending"

type AuthorizationStatus string

const Authorized AuthorizationStatus = "authorized"

type Stage string

const (
	Clear   Stage = "clear"
	Reverse Stage = "reverse"
	Refund  Stage = "refund"
)

type TransferKind string

func FromGenericTransferKind(value common.WalletTransferKind) TransferKind {
	switch value {
	case common.WalletTransfer_CardTopUp:
		return "card_top_up"
	case common.WalletTransfer_CardWithdraw:
		return "card_withdraw"
	case common.WalletTransfer_VirtualAccountTopUp:
		return "virtual_account_top_up"
	case common.WalletTransfer_VirtualAccountTransfer:
		return "virtual_account_transfer"
	default:
		return ""
	}
}
