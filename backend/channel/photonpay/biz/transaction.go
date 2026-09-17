package biz

import (
	"context"
)

type PhotonPayTransaction interface {
	InTx(context.Context, func(context.Context) error) error
}
