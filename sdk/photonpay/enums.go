package photonpay

const (
	ERROR_CODE_VCC1078 = "VCC1078" // 冻结解冻：卡状态异常
	ERROR_CODE_VCC1030 = "VCC1030" // 卡删除：卡状态无法
)

// type AccountEnum string

// const (
// 	ACCOUNT_ENUM_MEMBER AccountEnum = "member" // 会员光子易账户
// 	ACCOUNT_ENUM_MATRIX AccountEnum = "matrix" // Matrix账户
// 	ACCOUNT_ENUM_CARD   AccountEnum = "card"   // 常规卡账户
// )

// func (t AccountEnum) ToPrt() *AccountEnum {
// 	return &t
// }

// type AccountType string

// const (
// 	ACCOUNT_TYPE_FT10001 AccountType = "FT10001" // 可用金额
// 	ACCOUNT_TYPE_FT10002 AccountType = "FT10002" // 冻结金额
// 	ACCOUNT_TYPE_FT10003 AccountType = "FT10003" // 待结算金额
// 	ACCOUNT_TYPE_FT10004 AccountType = "FT10004" // 保证金金额
// )

// func (t AccountType) Valid() bool {
// 	switch t {
// 	case ACCOUNT_TYPE_FT10001, ACCOUNT_TYPE_FT10002, ACCOUNT_TYPE_FT10003, ACCOUNT_TYPE_FT10004:
// 		return true
// 	}
// 	return false
// }

// func (t AccountType) ToPrt() *AccountType {
// 	return &t
// }

// type CertType string

// const (
// 	CERT_TYPE_ID_CARD         CertType = "id_card"         // 身份证
// 	CERT_TYPE_PASSPORT        CertType = "passport"        // 护照
// 	CERT_TYPE_RESIDENT_PERMIT CertType = "resident_permit" // 居留许可证(永居、绿卡、工作签证）
// )

// func (t CertType) Valid() bool {
// 	switch t {
// 	case CERT_TYPE_ID_CARD, CERT_TYPE_PASSPORT, CERT_TYPE_RESIDENT_PERMIT:
// 		return true
// 	}
// 	return false
// }

// func (t CertType) ToPrt() *CertType {
// 	return &t
// }

// type CardholderStatus string

// const (
// 	CARDHOLDER_STATUS_NORMAL   CardholderStatus = "normal"   // 可用
// 	CARDHOLDER_STATUS_DISABLED CardholderStatus = "disabled" // 系统禁用
// 	CARDHOLDER_STATUS_PENDING  CardholderStatus = "pending"  // 审核中
// 	CARDHOLDER_STATUS_MODIFY   CardholderStatus = "modify"   // 待完善
// 	CARDHOLDER_STATUS_FAILED   CardholderStatus = "failed"   // 失败
// 	CARDHOLDER_STATUS_REJECTED CardholderStatus = "rejected" // 审核拒绝
// )

// func (t CardholderStatus) Valid() bool {
// 	switch t {
// 	case CARDHOLDER_STATUS_NORMAL, CARDHOLDER_STATUS_DISABLED, CARDHOLDER_STATUS_PENDING, CARDHOLDER_STATUS_MODIFY, CARDHOLDER_STATUS_REJECTED:
// 		return true
// 	}
// 	return false
// }

// func (t CardholderStatus) ToPrt() *CardholderStatus {
// 	return &t
// }

// type CardholderReviewStatus string

// const (
// 	CARDHOLDER_REVIEW_STATUS_APPROVED      CardholderReviewStatus = "approved"      // 可用
// 	CARDHOLDER_REVIEW_STATUS_PENDING       CardholderReviewStatus = "pending"       // 审核中
// 	CARDHOLDER_REVIEW_STATUS_MODIFY        CardholderReviewStatus = "modify"        // 待完善
// 	CARDHOLDER_REVIEW_STATUS_REJECTED      CardholderReviewStatus = "rejected"      // 审核拒绝
// 	CARDHOLDER_REVIEW_STATUS_CERT_REQUIRED CardholderReviewStatus = "cert_required" // 待上传身份信息
// )

// type CardFormFactorType string

