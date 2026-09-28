package enums

type WebhookEvent string

const (
	WebhookOpenCard      WebhookEvent = "issuing.open_card"
	WebhookCardOperate   WebhookEvent = "issuing.notify_card_operate"
	WebhookAuthorization WebhookEvent = "issuing.notify_auth_transaction"
	WebhookClearing      WebhookEvent = "issuing.notify_clear_transaction"
	WebhookTransfer      WebhookEvent = "issuing.notify_transfer"
)

func WebhookEvents() []WebhookEvent {
	return []WebhookEvent{WebhookOpenCard, WebhookCardOperate, WebhookAuthorization, WebhookClearing, WebhookTransfer}
}

func (event WebhookEvent) Valid() bool {
	switch event {
	case WebhookOpenCard, WebhookCardOperate, WebhookAuthorization, WebhookClearing, WebhookTransfer:
		return true
	default:
		return false
	}
}

// TODO: 确认文档枚举表与示例的大小写差异，暂按枚举表使用大写。
type WebhookAuthorizationType string

const (
	WebhookAuth     WebhookAuthorizationType = "AUTH"
	WebhookReversal WebhookAuthorizationType = "REVERSAL"
)

type WebhookAuthorizationStatus string

const (
	WebhookApproved WebhookAuthorizationStatus = "APPROVED"
	WebhookDeclined WebhookAuthorizationStatus = "DECLINED"
)

type WebhookClearType string

const (
	WebhookDebit  WebhookClearType = "DEBIT"
	WebhookCredit WebhookClearType = "CREDIT"
)

type WebhookClearStatus string

const (
	WebhookSettled WebhookClearStatus = "SETTLED"
)

type WebhookOperateType string

const (
	WebhookCardIn WebhookOperateType = "CARD_IN"
)

type WebhookTransferType string

const (
	WebhookTransferCard   WebhookTransferType = "CARD"
	WebhookTransferBudget WebhookTransferType = "BUDGET"
)

type WebhookServiceType string

const (
	WebhookServiceAirlines   WebhookServiceType = "AIRLINES"
	WebhookServiceCarRentals WebhookServiceType = "CAR_RENTALS"
	WebhookServiceHotel      WebhookServiceType = "HOTEL"
	WebhookServiceTravel     WebhookServiceType = "TRAVEL_GENERAL"
)
