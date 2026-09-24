package photonpay

import (
	"errors"
	"io"
	"unicode/utf8"
)

// 错误定义
var (
	ErrEmptyBalanceAccountID = errors.New("balanceAccountID不能为空")
	ErrEmptyCardID           = errors.New("cardID不能为空")
	ErrEmptyCardholderID     = errors.New("cardholderID不能为空")
	ErrEmptyParams           = errors.New("参数不能为空")
	ErrEmptyResponse         = errors.New("响应数据为空")
	ErrSignatureVerify       = errors.New("签名验证失败")
	ErrNoSignature           = errors.New("响应头中没有签名信息")
	ErrEmptyNonce            = errors.New("nonce不能为空")
	ErrEmptyRequestID        = errors.New("requestID不能为空")

	// api error
	ErrRepeatedRequest   = errors.New("重复请求")    //{"code":500,"message":"repeated request","success":false}
	ErrNotFound          = errors.New("未找到")     //{"code":500,"message":"not found","success":false}
	ErrInsufficientFunds = errors.New("余额不足")    //{"code":500,"message":"Insufficient funds","success":false}
	ErrServerInternal    = errors.New("服务器内部错误") //{"code":500,"message":"error.server.internal","success":false}
	ErrInvalidParameter  = errors.New("无效的参数")   //{"code":500,"message":"invalid parameter","success":false}

)

const (
	defaultCurrency = "USD"
)

type RestyRequestOptions struct {
	Method        string // 请求方法
	Path          string // 请求路径
	Token         string // token
	Authorization string // 签名认证
	Params        any    // 请求参数
	Result        any    // 响应数据
}

type AccountSingleRequest struct {
	Currency      *string `json:"currency,omitempty"`      // 币种[ISO4217]
	AccountNo     *string `json:"accountNo,omitempty"`     // 账户编号，系统内部唯一。如传入此参数，则精确查询，忽略currency、accountType、matrixAccount传值
	MemberID      *string `json:"memberId,omitempty"`      // 请求所指向的会员号，如不传则默认为token下的会员；若涉及matrix产品，可传入指定的连接会员号
	AccountType   *string `json:"accountType,omitempty"`   // 账户类型: FT10001 可用金额; FT10002 冻结金额; FT10003 待结算金额; FT10004 保证金金额; 默认为 FT10001
	MatrixAccount *string `json:"matrixAccount,omitempty"` // matrix账户号
}

func (r *AccountSingleRequest) Validate() error {
	if r.AccountNo == nil {
		if r.AccountType == nil {
			return errors.New("type不能为空")
		}
		if r.Currency == nil {
			return errors.New("currency不能为空")
		}
	}

	return nil
}

type AccountHistoryRequest struct {
	TransactedAtStart    string   `json:"transactedAtStart"`              // 起始动账日期，必填 2023-12-13T15:30:00
	TransactedAtEnd      string   `json:"transactedAtEnd"`                // 截止动账日期，必填 2023-12-13T15:30:00
	MemberID             *string  `json:"memberId,omitempty"`             // 请求所指向的会员号，如不传则默认为token下的会员；若涉及matrix产品，可传入指定的连接会员号
	AccountTransactionID *string  `json:"accountTransactionID,omitempty"` // 业务订单产生的账户交易ID
	Currency             *string  `json:"currency,omitempty"`             // 币种[ISO4217]
	AccountType          *string  `json:"accountType,omitempty"`          // 账户类型: FT10001 可用金额; FT10002 冻结金额; FT10003 待结算金额; FT10004 保证金金额; 默认为 FT10001
	MatrixAccount        *string  `json:"matrixAccount,omitempty"`        // matrix账户号
	AccountNo            *string  `json:"accountNo,omitempty"`            // 账户编号，系统内部唯一
	TxnType              []string `json:"txnType,omitempty"`              // 账户交易类型数组
	PageSize             *int64   `json:"pageSize,omitempty"`             // 分页大小
	PageIndex            *int64   `json:"pageIndex,omitempty"`            // 当前页/默认第一页
}

func (r *AccountHistoryRequest) Validate() error {
	if r.TransactedAtStart == "" {
		return errors.New("transactedAtStart不能为空")
	}
	if r.TransactedAtEnd == "" {
		return errors.New("transactedAtEnd不能为空")
	}
	if r.PageIndex != nil && *r.PageIndex <= 0 {
		return errors.New("pageIndex必须大于0")
	}
	if r.PageSize != nil && *r.PageSize <= 0 {
		return errors.New("pageSize必须大于0")
	}
	if r.AccountType != nil {
		return errors.New("AccountType不能为空")
	}
	return nil
}

