package biz

import "context"

type PingPongTransaction interface {
	InTx(context.Context, func(context.Context) error) error
}
