package enums

import generic "generic-mock/enums"

type WebhookDeliveryStatus string

const (
	WebhookDeliveryStatusPending   WebhookDeliveryStatus = "pending"
	WebhookDeliveryStatusSucceeded WebhookDeliveryStatus = "succeeded"
	WebhookDeliveryStatusFailed    WebhookDeliveryStatus = "failed"
)

func WebhookDeliveryStatusFromGeneric(value generic.WebhookDeliveryStatus) WebhookDeliveryStatus {
	switch value {
	case generic.WebhookDeliveryStatus_Succeeded:
		return WebhookDeliveryStatusSucceeded
	case generic.WebhookDeliveryStatus_Failed:
		return WebhookDeliveryStatusFailed
	default:
		return WebhookDeliveryStatusPending
	}
}

func WebhookDeliveryStatusToGeneric(value WebhookDeliveryStatus) generic.WebhookDeliveryStatus {
	switch value {
	case WebhookDeliveryStatusSucceeded:
		return generic.WebhookDeliveryStatus_Succeeded
	case WebhookDeliveryStatusFailed:
		return generic.WebhookDeliveryStatus_Failed
	default:
		return generic.WebhookDeliveryStatus_Pending
	}
}