type PagingVccCardholderRequest struct {
	PageIndex      *int64  `json:"pageIndex,omitempty"`      // 当前页/默认第一页
	PageSize       *int64  `json:"pageSize,omitempty"`       // 分页大小
	MemberID       *string `json:"memberId,omitempty"`       // 查询所指向的会员号，如不传则默认为token下的会员；如涉及matrix产品，如不传则指向会员和连接会员所有信息查询，或传入指定会员查询。
	MatrixAccount  *string `json:"matrixAccount,omitempty"`  // Matrix账户ID
	CreatedAtStart *string `json:"createdAtStart,omitempty"` // 起始时间 Example: createdAtStart=2022-03-18T08:43:28
	CreatedAtEnd   *string `json:"createdAtEnd,omitempty"`   // 结束时间 Example: createdAtEnd=2022-03-18T08:43:28
	CardholderID   *string `json:"cardholderId,omitempty"`   // 用卡人ID
	Status         *string `json:"status,omitempty"`         // 用卡人状态。normal - 可用，disabled - 系统禁用，pending - 审核中，modify - 待完善，rejected - 审核拒绝。
}

func (r *PagingVccCardholderRequest) Validate() error {
	if r.PageIndex != nil && *r.PageIndex <= 0 {
		return errors.New("pageIndex必须大于0")
	}
	if r.PageSize != nil && *r.PageSize <= 0 {
		return errors.New("pageSize必须大于0")
	}
	if r.Status != nil {
		return errors.New("Status不能为空")
	}
	return nil
}

type AddCardholderRequest struct {
	MemberID                   *string `json:"memberId,omitempty"`                   // 请求所指向的会员号，如不传则默认为token下的会员；若涉及matrix产品，可传入指定的连接会员号。
	MatrixAccount              *string `json:"matrixAccount,omitempty"`              // 如需在matrix下添加用卡人请填写matrix账户号 ,如不填写默认添加在member下。
	FirstName                  string  `json:"firstName"`                            // 名不能包含特殊字符。
	LastName                   string  `json:"lastName"`                             // 姓不能包含特殊字符。
	CardholderNameAbbreviation *string `json:"cardholderNameAbbreviation,omitempty"` // 用卡人姓名缩写,作为实体卡卡面上用卡人姓名，申请实体卡时必填 格式：（First Name/Last Name-名/姓） 长度：最多26个字符，包括一个“/ ” 要求：全大写字母。
	Email                      string  `json:"email"`                                // 邮箱。
	Mobile                     string  `json:"mobile"`                               // 手机号。
	MobilePrefix               string  `json:"mobilePrefix"`                         // 手机号前缀（包括区号），前面加 符号和1-4位的国家呼叫代码。不要包括连字符、空格或括号。
	DateOfBirth                string  `json:"dateOfBirth"`                          // 出生日期。<yyyy-MM-dd>
	CertType                   *string `json:"certType,omitempty"`                   // 身份证件类型。id_card：身份证，passport：护照，resident_permit：居留许可证(永居、绿卡、工作签证）。
	Portrait                   *string `json:"portrait,omitempty"`                   // 身份证件信息正面：请提供身份证件正面或护照首页照片。 仅限一张照片；PNG，JPG；单个文件最大6M
	ReverseSide                *string `json:"reverseSide,omitempty"`                // 身份证件信息反面：请提供身份证件反面照片。仅限一张照片；PNG，JPG；单个文件最大6M。
	NationalityCountryCode     string  `json:"nationalityCountryCode"`               // 国籍国家码二字码。
	ResidentialAddress         *string `json:"residentialAddress,omitempty"`         // 账单地址。建议填写账单地址。如需使用 Discover 的卡，此字段为必填。
	ResidentialCity            *string `json:"residentialCity,omitempty"`            // 账单地城市。建议填写账单地址所在城市。如需使用 Discover 的卡，此字段为必填。
	ResidentialCountryCode     *string `json:"residentialCountryCode,omitempty"`     // 账单地国家码二字码。建议填写账单地址国家码二字码。如需使用 Discover 的卡，此字段为必填。
	ResidentialPostalCode      *string `json:"residentialPostalCode,omitempty"`      // 账单地邮编。建议填写账单地址所属邮编。如需使用 Discover 的卡，此字段为必填。
	ResidentialState           *string `json:"residentialState,omitempty"`           // 账单地州省。建议填写账单地州省。如需使用 Discover 的卡，此字段为必填。
	CertCountryCode            *string `json:"certCountryCode,omitempty"`            // 证件签发国国家二字码。
	CertID                     *string `json:"certId,omitempty"`                     // 证件号。
}

type AddCardholderRequest1 struct {
	MemberID                   string // 请求所指向的会员号，如不传则默认为token下的会员；若涉及matrix产品，可传入指定的连接会员号。
	MatrixAccount              string // 如需在matrix下添加用卡人请填写matrix账户号 ,如不填写默认添加在member下。
	FirstName                  string // 名不能包含特殊字符。
	LastName                   string // 姓不能包含特殊字符。
	CardholderNameAbbreviation string // 用卡人姓名缩写,作为实体卡卡面上用卡人姓名，申请实体卡时必填 格式：（First Name/Last Name-名/姓） 长度：最多26个字符，包括一个“/ ” 要求：全大写字母。
	Email                      string // 邮箱。
	Mobile                     string // 手机号。
	MobilePrefix               string // 手机号前缀（包括区号），前面加 符号和1-4位的国家呼叫代码。不要包括连字符、空格或括号。
	DateOfBirth                string // 出生日期。<yyyy-MM-dd>
	CertType                   string // 身份证件类型。id_card：身份证，passport：护照，resident_permit：居留许可证(永居、绿卡、工作签证）。
	Portrait                   string // 身份证件信息正面：请提供身份证件正面或护照首页照片。 仅限一张照片；PNG，JPG；单个文件最大6M
	ReverseSide                string // 身份证件信息反面：请提供身份证件反面照片。仅限一张照片；PNG，JPG；单个文件最大6M。
	CertCountryCode            string // 证件签发国国家二字码。
	CertID                     string // 证件号。
}

