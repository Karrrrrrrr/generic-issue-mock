package photonpay

import (
	"context"
	"fmt"
	"net/http"
	"sdk/middleware/trace"
	"sdk/middleware/traffic"
	"strconv"
	"strings"
	"testing"
	"time"
	"tman/enums"
	"tman/pkg/types"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
)

var (
	pp  *PhotonPaySDK
	ctx = context.Background()
)

const (
	balanceAccountID1 = "FA-USD1981575716742991872" // 测试账号

	token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJyYW5kb21LZXkiOiIwMjk2YzlkNS1jYTBhLTQxZTItYTJmZi1kODNlMjZmNGY5OGMiLCJtZW1iZXJObyI6IjIwMjUxMDI0NTMwNTM3ODAiLCJhcHBJZCI6ImhUdlYyYmEyIiwidXNlclR5cGUiOiJvdXRlciIsInRpbWUiOnsibmFubyI6ODkxMDAwMDAwLCJkYXlPZlllYXIiOjE3NywiZGF5T2ZXZWVrIjoiRlJJREFZIiwibW9udGgiOiJKVU5FIiwiZGF5T2ZNb250aCI6MjYsInllYXIiOjIwMjYsIm1vbnRoVmFsdWUiOjYsImhvdXIiOjcsIm1pbnV0ZSI6Niwic2Vjb25kIjo1OCwiY2hyb25vbG9neSI6eyJjYWxlbmRhclR5cGUiOiJpc284NjAxIiwiaWQiOiJJU08ifX0sInVzZXJJZCI6IiJ9.6QL5SD9EjRtvecxvw08thOa8qw5hEAQCpvHFSB1G3jk"
)

func TestMain(m *testing.M) {
	conf := &testConf{
		appId:               "hTvV2ba2",
		appSecret:           "821b1b1ea7fb1c821ea81c908282d2908201164f",
		thirdPartyPublicKey: `MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCwxNbhQM1yOBgeqgfl6/wzp8f+84eY1scI+dvQotuVrMo5E6b5y4P2QnQTWiTlaHCKehiTjAiMCji4XbhJpTdNQiQ2OoQeOgEm9NSRiW/IKYIYFaLFjBo2mXN+AqEpP6dpP/SG/dG4THOmVR1nlaxz0uPncSUA6uVaDBDtEYE7zwIDAQAB`,
		publicKey:           `MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCXvtg1u4YnpX7pfG3pRh+aJCxLFaEeQm0+Va9vMJSBNf433MqXkw8r6At6hioD0aLMfa8ucwyAlMRwNP28Jk8n5cIe6C8F09qsPZDecA0UKLciZX8GeJBJnV10VW0XqCmJRels4fiecsHGc199YVNYXeRZkwf4jAa/FweD73hAOwIDAQAB`,
		privateKey: `-----BEGIN PRIVATE KEY-----
MIICdQIBADANBgkqhkiG9w0BAQEFAASCAl8wggJbAgEAAoGBAJe+2DW7hielful8
belGH5okLEsVoR5CbT5Vr28wlIE1/jfcypeTDyvoC3qGKgPRosx9ry5zDICUxHA0
/bwmTyflwh7oLwXT2qw9kN5wDRQotyJlfwZ4kEmdXXRVbReoKYlF6Wzh+J5ywcZz
X31hU1hd5FmTB/iMBr8XB4PveEA7AgMBAAECgYA6kO/FnUCj4J3g46NQGz5rMXbe
69QpZ53eJxf0pB1M2Vfqm46dfaanXYHAojNpEenxXrjUBpdWsRQ38lvT2D1IIqFl
PEU5dxLItwNC925G/z/krFJ8jHuppBq3SwW1DzrIQquqyXlCkgcd9vWNOTblhdE8
6bRmftYcMK8TB81S4QJBAMbp+82Tk1el24Z0k+HkeY0q9s8aYlqGc8RjuU0Jytcj
VKZgh+KjGvVM6GhK+hp8J0jMr1M3FUWemdkoNCboxgkCQQDDS2/IqcLNX8Aoz0+w
k9yZ5iBUrQ7X5d03zjI9g01AmXiPi1lq5jkRTpHU4IeUm+VE4EFi7CbUhPrY1PiN
TwUjAkBUFqw9Hsrl/ZaNA5FUqFp+RBBsQtIbRMWB20qFd8NJKYVqhFpNg/gshOOm
2zNZqOyOiQEBI8MZWE/fjnBllnfJAkA/+1P5GtzuagNLm3fRMvAgH4vSEgx94RoT
sWM2UfEaS+16ob2+zwQ9Tk9qvdDNeDGp2gqx/QpPr+164nM34H6fAkB/f567Kyz5
JBdFyxk6j4OAwapgKlESMnTON61n2PBD1d6VnzUtFS/gV5H7bgRrlQTtW6TH44Xg
nHySLt+myp8f
-----END PRIVATE KEY-----`,
		balanceAccount: balanceAccountID1,
		photonPayUrl:   "http://test-photonpay.ptmlink.top",
		debug:          true,
	}
	l := log.DefaultLogger
	pp = New(conf, WithResponseMiddlewares(trace.ResponseTracking(), traffic.PhotonTrafficLog(l)))
	m.Run()
}

