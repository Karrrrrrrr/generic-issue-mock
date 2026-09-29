package timefmt

import "time"

func Date(value time.Time) string {
	return value.Format(time.DateOnly)
}

func Month(value time.Time) string {
	return value.Format("01")
}

func Year(value time.Time) string {
	return value.Format("2006")
}

func CardExpiration(value time.Time) string {
	return value.Format("01/06")
}
