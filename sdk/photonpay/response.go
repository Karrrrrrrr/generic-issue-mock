package photonpay

import "encoding/json"

type AccessTokenResponse struct {
	ExpiresIn        int64  `json:"expiresIn"`
	RefreshExpiresIn int64  `json:"refreshExpiresIn"`
	RefreshToken     string `json:"refreshToken"`
	Token            string `json:"token"`
}

type IssueAuthCodeResponse struct {
	AuthCode string `json:"authCode,omitempty"`
	ReqID    string `json:"reqId,omitempty"`
}

type AccountSingleResponse struct {
	MemberID        string  `json:"memberId,omitempty"`        // 会员ID；如涉及matrix产品，返回对应的会员或连接会员信息。
	AccountNo       string  `json:"accountNo,omitempty"`       // 账户编号
	AccountType     string  `json:"accountType,omitempty"`     // 账户类型: FT10001 可用金额; FT10002 冻结金额; FT10003 待结算金额; FT10004 保证金金额
	Currency        string  `json:"currency,omitempty"`        // 币种
	RealTimeBalance float64 `json:"realTimeBalance,omitempty"` // 账户实时金额，正常面额
	ReturnedAt      string  `json:"returnedAt,omitempty"`      // 金额查询时间 Value: "2023-12-13T15:30:00"
}

type AccountHistoryResponse struct {
	MemberID         string  `json:"memberId,omitempty"`         // 会员ID；如涉及matrix产品，返回对应的会员或连接会员信息
	AccountHistoryNo string  `json:"accountHistoryNo,omitempty"` // 业务交易所产生的账户流水号 (唯一)
	AccountNo        string  `json:"accountNo,omitempty"`        // 账户编号
	Amount           float64 `json:"amount,omitempty"`           // 交易金额
	BalanceFund      float64 `json:"balanceFund,omitempty"`      // 交易后金额
	BatchID          string  `json:"batchId,omitempty"`          // 交易批次号
	Currency         string  `json:"currency,omitempty"`         // 币种
	TransactedAt     string  `json:"transactedAt,omitempty"`     // 交易时间，动账具体时间 Value: "2023-12-13T15:30:00"
	Sources          string  `json:"sources,omitempty"`          // 交易场景
	TransactionNotes string  `json:"transactionNotes,omitempty"` // 交易备注
	TxnType          string  `json:"txnType,omitempty"`          // 账户交易类型，详情可看光子易账户附录账户交易类型
	VoucherID        string  `json:"voucherId,omitempty"`        // 交易凭证号
}

type PagingVccCardholderResponse struct {
	CardholderID               string `json:"cardholderId,omitempty"`               // 用卡人ID。
	CreatedAt                  string `json:"createdAt,omitempty"`                  // 创建时间 Value: "2022-03-18T08:43:28"
	DateOfBirth                string `json:"dateOfBirth,omitempty"`                // 出生日期。 <yyyy-MM-dd>
	Email                      string `json:"email,omitempty"`                      // 邮件。
	FirstName                  string `json:"firstName,omitempty"`                  // 名字。
	LastName                   string `json:"lastName,omitempty"`                   // 姓氏。
	IsLegal                    string `json:"isLegal,omitempty"`                    // 是否法人 Enum: "Y" "N"
	CardholderNameAbbreviation string `json:"cardholderNameAbbreviation,omitempty"` // 用卡人姓名缩写。
	Mobile                     string `json:"mobile,omitempty"`                     // 手机号。
	MobilePrefix               string `json:"mobilePrefix,omitempty"`               // 手机号前缀。
	NationalityCountryCode     string `json:"nationalityCountryCode,omitempty"`     // 国籍国家码二字码。
	ResidentialAddress         string `json:"residentialAddress,omitempty"`         // 账单地址。
	ResidentialCity            string `json:"residentialCity,omitempty"`            // 账单地城市。
	ResidentialCountryCode     string `json:"residentialCountryCode,omitempty"`     // 账单地国家码二字码。
	ResidentialPostalCode      string `json:"residentialPostalCode,omitempty"`      // 账单地邮编。
	ResidentialState           string `json:"residentialState,omitempty"`           // 账单地州省。
	Status                     string `json:"status,omitempty"`                     // 用卡人状态。normal - 可用，disabled - 系统禁用，pending - 审核中，modify - 待完善，rejected - 审核拒绝
	CardholderReviewStatus     string `json:"cardholderReviewStatus,omitempty"`     // 用卡人信息审核状态。pending - 审核中，modify - 待完善，approved - 审核通过，rejected - 审核拒绝。
	IdInfoRequirement          string `json:"idInfoRequirement,omitempty"`          // 用卡人身份信息上传要求 Enum: "N" "Y"
	Reason                     string `json:"reason,omitempty"`                     // 原因。
	MemberID                   string `json:"memberId,omitempty"`                   // 会员ID；如涉及matrix产品，返回对应的连接会员信息。
	MatrixAccount              string `json:"matrixAccount,omitempty"`              // Matrix账户ID
	CertType                   string `json:"certType,omitempty"`                   // 身份证件类型。id_card：身份证，passport：护照，resident_permit：居留许可证(永居、绿卡、工作签证）。
	CertCountryCode            string `json:"certCountryCode,omitempty"`            // 证件签发国国家二字码。
	CertID                     string `json:"certId,omitempty"`                     // 证件号。
}

