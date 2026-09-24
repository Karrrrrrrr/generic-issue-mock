package enums

type CardControlPeriod string
type BalanceOperationType string

const (
	CardControlPeriodDay     CardControlPeriod    = "DAY"
	BalanceOperationIncrease BalanceOperationType = "INC"
	BalanceOperationDecrease BalanceOperationType = "DEC"
)