func (r *AddCardholderRequest) Validate() error {
	if r.FirstName == "" {
		return errors.New("FirstName不能为空")
	}
	if r.LastName == "" {
		return errors.New("LastName不能为空")
	}
	if r.Email == "" {
		return errors.New("email不能为空")
	}
	if r.Mobile == "" {
		return errors.New("手机号不能为空")
	}
	if r.MobilePrefix == "" {
		return errors.New("手机号前缀不能为空")
	}
	if r.DateOfBirth == "" {
		return errors.New("生日不能为空")
	}
	if len(r.NationalityCountryCode) != 2 {
		return errors.New("国家二字码错误")
	}
	if r.ResidentialCountryCode != nil && len(*r.ResidentialCountryCode) != 2 {
		return errors.New("residentialCountryCode必须是2位国家代码")
	}

	if r.CertCountryCode != nil && len(*r.CertCountryCode) != 2 {
		return errors.New("certCountryCode必须是2位国家代码")
	}
	if r.CertType == nil {
		return errors.New("CertType不能为空")
	}
	if r.ResidentialState != nil && utf8.RuneCountInString(*r.ResidentialState) > 32 {
		return errors.New("省州长度不能超过32")
	}
	return nil
}

type EditCardholderRequest struct {
	CardholderID               string  `json:"cardholderId"`                         // 用卡人ID
	DateOfBirth                *string `json:"dateOfBirth,omitempty"`                // 出生日期 <yyyy-MM-dd>
	CertType                   *string `json:"certType,omitempty"`                   // 身份证件类型
	Portrait                   *string `json:"portrait,omitempty"`                   // 身份证件正面照片
	ReverseSide                *string `json:"reverseSide,omitempty"`                // 身份证件反面照片
	Email                      *string `json:"email,omitempty"`                      // 邮件
	CardholderNameAbbreviation *string `json:"cardholderNameAbbreviation,omitempty"` // 用卡人姓名缩写
	Mobile                     *string `json:"mobile,omitempty"`                     // 手机号
	MobilePrefix               *string `json:"mobilePrefix,omitempty"`               // 手机号前缀
	NationalityCountryCode     *string `json:"nationalityCountryCode,omitempty"`     // 国籍国家二字码
	ResidentialAddress         *string `json:"residentialAddress,omitempty"`         // 账单地址
	ResidentialCity            *string `json:"residentialCity,omitempty"`            // 账单地城市
	ResidentialCountryCode     *string `json:"residentialCountryCode,omitempty"`     // 账单地国家码
	ResidentialPostalCode      *string `json:"residentialPostalCode,omitempty"`      // 账单地邮编
	ResidentialState           *string `json:"residentialState,omitempty"`           // 账单地州省
	CertCountryCode            *string `json:"certCountryCode,omitempty"`            // 证件签发国国家二字码
	CertID                     *string `json:"certId,omitempty"`                     // 证件号
}

func (r *EditCardholderRequest) Validate() error {
	if r.CardholderID == "" {
		return errors.New("cardholderId不能为空")
	}

	if r.Email != nil && *r.Email == "" {
		return errors.New("email不能为空")
	}

	if r.NationalityCountryCode != nil && len(*r.NationalityCountryCode) != 2 {
		return errors.New("nationalityCountryCode必须是2位国家代码")
	}

	if r.ResidentialCountryCode != nil && len(*r.ResidentialCountryCode) != 2 {
		return errors.New("residentialCountryCode必须是2位国家代码")
	}

	if r.CertCountryCode != nil && len(*r.CertCountryCode) != 2 {
		return errors.New("certCountryCode必须是2位国家代码")
	}

	if r.ResidentialState != nil && utf8.RuneCountInString(*r.ResidentialState) > 32 {
		return errors.New("省州长度不能超过32")
	}

	return nil
}

type CardBinRequest struct {
	MemberID       *string `json:"memberId,omitempty"`       // 请求所指向的会员号，如不传则默认为token下的会员；若涉及matrix产品，可传入指定的连接会员号
	CardFormFactor *string `json:"cardFormFactor,omitempty"` // 卡介质。值： virtual_card虚拟卡；physical_card实体卡。若为空，则查询结果显示所有虚拟卡及实体卡卡介质
	CardType       *string `json:"cardType,omitempty"`       // 卡类型。值: share 或 recharge 可为空。如为空，则查询结果将显示所有类型的卡
	CardScheme     *string `json:"cardScheme,omitempty"`     // 卡组织
	CardCurrency   *string `json:"cardCurrency,omitempty"`   // 卡币种
}

func (r *CardBinRequest) Validate() error {
	if r.CardFormFactor == nil {
		return errors.New("CardFormFactor不能为空")
	}
	if r.CardType == nil {
		return errors.New("CardType不能为空")
	}
	if r.CardScheme == nil {
		return errors.New("CardScheme不能为空")
	}
	return nil
}

