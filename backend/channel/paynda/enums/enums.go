package enums

import common "generic-mock/enums"

type CardStatus string

const (
	CardStatus_Active  CardStatus = "ACTIVE"
	CardStatus_Frozen  CardStatus = "FROZEN"
	CardStatus_Deleted CardStatus = "DELETED"
)

type TransferType string

const (
	TransferType_In  TransferType = "IN"
	TransferType_Out TransferType = "OUT"
)

func ConvertStringToTransferType(value string) (TransferType, bool) {
	switch value {
	case string(TransferType_In):
		return TransferType_In, true
	case string(TransferType_Out):
		return TransferType_Out, true
	default:
		return "", false
	}
}

type TransactionType string

const (
	TransactionType_Approved        TransactionType = "transaction.authentication.approved"
	TransactionType_Settled         TransactionType = "transaction.authentication.settled"
	TransactionType_ReversalSettled TransactionType = "transaction.authentication.reversal.settled"
	TransactionType_RefundSettled   TransactionType = "transaction.refund.settled"
)

type CardHolderStatus string

const CardHolderStatusNormal CardHolderStatus = "normal"

func ConvertGenericCardHolderStatusToCardHolderStatus(common.CardHolderStatus) CardHolderStatus {
	return CardHolderStatusNormal
}

type TransactionStatus string

const (
	TransactionStatusPending    TransactionStatus = "pending"
	TransactionStatusAuthorized TransactionStatus = "authorized"
	TransactionStatusSucceed    TransactionStatus = "succeed"
	TransactionStatusFailed     TransactionStatus = "failed"
	TransactionStatusVoid       TransactionStatus = "void"
)

func ConvertGenericTransactionStatusToTransactionStatus(value common.CardTransactionStatus) TransactionStatus {
	switch value {
	case common.TransactionStatus_PENDING:
		return TransactionStatusPending
	case common.TransactionStatus_AUTHORIZED:
		return TransactionStatusAuthorized
	case common.TransactionStatus_SUCCEED:
		return TransactionStatusSucceed
	case common.TransactionStatus_FAILED:
		return TransactionStatusFailed
	case common.TransactionStatus_VOID:
		return TransactionStatusVoid
	default:
		return TransactionStatusPending
	}
}

type WebhookEvent string

const (
	WebhookEventCardTransaction WebhookEvent = "CARD_TRANSACTION"
	WebhookEventCardStatus      WebhookEvent = "CARD_STATUS"
)

func (v WebhookEvent) Valid() bool {
	return v == WebhookEventCardTransaction || v == WebhookEventCardStatus
}

func WebhookEvents() []WebhookEvent {
	return []WebhookEvent{
		WebhookEventCardTransaction,
		WebhookEventCardStatus,
	}
}

func ConvertGenericTransactionTypeToTransactionType(value common.CardTransactionType) TransactionType {
	switch value {
	case common.CardTransactionType_CLEAR:
		return TransactionType_Settled
	case common.CardTransactionType_VOID:
		return TransactionType_ReversalSettled
	case common.CardTransactionType_REFUND:
		return TransactionType_RefundSettled
	default:
		return TransactionType_Approved
	}
}

const (
	SuccessCode    int64 = 200
	SuccessMessage       = "success"
)

func ConvertGenericCardStatusToCardStatus(value common.CardStatus) CardStatus {
	switch value {
	case common.CardStatus_Frozen:
		return CardStatus_Frozen
	case common.CardStatus_Deleted:
		return CardStatus_Deleted
	default:
		return CardStatus_Active
	}
}

func ConvertCardStatusToGenericCardStatus(value CardStatus) common.CardStatus {
	if value == CardStatus_Frozen {
		return common.CardStatus_Frozen
	}
	if value == CardStatus_Deleted {
		return common.CardStatus_Deleted
	}

	return common.CardStatus_Active
}

type WalletKind string

const (
	WalletKindAccount WalletKind = "account"
	WalletKindCard    WalletKind = "card"
)
