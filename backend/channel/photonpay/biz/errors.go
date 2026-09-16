package biz

import kratosErrors "github.com/go-kratos/kratos/v2/errors"

const (
	errorReasonDatabaseOperation  = "DATABASE_OPERATION_FAILED"
	errorMessageDatabaseOperation = "database operation failed"
	errorReasonResourceNotFound   = "RESOURCE_NOT_FOUND"
	errorMessageResourceNotFound  = "resource not found"
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
)
