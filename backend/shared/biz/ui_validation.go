package biz

import (
	"strings"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
	sharederrors "generic-mock/shared/errors"
)

type UIPageRequest struct {
	Offset int
	Limit  *int
}

func (req UIPageRequest) Validate() error {
	if req.Offset < 0 || (req.Limit != nil && *req.Limit <= 0) {
		return sharederrors.ErrInvalidUIRequest
	}
	return nil
}

type UITimeRange struct {
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

func (req UITimeRange) Validate() error {
	if (req.CreatedFrom != nil && req.CreatedFrom.IsZero()) ||
		(req.CreatedTo != nil && req.CreatedTo.IsZero()) ||
		(req.CreatedFrom != nil && req.CreatedTo != nil && req.CreatedFrom.After(*req.CreatedTo)) {
		return sharederrors.ErrInvalidUIRequest
	}
	return nil
}

func validUIIDs(values []*model.ID) bool {
	for _, value := range values {
		if value != nil && *value <= 0 {
			return false
		}
	}
	return true
}

func validUIOptionalText(value *string) bool {
	return value == nil || strings.TrimSpace(*value) != ""
}

func validUICurrency(value enums.Currency) bool {
	if len(value) != 3 {
		return false
	}
	for _, letter := range value {
		if letter < 'A' || letter > 'Z' {
			return false
		}
	}
	return true
}

func validUICardStatus(value enums.CardStatus) bool {
	switch value {
	case enums.CardStatus_Active, enums.CardStatus_Inactive, enums.CardStatus_Freezing,
		enums.CardStatus_Frozen, enums.CardStatus_Deleteing, enums.CardStatus_Deleted:
		return true
	default:
		return false
	}
}

func validUITransactionStatus(value enums.CardTransactionStatus) bool {
	switch value {
	case enums.TransactionStatus_PENDING, enums.TransactionStatus_AUTHORIZED,
		enums.TransactionStatus_SUCCEED, enums.TransactionStatus_FAILED, enums.TransactionStatus_VOID:
		return true
	default:
		return false
	}
}

func validUITransactionType(value enums.CardTransactionType) bool {
	switch value {
	case enums.CardTransactionType_AUTH, enums.CardTransactionType_CLEAR, enums.CardTransactionType_VOID,
		enums.CardTransactionType_REFUND, enums.CardTransactionType_VERIFICATION,
		enums.CardTransactionType_FundIn, enums.CardTransactionType_FundOut:
		return true
	default:
		return false
	}
}