type AddCardholderResponse struct {
	CardholderID           string `json:"cardholderId,omitempty"`           // 用卡人卡ID，每个用卡人的唯一编号
	MemberID               string `json:"memberId,omitempty"`               // 会员ID；如涉及matrix产品，返回对应的连接会员信息。
	Status                 string `json:"status,omitempty"`                 // 用卡人状态。Enum: "normal:可用" "disabled:失效" "failed:失败" "pending:审核中"
	CardholderReviewStatus string `json:"cardholderReviewStatus,omitempty"` // 用卡人信息审核状态。Enum: "pending:审核中" "modify:待修改" "approved:审核通过" "rejected:审核拒绝" "cert_required:待上传身份信息；可为空，表示无需审核且当前无审核状态"
	IdInfoRequirement      string `json:"idInfoRequirement,omitempty"`      // 用卡人身份信息上传要求 Enum: "N" "Y"
	Reason                 string `json:"reason,omitempty"`                 // 原因。
}

type EditCardholderResponse struct {
	CardholderID string `json:"cardholderId,omitempty"` // 用卡人卡ID，每个用卡人的唯一编号
}

type CardBinResponse struct {
	BillingAddressUpdatable string `json:"billingAddressUpdatable,omitempty"` // Enum: "Y" "N" 当值为'Y'此卡支持更新账单地址；值为'N'则此卡不支持变更账单地址
	CardBin                 string `json:"cardBin,omitempty"`                 // 卡bin
	CardCurrency            string `json:"cardCurrency,omitempty"`            // 卡bin支持的币种。多个币种用逗号分隔。
	CardScheme              string `json:"cardScheme,omitempty"`              // 卡组织 Enum: "MasterCard" "Discover"
	CardType                string `json:"cardType,omitempty"`                // 卡类型。share： 共享卡； recharge： 常规卡； share,recharge: 既支持常规卡也支持共享卡
	ExpiryDateCustomization string `json:"expiryDateCustomization,omitempty"` // Enum: "Y" "N" 有效期自定义
	RemainingAvailableCard  string `json:"remainingAvailableCard,omitempty"`  // 此卡BIN当前剩余可开卡的数量，Unlimited 为不限制
	AvailableCard           string `json:"availableCard,omitempty"`           // 此卡BIN 已开卡并可用的卡数量（不包含销卡及过期）
	CardFormFactor          string `json:"cardFormFactor,omitempty"`          // 卡介质。值：virtual_card 虚拟卡, physical_card 实体卡, virtual_card,physical_card 既支持虚拟卡也支持实体卡
}

