package biz

import (
	"github.com/shopspring/decimal"
)

func DefaultBalance() decimal.Decimal {
	return decimal.NewFromInt(1_000_000)
}
