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
	ErrCardClosed            = kratosErrors.BadRequest("CARD_CLOSED", "已注销或注销中的卡片不能恢复或冻结")
	ErrRequestResultNotFound = kratosErrors.NotFound("VCC1039", "invalid requestId")
	ErrDatabaseOperation     = kratosErrors.InternalServer(
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