type testConf struct {
	appId               string
	appSecret           string
	thirdPartyPublicKey string
	publicKey           string
	privateKey          string
	balanceAccount      string
	photonPayUrl        string
	debug               bool
}

func (c *testConf) GetAppId() string {
	return c.appId
}

func (c *testConf) GetAppSecret() string {
	return c.appSecret
}

func (c *testConf) GetThirdPartyPublicKey() string {
	return c.thirdPartyPublicKey
}
func (c *testConf) GetPublicKey() string {
	return c.publicKey
}
func (c *testConf) GetPrivateKey() string {
	return c.privateKey
}
func (c *testConf) GetBalanceAccount() string {
	return c.balanceAccount
}

func (c *testConf) GetUrl() string {
	return c.photonPayUrl
}

func (c *testConf) GetDebug() bool {
	return c.debug
}

// TestPhotonPaySDK_GetAccessToken 测试获取访问令牌
func TestPhotonPaySDK_GetAccessToken(t *testing.T) {
	result, err := pp.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("获取token失败: %v", err)
	}

	t.Logf("v: %+v\n", result)
}

// TestPhotonPaySDK_GetAccountSingle 测试获取账户钱包
func TestPhotonPaySDK_GetAccountSingle(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *AccountSingleRequest
	}{
		{
			name: "获取商户钱包列表",
			req: &AccountSingleRequest{
				Currency:    types.StringValueToPtr("USD"),
				AccountType: types.StringValueToPtr(enums.PHOTON_PAY_ACCOUNT_TYPE_ENUM_FT10001.ToString()),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			result, err := pp.GetAccountSingle(context.Background(), token, tt.req)
			assert.NoError(t, err)
			fmt.Printf("v: %+v\n", result)
		})
	}
}

// TestPhotonPaySDK_AccountHistory 光子易账户账单查询
func TestPhotonPaySDK_AccountHistory(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *AccountHistoryRequest
	}{
		{
			name: "光子易账户账单查询",
			req: &AccountHistoryRequest{
				TransactedAtStart: "2025-11-01T00:00:00",
				TransactedAtEnd:   "2025-11-30T23:59:59",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			result, err := pp.AccountHistory(context.Background(), token, tt.req)
			assert.NoError(t, err)
			fmt.Printf("v: %+v\n", result)
		})
	}
}

