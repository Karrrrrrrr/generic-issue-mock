package queryconv

import (
	"time"

	"generic-mock/pkg/timeparse"
)

type CardTransactionTimeRangeRequest struct {
	StartTime        *string
	StartCreatedDate *string
	StartPostingDate *string
	EndTime          *string
	EndCreatedDate   *string
	EndPostingDate   *string
}

func OptionalTime(value *string) (*time.Time, bool) {
	parsed, err := timeparse.ParseOptionalDate(value)
	if err != nil {
		return nil, false
	}
	return parsed, true
}

func CardTransactionTimeRange(req CardTransactionTimeRangeRequest) (*time.Time, *time.Time, bool) {
	fromValues := []*string{req.StartTime, req.StartCreatedDate, req.StartPostingDate}
	toValues := []*string{req.EndTime, req.EndCreatedDate, req.EndPostingDate}
	var from *time.Time
	var to *time.Time
	for _, value := range fromValues {
		parsed, valid := OptionalTime(value)
		if !valid {
			return nil, nil, false
		}
		if parsed != nil && (from == nil || parsed.After(*from)) {
			from = parsed
		}
	}
	for _, value := range toValues {
		parsed, valid := OptionalTime(value)
		if !valid {
			return nil, nil, false
		}
		if parsed != nil && (to == nil || parsed.Before(*to)) {
			to = parsed
		}
	}
	if from != nil && to != nil && from.After(*to) {
		return nil, nil, false
	}
	return from, to, true
}
