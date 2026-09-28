package biz

import (
	"context"

	"generic-mock/model"
)

type CardRepo interface {
	Create(context.Context, *model.Card) error
}