// TestPhotonPaySDK_AddCardholder 添加用卡人
func TestPhotonPaySDK_AddCardholder(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *AddCardholderRequest
	}{
		{
			name: "添加用卡人",
			req: &AddCardholderRequest{
				FirstName:              " ZHANG ",
				LastName:               " SI ",
				Email:                  "13800138003test@ptmlink.com",
				Mobile:                 "13800138003",
				MobilePrefix:           "+86",
				DateOfBirth:            "1990-01-01",
				NationalityCountryCode: "HK",
				CertType:               types.StringValueToPtr("id_card"),        // 身份证件类型。id_card：身份证，passport：护照，resident_permit：居留许可证(永居、绿卡、工作签证）。
				Portrait:               types.StringValueToPtr("f59omF3pUqckcg"), // 身份证件信息正面：请提供身份证件正面或护照首页照片。 仅限一张照片；PNG，JPG；单个文件最大6M
				ReverseSide:            types.StringValueToPtr("f59omF3pUqckcg"), // 身份证件信息反面：请提供身份证件反面照片。仅限一张照片；PNG，JPG；单个文件最大6M。
				ResidentialAddress:     types.StringValueToPtr("ewrwe"),          // 账单地址。建议填写账单地址。如需使用 Discover 的卡，此字段为必填。
				ResidentialCity:        types.StringValueToPtr("fsfs"),           // 账单地城市。建议填写账单地址所在城市。如需使用 Discover 的卡，此字段为必填。
				ResidentialCountryCode: types.StringValueToPtr("US"),             // 账单地国家码二字码。建议填写账单地址国家码二字码。如需使用 Discover 的卡，此字段为必填。
				ResidentialPostalCode:  types.StringValueToPtr("11111"),          // 账单地邮编。建议填写账单地址所属邮编。如需使用 Discover 的卡，此字段为必填。
				ResidentialState:       types.StringValueToPtr("AK"),             // 账单地州省。建议填写账单地州省。如需使用 Discover 的卡，此字段为必填。
				CertCountryCode:        types.StringValueToPtr("US"),             // 证件签发国国家二字码。
				CertID:                 types.StringValueToPtr("2454234234"),     // 证件号。
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			result, err := pp.AddCardholder(context.Background(), token, tt.req)
			if err != nil {
				t.Fatalf("添加用卡人失败: %v", err)
			}

			t.Logf("v: %+v\n", result)
		})
	}
}

// TestPhotonPaySDK_EditCardholder 更新用卡人
func TestPhotonPaySDK_EditCardholder(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *EditCardholderRequest
	}{
		// {
		// 	name: "用卡人赵星星",
		// 	req: &EditCardholderRequest{
		// 		CardholderID:       "CH1983724491523497984",
		// 		ResidentialAddress: types.StringValueToPtr("测试地址更新22222"),
		// 		// Email:              types.StringValueToPtr("13800138003-1test@ptmlink.com"),
		// 	},
		// },
		{
			name: "更新用卡人 zhang san",
			req: &EditCardholderRequest{
				CardholderID:       "CH1987772235057729536",
				ResidentialAddress: types.StringValueToPtr("测试地址更新22222"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			result, err := pp.EditCardholder(context.Background(), token, tt.req)
			if err != nil {
				t.Fatalf("更新用卡人: %v", err)
			}
			t.Logf("v: %+v\n", result)
		})
	}
}

// TestPhotonPaySDK_PagingVccCardholder 用卡人查询
func TestPhotonPaySDK_PagingVccCardholder(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *PagingVccCardholderRequest
	}{
		{
			name: "查询卡持有人列表",
			req:  &PagingVccCardholderRequest{},
		},
		// {
		// 	name: "用卡人赵星星",
		// 	req: &PagingVccCardholderRequest{
		// 		CardholderID: types.StringValueToPtr("CH1983724491523497984"),
		// 	},
		// },
		// {
		// 	name: "用卡人 zhang san",
		// 	req: &PagingVccCardholderRequest{
		// 		CardholderID: types.StringValueToPtr("CH1987772235057729536"),
		// 	},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			result, err := pp.PagingVccCardholder(context.Background(), token, tt.req)
			if err != nil {
				var kratosErr *errors.Error
				if errors.As(err, &kratosErr) {
					// 检查错误码是否为 1002 (invalid token)
					if kratosErr.Reason == "1002" {
						// 执行特定操作，例如：
						fmt.Println("检测到无效token错误，需要重新获取token")
						// 可以在这里执行重新获取token的逻辑
						// 或者记录日志、报警等操作
					}
				}
			}
			assert.NoError(t, err)
			fmt.Printf("v: %+v\n", result)
		})
	}
}

