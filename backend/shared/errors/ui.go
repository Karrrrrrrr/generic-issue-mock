package errors

import kratoserrors "github.com/go-kratos/kratos/v2/errors"

var (
	ErrInvalidUIRequest       = kratoserrors.BadRequest("INVALID_UI_REQUEST", "invalid UI request parameters")
	ErrInvalidUIConfiguration = kratoserrors.BadRequest("INVALID_UI_CONFIGURATION", "invalid shared UI configuration")
	ErrUIUnsupported          = kratoserrors.New(501, "UI_CAPABILITY_UNAVAILABLE", "UI capability is not configured")
	ErrUICardClosed           = kratoserrors.BadRequest("CARD_CLOSED", "closed cards cannot change status")
	ErrUITransferConflict     = kratoserrors.Conflict("TRANSFER_REQUEST_CONFLICT", "request ID has already been used for a different transfer")
	ErrUITransactionNotFound  = kratoserrors.NotFound("TRANSACTION_NOT_FOUND", "transaction not found in scope")
	ErrUIWebhookNotFound      = kratoserrors.NotFound("WEBHOOK_NOT_FOUND", "webhook not found in scope")
)