// const (
// 	CARD_FORM_FACTOR_TYPE_VIRTUAL_CARD     CardFormFactorType = "virtual_card"               // 虚拟卡
// 	CARD_FORM_FACTOR_TYPE_PHYSICAL_CARD    CardFormFactorType = "physical_card"              // 实体卡
// 	CARD_FORM_FACTOR_TYPE_VIRTUAL_PHYSICAL CardFormFactorType = "virtual_card,physical_card" // 虚拟卡+实体卡
// )

// func (t CardFormFactorType) Valid() bool {
// 	switch t {
// 	case CARD_FORM_FACTOR_TYPE_VIRTUAL_CARD, CARD_FORM_FACTOR_TYPE_PHYSICAL_CARD:
// 		return true
// 	}
// 	return false
// }

// func (t CardFormFactorType) ToPrt() *CardFormFactorType {
// 	return &t
// }

// type CardType string

// const (
// 	CARD_TYPE_SHARE          CardType = "share"          // 共享卡
// 	CARD_TYPE_RECHARGE       CardType = "recharge"       // 常规卡
// 	CARD_TYPE_SHARE_RECHARGE CardType = "share,recharge" // 共享卡+常规卡
// )

// func (t CardType) Valid() bool {
// 	switch t {
// 	case CARD_TYPE_SHARE, CARD_TYPE_RECHARGE:
// 		return true
// 	}
// 	return false
// }

// func (t CardType) ToPrt() *CardType {
// 	return &t
// }

// type TransactionLimitType string

// const (
// 	TRANSACTION_LIMIT_TYPE_LIMITED   TransactionLimitType = "limited"   // 有限制
// 	TRANSACTION_LIMIT_TYPE_UNLIMITED TransactionLimitType = "unlimited" // 无限制
// )

// func (t TransactionLimitType) Valid() bool {
// 	switch t {
// 	case TRANSACTION_LIMIT_TYPE_LIMITED, TRANSACTION_LIMIT_TYPE_UNLIMITED:
// 		return true
// 	}
// 	return false
// }

// func (t TransactionLimitType) ToPrt() *TransactionLimitType {
// 	return &t
// }

// type CardSchemeType string

// const (
// 	CARD_SCHEME_TYPE_MASTER   CardSchemeType = "MasterCard" // 万事达卡
// 	CARD_SCHEME_TYPE_VISA     CardSchemeType = "VISA"       // 发现卡
// 	CARD_SCHEME_TYPE_DISCOVER CardSchemeType = "Discover"   // 发现卡
// )

// func (t CardSchemeType) Valid() bool {
// 	switch t {
// 	case CARD_SCHEME_TYPE_MASTER, CARD_SCHEME_TYPE_VISA, CARD_SCHEME_TYPE_DISCOVER:
// 		return true
// 	}
// 	return false
// }

// func (t CardSchemeType) ToPrt() *CardSchemeType {
// 	return &t
// }

// type OpenCardStatus string

// const (
// 	OPEN_CARD_STATUS_PENDING          OpenCardStatus = "pending"          // 审核中
// 	OPEN_CARD_STATUS_PENDING_RECHARGE OpenCardStatus = "pending_recharge" // 待转入
// 	OPEN_CARD_STATUS_SUCCEED          OpenCardStatus = "succeed"          // 成功
// 	OPEN_CARD_STATUS_FAILED           OpenCardStatus = "failed"           // 失败
// )

// func (t OpenCardStatus) Valid() bool {
// 	switch t {
// 	case OPEN_CARD_STATUS_PENDING, OPEN_CARD_STATUS_PENDING_RECHARGE, OPEN_CARD_STATUS_SUCCEED, OPEN_CARD_STATUS_FAILED:
// 		return true
// 	}
// 	return false
// }

// type HandleCardType string

// const (
// 	HANDLE_CARD_TYPE_APPLY_CARD  HandleCardType = "apply_card"  // 申请开卡
// 	HANDLE_CARD_TYPE_CARD_UPDATE HandleCardType = "card_update" // 修改开卡信息
// 	HANDLE_CARD_TYPE_CARD_FREEZE HandleCardType = "card_freeze" // 冻结卡
// )