type OpenCardRequest struct {
	MemberID             *string  `json:"memberId,omitempty"`             // 请求所指向的会员号，如不传则默认为token下的会员；若涉及matrix产品，可传入指定的连接会员号
	MatrixAccount        *string  `json:"matrixAccount,omitempty"`        // 如果你想在matrix下申请卡，请填写matrix账号。如果不填，就在会员下默认创建
	CardBin              string   `json:"cardBin"`                        // 您需要填入你想开卡的卡bin信息，目前支持的卡bin信息可在卡bin接口中查询
	CardCurrency         string   `json:"cardCurrency"`                   // 卡本币
	CardExpirationDate   *int     `json:"cardExpirationDate,omitempty"`   // 卡有效期。您可填写此卡的有效期，以月为计数单位。如：您需要虚拟卡的有效期为1年，则输入“12”即可。如不上传，则由系统自动分配虚拟卡有效期。（最小12个月，最大35个月）
	CardScheme           string   `json:"cardScheme,omitempty"`           // 卡组织 Enum: "MasterCard" "Discover"
	CardType             string   `json:"cardType"`                       // 卡类型。share： 共享卡； recharge： 常规卡
	CardFormFactor       string   `json:"cardFormFactor,omitempty"`       // 卡介质。值： virtual_card虚拟卡；physical_card实体卡。不填默认虚拟卡
	CardholderID         string   `json:"cardholderId,omitempty"`         // 填入用卡人id后，卡将属于此用卡人。如不填则默认使用默认持卡人信息进行开卡。
	CardDesignID         *string  `json:"cardDesignId,omitempty"`         // 自定义卡面ID。如不传，使用默认卡面
	CardLogoID           *string  `json:"cardLogoId,omitempty"`           // 自定义卡logoID。如不传则不打印logo
	MaxOnDaily           *int64   `json:"maxOnDaily,omitempty"`           // 日交易限额。为空时不做限制，非空时大于1。默认最高限额为20000USD
	MaxOnMonthly         *int64   `json:"maxOnMonthly,omitempty"`         // 月交易限额。为空时不做限制，非空时大于1。默认最高限额为20000USD
	MaxOnPercent         *int64   `json:"maxOnPercent,omitempty"`         // 单笔交易最大金额。为空时不做限制，非空时大于1。默认最高限额为20000USD
	RechargeAmount       *float64 `json:"rechargeAmount,omitempty"`       // 转入金额  转入金额和到账金额择一填写即可
	RequestID            string   `json:"requestId"`                      // 幂等id,必填 商户请求流水号，每笔交易的唯一请求号，不可重复
	TransactionLimit     *float64 `json:"transactionLimit,omitempty"`     // 可交易额度
	TransactionLimitType *string  `json:"transactionLimitType,omitempty"` // 是否限制可交易额度 Enum: "limited" "unlimited"，recharge常规卡、share虚拟共享卡，默认unlimited；实体共享卡必为Limited
	AccountID            *string  `json:"accountId,omitempty"`            // 您需要用于转入的币种光子易账户ID号
	ArrivalAmount        *float64 `json:"arrivalAmount,omitempty"`        // 到账金额  转入金额和到账金额择一填写即可
	RecipientID          *string  `json:"recipientId,omitempty"`          // 收件人 ID，实体卡必填
}

func (r *OpenCardRequest) Validate() error {
	if r.CardBin == "" {
		return errors.New("CardBin不能为空")
	}
	if r.CardCurrency == "" {
		return errors.New("币种不能为空")
	}
	if r.CardScheme == "" {
		return errors.New("卡组织不能为空")
	}
	if r.CardType == "" {
		return errors.New("卡类型不能为空")
	}
	if r.CardFormFactor == "" {
		return errors.New("卡介质不能为空")
	}
	if r.CardholderID == "" {
		return errors.New("持卡人id不能为空")
	}
	if r.RequestID == "" {
		return errors.New("幂等id不能为空")
	}
	return nil
}

type GetRequestResultRequest struct {
	MemberID  *string `json:"memberId,omitempty"` // 请求所指向的会员号，如不传则默认为token下的会员；若涉及matrix产品，可传入指定的连接会员号
	RequestID string  `json:"requestId"`          // 幂等id,必填 商户请求流水号，每笔交易的唯一请求号，不可重复
	Type      string  `json:"type,omitempty"`     // 请求类型 Enum: "apply_card" "card_update" "card_freeze" 为空时默认查询开卡结果
}

func (r *GetRequestResultRequest) Validate() error {
	if r.RequestID == "" {
		return errors.New("requestId不能为空")
	}
	return nil
}

type GetCardDetailRequest struct {
	CardID string `json:"cardId,omitempty"` // 卡ID，每张卡的唯一编号
}

func (r *GetCardDetailRequest) Validate() error {
	if r.CardID == "" {
		return errors.New("CardID不能为空")
	}
	return nil
}