// TestPhotonPaySDK_GetCardBin 卡bin查询
func TestPhotonPaySDK_GetCardBin(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *CardBinRequest
	}{
		{
			name: "卡bin查询",
			req: &CardBinRequest{
				CardFormFactor: types.StringValueToPtr("virtual_card"),                                                    // 卡介质。值： virtual_card虚拟卡；physical_card实体卡。不填默认虚拟卡
				CardType:       types.StringValueToPtr(strings.ToLower(enums.PHOTON_PAY_CARD_TYPE_ENUM_SHARE.ToString())), // 卡类型。share： 共享卡； recharge： 常规卡
				CardScheme:     types.StringValueToPtr(""),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			result, err := pp.GetCardBin(context.Background(), token, tt.req)
			assert.NoError(t, err)
			fmt.Printf("v: %+v\n", result)
		})
	}
}

// TestPhotonPaySDK_OpenCard 单卡开卡
func TestPhotonPaySDK_OpenCard(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *OpenCardRequest
	}{
		{
			name: "单卡开卡",
			req: &OpenCardRequest{
				CardBin:              "404038",                                         // 您需要填入你想开卡的卡bin信息，目前支持的卡bin信息可在卡bin接口中查询
				CardCurrency:         "USD",                                            // 卡本币
				CardExpirationDate:   types.IntValueToPtr(24),                          // 卡有效期。您可填写此卡的有效期，以月为计数单位。如：您需要虚拟卡的有效期为1年，则输入“12”即可。如不上传，则由系统自动分配虚拟卡有效期。（最小12个月，最大35个月）
				CardScheme:           "VISA",                                           // 卡组织 Enum: "MasterCard" "Discover"
				CardType:             enums.PHOTON_PAY_CARD_TYPE_ENUM_SHARE.ToString(), // 卡类型。share： 共享卡； recharge： 常规卡
				CardFormFactor:       "virtual_card",                                   // 卡介质。值： virtual_card虚拟卡；physical_card实体卡。不填默认虚拟卡
				CardholderID:         "CH1981575723567030272",                          // 填入用卡人id后，卡将属于此用卡人。如不填则默认使用默认持卡人信息进行开卡。
				TransactionLimitType: types.StringValueToPtr("unlimited"),              // 是否限制可交易额度，Enum: "limited" "unlimited"recharge常规卡、share虚拟共享卡，默认unlimited；实体共享卡必为Limited
				AccountID:            types.StringValueToPtr(balanceAccountID1),        // 您需要用于转入的币种光子易账户ID号
				ArrivalAmount:        types.Float64ValueToPtr(1),                       // 到账金额  转入金额和到账金额择一填写即可
				RequestID:            strconv.Itoa(int(time.Now().Unix())),             // 幂等id,三方非必填,我方必填 商户请求流水号，每笔交易的唯一请求号，不可重复
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			result, err := pp.OpenCard(context.Background(), token, tt.req)
			assert.NoError(t, err)
			fmt.Printf("v: %+v\n", result)
		})
	}
}