// func (t HandleCardType) Valid() bool {
// 	switch t {
// 	case HANDLE_CARD_TYPE_APPLY_CARD, HANDLE_CARD_TYPE_CARD_UPDATE, HANDLE_CARD_TYPE_CARD_FREEZE:
// 		return true
// 	}
// 	return false
// }

// type CardStatus string

// const (
// 	CARD_STATUS_NORMAL           CardStatus = "normal"           //	可用
// 	CARD_STATUS_PENDING_RECHARGE CardStatus = "pending_recharge" //	待转入
// 	CARD_STATUS_UNACTIVATED      CardStatus = "unactivated"      //	未激活
// 	CARD_STATUS_FREEZING         CardStatus = "freezing"         //	冻结中
// 	CARD_STATUS_FROZEN           CardStatus = "frozen"           //	冻结
// 	CARD_STATUS_RISK_FROZEN      CardStatus = "risk_frozen"      //	风控冻结
// 	CARD_STATUS_SYSTEM_FROZEN    CardStatus = "system_frozen"    //	系统冻结
// 	CARD_STATUS_UNFREEZING       CardStatus = "unfreezing"       //	解冻中
// 	CARD_STATUS_EXPIRED          CardStatus = "expired"          //	过期
// 	CARD_STATUS_CANCELING        CardStatus = "canceling"        //	销卡中
// 	CARD_STATUS_CANCELLED        CardStatus = "cancelled"        //	销卡
// 	CARD_STATUS_RENEWING         CardStatus = "renewing"         //	续卡中
// 	CARD_STATUS_REPLACING        CardStatus = "replacing"        //	补卡中
// 	CARD_STATUS_LOST             CardStatus = "lost"             //	挂失
// 	CARD_STATUS_STOLEN           CardStatus = "stolen"           //	盗卡
// 	CARD_STATUS_PIN_LOST         CardStatus = "pin_lost"         //	密码锁定
// )

// func (t CardStatus) Valid() bool {
// 	switch t {
// 	case CARD_STATUS_NORMAL, CARD_STATUS_PENDING_RECHARGE, CARD_STATUS_UNACTIVATED, CARD_STATUS_FREEZING, CARD_STATUS_FROZEN, CARD_STATUS_RISK_FROZEN, CARD_STATUS_SYSTEM_FROZEN, CARD_STATUS_UNFREEZING, CARD_STATUS_EXPIRED, CARD_STATUS_CANCELING, CARD_STATUS_CANCELLED, CARD_STATUS_RENEWING, CARD_STATUS_REPLACING, CARD_STATUS_LOST, CARD_STATUS_STOLEN, CARD_STATUS_PIN_LOST:
// 		return true
// 	}
// 	return false
// }

// func (t CardStatus) ToPrt() *CardStatus {
// 	return &t
// }

// type RechargeStatus string

// const (
// 	RECHARGE_STATUS_PENDING RechargeStatus = "pending" // 待处理
// 	RECHARGE_STATUS_SUCCEED RechargeStatus = "succeed" // 成功
// 	RECHARGE_STATUS_FAILED  RechargeStatus = "failed"  // 失败
// )

// func (t RechargeStatus) Valid() bool {
// 	switch t {
// 	case RECHARGE_STATUS_PENDING, RECHARGE_STATUS_SUCCEED, RECHARGE_STATUS_FAILED:
// 		return true
// 	}
// 	return false
// }

// func (t RechargeStatus) ToPrt() *RechargeStatus {
// 	return &t
// }

// type TransactionLimitChangeType string

// const (
// 	TRANSACTION_LIMIT_CHANGE_TYPE_INCREASE TransactionLimitChangeType = "increase" // 增加
// 	TRANSACTION_LIMIT_CHANGE_TYPE_DECREASE TransactionLimitChangeType = "decrease" // 减少
// )

// func (t TransactionLimitChangeType) Valid() bool {
// 	switch t {
// 	case TRANSACTION_LIMIT_CHANGE_TYPE_INCREASE, TRANSACTION_LIMIT_CHANGE_TYPE_DECREASE:
// 		return true
// 	}
// 	return false
// }
// func (t TransactionLimitChangeType) ToPrt() *TransactionLimitChangeType {
// 	return &t
// }

