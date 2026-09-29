package enums

import (
	"slices"
	"strings"

	common "generic-mock/enums"
)

type CardStatus string

const (
	CardActive    CardStatus = "ACTIVE"
	CardFrozen    CardStatus = "REVOKED"
	CardCancelled CardStatus = "CANCELLED"
	CardInactive  CardStatus = "INACTIVE"
)

func ConvertGenericCardStatusToCardStatus(status common.CardStatus) CardStatus {
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

func ConvertCardStatusToGenericCardStatus(status CardStatus) common.CardStatus {
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

const (
	TransferKindCardTopUp              TransferKind = "card_top_up"
	TransferKindCardWithdraw           TransferKind = "card_withdraw"
	TransferKindVirtualAccountTopUp    TransferKind = "virtual_account_top_up"
	TransferKindVirtualAccountTransfer TransferKind = "virtual_account_transfer"
)

type AccountType string

const (
	AccountTypeBudget         AccountType = "budget"
	AccountTypeVirtualAccount AccountType = "virtual_account"
)

func ConvertStringToAccountType(value string) (AccountType, bool) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == string(AccountTypeBudget) {
		return AccountTypeBudget, true
	}
	if normalized == string(AccountTypeVirtualAccount) {
		return AccountTypeVirtualAccount, true
	}
	return "", false
}

type TransactionDirection string

const (
	TransactionDirectionCredit TransactionDirection = "CREDIT"
	TransactionDirectionDebit  TransactionDirection = "DEBIT"
)

func ConvertStringToTransactionDirection(value string) (TransactionDirection, bool) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if normalized == string(TransactionDirectionCredit) {
		return TransactionDirectionCredit, true
	}
	if normalized == string(TransactionDirectionDebit) {
		return TransactionDirectionDebit, true
	}
	return "", false
}

func ConvertStringToGenericTransactionType(value string) (common.CardTransactionType, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "auth", "authorization":
		return common.CardTransactionType_AUTH, true
	case "clear", "clearing", "settlement":
		return common.CardTransactionType_CLEAR, true
	case "void", "reverse", "reversal":
		return common.CardTransactionType_VOID, true
	case "refund":
		return common.CardTransactionType_REFUND, true
	default:
		return "", false
	}
}

func ConvertStringToGenericTransactionTypes(value string) ([]common.CardTransactionType, bool) {
	parts := strings.Split(value, ",")
	result := make([]common.CardTransactionType, 0, len(parts))
	for _, item := range parts {
		transactionType, valid := ConvertStringToGenericTransactionType(item)
		if !valid {
			return nil, false
		}
		result = append(result, transactionType)
	}
	return result, true
}

func ConvertStringToGenericTransactionStatus(value string) (common.CardTransactionStatus, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "pending":
		return common.TransactionStatus_PENDING, true
	case "authorized":
		return common.TransactionStatus_AUTHORIZED, true
	case "succeed", "success", "completed":
		return common.TransactionStatus_SUCCEED, true
	case "failed", "declined":
		return common.TransactionStatus_FAILED, true
	case "void", "reversed":
		return common.TransactionStatus_VOID, true
	default:
		return "", false
	}
}

func ConvertStringToGenericTransactionStatuses(value string) ([]common.CardTransactionStatus, bool) {
	parts := strings.Split(value, ",")
	result := make([]common.CardTransactionStatus, 0, len(parts))
	for _, item := range parts {
		status, valid := ConvertStringToGenericTransactionStatus(item)
		if !valid {
			return nil, false
		}
		result = append(result, status)
	}
	return result, true
}

func ConvertTransactionDirectionToGenericTransactionTypes(
	transactionTypes []common.CardTransactionType,
	value TransactionDirection,
) ([]common.CardTransactionType, bool) {
	var allowed []common.CardTransactionType
	switch value {
	case TransactionDirectionCredit:
		allowed = []common.CardTransactionType{common.CardTransactionType_REFUND}
	case TransactionDirectionDebit:
		allowed = []common.CardTransactionType{
			common.CardTransactionType_AUTH,
			common.CardTransactionType_CLEAR,
			common.CardTransactionType_VOID,
		}
	default:
		return nil, false
	}
	if len(transactionTypes) == 0 {
		return slices.Clone(allowed), true
	}
	allowedSet := make(map[common.CardTransactionType]struct{}, len(allowed))
	for _, item := range allowed {
		allowedSet[item] = struct{}{}
	}
	result := make([]common.CardTransactionType, 0, len(transactionTypes))
	for _, item := range transactionTypes {
		if _, ok := allowedSet[item]; ok {
			result = append(result, item)
		}
	}
	return result, true
}

func ConvertGenericTransferKindToTransferKind(value common.WalletTransferKind) TransferKind {
	switch value {
	case common.WalletTransfer_CardTopUp:
		return TransferKindCardTopUp
	case common.WalletTransfer_CardWithdraw:
		return TransferKindCardWithdraw
	case common.WalletTransfer_VirtualAccountTopUp:
		return TransferKindVirtualAccountTopUp
	case common.WalletTransfer_VirtualAccountTransfer:
		return TransferKindVirtualAccountTransfer
	default:
		return ""
	}
}
