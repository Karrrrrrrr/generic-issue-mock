package service

import (
	"time"
)

type UIListTimeRange struct {
	CreatedFrom *time.Time `form:"created_from" time_format:"2006-01-02T15:04:05Z07:00" time_utc:"1"`
	CreatedTo   *time.Time `form:"created_to" time_format:"2006-01-02T15:04:05Z07:00" time_utc:"1"`
}