// TestPhotonPaySDK_GetRequestResult 请求结果查询
func TestPhotonPaySDK_GetRequestResult(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *GetRequestResultRequest
	}{
		// {
		// 	name: "请求结果查询",
		// 	req: &GetRequestResultRequest{
		// 		RequestID: "1762776232", // 幂等id,三方非必填,我方必填 商户请求流水号，每笔交易的唯一请求号，不可重复
		// 		Type:      HANDLE_CARD_TYPE_APPLY_CARD,
		// 	},
		// },
		//{
		//	name: "请求结果查询-开卡",
		//	req: &GetRequestResultRequest{
		//		RequestID: "1762831072", // 幂等id,三方非必填,我方必填 商户请求流水号，每笔交易的唯一请求号，不可重复
		//		Type:      enums.PHOTON_PAY_HANDLE_CARD_TYPE_APPLY_CARD.ToString(),
		//	},
		//},
		{
			name: "请求结果查询-冻结卡",
			req: &GetRequestResultRequest{
				RequestID: "20464740907018895514",
				Type:      enums.PHOTON_PAY_HANDLE_CARD_TYPE_CARD_FREEZE.ToString(),
			},
		},
		// {
		// 	name: "给共享卡添加金额",
		// 	req: &GetRequestResultRequest{
		// 		RequestID: "1762832704", // 幂等id,三方非必填,我方必填 商户请求流水号，每笔交易的唯一请求号，不可重复
		// 		Type:      HANDLE_CARD_TYPE_APPLY_CARD,
		// 	},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			result, err := pp.GetRequestResult(context.Background(), token, tt.req)
			assert.NoError(t, err)
			fmt.Printf("v: %+v\n", result)
			fmt.Printf("status:%s", result.CardDetail.CardStatus)
		})
	}
}

// TestPhotonPaySDK_GetCardDetail 卡信息查询
func TestPhotonPaySDK_GetCardDetail(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *GetCardDetailRequest
	}{
		// {
		// 	name: "共享卡信息查询",
		// 	req: &GetCardDetailRequest{
		// 		CardID: "XR2013964643671740416",
		// 	},
		// },
		{
			name: "共享卡信息查询",
			req: &GetCardDetailRequest{
				CardID: "XR1993238479990951936",
			},
		},
		// {
		// 	name: "常规卡信息查询",
		// 	req: &GetCardDetailRequest{
		// 		CardID: "XR1985590717782695936",
		// 	},
		// },
		// {
		// 	name: "给共享卡添加金额",
		// 	req: &GetCardDetailRequest{
		// 		CardID: "XR1988090585541513216",
		// 	},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			result, err := pp.GetCardDetail(context.Background(), token, tt.req)
			assert.NoError(t, err)
			fmt.Printf("v: %+v\n", result)
		})
	}
}

// TestPhotonPaySDK_PagingVccCard 卡列表
func TestPhotonPaySDK_PagingVccCard(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *PagingVccCardRequest
	}{
		{
			name: "查询卡列表",
			req: &PagingVccCardRequest{
				PageIndex: types.Int64ValueToPtr(34),
				PageSize:  types.Int64ValueToPtr(10),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			result, err := pp.PagingVccCard(context.Background(), token, tt.req)
			assert.NoError(t, err)
			fmt.Printf("v: %+v\n", result)
		})
	}
}

// TestPhotonPaySDK_GetCvv CVV查询
func TestPhotonPaySDK_GetCvv(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *GetCvvRequest
	}{
		{
			name: "CVV查询",
			req: &GetCvvRequest{
				CardID: "XR1988149731867496448",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			result, err := pp.GetCvv(context.Background(), token, tt.req)
			assert.NoError(t, err)
			fmt.Printf("v: %+v\n", result)
		})
	}
}

// TestPhotonPaySDK_UpdateCard 卡更新
func TestPhotonPaySDK_UpdateCard(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *UpdateCardRequest
	}{
		{
			name: "更新卡信息",
			req: &UpdateCardRequest{
				CardID:       "XR2014619194435305472",
				RequestID:    strconv.Itoa(int(time.Now().Unix())),
				MaxOnMonthly: types.Int64ValueToPtr(0),
				MaxOnDaily:   types.Int64ValueToPtr(0),
				MaxOnPercent: types.Int64ValueToPtr(0),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			result, err := pp.UpdateCard(context.Background(), token, tt.req)
			assert.NoError(t, err)
			fmt.Printf("v: %+v\n", result)
		})
	}
}