type CardDetail struct {
	CardID                    string  `json:"cardId,omitempty"`                    // 卡ID，每张卡的唯一编号
	CardNo                    string  `json:"cardNo,omitempty"`                    // 卡号
	CardCurrency              string  `json:"cardCurrency,omitempty"`              // 卡本币
	CardScheme                string  `json:"cardScheme,omitempty"`                // 卡组织 Enum: "MasterCard" "Discover"
	CardStatus                string  `json:"cardStatus,omitempty"`                // 卡状态Enum: "normal" "pending_recharge" "freezing" "frozen" "risk_frozen" "system_frozen" "unfreezing" "expired" "canceling" "cancelled" "unactivated"
	CardFormFactor            string  `json:"cardFormFactor,omitempty"`            // 卡介质。值： virtual_card虚拟卡；physical_card实体卡。
	CardType                  string  `json:"cardType,omitempty"`                  // 卡类型。share： 共享卡； recharge： 常规卡。
	CVV                       string  `json:"cvv,omitempty"`                       // cvv，仅虚拟卡展示
	Email                     string  `json:"email,omitempty"`                     // 邮件
	ExpirationDate            string  `json:"expirationDate,omitempty"`            // 卡有效期，MM/YY，仅虚拟卡展示
	FirstName                 string  `json:"firstName,omitempty"`                 // 名字
	LastName                  string  `json:"lastName,omitempty"`                  // 姓氏
	MemberID                  string  `json:"memberId,omitempty"`                  // 会员ID；如涉及matrix产品，返回对应的连接会员信息。
	MatrixAccount             string  `json:"matrixAccount,omitempty"`             // matrix账户号
	MaxOnDaily                int64   `json:"maxOnDaily,omitempty"`                // 日交易限额
	MaxOnMonthly              int64   `json:"maxOnMonthly,omitempty"`              // 月交易限额
	MaxOnPercent              int64   `json:"maxOnPercent,omitempty"`              // 单笔交易最大金额
	Mobile                    string  `json:"mobile,omitempty"`                    // 手机号
	MobilePrefix              string  `json:"mobilePrefix,omitempty"`              // 手机号前缀
	Nationality               string  `json:"nationality,omitempty"`               // 国籍，国家码二字码
	TransactionLimitType      string  `json:"transactionLimitType,omitempty"`      // 是否限制可交易额度 Enum: "limited" "unlimited"
	AvailableTransactionLimit float64 `json:"availableTransactionLimit,omitempty"` // 可交易额度
	CardBalance               float64 `json:"cardBalance,omitempty"`               // 常规卡金额
	RecipientID               string  `json:"recipientId,omitempty"`               // 收件人ID
}
type OpenCardResponse struct {
	CardDetail CardDetail `json:"cardDetail,omitempty"` // 卡对象
	RequestID  string     `json:"requestId,omitempty"`  // 幂等id,必填 商户请求流水号，每笔交易的唯一请求号，不可重复
	Status     string     `json:"status,omitempty"`     // 请求状态 Enum: "pending" "pending_recharge" "succeed" "failed"
}

// GetCardDetailResponse 卡片详细信息
type GetCardDetailResponse struct {
	CardID                     string `json:"cardId"`                     // 卡ID，每张卡的唯一编号
	CardFormFactor             string `json:"cardFormFactor"`             // 卡介质。Enum: "virtual_card", "physical_card"
	CardNo                     string `json:"cardNo"`                     // 卡号
	CardCurrency               string `json:"cardCurrency"`               // 卡本币
	CardScheme                 string `json:"cardScheme"`                 // 卡组织。Enum: "MasterCard", "Discover"
	CardStatus                 string `json:"cardStatus"`                 // 卡状态。Enum: "normal", "pending_recharge", "freezing", "frozen", "risk_frozen", "system_frozen", "unfreezing", "expired", "canceling", "cancelled", "unactivated", "renewing", "replacing", "lost", "stolen", "pin_lost"
	CardType                   string `json:"cardType"`                   // 卡类型。Enum: "share", "recharge"
	CreatedAt                  string `json:"createdAt"`                  // 创建时间。格式: "2022-03-18T08:43:28"
	CardholderID               string `json:"cardholderId"`               // 用卡人ID
	CardholderNameAbbreviation string `json:"cardholderNameAbbreviation"` // 卡面上的用卡人姓名
	MemberID                   string `json:"memberId"`                   // 会员ID；如涉及matrix产品，返回对应的连接会员信息
	MatrixAccount              string `json:"matrixAccount"`              // matrix账户号
	Email                      string `json:"email"`                      // 邮件
	ExpirationDate             string `json:"expirationDate"`             // 卡有效期，MM/YY
	FirstName                  string `json:"firstName"`                  // 名字
	LastName                   string `json:"lastName"`                   // 姓氏
	MaskCardNo                 string `json:"maskCardNo"`                 // 掩码卡号
	MaxOnDaily                 int64  `json:"maxOnDaily"`                 // 日交易限额
	MaxOnMonthly               int64  `json:"maxOnMonthly"`               // 月交易限额
	MaxOnPercent               int64  `json:"maxOnPercent"`               // 单笔交易最大金额
	Mobile                     string `json:"mobile"`                     // 手机号
	MobilePrefix               string `json:"mobilePrefix"`               // 手机号前缀
	Nationality                string `json:"nationality"`                // 国籍，国家码二字码
	Nickname                   string `json:"nickname"`                   // 卡昵称
	TotalTransactionLimit      string `json:"totalTransactionLimit"`      // 总交易限额。此卡至今为止设置的所有交易限额汇总。此值不可作为卡内可用交易限额来参考。（bigdecimal 类型，使用 string 存储）
	TransactionLimitType       string `json:"transactionLimitType"`       // 是否限制可交易额度。Enum: "limited", "unlimited"
	AvailableTransactionLimit  string `json:"availableTransactionLimit"`  // 可交易额度。（bigdecimal 类型，使用 string 存储）
	BillingAddress             string `json:"billingAddress"`             // 账单地址信息中的详细地址
	BillingAddressUpdatable    string `json:"billingAddressUpdatable"`    // 当值为'Y'此卡支持更新账单地址；值为'N'则此卡不支持变更账单地址。Enum: "Y", "N"
	BillingCity                string `json:"billingCity"`                // 账单地址信息中的市
	BillingCountry             string `json:"billingCountry"`             // 账单地址信息中的国家
	BillingPostalCode          string `json:"billingPostalCode"`          // 账单地址信息中的邮编
	BillingState               string `json:"billingState"`               // 账单地址信息中的省/州
	CardBalance                string `json:"cardBalance"`                // 常规卡金额。（bigdecimal 类型，使用 string 存储）
	RecipientID                string `json:"recipientId"`                // 收件人ID
	ProduceStatus              string `json:"produceStatus"`              // 制卡状态。Enum: "pending:处理中", "produced:制卡完成"
	TrackingNumber             string `json:"trackingNumber"`             // 物流单号
	UpdatedAt                  string `json:"updateAt"`                   // 更新时间。格式: "2022-03-18T08:43:28"
}

