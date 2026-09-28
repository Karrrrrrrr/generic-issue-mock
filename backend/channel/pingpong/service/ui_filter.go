package service

import (
	"time"

	pingerrors "generic-mock/channel/pingpong/errors"
)

type UIListTimeRange struct {
	CreatedFrom *time.Time `form:"created_from" time_format:"2006-01-02T15:04:05Z07:00"`
	CreatedTo   *time.Time `form:"created_to" time_format:"2006-01-02T15:04:05Z07:00"`
}

func (req UIListTimeRange) Validate() error {
	if req.CreatedFrom != nil && req.CreatedFrom.IsZero() || req.CreatedTo != nil && req.CreatedTo.IsZero() {
		return pingerrors.ErrInvalid
	}
	if req.CreatedFrom != nil && req.CreatedTo != nil && req.CreatedFrom.After(*req.CreatedTo) {
		return pingerrors.ErrInvalid
	}
	return nil
}
