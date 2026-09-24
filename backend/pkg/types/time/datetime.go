package time

import (
	"fmt"
	"time"
)

type DateTime time.Time

func (value *DateTime) UnmarshalParam(text string) error {
	if value == nil {
		return fmt.Errorf("cannot unmarshal query parameter to *DateTime")
	}
	parsed, err := time.Parse(time.DateTime, text)
	if err != nil {
		return err
	}
	*value = DateTime(parsed)
	return nil
}

func (t DateTime) MarshalJSON() ([]byte, error) {
	s := time.Time(t).Format(time.DateTime)
	return []byte(fmt.Sprintf(`"%s"`, s)), nil
}

func (t *DateTime) UnmarshalJSON(s []byte) error {
	if t == nil {
		return fmt.Errorf("cannot unmarshal JSON to *DateTime")
	}
	if len(s) == 0 || s[0] != '"' || s[len(s)-1] != '"' {
		return fmt.Errorf("invalid date format")
	}
	str := string(s[1 : len(s)-1])
	tt, err := time.ParseInLocation(time.DateTime, str, time.Local)
	if err != nil {
		return err
	}
	*t = DateTime(tt)
	return nil
}