// TestPhotonPaySDK_EditCardBillingAddress 更新账单地址
func TestPhotonPaySDK_EditCardBillingAddress(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *EditCardBillingAddressRequest
	}{
		{
			name: "更新账单地址",
			req: &EditCardBillingAddressRequest{
				CardID:            "XR1994224038272045056",
				BillingAddress:    types.StringValueToPtr("ewiriweiio"),
				BillingCity:       types.StringValueToPtr("ewiriweiio"), // 账单地址信息中的市，建议填写账单地址所在城市
				BillingCountry:    types.StringValueToPtr("US"),         // 账单地址信息中的国家二字码，建议填写账单地址国家码二字码
				BillingPostalCode: types.StringValueToPtr("22222"),      // 账单地址信息中的邮编，建议填写账单地址所属邮编
				BillingState:      types.StringValueToPtr("AK"),         // 账单地址信息中的省/州，建议填写账单地州省
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			err := pp.EditCardBillingAddress(context.Background(), token, tt.req)
			assert.NoError(t, err)
		})
	}
}

// TestPhotonPaySDK_FreezeCard 冻结
func TestPhotonPaySDK_FreezeCard(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *FreezeCardRequest
	}{
		{
			name: "冻结卡",
			req: &FreezeCardRequest{
				CardID:    "XR1993238479990951936",
				RequestID: "20464740907018895514",
				Status:    "unfreeze",
			},
		},
		// {
		// 	name: "解冻卡",
		// 	req: &FreezeCardRequest{
		// 		CardID:    "XR1988434935609823232",
		// 		RequestID: strconv.Itoa(int(time.Now().Unix())),
		// 		Status:    "unfreeze",
		// 	},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			err := pp.FreezeCard(context.Background(), token, tt.req)
			assert.NoError(t, err)
		})
	}
}

// TestPhotonPaySDK_CancelCard 销卡
func TestPhotonPaySDK_CancelCard(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *CancelCardRequest
	}{
		{
			name: "销卡",
			req: &CancelCardRequest{
				CardID: "XR2012132673811648512",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			err := pp.CancelCard(context.Background(), token, tt.req)
			assert.NoError(t, err)
		})
	}
}

// TestPhotonPaySDK_PreRecharge 换汇询价
func TestPhotonPaySDK_PreRecharge(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *PreRechargeRequest
	}{
		// {
		// 	name: "共享卡-换汇询价",
		// 	req: &PreRechargeRequest{
		// 		RequestID:      strconv.Itoa(int(time.Now().Unix())),
		// 		AccountID:      balanceAccountID1,
		// 		CardID:         "XR1987853725120598016",
		// 		RechargeAmount: types.Float64ValueToPtr(100),
		// 	},
		// },
		{
			name: "常规卡-换汇询价",
			req: &PreRechargeRequest{
				RequestID:      strconv.Itoa(int(time.Now().Unix())),
				AccountID:      balanceAccountID1,
				CardID:         "XR1985590717782695936",
				RechargeAmount: types.Float64ValueToPtr(100),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			result, err := pp.PreRecharge(context.Background(), token, tt.req)
			assert.NoError(t, err)
			fmt.Printf("v: %+v\n", result)
		})
	}
}

// TestPhotonPaySDK_Recharge 转入下单
func TestPhotonPaySDK_Recharge(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *RechargeRequest
	}{
		{
			name: "转入下单",
			req: &RechargeRequest{
				RequestID: "1764294019", // 使用之前预充值的requestID
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			result, err := pp.Recharge(context.Background(), token, tt.req)
			assert.NoError(t, err)
			fmt.Printf("v: %+v\n", result)
		})
	}
}

