package enums

import common "generic-mock/enums"

const (
	AuthorizationStatusPending  AuthorizationStatus = "pending"
	AuthorizationStatusDeclined AuthorizationStatus = "declined"
	AuthorizationStatusVoided   AuthorizationStatus = "void"
	AuthorizationStatusUnknown  AuthorizationStatus = "unknown"
)

func ConvertGenericTransactionStatusToAuthorizationStatus(value common.CardTransactionStatus) AuthorizationStatus {
	switch value {
	case common.TransactionStatus_PENDING:
		return AuthorizationStatusPending
	case common.TransactionStatus_AUTHORIZED, common.TransactionStatus_SUCCEED:
		return AuthorizationStatus_Authorized
	case common.TransactionStatus_FAILED:
		return AuthorizationStatusDeclined
	case common.TransactionStatus_VOID:
		return AuthorizationStatusVoided
	default:
		return AuthorizationStatusUnknown
	}
}