// type FreezeCardStatus string

// const (
// 	FREEZE_CARD_STATUS_FREEZE   FreezeCardStatus = "freeze"   // 冻结卡
// 	FREEZE_CARD_STATUS_UNFREEZE FreezeCardStatus = "unfreeze" // 解冻卡
// )

// func (t FreezeCardStatus) Valid() bool {
// 	switch t {
// 	case FREEZE_CARD_STATUS_FREEZE, FREEZE_CARD_STATUS_UNFREEZE:
// 		return true
// 	}
// 	return false
// }
// func (t FreezeCardStatus) ToPrt() *FreezeCardStatus {
// 	return &t
// }

// type TransactionType string

// const (
// 	TRANSACTION_TYPE_AUTH                          TransactionType = "auth"                          // 消费
// 	TRANSACTION_TYPE_CORRECTIVE_AUTH               TransactionType = "corrective_auth"               // 纠正授权
// 	TRANSACTION_TYPE_VERIFICATION                  TransactionType = "verification"                  // 验证
// 	TRANSACTION_TYPE_VOID                          TransactionType = "void"                          // 撤销
// 	TRANSACTION_TYPE_REFUND                        TransactionType = "refund"                        // 退款
// 	TRANSACTION_TYPE_CORRECTIVE_REFUND             TransactionType = "corrective_refund"             // 校正退款
// 	TRANSACTION_TYPE_RECHARGE                      TransactionType = "recharge"                      // 转入
// 	TRANSACTION_TYPE_RECHARGE_RETURN               TransactionType = "recharge_return"               // 卡金额退还
// 	TRANSACTION_TYPE_DISCARD_RECHARGE_RETURN       TransactionType = "discard_recharge_return"       // 销卡退回
// 	TRANSACTION_TYPE_SERVICE_FEE                   TransactionType = "service_fee"                   // 服务费
// 	TRANSACTION_TYPE_REFUND_REVERSAL               TransactionType = "refund_reversal"               // 退款撤销
// 	TRANSACTION_TYPE_FUND_IN                       TransactionType = "fund_in"                       // 汇入
// 	TRANSACTION_TYPE_ATM_INQUIRY                   TransactionType = "atm_inquiry"                   // ATM查询
// 	TRANSACTION_TYPE_ATM_WITHDRAWALS               TransactionType = "atm_withdrawals"               // ATM提现
// 	TRANSACTION_TYPE_ATM_INQUIRY_FAILED_RETURN     TransactionType = "atm_inquiry_failed_return"     // ATM查询失败退回
// 	TRANSACTION_TYPE_ATM_WITHDRAWALS_FAILED_RETURN TransactionType = "atm_withdrawals_failed_return" // ATM提提现失败退回
// )

// func (t TransactionType) Valid() bool {
// 	switch t {
// 	case TRANSACTION_TYPE_AUTH, TRANSACTION_TYPE_CORRECTIVE_AUTH, TRANSACTION_TYPE_VERIFICATION, TRANSACTION_TYPE_VOID, TRANSACTION_TYPE_REFUND, TRANSACTION_TYPE_CORRECTIVE_REFUND, TRANSACTION_TYPE_RECHARGE, TRANSACTION_TYPE_RECHARGE_RETURN, TRANSACTION_TYPE_DISCARD_RECHARGE_RETURN, TRANSACTION_TYPE_SERVICE_FEE, TRANSACTION_TYPE_REFUND_REVERSAL, TRANSACTION_TYPE_FUND_IN, TRANSACTION_TYPE_ATM_INQUIRY, TRANSACTION_TYPE_ATM_WITHDRAWALS, TRANSACTION_TYPE_ATM_INQUIRY_FAILED_RETURN, TRANSACTION_TYPE_ATM_WITHDRAWALS_FAILED_RETURN:
// 		return true
// 	}
// 	return false
// }

// func (t TransactionType) ToPrt() *TransactionType {
// 	return &t
// }

// type TransactionStatus string

