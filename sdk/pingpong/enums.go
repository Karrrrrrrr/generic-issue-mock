package pingpong

// CardAction 卡片操作类型。
type CardAction string

const (
	CardActionFreeze   CardAction = "freeze"   // 冻结
	CardActionUnfreeze CardAction = "unfreeze" // 解冻
	CardActionClose    CardAction = "close"    // 注销
)

func (c CardAction) ToString() string {
	return string(c)
}

// CardFundingAction 卡片资金操作类型。
type CardFundingAction string

const (
	CardFundingTopUp    CardFundingAction = "top_up"   // 充值
	CardFundingWithdraw CardFundingAction = "withdraw" // 提现
)

func (c CardFundingAction) ToString() string {
	return string(c)
}

// CardStatus 卡片状态。
type CardStatus string

const (
	CardStatusInactive  CardStatus = "INACTIVE"  // 待激活
	CardStatusActive    CardStatus = "ACTIVE"    // 激活
	CardStatusFrozen    CardStatus = "REVOKED"   // 冻结
	CardStatusCancelled CardStatus = "CANCELLED" // 注销
)

func (c CardStatus) ToString() string {
	return string(c)
}
func (s CardStatus) Valid() bool {
	switch s {
	case CardStatusInactive, CardStatusActive, CardStatusFrozen, CardStatusCancelled:
		return true
	}
	return false
}

func CardStatusFromString(value string) CardStatus {
	switch value {
	case CardStatusInactive.ToString():
		return CardStatusInactive
	case CardStatusFrozen.ToString():
		return CardStatusFrozen
	case CardStatusActive.ToString():
		return CardStatusActive
	case CardStatusCancelled.ToString():
		return CardStatusCancelled
	default:
		return ""
	}
}

// FundingOrderStatus 资金订单状态。
type FundingOrderStatus string

const (
	FundingOrderStatusSuccess    FundingOrderStatus = "SUCCESS"    // 成功
	FundingOrderStatusFail       FundingOrderStatus = "FAIL"       // 失败
	FundingOrderStatusProcessing FundingOrderStatus = "PROCESSING" // 处理中
)

func (c FundingOrderStatus) ToString() string {
	return string(c)
}

// CardNetwork 卡组织。
type CardNetwork string

const (
	CardNetworkVisa       CardNetwork = "visa"
	CardNetworkMastercard CardNetwork = "mastercard"
	CardNetworkDiscover   CardNetwork = "discover"
)

func (c CardNetwork) ToString() string {
	return string(c)
}
