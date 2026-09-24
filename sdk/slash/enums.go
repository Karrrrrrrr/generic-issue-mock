package slash

const (
	COMMISSION_FREQUENCY_MONTHLY = "monthly" // 每月
	COMMISSION_FREQUENCY_YEARLY  = "yearly"  // 每年

	COMMISSION_TYPE_FLATFEE  = "flatFee"  // 固定金额
	COMMISSION_TYPE_TAKERATE = "takeRate" // 按比例

	// Card Status
	//active, paused, inactive, closed
	CARD_STATUS_ACTIVE   = "active"   // 正
	CARD_STATUS_PAUSED   = "paused"   // 冻结
	CARD_STATUS_INACTIVE = "inactive" // 停用(开卡时初始状态)
	CARD_STATUS_CLOSED   = "closed"   // 删除

	// Card Type
	CARD_TYPE_VIRTUAL = "virtual" // 开卡时卡类型：virtual

	// Restriction
	RESTRICTION_ALLOWLIST = "allowlist"
	RESTRICTION_BLACKLIST = "blacklist"

	// Card Product Status
	CARD_PRODUCT_STATUS_ACTIVE   = "active"
	CARD_PRODUCT_STATUS_INACTIVE = "inactive"

	// Virtual Account Type
	VIRTUAL_ACCOUNT_TYPE_PRIMARY = "primary"
	VIRTUAL_ACCOUNT_TYPE_DEFAULT = "default"
)

type CardStatusEnum string

const (
	CardStatusEnumUnknow   CardStatusEnum = "unknow"
	CardStatusEnumActive   CardStatusEnum = "active"
	CardStatusEnumPaused   CardStatusEnum = "paused"
	CardStatusEnumInactive CardStatusEnum = "inactive"
	CardStatusEnumClosed   CardStatusEnum = "closed"
)

func cardStatusToCardStatusEnum(status string) CardStatusEnum {
	switch CardStatusEnum(status) {
	case CardStatusEnumActive,
		CardStatusEnumPaused,
		CardStatusEnumInactive,
		CardStatusEnumClosed:
		return CardStatusEnum(status)
	default:
		return CardStatusEnumUnknow
	}
}