type CardDetailResponse struct {
	CreatedAt      string   `json:"createdAt"`             // 创建时间，格式: "2022-03-18T08:43:28"
	MemberID       string   `json:"memberId"`              // 会员ID；如涉及matrix产品，返回对应的连接会员信息。
	MatrixAccount  string   `json:"matrixAccount"`         // Matrix账户ID
	CardID         string   `json:"cardId"`                // 卡ID
	CardCurrency   string   `json:"cardCurrency"`          // 卡本币
	CardScheme     string   `json:"cardScheme"`            // 卡组织 Enum: "MasterCard" "Discover"
	CardFormFactor string   `json:"cardFormFactor"`        // 卡介质 Enum: "virtual_card" "physical_card"
	CardStatus     string   `json:"cardStatus"`            // 卡状态
	CardType       string   `json:"cardType"`              // 卡类型 Enum: "share" "recharge"
	MaskCardNo     string   `json:"maskCardNo"`            // 掩码卡号
	Nickname       string   `json:"nickname"`              // 昵称
	CardBalance    *float64 `json:"cardBalance,omitempty"` // 常规卡金额
}

type GetCvvResponse struct {
	CardID         string `json:"cardId"`                   // 卡ID
	CardNo         string `json:"cardNo,omitempty"`         // 卡号
	Cvv            string `json:"cvv"`                      // cvv
	ExpirationDate string `json:"expirationDate,omitempty"` // 卡有效期，MM/YY，仅虚拟卡展示
}

type PreRechargeResponse struct {
	AccountID              string  `json:"accountId,omitempty"`              // 账户ID
	ArrivalAmount          float64 `json:"arrivalAmount,omitempty"`          // 到账金额
	ArrivalAmountCurrency  string  `json:"arrivalAmountCurrency,omitempty"`  // 到账金额币种
	EffectiveQuotationTime int     `json:"effectiveQuotationTime,omitempty"` // 汇率报价有效时长
	ExchangeRate           float64 `json:"exchangeRate,omitempty"`           // 汇率
	QuotedAt               string  `json:"quotedAt,omitempty"`               // 汇率报价时间 Value: "2022-03-18T08:43:28"
	RechargeAmount         float64 `json:"rechargeAmount,omitempty"`         // 转入金额
	RechargeCurrency       string  `json:"rechargeCurrency,omitempty"`       // 转入币种
	RechargeFee            float64 `json:"rechargeFee,omitempty"`            // 转入手续费
	RechargeFeeCurrency    string  `json:"rechargeFeeCurrency,omitempty"`    // 转入手续费币种
	RequestID              string  `json:"requestId,omitempty"`              // 请保存此ID，带入转入下单接口完成转入。(所有交易以及请求的唯一标识，建议您保存。可用于对应的操作。)
}

