package biz

import (
	"context"
)

type SlashTransaction interface {
	InTx(context.Context, func(context.Context) error) error
}
