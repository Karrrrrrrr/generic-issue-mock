package biz

import (
	"context"
)

type PayndaTransaction interface {
	InTx(context.Context, func(context.Context) error) error
}
