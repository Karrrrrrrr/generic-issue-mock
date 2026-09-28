package biz

import "context"

type Transaction interface {
	InTx(context.Context, func(context.Context) error) error
	IsInTx(context.Context) bool
}