type RechargeResponse struct {
	ArrivalAmount         float64 `json:"arrivalAmount,omitempty"`         // 到账金额
	ArrivalAmountCurrency string  `json:"arrivalAmountCurrency,omitempty"` // 到账金额币种
	CardBalance           float64 `json:"cardBalance,omitempty"`           // 卡内金额
	CardID                string  `json:"cardId,omitempty"`                // 卡ID
	CreatedAt             string  `json:"createdAt,omitempty"`             // 创建时间 Value: "2022-03-18T08:43:28"
	ExchangeRate          float64 `json:"exchangeRate,omitempty"`          // 汇率
	RechargeAmount        float64 `json:"rechargeAmount,omitempty"`        // 转入金额
	RechargeCurrency      string  `json:"rechargeCurrency,omitempty"`      // 转入币种
	RechargeFee           float64 `json:"rechargeFee,omitempty"`           // 转入手续费
	RechargeFeeCurrency   string  `json:"rechargeFeeCurrency,omitempty"`   // 转入手续费币种
	Status                string  `json:"status,omitempty"`                // 状态 Enum: "failed" "succeed"
	TransactionID         string  `json:"transactionId,omitempty"`         // 交易ID
}

type RechargeReturnResponse struct {
	TransactionID   string  `json:"transactionId,omitempty"`   // 交易ID
	CardID          string  `json:"cardId,omitempty"`          // 卡ID
	CreatedAt       string  `json:"createdAt,omitempty"`       // 创建时间 Value: "2022-03-18T08:43:28"
	MaskCardNo      string  `json:"maskCardNo,omitempty"`      // 掩码卡号
	ArrivalAmount   float64 `json:"arrivalAmount,omitempty"`   // 到账金额
	ReturnFeeAmount float64 `json:"returnFeeAmount,omitempty"` // 转入退回手续费
	CardBalance     float64 `json:"cardBalance,omitempty"`     // 常规卡金额
	Status          string  `json:"status,omitempty"`          // 状态 Enum: "failed" "succeed"
}

type PagingIssuingHistoryResponse struct {
	CardID            string `json:"cardId,omitempty"`            // 卡ID
	CardType          string `json:"cardType,omitempty"`          // 卡类型 Enum: "share" "recharge"
	CardFormFactor    string `json:"cardFormFactor,omitempty"`    // 卡介质 Enum: "virtual_card" "physical_card"
	CreatedAt         string `json:"createdAt,omitempty"`         // 申请时间 Value: "2022-03-18T08:43:28"
	FeeType           string `json:"feeType,omitempty"`           // 收费类型 Enum: "create_card_fee" "discard_fee" "loss_reporting_fee" "renewal_fee" "replacement_fee" "standard_card_production_fee" "premium_card_production_fee"
	MemberID          string `json:"memberId,omitempty"`          // 会员ID；如涉及matrix产品，返回对应的连接会员信息
	MatrixAccount     string `json:"matrixAccount,omitempty"`     // Matrix账户ID
	MaskCardNo        string `json:"maskCardNo,omitempty"`        // 前六后四格式的掩码卡号
	Nickname          string `json:"nickname,omitempty"`          // 昵称
	ReferenceNo       string `json:"referenceNo,omitempty"`       // 此ID用于对应光子易账户动账id
	Status            string `json:"status,omitempty"`            // 状态 Enum: "pending" "succeed" "failed"
	ActualFeeAmount   string `json:"actualFeeAmount,omitempty"`   // 手续费金额
	ActualFeeCurrency string `json:"actualFeeCurrency,omitempty"` // 手续费币种
}