// TestPhotonPaySDK_RechargeReturn 卡金额退还
func TestPhotonPaySDK_RechargeReturn(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *RechargeReturnRequest
	}{
		{
			name: "卡金额退还",
			req: &RechargeReturnRequest{
				CardID:       "XR1985590717782695936",
				RequestID:    strconv.Itoa(int(time.Now().Unix())),
				ReturnAmount: 10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			result, err := pp.RechargeReturn(context.Background(), token, tt.req)
			assert.NoError(t, err)
			fmt.Printf("v: %+v\n", result)
		})
	}
}

// TestPhotonPaySDK_PagingIssuingHistory 卡历史明细
func TestPhotonPaySDK_PagingIssuingHistory(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *PagingIssuingHistoryRequest
	}{
		{
			name: "卡历史明细",
			req: &PagingIssuingHistoryRequest{
				CardID:    types.StringValueToPtr("XR1985590717782695936"),
				PageIndex: types.Int64ValueToPtr(1),
				PageSize:  types.Int64ValueToPtr(10),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			result, err := pp.PagingIssuingHistory(context.Background(), token, tt.req)
			assert.NoError(t, err)
			fmt.Printf("v: %+v\n", result)
		})
	}
}

// TestPhotonPaySDK_PagingRechargeCardFundsDetail 常规卡资金明细
func TestPhotonPaySDK_PagingRechargeCardFundsDetail(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *PagingRechargeCardFundsDetailRequest
	}{
		{
			name: "常规卡资金明细",
			req: &PagingRechargeCardFundsDetailRequest{
				CardID: types.StringValueToPtr("XR1985590717782695936"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			result, err := pp.PagingRechargeCardFundsDetail(context.Background(), token, tt.req)
			assert.NoError(t, err)
			fmt.Printf("v: %+v\n", result)
		})
	}
}

// TestPhotonPaySDK_PagingShareCardTxnLimitDetail 共享卡交易额度明细
func TestPhotonPaySDK_PagingShareCardTxnLimitDetail(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *PagingShareCardTxnLimitDetailRequest
	}{
		{
			name: "共享卡交易额度明细",
			req: &PagingShareCardTxnLimitDetailRequest{
				CardID: types.StringValueToPtr("XR1987853725120598016"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			result, err := pp.PagingShareCardTxnLimitDetail(context.Background(), token, tt.req)
			assert.NoError(t, err)
			fmt.Printf("v: %+v\n", result)
		})
	}
}

// TestPhotonPaySDK_PagingVccTradeOrder 交易明细
func TestPhotonPaySDK_PagingVccTradeOrder(t *testing.T) {
	// local, _ := time.LoadLocation("Asia/Shanghai")
	// startTime, _ := time.ParseInLocation(time.DateTime, "2026-01-04 17:00:00", local)
	// endTime, _ := time.ParseInLocation(time.DateTime, "2026-01-05 18:00:00", local)

	// CreatedAtStart := types.StringValueToPtr(startTime.UTC().Format("2006-01-02T15:04:05"))
	// CreatedAtEnd := types.StringValueToPtr(endTime.UTC().Format("2006-01-02T15:04:05"))
	// 测试用例设计
	tests := []struct {
		name string
		req  *PagingVccTradeOrderRequest
	}{
		{
			name: "交易明细",
			req: &PagingVccTradeOrderRequest{
				PageIndex: types.Uint32ValueToPtr(1),
				PageSize:  types.Uint32ValueToPtr(1000),
				// CreatedAtStart: CreatedAtStart,
				// CreatedAtEnd:   CreatedAtEnd,
				CardID: types.StringValueToPtr("XR2012132681235566592"),
				// TransactionID: types.StringValueToPtr("IT1996432796838526976"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			result, err := pp.PagingVccTradeOrder(context.Background(), token, tt.req)
			assert.NoError(t, err)
			fmt.Printf("v: %+v\n", result)
		})
	}
}