// const (
// 	TRANSACTION_STATUS_PENDING    TransactionStatus = "pending"    // 处理中
// 	TRANSACTION_STATUS_AUTHORIZED TransactionStatus = "authorized" // 已授权
// 	TRANSACTION_STATUS_SUCCEED    TransactionStatus = "succeed"    // 成功
// 	TRANSACTION_STATUS_FAILED     TransactionStatus = "failed"     // 失败
// 	TRANSACTION_STATUS_VOID       TransactionStatus = "void"       // 撤销
// 	TRANSACTION_STATUS_PROCESSING TransactionStatus = "processing" // 处理中
// )

// func (t TransactionStatus) Valid() bool {
// 	switch t {
// 	case TRANSACTION_STATUS_PENDING, TRANSACTION_STATUS_AUTHORIZED, TRANSACTION_STATUS_SUCCEED, TRANSACTION_STATUS_FAILED, TRANSACTION_STATUS_VOID, TRANSACTION_STATUS_PROCESSING:
// 		return true
// 	}
// 	return false
// }
// func (t TransactionStatus) ToPrt() *TransactionStatus {
// 	return &t
// }

// type BooleanType string

// const (
// 	Boolean_Type_Y BooleanType = "Y" // 是
// 	Boolean_Type_N BooleanType = "N" // 否
// )

// type SettleStatus string

// const (
// 	SETTLE_STATUS_UNSETTLED SettleStatus = "Unsettled" // 未结算
// 	SETTLE_STATUS_SETTLED   SettleStatus = "Settled"   // 已结算
// 	SETTLE_STATUS_NA        SettleStatus = "N/A"       // 未知
// )

// type CapitalFlowsType string

// const (
// 	CAPITAL_FLOWS_TYPE_TRANSFER_IN  CapitalFlowsType = "transfer_in"  // 转入
// 	CAPITAL_FLOWS_TYPE_TRANSFER_OUT CapitalFlowsType = "transfer_out" // 转出
// 	CAPITAL_FLOWS_TYPE_UNLIMITED    CapitalFlowsType = "unlimited"    // 不限(额度流向)
// )

// type FeeDeductionMethod string

// const (
// 	FEE_DEDUCTION_METHOD_INSIDE  FeeDeductionMethod = "inside"  // 内扣
// 	FEE_DEDUCTION_METHOD_OUTSIDE FeeDeductionMethod = "outside" // 外扣
// )

// type FeeType string

// const (
// 	FEE_TYPE_CREATE_CARD_FEE              FeeType = "create_card_fee"              // 新增会员卡费
// 	FEE_TYPE_DISCARD_FEE                  FeeType = "discard_fee"                  // 销卡费
// 	FEE_TYPE_LOSS_REPORTING_FEE           FeeType = "loss_reporting_fee"           // 挂失报告费
// 	FEE_TYPE_RENEWAL_FEE                  FeeType = "renewal_fee"                  // 续费
// 	FEE_TYPE_REPLACEMENT_FEE              FeeType = "replacement_fee"              // 替换费
// 	FEE_TYPE_STANDARD_CARD_PRODUCTION_FEE FeeType = "standard_card_production_fee" // 标准会员卡生产费
// 	FEE_TYPE_PREMIUM_CARD_PRODUCTION_FEE  FeeType = "premium_card_production_fee"  // 高级会员卡生产费
// )

// type TxnType string

// const (
// 	FEE_TYPE_AUTH   TxnType = "auth"   // 授权
// 	FEE_TYPE_VOID   TxnType = "void"   // 撤销
// 	FEE_TYPE_REFUND TxnType = "refund" // 退款
// )

// func (t TxnType) Valid() bool {
// 	switch t {
// 	case FEE_TYPE_AUTH, FEE_TYPE_VOID, FEE_TYPE_REFUND:
// 		return true
// 	}
// 	return false
// }

// type UploadFileBusinessKey string

// const (
// 	UPLOAD_FILE_BUSINESS_KEY_CARDHOLDER UploadFileBusinessKey = "issuing_cardholder_identity_certificate" // 持卡人身份凭证
// )

// func (t UploadFileBusinessKey) Valid() bool {
// 	switch t {
// 	case UPLOAD_FILE_BUSINESS_KEY_CARDHOLDER:
// 		return true
// 	}
// 	return false
// }
