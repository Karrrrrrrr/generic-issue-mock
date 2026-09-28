package service

import (
	"time"

	"generic-mock/shared/biz"
	sharederrors "generic-mock/shared/errors"
)

type PageRequest struct {
	PageNumber *int `form:"page_number" binding:"omitempty,min=1,max=1000000"`
	PageSize   *int `form:"page_size" binding:"omitempty,min=1,max=200"`
}

func (req PageRequest) toBizPage() (biz.UIPageRequest, error) {
	page := 1
	size := 20
	if req.PageNumber != nil {
		page = *req.PageNumber
	}
	if req.PageSize != nil {
		size = *req.PageSize
	}
	if page < 1 || page > 1000000 || size < 1 || size > 200 {
		return biz.UIPageRequest{}, sharederrors.ErrInvalidUIRequest
	}
	return biz.UIPageRequest{
		Offset: (page - 1) * size,
		Limit:  size,
	}, nil
}

type TimeRange struct {
	CreatedFrom *time.Time `form:"created_from" time_format:"2006-01-02T15:04:05Z07:00"`
	CreatedTo   *time.Time `form:"created_to" time_format:"2006-01-02T15:04:05Z07:00"`
}

type Page[Item any] struct {
	Items []Item `json:"items"`
	Total int64  `json:"total"`
}