type PagingVccCardRequest struct {
	PageIndex      *int64  `json:"pageIndex,omitempty"`      // 当前页/默认第一页
	PageSize       *int64  `json:"pageSize,omitempty"`       // 分页大小
	MemberID       *string `json:"memberId,omitempty"`       // 查询所指向的会员号，如不传则默认为token下的会员；如涉及matrix产品，如不传则指向会员和连接会员所有信息查询，或传入指定会员查询。
	MatrixAccount  *string `json:"matrixAccount,omitempty"`  // Matrix账户ID
	CardBin        *string `json:"cardBin,omitempty"`        // 卡bin。您可输入卡bin进行筛选。如需筛选多个卡bin，每个卡bin之间逗号分隔即可。
	CreatedAtStart *string `json:"createdAtStart,omitempty"` // 起始时间 Example: createdAtStart=2022-03-18T08:43:28
	CreatedAtEnd   *string `json:"createdAtEnd,omitempty"`   // 结束时间 Example: createdAtEnd=2022-03-18T08:43:28
	CardType       *string `json:"cardType,omitempty"`       // 卡类型。share： 共享卡； recharge： 常规卡。
	CardFormFactor *string `json:"cardFormFactor,omitempty"` // 卡介质。值： virtual_card虚拟卡，physical_card实体卡。
	CardStatus     *string `json:"cardStatus,omitempty"`     // 卡状态
}

func (r *PagingVccCardRequest) Validate() error {
	if r.PageIndex != nil && *r.PageIndex <= 0 {
		return errors.New("pageIndex必须大于0")
	}
	if r.PageSize != nil && *r.PageSize <= 0 {
		return errors.New("pageSize必须大于0")
	}
	if r.CardType != nil {
		return errors.New("卡类型不能为空")
	}
	if r.CardFormFactor != nil {
		return errors.New("卡介质不能为空")
	}
	if r.CardStatus != nil {
		return errors.New("卡状态不能为空")
	}
	return nil
}

type GetCvvRequest struct {
	CardID string `json:"cardId,omitempty"` // 卡ID，每张卡的唯一编号
}

func (r *GetCvvRequest) Validate() error {
	if r.CardID == "" {
		return errors.New("CardID不能为空")
	}
	return nil
}

type UpdateCardRequest struct {
	CardID                     string   `json:"cardId"`                               // 卡ID，必填
	CardFormFactor             *string  `json:"cardFormFactor,omitempty"`             // 卡介质。值： virtual_card虚拟卡；physical_card实体卡。不填默认虚拟卡
	RequestID                  string   `json:"requestId"`                            // 商户请求流水号，每笔交易的唯一请求号，必填
	MaxOnDaily                 *int64   `json:"maxOnDaily,omitempty"`                 // 日交易限额，为空时不更新。已销的卡无法调整此限额
	MaxOnMonthly               *int64   `json:"maxOnMonthly,omitempty"`               // 月交易限额，为空时不更新。已销的卡无法调整此限额
	MaxOnPercent               *int64   `json:"maxOnPercent,omitempty"`               // 单笔交易最大金额，为空时不更新。已销的卡无法调整此限额
	Nickname                   *string  `json:"nickname,omitempty"`                   // 昵称，为空时不更新
	TransactionLimit           *float64 `json:"transactionLimit,omitempty"`           // 可交易额度变动金额，为空时不更新 (需遵循 ISO4217规范)
	TransactionLimitChangeType *string  `json:"transactionLimitChangeType,omitempty"` // 可交易额度变动类型，销卡后不可再操作调增，可以调减
	TransactionLimitType       *string  `json:"transactionLimitType,omitempty"`       // 是否限制可交易额度，为空时不更新，实体共享卡必为Limited
}

func (r *UpdateCardRequest) Validate() error {
	if r.CardID == "" {
		return errors.New("cardId不能为空")
	}
	if r.RequestID == "" {
		return errors.New("requestId不能为空")
	}
	if r.CardFormFactor != nil {
		return errors.New("卡介质不能为空")
	}
	if r.TransactionLimitChangeType != nil {
		return errors.New("可交易额度变动类型不能为空")
	}
	if r.TransactionLimitType != nil {
		return errors.New("可交易额度类型不能为空")
	}
	return nil
}

type EditCardBillingAddressRequest struct {
	CardID            string  `json:"cardId"`                      // 卡ID，必填
	BillingAddress    *string `json:"billingAddress,omitempty"`    // 账单地址信息中的详细地址，建议填写账单地址
	BillingCity       *string `json:"billingCity,omitempty"`       // 账单地址信息中的市，建议填写账单地址所在城市
	BillingCountry    *string `json:"billingCountry,omitempty"`    // 账单地址信息中的国家二字码，建议填写账单地址国家码二字码
	BillingPostalCode *string `json:"billingPostalCode,omitempty"` // 账单地址信息中的邮编，建议填写账单地址所属邮编
	BillingState      *string `json:"billingState,omitempty"`      // 账单地址信息中的省/州，建议填写账单地州省
}

func (r *EditCardBillingAddressRequest) Validate() error {
	if r.CardID == "" {
		return errors.New("cardId不能为空")
	}

	if r.BillingCountry != nil && len(*r.BillingCountry) != 2 {
		return errors.New("国家二字码错误")
	}

	return nil
}

