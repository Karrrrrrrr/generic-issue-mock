package queryconv

import (
	"time"

	"generic-mock/pkg/timeparse"
)

func RequiredDate(value string) (*time.Time, bool) {
	parsed, err := timeparse.ParseDate(value)
	if err != nil {
		return nil, false
	}
	return parsed, true
}

func OptionalTime(value *string) (*time.Time, bool) {
	parsed, err := timeparse.ParseOptionalDate(value)
	if err != nil {
		return nil, false
	}
	return parsed, true
}