type PagingRechargeCardFundsDetailResponse struct {
	CreatedAt          string  `json:"createdAt,omitempty"`          // 创建时间 Value: "2022-03-18T08:43:28"
	CardID             string  `json:"cardId,omitempty"`             // 卡ID
	TransactionID      string  `json:"transactionId,omitempty"`      // 交易ID
	TransactionType    string  `json:"transactionType,omitempty"`    // 交易类型
	Amount             float64 `json:"amount,omitempty"`             // 金额
	CapitalFlows       string  `json:"capitalFlows,omitempty"`       // 资金流 Enum: "transfer_in" "transfer_out"
	CardBalance        float64 `json:"cardBalance,omitempty"`        // 卡金额
	CardChangeAmount   float64 `json:"cardChangeAmount,omitempty"`   // 总变动金额
	CardCurrency       string  `json:"cardCurrency,omitempty"`       // 卡本币
	FeeAmount          float64 `json:"feeAmount,omitempty"`          // 手续费金额
	FeeDeductionMethod string  `json:"feeDeductionMethod,omitempty"` // 手续费扣费模式 Enum: "inside" "outside"
	MaskCardNo         string  `json:"maskCardNo,omitempty"`         // 掩码卡号
	CardFormFactor     string  `json:"cardFormFactor,omitempty"`     // 卡介质 Enum: "virtual_card" "physical_card"
	MemberID           string  `json:"memberId,omitempty"`           // 会员ID；如涉及matrix产品，返回对应的连接会员信息
	MatrixAccount      string  `json:"matrixAccount,omitempty"`      // matrix账户号
}
type PagingShareCardTxnLimitDetailResponse struct {
	CreatedAt                 string  `json:"createdAt,omitempty"`                 // 创建时间 Value: "2022-03-18T08:43:28"
	CardID                    string  `json:"cardId,omitempty"`                    // 卡ID
	TransactionID             string  `json:"transactionId,omitempty"`             // 交易ID
	TransactionType           string  `json:"transactionType,omitempty"`           // 交易类型
	Amount                    float64 `json:"amount,omitempty"`                    // 金额
	AvailableTransactionLimit float64 `json:"availableTransactionLimit,omitempty"` // 可用交易额度
	CapitalFlows              string  `json:"capitalFlows,omitempty"`              // 额度流向 Enum: "transfer_in" "transfer_out" "unlimited"
	CardCurrency              string  `json:"cardCurrency,omitempty"`              // 卡本币
	ChangeAmount              float64 `json:"changeAmount,omitempty"`              // 总变动金额
	FeeAmount                 float64 `json:"feeAmount,omitempty"`                 // 手续费金额
	MaskCardNo                string  `json:"maskCardNo,omitempty"`                // 掩码卡号
	CardFormFactor            string  `json:"cardFormFactor,omitempty"`            // 卡介质 Enum: "virtual_card" "physical_card"
	TxnLimitMethod            string  `json:"txnLimitMethod,omitempty"`            // 交易额度模式 Enum: "txn" "txn_fee"
	MemberID                  string  `json:"memberId,omitempty"`                  // 会员ID；如涉及matrix产品，返回对应的连接会员信息
	MatrixAccount             string  `json:"matrixAccount,omitempty"`             // Matrix账户ID
}

