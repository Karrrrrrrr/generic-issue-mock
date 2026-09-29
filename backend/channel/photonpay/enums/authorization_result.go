package enums

type AuthorizationResultCode string

const (
	AuthorizationResultCodePass            AuthorizationResultCode = "0000"
	AuthorizationResultCodeInvalidMerchant AuthorizationResultCode = "4001"
	AuthorizationResultCodeInvalidAmount   AuthorizationResultCode = "4002"
	AuthorizationResultCodeInsufficient    AuthorizationResultCode = "4003"
	AuthorizationResultCodeRiskControl     AuthorizationResultCode = "4004"
	AuthorizationResultCodeLimitExceeded   AuthorizationResultCode = "4005"
)

func ConvertAuthorizationResultCodeToMessage(value string) string {
	switch AuthorizationResultCode(value) {
	case AuthorizationResultCodePass:
		return "third-party authorization approved"
	case AuthorizationResultCodeInvalidMerchant:
		return "third-party authorization declined: invalid merchant"
	case AuthorizationResultCodeInvalidAmount:
		return "third-party authorization declined: invalid amount"
	case AuthorizationResultCodeInsufficient:
		return "third-party authorization declined: insufficient balance"
	case AuthorizationResultCodeRiskControl:
		return "third-party authorization declined: risk control"
	case AuthorizationResultCodeLimitExceeded:
		return "third-party authorization declined: limit exceeded"
	default:
		return "third-party authorization declined"
	}
}
