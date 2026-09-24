package service

import (
	channelEnums "generic-mock/channel/photonpay/enums"
	"generic-mock/enums"
	"time"
)

type UIListTimeRange struct {
	CreatedFrom *time.Time `form:"created_from" time_format:"2006-01-02T15:04:05Z07:00" time_utc:"1"`
	CreatedTo   *time.Time `form:"created_to" time_format:"2006-01-02T15:04:05Z07:00" time_utc:"1"`
}

func uiCardStatuses(value *channelEnums.CardStatus) []enums.CardStatus {
	if value == nil {
		return nil
	}
	var result []enums.CardStatus
	for _, candidate := range []enums.CardStatus{
		enums.CardStatus_Active,
		enums.CardStatus_Frozen,
		enums.CardStatus_Frozing,
		enums.CardStatus_Deleted,
		enums.CardStatus_Inactive,
	} {
		if channelEnums.CardStatusFromGeneric(candidate) == *value {
			result = append(result, candidate)
		}
	}
	return result
}

func uiTransactionTypes(value *channelEnums.TransactionType) []enums.CardTransactionType {
	if value == nil {
		return nil
	}
	var result []enums.CardTransactionType
	for _, candidate := range []enums.CardTransactionType{
		enums.CardTransactionType_AUTH,
		enums.CardTransactionType_CLEAR,
		enums.CardTransactionType_VOID,
		enums.CardTransactionType_REFUND,
		enums.CardTransactionType_VERIFICATION,
	} {
		if channelEnums.TransactionTypeFromGeneric(candidate) == *value {
			result = append(result, candidate)
		}
	}
	return result
}

func uiTransactionStatuses(value *channelEnums.TransactionStatus) []enums.CardTransactionStatus {
	if value == nil {
		return nil
	}
	var result []enums.CardTransactionStatus
	for _, candidate := range []enums.CardTransactionStatus{
		enums.TransactionStatus_PENDING,
		enums.TransactionStatus_AUTHORIZED,
		enums.TransactionStatus_SUCCEED,
		enums.TransactionStatus_FAILED,
		enums.TransactionStatus_VOID,
	} {
		if channelEnums.TransactionStatusFromGeneric(candidate) == *value {
			result = append(result, candidate)
		}
	}
	return result
}
