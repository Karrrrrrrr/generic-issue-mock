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
	TransactionType_AuthenticationApproved        TransactionType = "transaction.authentication.approved"
	TransactionType_AuthenticationDeclined        TransactionType = "transaction.authentication.declined"
	TransactionType_AuthenticationSettled         TransactionType = "transaction.authentication.settled"
	TransactionType_AuthenticationReversalPending TransactionType = "transaction.authentication.reversal.pending"
	TransactionType_AuthenticationReversalSettled TransactionType = "transaction.authentication.reversal.settled"
	TransactionType_AuthenticationReversalExpired TransactionType = "transaction.authentication.reversal.expired"
	TransactionType_RefundApproved                TransactionType = "transaction.refund.approved"
	TransactionType_RefundSettled                 TransactionType = "transaction.refund.settled"
	TransactionType_RefundDeclined                TransactionType = "transaction.refund.declined"
	TransactionType_RefundReversal                TransactionType = "transaction.refund.reversal"
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

type WebhookType string

const (
	WebhookTypeCardTransaction WebhookType = "CARD_TRANSACTION"
	WebhookTypeCardStatus      WebhookType = "CARD_STATUS"
)

func (v WebhookType) Valid() bool {
	return v == WebhookTypeCardTransaction || v == WebhookTypeCardStatus
}

func WebhookTypes() []WebhookType {
	return []WebhookType{
		WebhookTypeCardTransaction,
		WebhookTypeCardStatus,
	}
}

func ConvertGenericTransactionTypeToTransactionType(
	value common.CardTransactionType,
) TransactionType {
	switch value {
	case common.CardTransactionType_AUTH:
		return TransactionType_AuthenticationApproved
	case common.CardTransactionType_CLEAR:
		return TransactionType_AuthenticationSettled
	case common.CardTransactionType_VOID:
		return TransactionType_AuthenticationReversalSettled
	case common.CardTransactionType_REFUND:
		return TransactionType_RefundSettled
	default:
		return TransactionType_AuthenticationApproved
	}
}

func ConvertGenericCardTransactionToTransactionType(
	value common.CardTransactionType,
	status common.CardTransactionStatus,
) TransactionType {
	switch value {
	case common.CardTransactionType_AUTH:
		if status == common.TransactionStatus_FAILED {
			return TransactionType_AuthenticationDeclined
		}
		return TransactionType_AuthenticationApproved
	case common.CardTransactionType_CLEAR:
		return TransactionType_AuthenticationSettled
	case common.CardTransactionType_VOID:
		switch status {
		case common.TransactionStatus_PENDING:
			return TransactionType_AuthenticationReversalPending
		case common.TransactionStatus_FAILED:
			return TransactionType_AuthenticationReversalExpired
		default:
			return TransactionType_AuthenticationReversalSettled
		}
	case common.CardTransactionType_REFUND:
		switch status {
		case common.TransactionStatus_PENDING, common.TransactionStatus_AUTHORIZED:
			return TransactionType_RefundApproved
		case common.TransactionStatus_FAILED:
			return TransactionType_RefundDeclined
		case common.TransactionStatus_VOID:
			return TransactionType_RefundReversal
		default:
			return TransactionType_RefundSettled
		}
	default:
		return TransactionType_AuthenticationApproved
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