type PagingVccTradeOrderResponse struct {
	MemberID                        string          `json:"memberId,omitempty"`                        // 会员ID；如涉及matrix产品，返回对应的连接会员信息。
	MatrixAccount                   string          `json:"matrixAccount,omitempty"`                   // Matrix账户ID
	CreatedAt                       string          `json:"createdAt,omitempty"`                       // 创建时间(入账时间) Value: "2022-03-18T08:43:28"
	CardID                          string          `json:"cardId,omitempty"`                          // 卡ID
	CardType                        string          `json:"cardType,omitempty"`                        // 卡类型 Enum: "share" "recharge"
	CardFormFactor                  string          `json:"cardFormFactor,omitempty"`                  // 卡介质 Enum: "virtual_card" "physical_card"
	CardCurrency                    string          `json:"cardCurrency,omitempty"`                    // 卡本币
	TransactionID                   string          `json:"transactionId,omitempty"`                   // 交易ID
	OriginTransactionID             string          `json:"originTransactionID,omitempty"`             // 原始交易ID
	RequestID                       string          `json:"requestId,omitempty"`                       // 商户请求ID (所有交易以及请求的唯一标识，建议您保存。可用于对应的操作。)
	TransactionType                 string          `json:"transactionType,omitempty"`                 // 交易类型
	Status                          string          `json:"status,omitempty"`                          // 状态 Enum: "pending" "authorized" "succeed" "failed" "void" "processing"
	Code                            string          `json:"code,omitempty"`                            // 状态码
	Msg                             string          `json:"msg,omitempty"`                             // 状态描述
	Mcc                             string          `json:"mcc,omitempty"`                             // 商户类别码
	AuthCode                        string          `json:"authCode,omitempty"`                        // 授权码
	SettleStatus                    string          `json:"settleStatus,omitempty"`                    // 结算状态 Enum: "Unsettled" "Settled"
	TxnDate                         string          `json:"txnDate,omitempty"`                         // 交易时间 Value: "2022-03-18T08:43:28"
	TransactionAmount               float64         `json:"transactionAmount,omitempty"`               // 交易金额
	TransactionCurrency             string          `json:"transactionCurrency,omitempty"`             // 交易币种
	TxnPrincipalChangeAccount       string          `json:"txnPrincipalChangeAccount,omitempty"`       // 交易本金变动账户 Enum: "member" "matrix" "card"
	TxnPrincipalChangeAmount        float64         `json:"txnPrincipalChangeAmount,omitempty"`        // 交易本金金额
	TxnPrincipalChangeCurrency      string          `json:"txnPrincipalChangeCurrency,omitempty"`      // 交易本金变动币种
	TxnPrincipalChangeSettledAmount float64         `json:"txnPrincipalChangeSettledAmount,omitempty"` // 交易本金已结算金额
	SettleSpreadChangeAccount       string          `json:"settleSpreadChangeAccount,omitempty"`       // 结算价差变动账户 Enum: "member" "matrix" "card"
	SettleSpreadChangeCurrency      string          `json:"settleSpreadChangeCurrency,omitempty"`      // 结算价差变动币种
	FeeDeductionAccount             string          `json:"feeDeductionAccount,omitempty"`             // 手续费扣费账户 Enum: "member" "matrix" "card"
	FeeDeductionAmount              float64         `json:"feeDeductionAmount,omitempty"`              // 手续费扣费金额
	FeeDeductionCurrency            string          `json:"feeDeductionCurrency,omitempty"`            // 手续费扣费币种
	FeeDetailJson                   json.RawMessage `json:"feeDetailJson,omitempty"`                   // 手续费扣费金额明细
	FeeReturnAccount                string          `json:"feeReturnAccount,omitempty"`                // 手续费返还账户 Enum: "member" "matrix" "card"
	FeeReturnAmount                 float64         `json:"feeReturnAmount,omitempty"`                 // 手续费返还金额
	FeeReturnCurrency               string          `json:"feeReturnCurrency,omitempty"`               // 手续费返还币种
	FeeReturnDetailJson             json.RawMessage `json:"feeReturnDetailJson,omitempty"`             // 手续费返还金额明细
	ArrivalAccount                  string          `json:"arrivalAccount,omitempty"`                  // 到账账户 Enum: "member" "matrix" "card"
	ArrivalAmount                   float64         `json:"arrivalAmount,omitempty"`                   // 到账金额
	MaskCardNo                      string          `json:"maskCardNo,omitempty"`                      // 前六后四格式的掩码卡号
	MerchantNameLocation            string          `json:"merchantNameLocation,omitempty"`            // 商户名称
	MerchantLocation                string          `json:"merchantLocation,omitempty"`                // 商户所在国家
	Country                         string          `json:"country,omitempty"`                         // 国家
}

type GetWebhookNotificationResponse struct {
	Categories []WebhookCategory `json:"categories"`
}

// WebhookCategory Webhook类目
type WebhookCategory struct {
	Category string         `json:"category"` // Webhook类目编码
	Name     string         `json:"name"`     // Webhook类目名称
	Topics   []WebhookTopic `json:"topics"`   // 类目下的Webhook通知Topic
}

// WebhookTopic Webhook通知Topic
type WebhookTopic struct {
	TopicCode string            `json:"topicCode"` // Webhook通知Topic的唯一标识
	Name      string            `json:"name"`      // Webhook通知Topic的名称
	Templates []WebhookTemplate `json:"templates"` // Topic下的通知模板
}

// WebhookTemplate Webhook通知模板
type WebhookTemplate struct {
	Version       string `json:"version"`       // Webhook通知模板版本编号
	TemplateCode  string `json:"templateCode"`  // Webhook通知模板版本编码(唯一标识)
	ContentSample string `json:"contentSample"` // Webhook通知模板的内容样板
	Active        bool   `json:"active"`        // 标注当前模板和其对应的Webhook通知是否已生效
}
