package biz

import kratosErrors "github.com/go-kratos/kratos/v2/errors"

const (
	errorReasonDatabaseOperation  = "PAYNDA_DATABASE_OPERATION_FAILED"
	errorMessageDatabaseOperation = "database operation failed"
	errorReasonResourceNotFound   = "PAYNDA_RESOURCE_NOT_FOUND"
	errorMessageResourceNotFound  = "resource not found"
	errorReasonInvalidOperation   = "PAYNDA_INVALID_OPERATION"
	errorMessageInvalidOperation  = "invalid operation"
)

var (
	ErrDatabaseOperation = kratosErrors.InternalServer(
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
)
