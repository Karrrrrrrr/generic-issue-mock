package enums

type AccountStatus string
type AccountType string
type BalanceType string
type AuthorizationWebhookStatus string

const (
	AccountStatusOpen            AccountStatus              = "open"
	AccountTypeDebit             AccountType                = "debit"
	BalanceTypeCash              BalanceType                = "cash"
	AuthorizationWebhookDisabled AuthorizationWebhookStatus = "disabled"
)