type FreezeCardRequest struct {
	CardID    string `json:"cardId"`    // 卡ID，必填
	RequestID string `json:"requestId"` // 商户请求流水号，每笔交易的唯一请求号，必填
	Status    string `json:"status"`    // freeze：冻结卡； unfreeze：解冻卡，必填
}

func (r *FreezeCardRequest) Validate() error {
	if r.CardID == "" {
		return errors.New("cardId不能为空")
	}

	if r.RequestID == "" {
		return errors.New("requestId不能为空")
	}

	if r.Status == "" {
		return errors.New("type不能为空")
	}

	return nil
}

type CancelCardRequest struct {
	CardID string `json:"cardId"` // 卡ID，必填
}

func (r *CancelCardRequest) Validate() error {
	if r.CardID == "" {
		return errors.New("CardID不能为空")
	}
	return nil
}

type PreRechargeRequest struct {
	MemberID       *string  `json:"memberId,omitempty"`       // 请求所指向的会员号，如不传则默认为token下的会员；若涉及matrix产品，可传入指定的连接会员号。
	RequestID      string   `json:"requestId"`                // 商户请求流水号，每笔交易的唯一请求号，不可重复
	AccountID      string   `json:"accountId"`                // 每个账户的唯一ID号。您可在光子易账户接口中进行查询您的账户ID。
	CardID         string   `json:"cardId"`                   // 卡ID
	RechargeAmount *float64 `json:"rechargeAmount,omitempty"` // 转入金额，您想从币种光子易账户金额往卡里转入的资金。我们将为您计算出最后实际到卡内的金额。转入金额和到账金额择一填写即可 (需遵循 ISO4217规范)。
	ArrivalAmount  *float64 `json:"arrivalAmount,omitempty"`  // 到账金额，您希望转入到账金额是多少，我们将为您计算出需要扣取您所选光子易账户的金额。转入金额和到账金额择一填写即可 (需遵循 ISO4217规范)。
}

func (r *PreRechargeRequest) Validate() error {
	if r.RequestID == "" {
		return errors.New("requestId不能为空")
	}
	if r.AccountID == "" {
		return errors.New("accountId不能为空")
	}
	if r.CardID == "" {
		return errors.New("cardId不能为空")
	}
	if r.RechargeAmount == nil && r.ArrivalAmount == nil {
		return errors.New("转入金额和到账金额必须填写其中一个")
	}
	if r.RechargeAmount != nil && r.ArrivalAmount != nil {
		return errors.New("转入金额和到账金额只能填写其中一个")
	}
	return nil
}

type RechargeRequest struct {
	MemberID  *string `json:"memberId,omitempty"` // 请求所指向的会员号，如不传则默认为token下的会员；若涉及matrix产品，可传入指定的连接会员号。
	RequestID string  `json:"requestId"`          // 商户请求流水号，每笔交易的唯一请求号，不可重复
}

func (r *RechargeRequest) Validate() error {
	if r.RequestID == "" {
		return errors.New("requestId不能为空")
	}
	return nil
}

type RechargeReturnRequest struct {
	CardID       string  `json:"cardId"`       // 卡ID，必填
	RequestID    string  `json:"requestId"`    // 商户请求流水号，每笔交易的唯一请求号，不可重复
	ReturnAmount float64 `json:"returnAmount"` // 退回金额，必填 (需遵循 ISO4217规范)
}

func (r *RechargeReturnRequest) Validate() error {
	if r.CardID == "" {
		return errors.New("cardId不能为空")
	}
	if r.RequestID == "" {
		return errors.New("requestId不能为空")
	}
	if r.ReturnAmount <= 0 {
		return errors.New("退回金额必须大于0")
	}
	return nil
}

type PagingIssuingHistoryRequest struct {
	PageIndex      *int64  `json:"pageIndex,omitempty"`      // 当前页/默认第一页
	PageSize       *int64  `json:"pageSize,omitempty"`       // 分页大小
	MemberID       *string `json:"memberId,omitempty"`       // 查询所指向的会员号，如不传则默认为token下的会员；如涉及matrix产品，如不传则指向会员和连接会员所有信息查询，或传入指定会员查询。
	MatrixAccount  *string `json:"matrixAccount,omitempty"`  // Matrix账户ID
	CreatedAtStart *string `json:"createdAtStart,omitempty"` // 起始时间 Example: createdAtStart=2022-03-18T08:43:28
	CreatedAtEnd   *string `json:"createdAtEnd,omitempty"`   // 结束时间 Example: createdAtEnd=2022-03-18T08:43:28
	CardID         *string `json:"cardId,omitempty"`         // 卡ID
	CardFormFactor *string `json:"cardFormFactor,omitempty"` // 卡介质。值： virtual_card虚拟卡；physical_card实体卡。可为空。若为空，则查询结果显示所有虚拟卡及实体卡卡介质。
	Status         *string `json:"status,omitempty"`         // 状态 Enum: "pending" "succeed" "failed"
}

