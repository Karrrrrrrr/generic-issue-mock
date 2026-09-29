package enums

import (
	"strings"

	generic "generic-mock/enums"
)

type ResponseCode string

const (
	ResponseCode_Success       ResponseCode = "0000"
	ResponseCode_BadInput      ResponseCode = "4000"
	ResponseCode_NotFound      ResponseCode = "VCC1039"
	ResponseCode_InternalError ResponseCode = "5000"
)

type AccountType string

const (
	AccountType_Available AccountType = "FT10001"
)

type CardType string

const (
	CardType_Share    CardType = "share"
	CardType_Recharge CardType = "recharge"
)

type CardFormFactor string

const (
	CardFormFactor_Virtual  CardFormFactor = "virtual_card"
	CardFormFactor_Physical CardFormFactor = "physical_card"
)

type CardStatus string

const (
	CardStatus_Normal    CardStatus = "normal"
	CardStatus_Freezing  CardStatus = "freezing"
	CardStatus_Frozen    CardStatus = "frozen"
	CardStatus_Cancelled CardStatus = "cancelled"
)

type CardHolderStatus string

const CardHolderStatus_Normal CardHolderStatus = "normal"

type CardHolderReviewStatus string

const CardHolderReviewStatus_Approved CardHolderReviewStatus = "approved"

type AuthorizationStatus string

const AuthorizationStatus_Authorized AuthorizationStatus = "authorized"

type WebhookNotificationCategory string

const (
	WebhookNotificationCategoryIssuing           WebhookNotificationCategory = "issuing"
	WebhookNotificationCategoryIssuingSettlement WebhookNotificationCategory = "issuing_settlement"
	WebhookNotificationCategoryIssuingCard       WebhookNotificationCategory = "issuing_card"
	WebhookSuccessCode                                                       = "0"
	WebhookSuccessMessage                                                    = "success"
	WebhookBalanceAccountCard                                                = "card"
)

type WebhookEvent string

const (
	WebhookEventAuth                   WebhookEvent = "auth"
	WebhookEventVerification           WebhookEvent = "verification"
	WebhookEventVoid                   WebhookEvent = "void"
	WebhookEventRefund                 WebhookEvent = "refund"
	WebhookEventCardStatusUpdate       WebhookEvent = "card_status_update"
	WebhookEventCardholderStatusUpdate WebhookEvent = "cardholder_status_update"
)

func ConvertGenericTransactionTypeToWebhookEvent(value generic.CardTransactionType) WebhookEvent {
	switch value {
	case generic.CardTransactionType_VERIFICATION:
		return WebhookEventVerification
	case generic.CardTransactionType_VOID:
		return WebhookEventVoid
	case generic.CardTransactionType_REFUND:
		return WebhookEventRefund
	default:
		return WebhookEventAuth
	}
}

func (v WebhookEvent) Valid() bool {
	switch v {
	case WebhookEventAuth,
		WebhookEventVerification,
		WebhookEventVoid,
		WebhookEventRefund,
		WebhookEventCardStatusUpdate:
		return true
	default:
		return false
	}
}

func WebhookEvents() []WebhookEvent {
	return []WebhookEvent{
		WebhookEventAuth,
		WebhookEventVerification,
		WebhookEventVoid,
		WebhookEventRefund,
		WebhookEventCardStatusUpdate,
	}
}

type TransactionStatus string

const (
	TransactionStatus_Pending    TransactionStatus = "pending"
	TransactionStatus_Authorized TransactionStatus = "authorized"
	TransactionStatus_Succeed    TransactionStatus = "succeed"
	TransactionStatus_Failed     TransactionStatus = "failed"
	TransactionStatus_Void       TransactionStatus = "void"
)

type TransactionType string

const (
	TransactionType_Auth   TransactionType = "auth"
	TransactionType_Clear  TransactionType = "clear"
	TransactionType_Void   TransactionType = "void"
	TransactionType_Refund TransactionType = "refund"
)

type FreezeStatus string

const (
	FreezeStatus_Freeze   FreezeStatus = "freeze"
	FreezeStatus_Unfreeze FreezeStatus = "unfreeze"
)

type RequestResultType string

const (
	RequestResultType_ApplyCard  RequestResultType = "apply_card"
	RequestResultType_CardUpdate RequestResultType = "card_update"
	RequestResultType_CardFreeze RequestResultType = "card_freeze"
)

type OperationStatus string

const (
	OperationStatus_Succeed OperationStatus = "succeed"
)

const (
	MemberID                        = "photonpay-mock-member"
	AccountNumber                   = "photonpay-mock-account"
	CardScheme                      = generic.CardScheme_MasterCard
	DefaultMobilePrefix             = "+1"
	DefaultNationalityCountryCode   = "US"
	RemainingAvailableCardUnlimited = "Unlimited"
	CardNumberMask                  = "******"
)

func ConvertGenericCardTypeToCardType(value generic.CardType) CardType {
	if value == generic.CardType_Share {
		return CardType_Share
	}

	return CardType_Recharge
}

func ConvertCardTypeToGenericCardType(value CardType) generic.CardType {
	if value == CardType_Share {
		return generic.CardType_Share
	}

	return generic.CardType_Single
}

func ConvertGenericCardFormTypeToCardFormFactor(value generic.CardFormType) CardFormFactor {
	if value == generic.CardFormType_Physical {
		return CardFormFactor_Physical
	}

	return CardFormFactor_Virtual
}

func ConvertCardFormFactorToGenericCardFormType(value CardFormFactor) generic.CardFormType {
	if value == CardFormFactor_Physical {
		return generic.CardFormType_Physical
	}

	return generic.CardFormType_Virtual
}