// TestPhotonPaySDK_SandBoxTransaction 交易模拟
func TestPhotonPaySDK_SandBoxTransaction(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *SandBoxTransactionRequest
	}{
		{
			name: "交易模拟",
			req: &SandBoxTransactionRequest{
				RequestID: strconv.Itoa(int(time.Now().Unix())), // 商户请求流水号，必填
				// CardID:    "XR1994219563427827712",              // 曹晴晴测试
				CardID:         "XR2034188055010607104", // 每张卡的唯一 ID 号，必填
				Cvv:            "123",                   // CVV，必填
				ExpirationDate: "12/24",                 // 卡的有效期 MM/YY，必填
				// OriginTransactionID: "IT2004957183892000768",              // 当交易类型为 void 或 refund 时要填写
				TxnCurrency:      "USD",     // ISO 4217 货币代码，必填
				TxnAmount:        5,         // 交易金额，必填
				TxnType:          "auth",    // 交易类型 auth/void/refund,必填
				Mcc:              "1234",    // MCC，必填
				MerchantName:     "KFC",     // 商户名称，必填
				MerchantCountry:  "HK",      // 商户国家二字码，必填
				MerchantCity:     "jiulong", // 商户城市，必填
				MerchantPostcode: "999077",  // 商户邮编，必填
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			err := pp.SandBoxTransaction(context.Background(), token, tt.req)
			assert.NoError(t, err)
		})
	}
}

// TestPhotonPaySDK_PagingVccTradeOrder 文件上传
func TestPhotonPaySDK_ApiUploadFile(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "文件上传",
			url:  "https://static-test.ptmlink.com/admin/File/2024/11/18/133340.313ZlFV_HmRIdmi_dsrQGoOO4.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			parts := strings.Split(tt.url, "/")
			fileName := parts[len(parts)-1]
			// 获取文件
			file, _ := http.Get(tt.url)
			defer file.Body.Close()

			req := &ApiUploadFileRequest{
				FileReader:  file.Body,
				FileName:    fileName,
				BusinessKey: enums.PHOTON_PAY_UPLOAD_FILE_BUSINESS_KEY_CARDHOLDER.ToString(),
			}
			// 调用被测方法
			result, err := pp.ApiUploadFile(context.Background(), token, req)
			assert.NoError(t, err)
			fmt.Printf("v: %+v\n", result)
		})
	}
}

// TestPhotonPaySDK_GetWebhookNotification 查询 Webhook 订阅通知
func TestPhotonPaySDK_GetWebhookNotification(t *testing.T) {
	// 调用被测方法
	result, err := pp.GetWebhookNotification(context.Background(), token)
	assert.NoError(t, err)
	fmt.Printf("v:\n %+v\n", result)
}

// TestPhotonPaySDK_SetWebhookNotification 订阅 Webhook 通知
func TestPhotonPaySDK_SetWebhookNotification(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *WebhookNotificationRequest
	}{
		{
			name: "订阅 持卡人审核状态通知",
			req: &WebhookNotificationRequest{
				TopicCode:    "issuing_cardholder_review_status_update_topic",
				TemplateCode: "issuing_cardholder_review_status_update_template",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			err := pp.SetWebhookNotification(context.Background(), token, tt.req)
			assert.NoError(t, err)
		})
	}
}

// TestPhotonPaySDK_DelWebhookNotification 取消订阅 Webhook 通知
func TestPhotonPaySDK_DelWebhookNotification(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		req  *WebhookNotificationRequest
	}{
		{
			name: "取消订阅 持卡人审核状态通知",
			req: &WebhookNotificationRequest{
				TopicCode:    "issuing_cardholder_review_status_update_topic",
				TemplateCode: "issuing_cardholder_review_status_update_template",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测方法
			err := pp.DelWebhookNotification(context.Background(), token, tt.req)
			assert.NoError(t, err)
		})
	}
}
