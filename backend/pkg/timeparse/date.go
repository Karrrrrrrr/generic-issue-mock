package timeparse

import (
	"strings"
	"time"
)

var dateLayouts = []string{
	"2006-01-02",
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006/01/02",
	"01/02/2006",
}

func ParseDate(value string) (*time.Time, error) {
	for _, layout := range dateLayouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return &parsed, nil
		}
	}

	return nil, ErrUnsupportedDateLayout
}

func ParseOptionalDate(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	if strings.TrimSpace(*value) == "" {
		return nil, ErrUnsupportedDateLayout
	}
	return ParseDate(*value)
}