// 卡历史明细
func (r *PagingIssuingHistoryRequest) Validate() error {
	// if r.PageIndex != nil && *r.PageIndex <= 0 {
	// 	return errors.New("pageIndex必须大于0")
	// }
	// if r.PageSize != nil && *r.PageSize <= 0 {
	// 	return errors.New("pageSize必须大于0")
	// }
	// if r.CardFormFactor != nil  {
	// 	return errors.New("卡介质不能为空")
	// }

	// if r.Status.Valid() {
	// 	return errors.New("type不能为空")
	// }

	return nil
}

type PagingRechargeCardFundsDetailRequest struct {
	PageIndex       *int64  `json:"pageIndex,omitempty"`       // 当前页/默认第一页
	PageSize        *int64  `json:"pageSize,omitempty"`        // 分页大小
	MemberID        *string `json:"memberId,omitempty"`        // 查询所指向的会员号，如不传则默认为token下的会员；如涉及matrix产品，如不传则指向会员和连接会员所有信息查询，或传入指定会员查询。
	MatrixAccount   *string `json:"matrixAccount,omitempty"`   // Matrix账户ID
	CreatedAtStart  *string `json:"createdAtStart,omitempty"`  // 起始时间 Example: createdAtStart=2022-03-18T08:43:28
	CreatedAtEnd    *string `json:"createdAtEnd,omitempty"`    // 结束时间 Example: createdAtEnd=2022-03-18T08:43:28
	TransactionID   *string `json:"transactionId,omitempty"`   // 交易ID
	CardID          *string `json:"cardId,omitempty"`          // 卡ID
	CardFormFactor  *string `json:"cardFormFactor,omitempty"`  // 卡介质。值： virtual_card虚拟卡；physical_card实体卡。可为空。若为空，则查询结果显示所有虚拟卡及实体卡卡介质。
	TransactionType *string `json:"transactionType,omitempty"` // 交易类型
}

// 常规卡资金明细
func (r *PagingRechargeCardFundsDetailRequest) Validate() error {
	if r.PageIndex != nil && *r.PageIndex <= 0 {
		return errors.New("pageIndex必须大于0")
	}
	if r.PageSize != nil && *r.PageSize <= 0 {
		return errors.New("pageSize必须大于0")
	}
	if r.CardFormFactor != nil {
		return errors.New("卡介质不能为空")
	}
	if r.TransactionType != nil {
		return errors.New("交易类型不能为空")
	}
	return nil
}

type PagingShareCardTxnLimitDetailRequest struct {
	PageIndex       *int64  `json:"pageIndex,omitempty"`       // 当前页/默认第一页
	PageSize        *int64  `json:"pageSize,omitempty"`        // 分页大小
	MemberID        *string `json:"memberId,omitempty"`        // 查询所指向的会员号，如不传则默认为token下的会员；如涉及matrix产品，如不传则指向会员和连接会员所有信息查询，或传入指定会员查询。
	MatrixAccount   *string `json:"matrixAccount,omitempty"`   // Matrix账户ID
	CreatedAtStart  *string `json:"createdAtStart,omitempty"`  // 起始时间 Example: createdAtStart=2022-03-18T08:43:28
	CreatedAtEnd    *string `json:"createdAtEnd,omitempty"`    // 结束时间 Example: createdAtEnd=2022-03-18T08:43:28
	TransactionID   *string `json:"transactionId,omitempty"`   // 交易ID
	CardID          *string `json:"cardId,omitempty"`          // 卡ID
	CardFormFactor  *string `json:"cardFormFactor,omitempty"`  // 卡介质。值： virtual_card虚拟卡；physical_card实体卡。可为空。若为空，则查询结果显示所有虚拟卡及实体卡卡介质。
	TransactionType *string `json:"transactionType,omitempty"` // 交易类型 Enum: "auth" "verification" "auth_failed_return" "void" "refund" "refund_reversal" "corrective_auth" "corrective_refund" "corrective_refund_void" "limit_adjustment" "service_fee" "fund_in" "settlement_spread" "atm_inquiry" "atm_withdrawals" "atm_inquiry_failed_return" "atm_withdrawals_failed_return"
}

func (r *PagingShareCardTxnLimitDetailRequest) Validate() error {
	if r.PageIndex != nil && *r.PageIndex <= 0 {
		return errors.New("pageIndex必须大于0")
	}
	if r.PageSize != nil && *r.PageSize <= 0 {
		return errors.New("pageSize必须大于0")
	}
	if r.CardFormFactor != nil {
		return errors.New("卡介质不能为空")
	}
	if r.TransactionType != nil {
		return errors.New("交易类型不能为空")
	}
	return nil
}

