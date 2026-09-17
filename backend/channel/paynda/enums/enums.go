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

func TransferTypeFromString(value string) (TransferType, bool) {
	switch TransferType(value) {
	case TransferType_In:
		return TransferType_In, true
	case TransferType_Out:
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

type WebhookEvent string

const (
	WebhookEventCardTransaction WebhookEvent = "CARD_TRANSACTION"
	WebhookEventCardStatus      WebhookEvent = "CARD_STATUS"
)

func (v WebhookEvent) Valid() bool {
	return v == WebhookEventCardTransaction
}

func WebhookEvents() []WebhookEvent {
	return []WebhookEvent{WebhookEventCardTransaction}
}

func TransactionTypeFromGeneric(value common.CardTransactionType) TransactionType {
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
	DefaultCardBin       = "523456"
	SuccessCode    int64 = 200
	SuccessMessage       = "success"
)

func CardStatusFromGeneric(value common.CardStatus) CardStatus {
	switch value {
	case common.CardStatus_Frozen:
		return CardStatus_Frozen
	case common.CardStatus_Deleted:
		return CardStatus_Deleted
	default:
		return CardStatus_Active
	}
}

func CardStatusToGeneric(value CardStatus) common.CardStatus {
	if value == CardStatus_Frozen {
		return common.CardStatus_Frozen
	}
	if value == CardStatus_Deleted {
		return common.CardStatus_Deleted
	}

	return common.CardStatus_Active
}