func ConvertGenericCardStatusToCardStatus(value generic.CardStatus) CardStatus {
	switch value {
	case generic.CardStatus_Freezing:
		return CardStatus_Freezing
	case generic.CardStatus_Frozen:
		return CardStatus_Frozen
	case generic.CardStatus_Deleted:
		return CardStatus_Cancelled
	default:
		return CardStatus_Normal
	}
}

func ConvertCardStatusToGenericCardStatus(value CardStatus) generic.CardStatus {
	switch value {
	case CardStatus_Freezing:
		return generic.CardStatus_Freezing
	case CardStatus_Frozen:
		return generic.CardStatus_Frozen
	case CardStatus_Cancelled:
		return generic.CardStatus_Deleted
	default:
		return generic.CardStatus_Active
	}
}

func ConvertFreezeStatusToGenericCardStatus(value FreezeStatus) generic.CardStatus {
	if value == FreezeStatus_Unfreeze {
		return generic.CardStatus_Active
	}

	return generic.CardStatus_Frozen
}

func ConvertGenericCardHolderStatusToCardHolderStatus(_ generic.CardHolderStatus) CardHolderStatus {
	return CardHolderStatus_Normal
}

func ConvertGenericCardHolderReviewStatusToCardHolderReviewStatus(_ generic.CardHolderReviewStatus) CardHolderReviewStatus {
	return CardHolderReviewStatus_Approved
}

func ConvertGenericTransactionStatusToTransactionStatus(value generic.CardTransactionStatus) TransactionStatus {
	switch value {
	case generic.TransactionStatus_PENDING:
		return TransactionStatus_Pending
	case generic.TransactionStatus_AUTHORIZED:
		return TransactionStatus_Authorized
	case generic.TransactionStatus_FAILED:
		return TransactionStatus_Failed
	case generic.TransactionStatus_VOID:
		return TransactionStatus_Void
	default:
		return TransactionStatus_Succeed
	}
}

func ConvertTransactionStatusToGenericTransactionStatus(value TransactionStatus) (generic.CardTransactionStatus, bool) {
	switch value {
	case TransactionStatus_Pending:
		return generic.TransactionStatus_PENDING, true
	case TransactionStatus_Authorized:
		return generic.TransactionStatus_AUTHORIZED, true
	case TransactionStatus_Succeed:
		return generic.TransactionStatus_SUCCEED, true
	case TransactionStatus_Failed:
		return generic.TransactionStatus_FAILED, true
	case TransactionStatus_Void:
		return generic.TransactionStatus_VOID, true
	default:
		return "", false
	}
}

func ConvertStringToTransactionStatus(value string) (TransactionStatus, bool) {
	normalized := strings.TrimSpace(value)
	switch normalized {
	case string(TransactionStatus_Pending):
		return TransactionStatus_Pending, true
	case string(TransactionStatus_Authorized):
		return TransactionStatus_Authorized, true
	case string(TransactionStatus_Succeed):
		return TransactionStatus_Succeed, true
	case string(TransactionStatus_Failed):
		return TransactionStatus_Failed, true
	case string(TransactionStatus_Void):
		return TransactionStatus_Void, true
	default:
		return "", false
	}
}

func ConvertStringToGenericTransactionStatuses(value string) ([]generic.CardTransactionStatus, bool) {
	parts := strings.Split(value, ",")
	result := make([]generic.CardTransactionStatus, 0, len(parts))
	for _, item := range parts {
		status, valid := ConvertStringToTransactionStatus(item)
		if !valid {
			return nil, false
		}
		genericStatus, valid := ConvertTransactionStatusToGenericTransactionStatus(status)
		if !valid {
			return nil, false
		}
		result = append(result, genericStatus)
	}
	return result, true
}

func ConvertGenericTransactionTypeToTransactionType(value generic.CardTransactionType) TransactionType {
	switch value {
	case generic.CardTransactionType_CLEAR:
		return TransactionType_Clear
	case generic.CardTransactionType_VOID:
		return TransactionType_Void
	case generic.CardTransactionType_REFUND:
		return TransactionType_Refund
	default:
		return TransactionType_Auth
	}
}

func ConvertTransactionTypeToGenericTransactionType(value TransactionType) (generic.CardTransactionType, bool) {
	switch value {
	case TransactionType_Auth:
		return generic.CardTransactionType_AUTH, true
	case TransactionType_Clear:
		return generic.CardTransactionType_CLEAR, true
	case TransactionType_Void:
		return generic.CardTransactionType_VOID, true
	case TransactionType_Refund:
		return generic.CardTransactionType_REFUND, true
	default:
		return "", false
	}
}

func ConvertStringToTransactionType(value string) (TransactionType, bool) {
	normalized := strings.TrimSpace(value)
	switch normalized {
	case string(TransactionType_Auth):
		return TransactionType_Auth, true
	case string(TransactionType_Clear):
		return TransactionType_Clear, true
	case string(TransactionType_Void):
		return TransactionType_Void, true
	case string(TransactionType_Refund):
		return TransactionType_Refund, true
	default:
		return "", false
	}
}

func ConvertStringToGenericTransactionTypes(value string) ([]generic.CardTransactionType, bool) {
	parts := strings.Split(value, ",")
	result := make([]generic.CardTransactionType, 0, len(parts))
	for _, item := range parts {
		transactionType, valid := ConvertStringToTransactionType(item)
		if !valid {
			return nil, false
		}
		genericType, valid := ConvertTransactionTypeToGenericTransactionType(transactionType)
		if !valid {
			return nil, false
		}
		result = append(result, genericType)
	}
	return result, true
}

type WalletKind string

const (
	WalletKindAccount        WalletKind = "account"
	WalletKindVirtualAccount WalletKind = "virtual_account"
	WalletKindCard           WalletKind = "card"
)
