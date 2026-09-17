package enums

import generic "generic-mock/enums"

type CardStatus string

const (
	CardStatus_Active   CardStatus = "active"
	CardStatus_Paused   CardStatus = "paused"
	CardStatus_Inactive CardStatus = "inactive"
	CardStatus_Closed   CardStatus = "closed"
)

type CardType string

const CardType_Virtual CardType = "virtual"

type CardProductStatus string

const (
	CardProductStatus_Active   CardProductStatus = "active"
	CardProductStatus_Inactive CardProductStatus = "inactive"
)

type TransactionStatus string

const (
	TransactionStatus_Pending TransactionStatus = "pending"
	TransactionStatus_Posted  TransactionStatus = "posted"
	TransactionStatus_Failed  TransactionStatus = "failed"
)

type TransactionType string

const (
	TransactionType_Authorization TransactionType = "auth"
	TransactionType_Clearing      TransactionType = "clear"
	TransactionType_Void          TransactionType = "void"
	TransactionType_Refund        TransactionType = "refund"
)

type CardHolderStatus string

const CardHolderStatus_Normal CardHolderStatus = "normal"

type WebhookEvent string

const (
	WebhookEventTransactionCreate WebhookEvent = "aggregated_transaction.create"
	WebhookEventTransactionUpdate WebhookEvent = "aggregated_transaction.update"
	WebhookEventCardCreate        WebhookEvent = "card_creation.event"
	WebhookEventCardUpdate        WebhookEvent = "card.update"
	WebhookEventCardDelete        WebhookEvent = "card.delete"
)

func (v WebhookEvent) Valid() bool {
	switch v {
	case WebhookEventTransactionCreate,
		WebhookEventTransactionUpdate,
		WebhookEventCardCreate,
		WebhookEventCardUpdate,
		WebhookEventCardDelete:
		return true
	default:
		return false
	}
}

func WebhookEvents() []WebhookEvent {
	return []WebhookEvent{
		WebhookEventTransactionCreate,
		WebhookEventTransactionUpdate,
		WebhookEventCardCreate,
		WebhookEventCardUpdate,
		WebhookEventCardDelete,
	}
}

func CardStatusFromGeneric(value generic.CardStatus) CardStatus {
	switch value {
	case generic.CardStatus_Frozen:
		return CardStatus_Paused
	case generic.CardStatus_Deleted:
		return CardStatus_Closed
	case generic.CardStatus_Inactive:
		return CardStatus_Inactive
	default:
		return CardStatus_Active
	}
}

func CardStatusToGeneric(value CardStatus) generic.CardStatus {
	switch value {
	case CardStatus_Paused:
		return generic.CardStatus_Frozen
	case CardStatus_Closed:
		return generic.CardStatus_Deleted
	case CardStatus_Inactive:
		return generic.CardStatus_Inactive
	default:
		return generic.CardStatus_Active
	}
}

func TransactionStatusFromGeneric(value generic.CardTransactionStatus) TransactionStatus {
	if value == generic.TransactionStatus_FAILED {
		return TransactionStatus_Failed
	}
	if value == generic.TransactionStatus_PENDING {
		return TransactionStatus_Pending
	}

	return TransactionStatus_Posted
}

func TransactionStatusToGeneric(value TransactionStatus) generic.CardTransactionStatus {
	switch value {
	case TransactionStatus_Pending:
		return generic.TransactionStatus_PENDING
	case TransactionStatus_Failed:
		return generic.TransactionStatus_FAILED
	default:
		return generic.TransactionStatus_SUCCEED
	}
}

func TransactionTypeFromGeneric(value generic.CardTransactionType) TransactionType {
	switch value {
	case generic.CardTransactionType_CLEAR:
		return TransactionType_Clearing
	case generic.CardTransactionType_VOID:
		return TransactionType_Void
	case generic.CardTransactionType_REFUND:
		return TransactionType_Refund
	default:
		return TransactionType_Authorization
	}
}

func TransactionTypeToGeneric(value TransactionType) generic.CardTransactionType {
	switch value {
	case TransactionType_Clearing:
		return generic.CardTransactionType_CLEAR
	case TransactionType_Void:
		return generic.CardTransactionType_VOID
	case TransactionType_Refund:
		return generic.CardTransactionType_REFUND
	default:
		return generic.CardTransactionType_AUTH
	}
}

func CardHolderStatusFromGeneric(_ generic.CardHolderStatus) CardHolderStatus {
	return CardHolderStatus_Normal
}

type WalletKind string

const (
	WalletKindAccount        WalletKind = "account"
	WalletKindVirtualAccount WalletKind = "virtual_account"
	WalletKindCard           WalletKind = "card"
)
