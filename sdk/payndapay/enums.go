package payndapay

type BalanceOperationType string

const (
	BalanceOperationType_INC BalanceOperationType = "INC" // 增加
	BalanceOperationType_DEC BalanceOperationType = "DEC" // 减少
)

func (t BalanceOperationType) Valid() bool {
	switch t {
	case BalanceOperationType_INC, BalanceOperationType_DEC:
		return true
	}
	return false
}

type TransferType string

const (
	TransferType_IN  TransferType = "IN"  // 充值
	TransferType_OUT TransferType = "OUT" // 提现
)

func (t TransferType) Valid() bool {
	switch t {
	case TransferType_IN, TransferType_OUT:
		return true
	}
	return false
}

type CardControlPeriod string

const (
	CardControlPeriod_DAY   CardControlPeriod = "DAY"   // 按天
	CardControlPeriod_MONTH CardControlPeriod = "MONTH" // 按月
	CardControlPeriod_TOTAL CardControlPeriod = "TOTAL" // 总
	CardControlPeriod_ONCE  CardControlPeriod = "ONCE"  // 一次
)

func (t CardControlPeriod) Valid() bool {
	switch t {
	case CardControlPeriod_DAY, CardControlPeriod_MONTH, CardControlPeriod_TOTAL, CardControlPeriod_ONCE:
		return true
	}
	return false
}

type CreditLimitType string

const (
	CreditLimitType_SHARED      = "SHARED"      // 共享额度
	CreditLimitType_INDEPENDENT = "INDEPENDENT" // 独立额度
)

func (t CreditLimitType) Valid() bool {
	switch t {
	case CreditLimitType_SHARED, CreditLimitType_INDEPENDENT:
		return true
	}
	return false
}

const (
	TRANSACTION_AUTHENTICATION_APPROVED         = "transaction.authentication.approved"         // 卡片授权尝试已获批准。当使用卡片且供应商收到支付卡片详情后，商户授权了该交易时发生。
	TRANSACTION_AUTHENTICATION_DECLINED         = "transaction.authentication.declined"         // 卡片授权尝试被拒绝。
	TRANSACTION_AUTHENTICATION_SETTLED          = "transaction.authentication.settled"          // 先前批准的交易已成功结算。
	TRANSACTION_AUTHENTICATION_REVERSAL_PENDING = "transaction.authentication.reversal.pending" // 先前批准的交易在结算完成前正被供应商作废。
	TRANSACTION_AUTHENTICATION_REVERSAL_SETTLED = "transaction.authentication.reversal.settled" // 先前批准的交易在结算过程之前已被供应商作废。
	TRANSACTION_AUTHENTICATION_REVERSAL_EXPIRED = "transaction.authentication.reversal.expired" // 先前批准的交易由于 7 天后未收到结算或撤销而已过期。
	TRANSACTION_REFUND_APPROVED                 = "transaction.refund.approved"                 // 尝试向卡片退款已发生并已获批准，但尚未完成。
	TRANSACTION_REFUND_SETTLED                  = "transaction.refund.settled"                  // 尝试向卡片退款已发生并已成功批准和结算。
	TRANSACTION_REFUND_DECLINED                 = "transaction.refund.declined"                 // 尝试向卡片退款已发生但被拒绝。
	TRANSACTION_REFUND_REVERSAL                 = "transaction.refund.reversal"                 // 退款批准后，退款已被作废并撤销。
)

// 三方卡状态
const (
	CARD_STATUS_DELETED  = "DELETED"
	CARD_STATUS_ACTIVE   = "ACTIVE"
	CARD_STATUS_FROZEN   = "FROZEN"
	CARD_STATUS_EXPIRED  = "EXPIRED"
	CARD_STATUS_BLOCK    = "BLOCK"
	CARD_STATUS_UNACTIVE = "UNACTIVE"
	CARD_STATUS_UNKNOWN  = "UNKNOWN"
)
