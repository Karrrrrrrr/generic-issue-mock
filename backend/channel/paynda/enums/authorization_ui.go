package enums

import common "generic-mock/enums"

type AuthorizationStatus string

const (
	AuthorizationStatusPending    AuthorizationStatus = "pending"
	AuthorizationStatusAuthorized AuthorizationStatus = "authorized"
	AuthorizationStatusDeclined   AuthorizationStatus = "declined"
	AuthorizationStatusVoided     AuthorizationStatus = "void"
	AuthorizationStatusUnknown    AuthorizationStatus = "unknown"
)

func AuthorizationStatusFromGeneric(value common.CardTransactionStatus) AuthorizationStatus {
	switch value {
	case common.TransactionStatus_PENDING:
		return AuthorizationStatusPending
	case common.TransactionStatus_AUTHORIZED, common.TransactionStatus_SUCCEED:
		return AuthorizationStatusAuthorized
	case common.TransactionStatus_FAILED:
		return AuthorizationStatusDeclined
	case common.TransactionStatus_VOID:
		return AuthorizationStatusVoided
	default:
		return AuthorizationStatusUnknown
	}
}
