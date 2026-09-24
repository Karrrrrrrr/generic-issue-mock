package time

import (
	"encoding/json"
	"fmt"
	"time"
)

const isoDateTimeLayout = "2006-01-02T15:04:05"

type ISODateTime time.Time

func (value ISODateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(value).Format(isoDateTimeLayout))
}

func (value *ISODateTime) UnmarshalJSON(data []byte) error {
	if value == nil {
		return fmt.Errorf("cannot unmarshal JSON to *ISODateTime")
	}
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("invalid date format")
	}
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}
	parsed, err := time.ParseInLocation(isoDateTimeLayout, text, time.Local)
	if err != nil {
		return err
	}
	*value = ISODateTime(parsed)
	return nil
}
