package errors

import kratoserrors "github.com/go-kratos/kratos/v2/errors"

var (
	ErrInvalidIssueRequest    = kratoserrors.BadRequest("INVALID_ISSUE_REQUEST", "invalid card issuance request")
	ErrInvalidCardHolder      = kratoserrors.BadRequest("INVALID_CARD_HOLDER", "invalid card holder selection or details")
	ErrInvalidCardProduct     = kratoserrors.BadRequest("INVALID_CARD_PRODUCT", "invalid card product or issuing sequence")
	ErrInvalidWallet          = kratoserrors.BadRequest("INVALID_CARD_WALLET", "invalid card wallet scope, type or currency")
	ErrAccountNotFound        = kratoserrors.NotFound("ACCOUNT_NOT_FOUND", "account not found in channel")
	ErrCardHolderNotFound     = kratoserrors.NotFound("CARD_HOLDER_NOT_FOUND", "card holder not found in account and channel")
	ErrCardProductNotFound    = kratoserrors.NotFound("CARD_PRODUCT_NOT_FOUND", "card product not found in channel")
	ErrVirtualAccountNotFound = kratoserrors.NotFound("VIRTUAL_ACCOUNT_NOT_FOUND", "virtual account not found in account and channel")
	ErrDatabaseOperation      = kratoserrors.InternalServer("DATABASE_OPERATION_FAILED", "database operation failed")
	ErrNestedTransaction      = kratoserrors.BadRequest("NESTED_ISSUANCE_TRANSACTION", "card issuance must own its transaction before notifying")
)
