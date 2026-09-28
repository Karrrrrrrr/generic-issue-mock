package errors

import kratoserrors "github.com/go-kratos/kratos/v2/errors"

var (
	ErrInvalidIssueRequest       = kratoserrors.BadRequest("INVALID_ISSUE_REQUEST", "invalid card issuance request")
	ErrInvalidCardHolder         = kratoserrors.BadRequest("INVALID_CARD_HOLDER", "invalid card holder selection or details")
	ErrInvalidCardProduct        = kratoserrors.BadRequest("INVALID_CARD_PRODUCT", "invalid card product or issuing sequence")
	ErrInvalidWallet             = kratoserrors.BadRequest("INVALID_CARD_WALLET", "invalid card wallet scope, type or currency")
	ErrAccountNotFound           = kratoserrors.NotFound("ACCOUNT_NOT_FOUND", "account not found in channel")
	ErrCardHolderNotFound        = kratoserrors.NotFound("CARD_HOLDER_NOT_FOUND", "card holder not found in account and channel")
	ErrCardProductNotFound       = kratoserrors.NotFound("CARD_PRODUCT_NOT_FOUND", "card product not found in channel")
	ErrVirtualAccountNotFound    = kratoserrors.NotFound("VIRTUAL_ACCOUNT_NOT_FOUND", "virtual account not found in account and channel")
	ErrDatabaseOperation         = kratoserrors.InternalServer("DATABASE_OPERATION_FAILED", "database operation failed")
	ErrNestedTransaction         = kratoserrors.BadRequest("NESTED_ISSUANCE_TRANSACTION", "card issuance must own its transaction before notifying")
	ErrInvalidSimulationRequest  = kratoserrors.BadRequest("INVALID_SIMULATION_REQUEST", "invalid card transaction simulation request")
	ErrCardNotFound              = kratoserrors.NotFound("CARD_NOT_FOUND", "card not found in account and channel")
	ErrCardNotActive             = kratoserrors.BadRequest("CARD_NOT_ACTIVE", "card must be active for this operation")
	ErrAuthorizationNotFound     = kratoserrors.NotFound("AUTHORIZATION_NOT_FOUND", "authorization not found for card in account and channel")
	ErrInvalidAuthorization      = kratoserrors.BadRequest("INVALID_AUTHORIZATION", "invalid authorization for simulation")
	ErrInsufficientCardBalance   = kratoserrors.BadRequest("INSUFFICIENT_CARD_BALANCE", "insufficient available card wallet balance")
	ErrSimulationRequestConflict = kratoserrors.Conflict("SIMULATION_REQUEST_CONFLICT", "simulation request ID has already been used with different parameters")
)
