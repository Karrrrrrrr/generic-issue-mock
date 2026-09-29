package timeconv

import "time"

const cardExpirationLayout = "01/06"

func ParseCardExpiration(value string, defaultValue time.Time) (time.Time, bool) {
	if value == "" {
		return defaultValue, true
	}
	parsed, err := time.Parse(cardExpirationLayout, value)
	if err != nil {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}