type PagingVccTradeOrderRequest struct {
	PageIndex       *uint32 `json:"pageIndex,omitempty"`       // 当前页/默认第一页
	PageSize        *uint32 `json:"pageSize,omitempty"`        // 分页大小
	MemberID        *string `json:"memberId,omitempty"`        // 查询所指向的会员号，如不传则默认为token下的会员；如涉及matrix产品，如不传则指向会员和连接会员所有信息查询，或传入指定会员查询。
	MatrixAccount   *string `json:"matrixAccount,omitempty"`   // Matrix账户ID
	CreatedAtStart  *string `json:"createdAtStart,omitempty"`  // 起始时间 Example: createdAtStart=2022-03-18T08:43:28
	CreatedAtEnd    *string `json:"createdAtEnd,omitempty"`    // 结束时间 Example: createdAtEnd=2022-03-18T08:43:28
	CardID          *string `json:"cardId,omitempty"`          // 卡ID
	CardType        *string `json:"cardType,omitempty"`        // 卡类型。值: share 或 recharge 可为空。如为空，则查询结果将显示所有类型的卡。
	CardFormFactor  *string `json:"cardFormFactor,omitempty"`  // 卡介质。值： virtual_card虚拟卡；physical_card实体卡。可为空。若为空，则查询结果显示所有虚拟卡及实体卡卡介质。
	RequestID       *string `json:"requestId,omitempty"`       // 商户请求流水号，每笔交易的唯一请求号，不可重复。
	TransactionID   *string `json:"transactionId,omitempty"`   // 交易ID
	TransactionType *string `json:"transactionType,omitempty"` // 您可对交易类型进行筛选，多个交易类型用逗号分隔。
	Status          *string `json:"status,omitempty"`          // 您可对交易状态进行筛选，多个交易状态用逗号分隔。
}

// 交易明细
func (r *PagingVccTradeOrderRequest) Validate() error {
	if r.PageIndex != nil && *r.PageIndex <= 0 {
		return errors.New("pageIndex必须大于0")
	}
	if r.PageSize != nil && *r.PageSize <= 0 {
		return errors.New("pageSize必须大于0")
	}
	if r.CardFormFactor != nil {
		return errors.New("卡介质不能为空")
	}
	if r.TransactionType != nil {
		return errors.New("交易类型不能为空")
	}
	return nil
}

type SandBoxTransactionRequest struct {
	RequestID           string  `json:"requestId"`           // 商户请求流水号，必填
	CardID              string  `json:"cardID"`              // 每张卡的唯一 ID 号，必填
	Cvv                 string  `json:"cvv"`                 // CVV，必填
	ExpirationDate      string  `json:"expirationDate"`      // 卡的有效期 MM/YY，必填
	OriginTransactionID string  `json:"originTransactionId"` // 当交易类型为 void 或 refund 时要填写
	TxnCurrency         string  `json:"txnCurrency"`         // ISO 4217 货币代码，必填
	TxnAmount           float64 `json:"txnAmount"`           // 交易金额，必填
	TxnType             string  `json:"txnType"`             // 交易类型 auth/void/refund，必填
	Mcc                 string  `json:"mcc"`                 // MCC，必填
	MerchantName        string  `json:"merchantName"`        // 商户名称，必填
	MerchantCountry     string  `json:"merchantCountry"`     // 商户国家二字码，必填
	MerchantCity        string  `json:"merchantCity"`        // 商户城市，必填
	MerchantPostcode    string  `json:"merchantPostcode"`    // 商户邮编，必填
}

func (r *SandBoxTransactionRequest) Validate() error {
	if r.RequestID == "" {
		return errors.New("requestId不能为空")
	}
	if r.CardID == "" {
		return errors.New("cardId不能为空")
	}
	if r.Cvv == "" {
		return errors.New("cvv不能为空")
	}
	if r.ExpirationDate == "" {
		return errors.New("expirationDate不能为空")
	}
	if len(r.ExpirationDate) != 5 || r.ExpirationDate[2] != '/' {
		return errors.New("expirationDate格式必须为MM/YY")
	}
	if r.TxnCurrency == "" {
		return errors.New("txnCurrency不能为空")
	}
	if r.TxnAmount == 0 {
		return errors.New("txnAmount不能为0")
	}
	if r.TxnType == "" {
		return errors.New("卡介质不能为空")
	}
	if r.Mcc == "" {
		return errors.New("mcc不能为空")
	}
	if r.MerchantName == "" {
		return errors.New("merchantName不能为空")
	}
	if len(r.MerchantCountry) != 2 {
		return errors.New("merchantCountry必须为2位国家代码")
	}
	if r.MerchantCity == "" {
		return errors.New("merchantCity不能为空")
	}
	if r.MerchantPostcode == "" {
		return errors.New("merchantPostcode不能为空")
	}
	return nil
}

type ApiUploadFileRequest struct {
	FileReader  io.ReadCloser `json:"fileReader"`  // 文件，必填
	FileName    string        `json:"fileName"`    // 文件名，必填
	BusinessKey string        `json:"businessKey"` // 业务类型，必填
}

func (r *ApiUploadFileRequest) Validate() error {
	if r.FileReader == nil {
		return errors.New("fileReader不能为空")
	}
	if r.FileName == "" {
		return errors.New("文件名不能为空")
	}
	if r.BusinessKey == "" {
		return errors.New("文件上传业务类型不能为空")
	}
	return nil
}

type WebhookNotificationRequest struct {
	TopicCode    string `json:"topicCode"`    // Webhook 通知 Topic 的唯一标识
	TemplateCode string `json:"templateCode"` // Webhook 通知模版的唯一标识，该模版必须归属于指定的 Topic
}

func (r *WebhookNotificationRequest) Validate() error {
	if r.TopicCode == "" {
		return errors.New("Topic标识不能为空")
	}
	if r.TemplateCode == "" {
		return errors.New("模版的标识不能为空")
	}
	return nil
}
