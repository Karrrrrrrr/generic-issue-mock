package enums

type Currency string

const (
	Currency_USD Currency = "USD"
	Currency_GBP Currency = "GBP"
	Currency_JPY Currency = "JPY"
	Currency_CNY Currency = "CNY"
)

type WalletType string

const (
	WalletType_Card           WalletType = "card"
	WalletType_VirtualAccount WalletType = "virtual_account"
	WalletType_Account        WalletType = "account"
)

type CardType string

const (
	CardType_Share  CardType = "share"  // 表示是虚拟账户卡, 可以共享余额
	CardType_Single CardType = "single" // 表示是普通卡, 不可以共享余额
)

type CardFormType string

const (
	CardFormType_Virtual  CardFormType = "virtual"
	CardFormType_Physical CardFormType = "physical"
)

type CardScheme string

const (
	CardScheme_Visa       CardScheme = "VISA"
	CardScheme_MasterCard CardScheme = "MasterCard"
)

type CardTransactionStatus string

const (
	TransactionStatus_PENDING    CardTransactionStatus = "pending"    // 处理中
	TransactionStatus_AUTHORIZED CardTransactionStatus = "authorized" // 已授权
	TransactionStatus_SUCCEED    CardTransactionStatus = "succeed"    // 成功
	TransactionStatus_FAILED     CardTransactionStatus = "failed"     // 失败
	TransactionStatus_VOID       CardTransactionStatus = "void"       // 撤销
	//TransactionStatus_PROCESSING CardTransactionStatus = "processing" // 处理中
)

type WebhookDeliveryStatus string

const (
	WebhookDeliveryStatus_Pending   WebhookDeliveryStatus = "pending"
	WebhookDeliveryStatus_Succeeded WebhookDeliveryStatus = "succeeded"
	WebhookDeliveryStatus_Failed    WebhookDeliveryStatus = "failed"
)

type CardTransactionType string

const (
	CardTransactionType_AUTH         CardTransactionType = "auth"         // 消费
	CardTransactionType_CLEAR        CardTransactionType = "clear"        // 清算
	CardTransactionType_VERIFICATION CardTransactionType = "verification" // 验证
	CardTransactionType_VOID         CardTransactionType = "void"         // 撤销
	CardTransactionType_REFUND       CardTransactionType = "refund"       // 退款
	CardTransactionType_FundIn       CardTransactionType = "fund_in"      // 资金转入
	CardTransactionType_FundOut      CardTransactionType = "fund_out"     // 资金转出
	//CardTransactionType_CORRECTIVE_AUTH               CardTransactionType = "corrective_auth"               // 纠正授权
	//CardTransactionType_CORRECTIVE_REFUND             CardTransactionType = "corrective_refund"             // 校正退款
	//CardTransactionType_CORRECTIVE_REFUND_VOID        CardTransactionType = "corrective_refund_void"        // 校正退款撤销
	//CardTransactionType_RECHARGE                      CardTransactionType = "recharge"                      // 转入
	//CardTransactionType_RECHARGE_RETURN               CardTransactionType = "recharge_return"               // 卡金额退还
	//CardTransactionType_DISCARD_RECHARGE_RETURN       CardTransactionType = "discard_recharge_return"       // 销卡退回
	//CardTransactionType_SERVICE_FEE                   CardTransactionType = "service_fee"                   // 服务费
	//CardTransactionType_REFUND_REVERSAL               CardTransactionType = "refund_reversal"               // 退款撤销
	//CardTransactionType_FUND_IN                       CardTransactionType = "fund_in"                       // 汇入
	//CardTransactionType_ATM_INQUIRY                   CardTransactionType = "atm_inquiry"                   // ATM查询
	//CardTransactionType_ATM_WITHDRAWALS               CardTransactionType = "atm_withdrawals"               // ATM提现
	//CardTransactionType_ATM_INQUIRY_FAILED_RETURN     CardTransactionType = "atm_inquiry_failed_return"     // ATM查询失败退回
	//CardTransactionType_ATM_WITHDRAWALS_FAILED_RETURN CardTransactionType = "atm_withdrawals_failed_return" // ATM提提现失败退回
)

type CardStatus string

const (
	CardStatus_Inactive  CardStatus = "inactive" // 实体卡专用
	CardStatus_Active    CardStatus = "active"   // 正常
	CardStatus_Frozing   CardStatus = "frozing"  // 冻结中
	CardStatus_Frozen    CardStatus = "frozen"   // 冻结
	CardStatus_Deleteing CardStatus = "deleting" // 删除中
	CardStatus_Deleted   CardStatus = "deleted"  // 已删除
)

type Channel string

const (
	Channel_PhotonPay Channel = "photonpay"
	Channel_Paynda    Channel = "paynda"
	Channel_Slash     Channel = "slash"
)

type OperationType string

const (
	OperationType_OpenCard   OperationType = "open_card"
	OperationType_UpdateCard OperationType = "update_card"
	OperationType_FreezeCard OperationType = "freeze_card"
	OperationType_CancelCard OperationType = "cancel_card"
)

type OperationStatus string

const (
	OperationStatus_Succeed OperationStatus = "succeed"
)

type CardHolderStatus string

const (
	CardHolderStatus_Normal CardHolderStatus = "normal"
)

type CardHolderReviewStatus string

const (
	CardHolderReviewStatus_Approved CardHolderReviewStatus = "approved"
)
