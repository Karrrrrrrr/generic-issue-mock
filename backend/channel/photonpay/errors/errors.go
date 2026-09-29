package errors

import kratosErrors "github.com/go-kratos/kratos/v2/errors"

const (
	errorReasonDatabaseOperation   = "DATABASE_OPERATION_FAILED"
	errorMessageDatabaseOperation  = "database operation failed"
	errorReasonResourceNotFound    = "RESOURCE_NOT_FOUND"
	errorMessageResourceNotFound   = "resource not found"
	errorReasonInvalidOperation    = "INVALID_OPERATION"
	errorMessageInvalidOperation   = "invalid operation"
	errorReasonInvalidDateOfBirth  = "INVALID_DATE_OF_BIRTH"
	errorMessageInvalidDateOfBirth = "invalid date of birth"
)

var (
	ErrCardClosed                    = kratosErrors.BadRequest("CARD_CLOSED", "已注销或注销中的卡片不能恢复或冻结")
	ErrRequestResultNotFound         = kratosErrors.NotFound("VCC1039", "invalid parameter[invalid requestId]")
	ErrInvalidCardholderID           = kratosErrors.NotFound("VCC1039", "invalid parameter[invalid cardholderId]")
	ErrCardBinNotFound               = kratosErrors.NotFound("PHOTONPAY_CARD_BIN_NOT_FOUND", "photonpay card product not found for cardBin")
	ErrDefaultVirtualAccountNotFound = kratosErrors.NotFound("PHOTONPAY_DEFAULT_VIRTUAL_ACCOUNT_NOT_FOUND", "photonpay default virtual account not found")
	ErrDatabaseOperation             = kratosErrors.InternalServer(
		errorReasonDatabaseOperation,
		errorMessageDatabaseOperation,
	)
	ErrResourceNotFound = kratosErrors.NotFound(
		errorReasonResourceNotFound,
		errorMessageResourceNotFound,
	)
	ErrInvalidOperation = kratosErrors.BadRequest(
		errorReasonInvalidOperation,
		errorMessageInvalidOperation,
	)
	ErrInvalidDateOfBirth = kratosErrors.BadRequest(
		errorReasonInvalidDateOfBirth,
		errorMessageInvalidDateOfBirth,
	)
)

var ErrCardNotActive = kratosErrors.BadRequest("CARD_NOT_ACTIVE", "activate card before adjusting funds")
