package payndapay

import (
	"context"
	"fmt"
	"os"
	"sdk/crypto"
	"sdk/middleware/trace"
	"sdk/middleware/traffic"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

var (
	pp    *PayndaPaySDK
	ppSDK PayndaPaySDKInterface
	ctx   = context.Background()
)

const (
	pd1BalanceAccountID1    = "1960889588797104130"
	pd1BalanceAccountID2    = "1963485582302380034" // 测试账号
	pd1BalanceAccountID4    = "1961375380007002114" // pre
	pd1BalanceAccountID3    = "1960651159362560002"
	pd1BalanceAccountIDProd = "1960642666765168642"

	pd2BalanceAccountIDTest = "2092152689419726849" // 测试账号
	pd2BalanceAccountIDPre  = "2096915104707813377" // pd2 pre
)

func TestMain(m *testing.M) {
	conf := &testConf{
		appID:          "appId_WaXRKyF5s8JH",
		appSecret:      "appSecret_JaARu7mS8V3N",
		balanceAccount: pd1BalanceAccountID2, // pd1BalanceAccountID2,
		payndaPayUrl:   "http://payndapay.ptmlink.top",
		debug:          true,
	}

	l := log.DefaultLogger
	pp = New(conf, WithResponseMiddlewares(trace.ResponseTracking(), traffic.PayndaTrafficLog(l)))
	ppSDK = pp

	m.Run()
}

// 测试 GenerateSignature 函数
func TestGenerateSignatureFunc(t *testing.T) {
	params := &crypto.SignatureParams{
		AppID:     "appId_WaXRKyF5s8JH",
		AppSecret: "appSecret_JaARu7mS8V3N",
		Path:      "/test/path",
		Nonce:     "test_nonce",
		Timestamp: "1234567890",
	}

	signature := crypto.GenerateSignature(params)
	if signature == "" {
		t.Error("Expected non-empty signature")
	}

	// 测试相同参数生成相同签名
	signature2 := crypto.GenerateSignature(params)
	if signature != signature2 {
		t.Errorf("Expected same signatures, got %s and %s", signature, signature2)
	}

	// 测试不同参数生成不同签名
	params.Nonce = "different_nonce"
	signature3 := crypto.GenerateSignature(params)
	if signature == signature3 {
		t.Errorf("Expected different signatures, got %s and %s", signature, signature3)
	}
}

type testConf struct {
	appID          string
	appSecret      string
	balanceAccount string
	payndaPayUrl   string
	debug          bool
	disabledBins   string
}

func (c *testConf) GetAppId() string {
	return c.appID
}

func (c *testConf) GetAppSecret() string {
	return c.appSecret
}
func (c *testConf) GetBalanceAccount() string {
	return c.balanceAccount
}

func (c *testConf) GetUrl() string {
	return c.payndaPayUrl
}

func (c *testConf) GetDebug() bool {
	return c.debug
}

func (c *testConf) GetDisabledBins() string {
	return c.disabledBins
}

func TestPayndaPaySDKValidCardBin(t *testing.T) {
	sdk := New(&testConf{disabledBins: "123456,654321"})

	assert.False(t, sdk.ValidCardBin(context.Background(), "123456"))
	assert.True(t, sdk.ValidCardBin(context.Background(), "999999"))
}

func TestPayndaPaySDKCardTransferRejectsDisabledBin(t *testing.T) {
	sdk := New(&testConf{disabledBins: "123456"})
	req := &CardTransferRequest{CardID: "card-id", CardBin: "123456", Amount: decimal.NewFromInt(1), RequestID: "request-id"}

	transferIn, err := sdk.CardTransferIn(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, CardTransferStatusFailed, transferIn.Status)

	transferOut, err := sdk.CardTransferOut(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, CardTransferStatusFailed, transferOut.Status)
}

// payndapay.MerchantWallet
//
//		{
//		 ID: "1960644045659385857",
//		 CreateTime: "2025-08-27 10:02:22",
//		 UpdateTime: "2025-08-28 02:58:16",
//		 Name: "",
//		 MerchantId: "1960644045655191554",
//		 Currency: "USD",
//		 Amount: "74.7500"
//	}
func TestPayndaPaySDK_GetMerchantWallet(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string // 测试用例名称

	}{
		{
			name: "成功获取商户钱包列表",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 Mock 对象

			// 调用被测方法
			result, err := pp.GetMerchantWallet(context.Background())
			assert.NoError(t, err)
			for _, v := range result {
				fmt.Printf("v: %+v\n", v)
			}
		})
	}
}

func TestPayndaPaySDK_GetBalanceAccount(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name string
		// accountID  string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "获取资金账户2",
			// accountID:  "1961375380007002114",
			wantErr:    false,
			wantErrMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := pp.GetBalanceAccount(context.Background())
			if tt.wantErr && err != nil {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			fmt.Println(result, err)

		})
	}
}

// [
//
//	{ID: "1960889588797104130", CreateTime: "2025-08-28 02:18:04", UpdateTime: "2025-08-28 02:18:04", Name: "lll", MerchantId: "1960644045655191554"},
//	{ID: "1960879376946126849", CreateTime: "2025-08-28 01:37:29", UpdateTime: "2025-08-28 01:37:29", Name: "ll", MerchantId: "1960644045655191554"}
//	{ID: "1960651159362560002", CreateTime: "2025-08-27 10:30:38", UpdateTime: "2025-08-27 10:30:38", Name: "ZXX", MerchantId: "1960644045655191554"}]
func TestPayndaPaySDK_GetBalanceAccounts(t *testing.T) {
	// 测试用例设计
	tests := []struct {
		name       string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "成功查询资金账户列表",
			wantErr:    false,
			wantErrMsg: "",
		},
		// {
		// 	name:       "查询资金账户列表失败",
		// 	wantErr:    true,
		// 	wantErrMsg: "无效的请求参数",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := pp.GetBalanceAccounts(context.Background())
			if tt.wantErr && err != nil {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			for _, v := range result {
				fmt.Printf("%+v", v)
			}

		})
	}
}

// {ID:1960889588797104131 CreateTime:2025-08-28 02:18:04 UpdateTime:2025-08-28 02:18:04 MerchantId:1960644045655191554 BalanceAccountId:1960889588797104130 Currency:USD Amount:0.0000 FrozenAmount:0.0000}
// {ID:1960879376950321153 CreateTime:2025-08-28 01:37:29 UpdateTime:2025-08-28 08:21:17 MerchantId:1960644045655191554 BalanceAccountId:1960879376946126849 Currency:USD Amount:4.9000 FrozenAmount:0.0000}
// {ID:1960651159370948610 CreateTime:2025-08-27 10:30:38 UpdateTime:2025-08-27 10:34:01 MerchantId:1960644045655191554 BalanceAccountId:1960651159362560002 Currency:USD Amount:0.0000 FrozenAmount:0.0000}
func TestPayndaPaySDK_GetBalanceAccountWallets(t *testing.T) {

	tests := []struct {
		name       string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "获取资金账户2",
			wantErr:    false,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := pp.GetBalanceAccountWallets(ctx)
			if tt.wantErr && err != nil {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			for _, wallet := range got {
				fmt.Printf("%+v\n", wallet)
			}
		})
	}
}

// {ID:f259548e03d95a5a34ed2746b8d67295 CardBin:E0000037 Name:MCUK23800300 Type:CARD_CLASS CardClass:E0000037 SupportCurrencies:[USD]}
// {ID:7712f27f2df6c57076ada93dbc68bc69 CardBin:E0000038 Name:MCUK24600100 Type:CARD_CLASS CardClass:E0000038 SupportCurrencies:[USD]}
// {ID:18dc3b493b9aa12819a0c0dfe7d1fc3c CardBin:E0000039 Name:MCUK23501901 Type:CARD_CLASS CardClass:E0000039 SupportCurrencies:[USD]}
// {ID:5af8da34bb6d395604ccaa690ef8d71b CardBin:E0000040 Name:MCUK22360003 Type:CARD_CLASS CardClass:E0000040 SupportCurrencies:[USD]}
// {ID:11ceed897858c77b61eeb31d6ddd6f81 CardBin:E0000041 Name:MCUK22343001 Type:CARD_CLASS CardClass:E0000041 SupportCurrencies:[USD]}
// {ID:ad9e74f130030788d1afe3613fbb1ed0 CardBin:G0000002 Name:MCUK53839503 Type:CARD_CLASS CardClass:G0000002 SupportCurrencies:[USD]}
// {ID:a5ed01bc990678af21e4aeba6e71ad79 CardBin:G0000003 Name:MCUK53700201 Type:CARD_CLASS CardClass:G0000003 SupportCurrencies:[USD]}
// {ID:a4f944d97536e00c89c2a92a62d77ef2 CardBin:G0000004 Name:MCUK53506700 Type:CARD_CLASS CardClass:G0000004 SupportCurrencies:[USD]}
// {ID:2b2d82634e2577107b0c12607b776bf3 CardBin:G0000005 Name:MCUK53209102 Type:CARD_CLASS CardClass:G0000005 SupportCurrencies:[USD]}
// {ID:0c1076a672a10ce79efa02b16c0efe4c CardBin:G0000006 Name:MCUK23409102 Type:CARD_CLASS CardClass:G0000006 SupportCurrencies:[USD]}
// {ID:4f8a1c98e8deb4963a98ff687c4bac7a CardBin:G0000007 Name:MCUK23203903 Type:CARD_CLASS CardClass:G0000007 SupportCurrencies:[USD]}
// {ID:5308bcfd786a35126f2f83834e4c7652 CardBin:G0000008 Name:MCUK23600701 Type:CARD_CLASS CardClass:G0000008 SupportCurrencies:[USD]}

func TestPayndaPaySDK_GetCardBin(t *testing.T) {

	type args struct {
		creditLimitType CreditLimitType
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		// TODO: Add test cases.
		{
			name: "卡BIN查询-INDEPENDENT",
			args: args{
				creditLimitType: CreditLimitType_INDEPENDENT,
			},
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name: "卡BIN查询-SHARED",
			args: args{
				creditLimitType: CreditLimitType_SHARED,
			},
			wantErr:    false,
			wantErrMsg: "",
		},

		{
			name: "卡BIN查询-SHARED-错误参数",
			args: args{
				creditLimitType: "SHARED-ERROR",
			},
			wantErr:    true,
			wantErrMsg: `请求失败: {"code":500,"message":"error.server.internal","success":false}`,
		},
		{
			name: "卡BIN查询-BIN_type为空",
			args: args{
				creditLimitType: "",
			},
			wantErr:    false,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pp.GetCardBin(ctx, tt.args.creditLimitType)
			if tt.wantErr && err != nil {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			for _, v := range got {
				fmt.Printf("%s:%+v\n", tt.args.creditLimitType, *v)
			}
		})
	}
}
func TestPayndaPaySDK_GetCardholders(t *testing.T) {
	tests := []struct {
		name string

		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "balanceAccountID2 cardholders",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := pp.GetCardholders(ctx)
			if tt.wantErr && err != nil {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			for _, v := range got {
				t.Logf("%+v", v)
			}
		})
	}
}

func TestPayndaPaySDK_GetCardholder(t *testing.T) {

	type args struct {
		cardholderID string
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		// TODO: Add test cases.
		{
			name: "get cardholder success",
			args: args{
				cardholderID: "1963491533516439554",
			},
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name: "get cardholder cardholderID is empty",
			args: args{
				cardholderID: "",
			},
			wantErr:    true,
			wantErrMsg: "cardholderID不能为空",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := pp.GetCardholder(ctx, tt.args.cardholderID)
			if tt.wantErr && err != nil {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			fmt.Printf("%+v\n", got)
		})
	}
}

// 1963073068846546945
// Cardholder {ID: "1961337063643258882", CreateTime: "2025-08-29 07:56:10", UpdateTime: "2025-08-29 07:56:10", MerchantId: "1960644045655191554", BalanceAccountId: "1960889588797104130", FirstName: "zhang", LastName: "san", MobilePrefix: "86", Mobile: "13800000000", Email: "zhangsan@example.com", BillingAddressLine1: "", BillingAddressLine2: "", BillingCity: "", BillingCountryCode: "", BillingPostalCode: "", BillingState: "", UnlimitedBalance: false}
// Cardholder {ID:1963491533516439554 CreateTime:2025-08-29 08:10:14 UpdateTime:2025-08-29 08:10:14 MerchantId:1960644045655191554 BalanceAccountId:1960879376946126849 FirstName:zhang LastName:si MobilePrefix:86 Mobile:13800000000 Email:zhangsi@example.com BillingAddressLine1: BillingAddressLine2: BillingCity: BillingCountryCode: BillingPostalCode: BillingState: UnlimitedBalance:false}
func TestPayndaPaySDK_CreateCardholder(t *testing.T) {
	type args struct {
		param *CardHolderRequest
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "参数为空",
			args: args{
				param: nil,
			},
			wantErr:    true,
			wantErrMsg: "参数不能为空",
		},
		{
			name: "参数验证失败-firstName不能为空",
			args: args{
				param: &CardHolderRequest{
					Nonce:        genNonce(),
					CardholderID: "123456",
					FirstName:    "",
					LastName:     "1214123",
				},
			},
			wantErr:    true,
			wantErrMsg: "firstName不能为空",
		},
		{
			name: "参数验证失败-lastName不能为空",
			args: args{
				param: &CardHolderRequest{
					Nonce:        genNonce(),
					CardholderID: "123456",
					FirstName:    "张三",
					LastName:     "",
				},
			},
			wantErr:    true,
			wantErrMsg: "lastName不能为空",
		},
		{
			name: "参数验证失败-email格式不正确",
			args: args{
				param: &CardHolderRequest{
					Nonce:        genNonce(),
					CardholderID: "123456",
					FirstName:    "张三",
					LastName:     "张三",
					Email:        "123456",
				},
			},
			wantErr:    true,
			wantErrMsg: "email格式不正确",
		},
		{
			name: "参数验证失败-mobilePrefix为空",
			args: args{
				param: &CardHolderRequest{
					Nonce:        genNonce(),
					CardholderID: "123456",
					FirstName:    "张三",
					LastName:     "张三",
					Mobile:       "123456",
				},
			},
			wantErr:    true,
			wantErrMsg: "mobilePrefix不能为空",
		},
		{
			name: "参数验证失败-billingCountryCode错误",
			args: args{
				param: &CardHolderRequest{
					Nonce:              genNonce(),
					CardholderID:       "123456",
					FirstName:          "张三",
					LastName:           "张三",
					MobilePrefix:       "86",
					Mobile:             "123456",
					BillingCountryCode: "CNd123",
				},
			},
			wantErr:    true,
			wantErrMsg: "billingCountryCode错误",
		},
		{
			name: "参数验证失败-name长度超过64",
			args: args{
				param: &CardHolderRequest{
					Nonce:        genNonce(),
					FirstName:    "adfdlajfqfdadfewqfdadfewqfadfewqfadfewqfadfewqfadfewqfadfewqfadfewqfa",
					LastName:     "adfdlajfqfdadfewqfdadfewqfadfewqfadfewqfadfewqfadfewqfadfewqfadfewqfa",
					Email:        "zhangsi@examp111le.com",
					Mobile:       "13830000000",
					MobilePrefix: "86"},
			},
			wantErr:    true,
			wantErrMsg: "name长度超过64",
		},
		{
			name: "创建成功",
			args: args{
				param: &CardHolderRequest{
					Nonce:        genNonce(),
					FirstName:    "zhang1",
					LastName:     "si3",
					Email:        "zhangsi@examp111le.com",
					Mobile:       "13830000000",
					MobilePrefix: "86",
				},
			},
			wantErr:    false,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pp.CreateCardHolder(ctx, tt.args.param)
			if tt.wantErr && err != nil {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if got != nil {
				fmt.Printf("%+v\n", *got)
			}
		})
	}
}
func TestPayndaPaySDK_UpdateCardholder(t *testing.T) {

	type args struct {
		param *CardHolderRequest
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "参数为空",
			args: args{
				param: nil,
			},
			wantErr:    true,
			wantErrMsg: "参数不能为空",
		},
		{
			name: "参数验证失败-cardholderID不能为空",
			args: args{
				param: &CardHolderRequest{
					Nonce:        genNonce(),
					CardholderID: "",
				},
			},
			wantErr:    true,
			wantErrMsg: "cardholderID不能为空",
		},
		{
			name: "参数验证失败-firstName不能为空",
			args: args{
				param: &CardHolderRequest{
					Nonce:        genNonce(),
					CardholderID: "123456",
					FirstName:    "",
					LastName:     "1214123",
				},
			},
			wantErr:    true,
			wantErrMsg: "firstName不能为空",
		},
		{
			name: "参数验证失败-lastName不能为空",
			args: args{
				param: &CardHolderRequest{
					Nonce:        genNonce(),
					CardholderID: "123456",
					FirstName:    "张三",
					LastName:     "",
				},
			},
			wantErr:    true,
			wantErrMsg: "lastName不能为空",
		},
		{
			name: "参数验证失败-email格式不正确",
			args: args{
				param: &CardHolderRequest{
					Nonce:        genNonce(),
					CardholderID: "123456",
					FirstName:    "张三",
					LastName:     "张三",
					Email:        "123456",
				},
			},
			wantErr:    true,
			wantErrMsg: "email格式不正确",
		},
		{
			name: "参数验证失败-mobilePrefix为空",
			args: args{
				param: &CardHolderRequest{
					Nonce:        genNonce(),
					CardholderID: "123456",
					FirstName:    "张三",
					LastName:     "张三",
					Mobile:       "123456",
				},
			},
			wantErr:    true,
			wantErrMsg: "mobilePrefix不能为空",
		},
		// {
		// 	name: "参数验证失败-billingCountryCode错误",
		// 	args: args{
		// 		param: &CardHolderRequest{
		// 			Nonce:              genNonce(),
		// 			CardholderID:       "123456",
		// 			FirstName:          "张三",
		// 			LastName:           "张三",
		// 			MobilePrefix:       "86",
		// 			Mobile:             "123456",
		// 			BillingCountryCode: "CNd123",
		// 		},
		// 	},
		// 	wantErr:    true,
		// 	wantErrMsg: "billingCountryCode错误",
		// },

		{
			name: "更新成功",
			args: args{
				param: &CardHolderRequest{
					Nonce:        genNonce(),
					CardholderID: "1966074944311582721",
					Email:        "12345@1234.com",
					FirstName:    "zhang",
					LastName:     "sii",
					Mobile:       "138000000011",
					MobilePrefix: "86",
				},
			},
			wantErr:    false,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pp.UpdateCardholder(ctx, tt.args.param)
			if tt.wantErr && err != nil {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestPayndaPaySDK_DeleteCardholder(t *testing.T) {

	type args struct {
		cardholderID string
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "参数验证失败-cardholderID不能为空",
			args: args{
				cardholderID: "",
			},
			wantErr:    true,
			wantErrMsg: "cardholderID不能为空",
		},

		{
			name: "删除成功",
			args: args{
				cardholderID: "1961340605472940033",
			},
			wantErr:    false,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := pp.DeleteCardholder(ctx, tt.args.cardholderID)
			if tt.wantErr && err != nil {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestPayndaPaySDK_GetCardholderWallet(t *testing.T) {

	type args struct {
		cardholderID string
	}
	tests := []struct {
		name string

		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "GetCardholderWallet-cardholderID为空",
			args: args{
				cardholderID: "",
			},
			wantErr:    true,
			wantErrMsg: "cardholderID不能为空",
		},
		{
			name: "GetCardholderWallet",
			args: args{
				cardholderID: "1963491533516439554",
			},
			wantErr:    false,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := pp.GetCardholderWallet(ctx, tt.args.cardholderID)
			if tt.wantErr && err != nil {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			for _, v := range got {
				fmt.Printf("%+v\n", v)
			}
		})
	}
}

func TestPayndaPaySDK_UpdateCardholderWallet(t *testing.T) {

	type args struct {
		param *CardHolderWalletRequest
	}
	tests := []struct {
		name string

		args       args
		wantErr    bool
		wantErrMsg string
	}{

		{
			name: "UpdateCardholderWallet-cardholderID-empty",
			args: args{
				param: &CardHolderWalletRequest{
					Nonce:        genNonce(),
					CardholderID: "",
					Amount:       "100",
					Currency:     "USD",
					Type:         "INC",
				},
			},
			wantErr:    true,
			wantErrMsg: "cardholderID不能为空",
		},
		{
			name: "UpdateCardholderWallet-currency-empty",
			args: args{
				param: &CardHolderWalletRequest{
					Nonce:        genNonce(),
					CardholderID: "1960889588797104130",
					Amount:       "100",
					Currency:     "",
					Type:         "INC",
				},
			},
			wantErr:    true,
			wantErrMsg: "currency不能为空",
		},
		{
			name: "UpdateCardholderWallet-type-empty",
			args: args{
				param: &CardHolderWalletRequest{
					Nonce:        genNonce(),
					CardholderID: "1960889588797104130",
					Amount:       "100",
					Currency:     "CNY",
					Type:         "",
				},
			},
			wantErr:    true,
			wantErrMsg: "type不能为空",
		},

		{
			name: "UpdateCardholderWallet-type-error",
			args: args{
				param: &CardHolderWalletRequest{
					Nonce:        genNonce(),
					CardholderID: "1960889588797104130",
					Amount:       "100",
					Currency:     "CNY",
					Type:         "ADD",
				},
			},
			wantErr:    true,
			wantErrMsg: "type错误,不支持的类型:ADD",
		},
		{
			name: "UpdateCardholderWallet-amount-error",
			args: args{
				param: &CardHolderWalletRequest{
					Nonce:        genNonce(),
					CardholderID: "1960889588797104130",
					Amount:       "-10",
					Currency:     "USD",
					Type:         "INC",
				},
			},
			wantErr:    true,
			wantErrMsg: "amount必须大于0",
		},

		{
			name: "UpdateCardholderWallet-增加",
			args: args{
				param: &CardHolderWalletRequest{
					Nonce:        genNonce(),
					CardholderID: "1963491533516439554",
					Currency:     "USD",
					Amount:       "2",
					Type:         "INC",
				},
			},
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name: "UpdateCardholderWallet-减少",
			args: args{
				param: &CardHolderWalletRequest{
					Nonce:        genNonce(),
					CardholderID: "1963491533516439554",
					Currency:     "USD",
					Amount:       "10",
					Type:         "DEC",
				},
			},
			wantErr:    false,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := pp.UpdateCardholderWallet(ctx, tt.args.param)
			if tt.wantErr && err != nil {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if got != nil {
				fmt.Printf("%+v\n", *got)
			}
		})
	}
}

// {
// Card:{ID:1961350369053483010 CreateTime:2025-08-29 08:49:02 UpdateTime:2025-08-29 08:49:02 MerchantId:1960644045655191554 BalanceAccountId:1960879376946126849 CardholderId:1963491533516439554 Status:ACTIVE MaskCardNo:238003 **** **** 0079 CardScheme: CardBin:MCUK23800300 Currency:USD FirstName:zhang LastName:si MobilePrefix:86 Mobile:13800000000 Email:zhangsi@example.com BillingCountryCode: BillingAddressLine1: BillingAddressLine2: BillingCity: BillingPostalCode: BillingState: SingleUse:false CreditLimitType:INDEPENDENT AmountUsed: Amount:}
// CardSensitive:{ID:1961350369061871617 CreateTime:2025-08-29 08:49:02 UpdateTime:2025-08-29 08:49:02 CardID:1961350369053483010 CVV:230 ExpirationDate:06/26 CardNo:2380030012470079}
// CardBalance:{ID:1961350369095426050 CreateTime:2025-08-29 08:49:02 UpdateTime:2025-08-29 08:49:02 AmountUsed:0 AmountFrozen:0 AvailableAmount:0.5 Amount:0.5}
// }
func TestPayndaPaySDK_CreateCard(t *testing.T) {

	type args struct {
		param *CardCreateRequest
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{

		{
			name: "CreateCard-cardholder-empty",
			args: args{
				param: &CardCreateRequest{
					Nonce:          genNonce(),
					RequestID:      genNonce(),
					CardholderID:   "",
					CardBinID:      "f259548e03d95a5a34ed2746b8d67295",
					Amount:         "0.5",
					Currency:       "USD",
					ExpirationDate: "2025-08-29",
				},
			},
			wantErr:    true,
			wantErrMsg: "cardholderID不能为空",
		},
		{
			name: "CreateCard-cardBin-empty",
			args: args{
				param: &CardCreateRequest{
					Nonce:     genNonce(),
					RequestID: genNonce(),

					CardholderID: "1963491533516439554",
					CardBinID:    "",
					Amount:       "0.5",
					Currency:     "USD",
				},
			},
			wantErr:    true,
			wantErrMsg: "cardBinID不能为空",
		},
		// {
		// 	name: "CreateCard-cardBin-not-exist",
		// 	args: args{
		// 		param: &CardCreateRequest{
		//
		// 			CardholderID:     "1963491533516439554",
		// 			CardBinID:        "123456",
		// 			Amount:           "0.5",
		// 			Currency:         "CNY",
		// 		},
		// 	},
		// 	wantErr:    true,
		// 	wantErrMsg: "cardBinID不存在",
		// },
		{
			name: "CreateCard-amount-invalid",
			args: args{
				param: &CardCreateRequest{
					Nonce:     genNonce(),
					RequestID: genNonce(),

					CardholderID: "1963491533516439554",
					CardBinID:    "f259548e03d95a5a34ed2746b8d67295",
					Amount:       "0.0",
					Currency:     "USD",
				},
			},
			wantErr:    true,
			wantErrMsg: "amount必须大于0",
		},
		{
			name: "CreateCard-currency-invalid",
			args: args{
				param: &CardCreateRequest{
					Nonce:        genNonce(),
					RequestID:    genNonce(),
					CardholderID: "1963491533516439554",
					CardBinID:    "f259548e03d95a5a34ed2746b8d67295",
					Amount:       "100.0",
					Currency:     "ABC",
				},
			},
			wantErr:    true,
			wantErrMsg: "currency错误,只支持USD",
		},
		{
			name: "CreateCard-currency-empty",
			args: args{
				param: &CardCreateRequest{
					Nonce:        genNonce(),
					RequestID:    genNonce(),
					CardholderID: "1963491533516439554",
					CardBinID:    "f259548e03d95a5a34ed2746b8d67295",
					Amount:       "100.0",
					Currency:     "",
				},
			},
			wantErr:    true,
			wantErrMsg: "currency不能为空",
		},
		{
			name: "CreateCard-ExpirationDate格式错误",
			args: args{
				param: &CardCreateRequest{
					Nonce:          genNonce(),
					RequestID:      genNonce(),
					Amount:         "0.5",
					CardBinID:      "f259548e03d95a5a34ed2746b8d67295",
					CardholderID:   "1963491533516439554",
					Currency:       "USD",
					ExpirationDate: "2026-07",
					// SingleUse:                  false,
					// TransactionCountLimitTotal: 0,
				},
			},
			wantErr:    true,
			wantErrMsg: "expirationDate格式不正确, 应为MM/YY",
		},
		{
			name: "CreateCard-success",
			args: args{
				param: &CardCreateRequest{
					Nonce:          genNonce(),
					RequestID:      genNonce(),
					Amount:         "0.5",
					CardBinID:      "f259548e03d95a5a34ed2746b8d67295",
					CardholderID:   "1963491533516439554",
					Currency:       "USD",
					ExpirationDate: "07/26",
					// SingleUse:                  false,
					// TransactionCountLimitTotal: 0,
				},
			},
			wantErr:    false,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pp.CreateCard(ctx, tt.args.param)
			if tt.wantErr && err != nil {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if got != nil {
				fmt.Printf("%+v\n", *got)
			}
			r, err := pp.QueryRequestResult(ctx, tt.args.param.RequestID)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if r != nil {
				fmt.Printf("%+v\n", *r)
			}
		})
	}
}

func TestPayndaPaySDK_GetCards(t *testing.T) {
	tests := []struct {
		name string

		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "GetCards",
			wantErr:    false,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.Create("payndapay_cards.txt")
			if err != nil {
				t.Fatalf("create output file failed: %v", err)
			}
			defer file.Close()

			current := 1
			pageSize := 500
			for {
				got, err := pp.ListCardByPaginate(ctx, &CardPaginateRequest{
					Current:  current,
					PageSize: pageSize,
				})
				if tt.wantErr && err != nil {
					assert.Equal(t, tt.wantErrMsg, err.Error())
				} else if !tt.wantErr && err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if got == nil {
					break
				}
				for _, card := range got.Records {
					if _, err := fmt.Fprintf(file, "%+v\n", *card); err != nil {
						t.Fatalf("write output file failed: %v", err)
					}
				}
				if got.Current == got.Pages {
					break
				}
				current++
			}

		})
	}
}

func TestPayndaPaySDK_GetCard(t *testing.T) {
	type args struct {
		cardID string
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{

		{
			name: "getcard-cardID-empty",
			args: args{
				cardID: "",
			},
			wantErr:    true,
			wantErrMsg: "cardID不能为空",
		},
		{
			name: "getcard-success",
			args: args{
				cardID: "2095804386448269314", //1963497742025883650,1963497548190318594
			},
			wantErr:    false,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pp.GetCard(ctx, tt.args.cardID)
			if tt.wantErr && err != nil {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if got != nil {
				fmt.Printf("%+v\n", *got)
			}
		})
	}
}

func TestPayndaPaySDK_GetCardBalance(t *testing.T) {
	type args struct {
		cardID string
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{

		{
			name: "GetCardBalance-cardID-empty",
			args: args{
				cardID: "",
			},
			wantErr:    true,
			wantErrMsg: "cardID不能为空",
		},

		{
			name: "GetCardBalance-success",
			args: args{
				cardID: "2096874346733985794", //1963497742025883650,1963497548190318594
			},
			wantErr:    false,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pp.GetCardBalance(ctx, tt.args.cardID)
			if tt.wantErr && err != nil {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if got != nil {
				fmt.Printf("%+v\n", *got)
			}
		})
	}
}

func TestPayndaPaySDK_GetCardSensitive(t *testing.T) {
	type args struct {
		cardID string
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{

		{
			name: "GetCardSensitive-cardID-empty",
			args: args{
				cardID: "",
			},
			wantErr:    true,
			wantErrMsg: "cardID不能为空",
		},

		{
			name: "GetCardSensitive-success",
			args: args{
				cardID: "1963497742025883650",
			},
			wantErr:    false,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pp.GetCardSensitive(ctx, tt.args.cardID)
			if tt.wantErr && err != nil {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if got != nil {
				fmt.Printf("%+v\n", *got)
			}
		})
	}
}

func TestPayndaPaySDK_FrozenCard(t *testing.T) {
	type args struct {
		cardID string
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "frozencard-cardID-empty",
			args: args{
				cardID: "2085178822566473730",
			},
			wantErr:    true,
			wantErrMsg: "cardID不能为空",
		},
		{
			name: "frozencard-success",
			args: args{
				cardID: "2007726970992115713",
			},
			wantErr:    false,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pp.FrozenCard(ctx, tt.args.cardID)
			if tt.wantErr && err != nil {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestPayndaPaySDK_UnfrozenCard(t *testing.T) {
	type args struct {
		cardID string
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "unfrozencard-cardID-empty",
			args: args{
				cardID: "2085178822566473730",
			},
			wantErr:    true,
			wantErrMsg: "cardID不能为空",
		},
		{
			name: "unfrozencard",
			args: args{
				cardID: "1963497742025883650",
			},
			wantErr:    false,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pp.UnfrozenCard(ctx, tt.args.cardID)
			if tt.wantErr && err != nil {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestPayndaPaySDK_ReleaseCard(t *testing.T) {
	type args struct {
		cardID string
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "releasecard-cardID-empty",
			args: args{
				cardID: "",
			},
			wantErr:    true,
			wantErrMsg: "cardID不能为空",
		},

		{
			name: "releasecard",
			args: args{
				cardID: "1961350369053483010",
			},
			wantErr:    false,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pp.ReleaseCard(ctx, "", tt.args.cardID)
			if tt.wantErr && err != nil {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestPayndaPaySDK_GetCardControls(t *testing.T) {
	type args struct {
		cardID string
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "GetCardControls-cardID-empty",
			args: args{
				cardID: "",
			},
			wantErr:    true,
			wantErrMsg: "cardID不能为空",
		},

		{
			name: "GetCardControls-success",
			args: args{
				cardID: "1963497742025883650",
			},
			wantErr:    false,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pp.GetCardControls(ctx, tt.args.cardID)
			if tt.wantErr && err != nil {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			for _, v := range got {
				fmt.Printf("%+v\n", *v)
			}
		})
	}
}

func TestPayndaPaySDK_CardTransfer(t *testing.T) {
	type args struct {
		req *CardBalanceTransferRequest
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		// {
		// 	name: "CardTransfer-cardID-empty",
		// 	args: args{
		// 		req: &CardBalanceTransferRequest{
		// 			Nonce:     genNonce(),
		// 			RequestID: genNonce(),
		// 			CardID:    "",
		// 			Amount:    "100",
		// 			Type:      "IN",
		// 		},
		// 	},
		// 	wantErr:    true,
		// 	wantErrMsg: "cardID不能为空",
		// },
		// {
		// 	name: "CardTransfer-type-empty",
		// 	args: args{
		// 		req: &CardBalanceTransferRequest{
		// 			Nonce:     genNonce(),
		// 			RequestID: genNonce(),
		// 			CardID:    "1963497742025883650",
		// 			Amount:    "100",
		// 			Type:      "",
		// 		},
		// 	},
		// 	wantErr:    true,
		// 	wantErrMsg: "type不能为空",
		// },

		// {
		// 	name: "CardTransfer-type-error",
		// 	args: args{
		// 		req: &CardBalanceTransferRequest{
		// 			Nonce:     genNonce(),
		// 			RequestID: genNonce(),
		// 			CardID:    "1963497742025883650",
		// 			Amount:    "100",
		// 			Type:      "ADD",
		// 		},
		// 	},
		// 	wantErr:    true,
		// 	wantErrMsg: "type错误,不支持的类型:ADD",
		// },
		// {
		// 	name: "CardTransfer-amount-error",
		// 	args: args{
		// 		req: &CardBalanceTransferRequest{
		// 			Nonce:     genNonce(),
		// 			RequestID: genNonce(),
		// 			CardID:    "1963497742025883650",
		// 			Amount:    "-10",
		// 			Type:      "IN",
		// 		},
		// 	},
		// 	wantErr:    true,
		// 	wantErrMsg: "amount必须大于0",
		// },

		{
			name: "CardTransfer-in-success",
			args: args{
				req: &CardBalanceTransferRequest{
					Nonce:     genNonce(),
					RequestID: genNonce(),
					Amount:    "1",
					CardID:    "1965334766265634818",
					Type:      TransferType_OUT,
				},
			},
			wantErr:    true,
			wantErrMsg: "",
		},
		// {
		// 	name: "CardTransfer-out-success",
		// 	args: args{
		// 		req: &CardBalanceTransferRequest{
		// 			Amount:           1,
		// 			CardID:           "1961348610386010114",
		// 			Type:             "OUT",
		// 		},
		// 	},
		// 	wantErr:    false,
		// 	wantErrMsg: "请求失败: {"code":500,"message":"repeated request","success":false}",
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pp.CardTransfer(ctx, tt.args.req)
			if tt.wantErr && err != nil {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if got != nil {
				fmt.Printf("%+v\n", *got)
			}
			result, err := pp.QueryTransfer(ctx, tt.args.req.RequestID)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			fmt.Printf("%+v\n", result)
		})
	}
}

func TestPayndaPaySDK_CardTransferIn(t *testing.T) {
	tests := []struct {
		name       string
		request    *CardTransferRequest
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "params empty",
			wantErr:    true,
			wantErrMsg: ErrEmptyParams.Error(),
		},
		{
			name: "transfer in",
			request: &CardTransferRequest{
				CardID:    "2093263050503995393",
				RequestID: "20932630505039953932",
				Amount:    decimal.NewFromInt(1),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ppSDK.CardTransferIn(ctx, tt.request)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				assert.Equal(t, tt.wantErrMsg, err.Error())
				return
			}
			assert.NoError(t, err)
			t.Logf("CardTransferIn result: %+v", result)
		})
	}
}

func TestPayndaPaySDK_CardTransferOut(t *testing.T) {
	tests := []struct {
		name       string
		request    *CardTransferRequest
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "params empty",
			wantErr:    true,
			wantErrMsg: ErrEmptyParams.Error(),
		},
		{
			name: "transfer out",
			request: &CardTransferRequest{
				CardID:    "2093263050503995393",
				RequestID: "20932630505039953921",
				Amount:    decimal.NewFromInt(1),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ppSDK.CardTransferOut(ctx, tt.request)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				assert.Equal(t, tt.wantErrMsg, err.Error())
				return
			}
			assert.NoError(t, err)
			t.Logf("CardTransferOut result: %+v", result)
		})
	}
}

func TestPayndaPaySDK_QueryCardTransfer(t *testing.T) {
	tests := []struct {
		name       string
		requestID  string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "request ID empty",
			wantErr:    true,
			wantErrMsg: ErrEmptyRequestID.Error(),
		},
		{
			name:      "query transfer",
			requestID: "2093278518984458240",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ppSDK.QueryCardTransfer(ctx, tt.requestID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				assert.Equal(t, tt.wantErrMsg, err.Error())
				return
			}
			assert.NoError(t, err)
			t.Logf("QueryCardTransfer result: %+v", result)
		})
	}
}

func TestPayndaPaySDK_GetCardTransactions(t *testing.T) {
	type args struct {
		req *CardTransactionsRequest
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "获取卡交易记录",
			args: args{
				req: &CardTransactionsRequest{
					// CardholderID:         "1963491533516439554",
					CardID:   "2006183096251953154",
					Current:  1,
					PageSize: 10,
					// TransactionTimeStart: "2025-09-01 07:10:23",
					// TransactionTimeEnd:   "2025-12-16 07:10:23",
				},
			},
			wantErr:    true,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pp.GetCardTransactions(ctx, tt.args.req)
			if tt.wantErr && err != nil {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			for _, v := range got.Records {
				fmt.Printf("%+v\n", v)
			}
		})
	}
}

func TestPayndaPaySDK_CreateBalanceAccount(t *testing.T) {

	type args struct {
		name  string
		nonce string
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		// TODO: Add test cases.
		{
			name: "创建账户",
			args: args{
				name:  "测试账户",
				nonce: genNonce(),
			},
			wantErr:    false,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pp.CreateBalanceAccount(ctx, tt.args.name, tt.args.nonce)
			if tt.wantErr && err != nil {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			fmt.Printf("%+v\n", got)
		})
	}
}

func TestPayndaPaySDK_UpdateBalanceAccount(t *testing.T) {

	type args struct {
		name  string
		nonce string
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		// TODO: Add test cases.
		{
			name: "创建账户",
			args: args{
				name:  "Product Account",
				nonce: genNonce(),
			},
			wantErr:    false,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pp.UpdateBalanceAccount(ctx, tt.args.name, tt.args.nonce)
			if tt.wantErr && err != nil {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestPayndaPaySDK_DeleteBalanceAccount(t *testing.T) {
	type args struct {
		nonce string
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		// TODO: Add test cases.
		{
			name: "删除成功",
			args: args{
				nonce: fmt.Sprint(time.Now().Unix()),
			},
			wantErr:    false,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// err := pp.DeleteBalanceAccount(ctx, tt.args.nonce)
			// if tt.wantErr && err != nil {
			// 	assert.Equal(t, tt.wantErrMsg, err.Error())
			// } else if !tt.wantErr && err != nil {
			// 	t.Errorf("Expected no error, got %v", err)
			// }
		})
	}
}

func TestPayndaPaySDK_IndependentCard(t *testing.T) {
	// 查询资金账户钱包信息
	wallet, err := pp.GetBalanceAccountWallets(ctx)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("wallet: %+v\n", wallet[0])
	// 查询持卡人钱包信息
	cardHolderID := "1963491533516439554"
	cardholderWallet, err := pp.GetCardholderWallet(ctx, cardHolderID)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("cardholderWallet: %+v\n", cardholderWallet[0])
	// 查询卡信息
	cardID := "1963497742025883650"
	card, err := pp.GetCard(ctx, cardID)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("card: %+v\n", card)
	// 查询卡资金信息
	cardBalance, err := pp.GetCardBalance(ctx, cardID)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("cardBalance: %+v\n", cardBalance)
	// 入金
	tsf, err := pp.CardTransfer(ctx, &CardBalanceTransferRequest{
		Nonce:  genNonce(),
		CardID: cardID,
		Amount: "2.2",
		Type:   "OUT",
	})
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("tsf: %+v\n", tsf)

	// 查询资金账户钱包信息
	wallet, err = pp.GetBalanceAccountWallets(ctx)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("wallet: %+v\n", wallet[0])
	// 查询持卡人钱包信息
	cardholderWallet, err = pp.GetCardholderWallet(ctx, cardHolderID)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("cardholderWallet: %+v\n", cardholderWallet[0])
	// 查询卡信息
	card, err = pp.GetCard(ctx, cardID)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("card: %+v\n", card)
	// 查询卡资金信息
	cardBalance, err = pp.GetCardBalance(ctx, cardID)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("cardBalance: %+v\n", cardBalance)
	// 出金
	tsf, err = pp.CardTransfer(ctx, &CardBalanceTransferRequest{
		Nonce:  genNonce(),
		Amount: "2.4",
		CardID: cardID,
		Type:   "IN",
	})
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("tsf: %+v\n", tsf)

	// 查询资金账户钱包信息
	wallet, err = pp.GetBalanceAccountWallets(ctx)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("wallet: %+v\n", wallet[0])
	// 查询持卡人钱包信息
	cardholderWallet, err = pp.GetCardholderWallet(ctx, cardHolderID)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("cardholderWallet: %+v\n", cardholderWallet[0])
	// 查询卡信息
	card, err = pp.GetCard(ctx, cardID)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("card: %+v\n", card)
	// 查询卡资金信息
	cardBalance, err = pp.GetCardBalance(ctx, cardID)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("cardBalance: %+v\n", cardBalance)
}

func TestPayndaPaySDK_BalanceAccountWalletTransfer(t *testing.T) {

	type args struct {
		param *BalanceAccountWalletTransferRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *BalanceAccountWalletTransfer
		wantErr bool
	}{

		{
			name: "账户2入金",
			args: args{
				param: &BalanceAccountWalletTransferRequest{
					Nonce:     fmt.Sprint(time.Now().Unix()),
					RequestID: genNonce(),
					Amount:    "19880.0000",
					Currency:  "USD",
					Type:      "IN",
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pp.BalanceAccountWalletTransfer(ctx, tt.args.param)
			if tt.wantErr && err != nil {
				t.Errorf("PayndaPaySDK.BalanceAccountWalletTransfer() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			fmt.Printf("%+v\n", got)

			r, err := pp.QueryRequestResult(ctx, tt.args.param.RequestID)
			if err != nil {
				t.Error(err)
			}
			fmt.Printf("%+v\n", r)
		})
	}
}

func TestPayndaPaySDK_CradholderWallet(t *testing.T) {
	// 查询资金账户钱包信息
	wallet, err := pp.GetBalanceAccountWallets(ctx)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("wallet: %+v\n", wallet[0])
	// 查询持卡人钱包信息
	cardHolderID := "1963491533516439554"
	cardholderWallet, err := pp.GetCardholderWallet(ctx, cardHolderID)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("cardholderWallet: %+v\n", cardholderWallet[0])
	// 查询卡信息
	cardID := "1963497742025883650"
	card, err := pp.GetCard(ctx, cardID)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("card: %+v\n", card)
	// 查询卡资金信息
	cardBalance, err := pp.GetCardBalance(ctx, cardID)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("cardBalance: %+v\n", cardBalance)
	// 增加额度
	ucw, err := pp.UpdateCardholderWallet(ctx, &CardHolderWalletRequest{
		Nonce:        genNonce(),
		CardholderID: cardHolderID,
		Amount:       "3000",
		Currency:     "USD",
		Type:         "INC",
	})
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("ucw: %+v\n", ucw)

	// 查询资金账户钱包信息
	wallet, err = pp.GetBalanceAccountWallets(ctx)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("wallet: %+v\n", wallet[0])
	// 查询持卡人钱包信息
	cardholderWallet, err = pp.GetCardholderWallet(ctx, cardHolderID)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("cardholderWallet: %+v\n", cardholderWallet[0])
	// 查询卡信息
	card, err = pp.GetCard(ctx, cardID)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("card: %+v\n", card)
	// 查询卡资金信息
	cardBalance, err = pp.GetCardBalance(ctx, cardID)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("cardBalance: %+v\n", cardBalance)
	// 减少额度
	ucw, err = pp.UpdateCardholderWallet(ctx, &CardHolderWalletRequest{
		Nonce:        genNonce(),
		CardholderID: cardHolderID,
		Amount:       "2500",
		Currency:     "USD",
		Type:         "INC",
	})
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("ucw: %+v\n", ucw)

	// 查询资金账户钱包信息
	wallet, err = pp.GetBalanceAccountWallets(ctx)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("wallet: %+v\n", wallet[0])
	// 查询持卡人钱包信息
	cardholderWallet, err = pp.GetCardholderWallet(ctx, cardHolderID)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("cardholderWallet: %+v\n", cardholderWallet[0])
	// 查询卡信息
	card, err = pp.GetCard(ctx, cardID)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("card: %+v\n", card)
	// 查询卡资金信息
	cardBalance, err = pp.GetCardBalance(ctx, cardID)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("cardBalance: %+v\n", cardBalance)
}

func TestPayndaPaySDK_UpdateCardControls(t *testing.T) {
	type args struct {
		req *CardControlRequest
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "UpdateCardControls-cardID-empty",
			args: args{
				req: &CardControlRequest{
					Nonce: genNonce(),

					CardID:           "",
					Period:           "DAY",
					TransactionCount: 0,
					Amount:           "1000",
				},
			},
			wantErr:    true,
			wantErrMsg: "cardID不能为空",
		},
		{
			name: "UpdateCardControls-period-empty",
			args: args{
				req: &CardControlRequest{
					Nonce: genNonce(),

					CardID:           "1963497742025883650",
					Period:           "",
					TransactionCount: 0,
					Amount:           "1000",
				},
			},
			wantErr:    true,
			wantErrMsg: "period不能为空",
		},
		{
			name: "UpdateCardControls-period-invalid",
			args: args{
				req: &CardControlRequest{
					Nonce: genNonce(),

					CardID:           "1963497742025883650",
					Period:           "a",
					TransactionCount: 0,
					Amount:           "1000",
				},
			},
			wantErr:    true,
			wantErrMsg: "period错误,不支持的值:a",
		},
		{
			name: "UpdateCardControls-transactionCount-invalid",
			args: args{
				req: &CardControlRequest{
					Nonce: genNonce(),

					CardID:           "1963497742025883650",
					Period:           "DAY",
					TransactionCount: -1,
					Amount:           "1000",
				},
			},
			wantErr:    true,
			wantErrMsg: "transactionCount不能小于0",
		},
		{
			name: "UpdateCardControls-amount-invalid",
			args: args{
				req: &CardControlRequest{
					Nonce: genNonce(),

					CardID:           "1963497742025883650",
					Period:           "DAY",
					TransactionCount: 100,
					Amount:           "-1000",
				},
			},
			wantErr:    true,
			wantErrMsg: "amount不能小于0",
		},
		{
			name: "UpdateCardControls-success",
			args: args{
				req: &CardControlRequest{
					Nonce:            genNonce(),
					CardID:           "1963497742025883650",
					Period:           "DAY",
					TransactionCount: 0,
					Amount:           "1000",
				},
			},
			wantErr:    false,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		err := pp.UpdateCardControls(ctx, tt.args.req)
		if tt.wantErr && err != nil {
			assert.EqualError(t, err, tt.wantErrMsg)
		} else if !tt.wantErr && err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

	}
}

func TestPayndaPaySDK_UpdateCardBalance(t *testing.T) {
	type args struct {
		req *CardBalanceUpdateRequest
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "UpdateCardBalance-fail-empty-cardID",
			args: args{
				req: &CardBalanceUpdateRequest{
					Nonce:  genNonce(),
					CardID: "",
					Amount: "100.00",
				},
			},
			wantErr:    true,
			wantErrMsg: "cardID不能为空",
		},
		{
			name: "UpdateCardBalance-fail-invalid-amount",
			args: args{
				req: &CardBalanceUpdateRequest{
					Nonce:  genNonce(),
					CardID: "1963497742025883650",
					Amount: "0.00",
				},
			},
			wantErr:    true,
			wantErrMsg: "amount必须大于0",
		},
		{
			name: "UpdateCardBalance-success",
			args: args{
				req: &CardBalanceUpdateRequest{
					Nonce:  genNonce(),
					CardID: "1963497742025883650",
					Amount: "100",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pp.UpdateCardBalance(ctx, tt.args.req)
			if tt.wantErr && err != nil {
				assert.EqualError(t, err, tt.wantErrMsg)
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if got != nil {
				fmt.Printf("%+v\n", *got)
			}
		})
	}
}

func TestPayndaPaySDK_QueryCardBalanceUpdate(t *testing.T) {

	tests := []struct {
		name    string
		want    []*CardBalanceUpdateHistory
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "query card balance update",

			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := pp.QueryCardBalanceUpdate(ctx)
			if tt.wantErr && err != nil {
				t.Errorf("PayndaPaySDK.QueryCardBalanceUpdate() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			for _, v := range got {
				fmt.Printf("%+v\n", v)
			}
		})
	}
}

// func genNonce() string {
// 	return fmt.Sprint(time.Now().UnixNano())
// }

func TestPayndaPaySDK_BalanceTransToAccount(t *testing.T) {
	// cardholders, err := pp.GetCardholders(ctx)
	// if err != nil {
	// 	t.Errorf("GetCardholders() error = %v", err)
	// 	return
	// }
	// for _, cardholder := range cardholders {
	cards, err := pp.GetCards(ctx)
	if err != nil {
		t.Errorf("GetCards() error = %v", err)
		return
	}
	for _, card := range cards {
		balance, err := pp.GetCardBalance(ctx, card.ID)
		if err != nil {
			t.Errorf("BalanceTransToAccount() error = %v", err)
			return
		}
		amount, _ := decimal.NewFromString(balance.Amount)
		if amount.LessThan(decimal.NewFromFloat(0.001)) {
			continue
		}
		if card.Status != CARD_STATUS_ACTIVE {
			pp.UnfrozenCard(ctx, card.ID)
		}

		if card.Status == "FROZEN" {
			pp.UnfrozenCard(ctx, card.ID)
		}
		_, err = pp.CardTransfer(ctx,
			&CardBalanceTransferRequest{
				Nonce:     genNonce(),
				RequestID: genNonce(),
				CardID:    card.ID,
				Amount:    amount.String(),
				Type:      "OUT",
			})
		if err != nil {
			t.Errorf("UpdateCardBalance() error = %v", err)
			return
		}
		fmt.Printf("%+v\n", balance)

	}
}

func TestPayndaPaySDK_ListCards(t *testing.T) {
	// cardholders, err := pp.GetCardholders(ctx)
	// if err != nil {
	// 	t.Errorf("GetCardholders() error = %v", err)
	// 	return
	// }
	// for _, cardholder := range cardholders {
	got, err := pp.GetBalanceAccountWallets(ctx)
	if err != nil {
		t.Errorf("GetBalanceAccountWallets() error = %v", err)
		return
	}
	for _, wallet := range got {
		fmt.Printf("%+v\n", wallet)
	}
	cards, err := pp.GetCards(ctx)
	if err != nil {
		t.Errorf("GetCards() error = %v", err)
		return
	}
	for _, card := range cards {
		balance, err := pp.GetCardBalance(ctx, card.ID)
		if err != nil {
			t.Errorf("BalanceTransToAccount() error = %v", err)
			return
		}
		amount, _ := decimal.NewFromString(balance.Amount)
		if amount.LessThan(decimal.NewFromFloat(0.001)) {
			continue
		}
		// _, err = pp.CardTransfer(ctx,
		// 	&CardBalanceTransferRequest{
		// 		Nonce:  genNonce(),
		// 		CardID: card.ID,
		// 		Amount: amount.String(),
		// 		Type:   "OUT",
		// 	})
		// if err != nil {
		// 	t.Errorf("UpdateCardBalance() error = %v", err)
		// 	return
		// }
		fmt.Printf("%+v\n", balance)

	}
}

// 1966081834708377600
func TestPayndaPaySDK_QueryRequestResult(t *testing.T) {
	got, err := pp.QueryRequestResult(ctx, "2085633928568049665")
	if err != nil {
		t.Errorf("QueryRequestResult() error = %v", err)
		return
	}
	fmt.Printf("%+v\n", got)
}
func TestPayndaPaySDK_QueryCardTransaction(t *testing.T) {
	tests := []struct {
		name        string
		transaction string
		wantErr     bool
		wantErrMsg  string
	}{
		{
			name:        "query transaction",
			transaction: "2023794740516564994",
			wantErr:     false,
			wantErrMsg:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pp.QueryCardTransaction(ctx, tt.transaction)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				assert.Equal(t, tt.wantErrMsg, err.Error())
				return
			}
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			if got != nil {
				fmt.Printf("%+v\n", *got)
			}
		})
	}
}

func TestPayndaPaySDK_QueryTransfer(t *testing.T) {
	type args struct {
		req *CardBalanceTransferRequest
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "CardTransfer-in-success",
			args: args{
				req: &CardBalanceTransferRequest{
					Nonce:     genNonce(),
					RequestID: "2083067679084929026",
				},
			},
			wantErr:    true,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := pp.QueryTransfer(ctx, tt.args.req.RequestID)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			fmt.Printf("%+v\n", result)
		})
	}
}

func TestPayndaPaySDK_QueryCreateCardResult(t *testing.T) {
	tests := []struct {
		name       string
		args       string
		wantErr    bool
		wantErrMsg string
	}{
		// {
		// 	name:       "create card result in success",
		// 	args:       "2085633928568049665",
		// 	wantErr:    false,
		// 	wantErrMsg: "",
		// },
		{
			name:       "create card request id not found",
			args:       "2085633928213319665",
			wantErr:    true,
			wantErrMsg: "",
		},
		// {
		// 	name:       "create card result in failed",
		// 	args:       "2082732194991484929",
		// 	wantErr:    true,
		// 	wantErrMsg: "",
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pp.QueryCreateCardResult(ctx, tt.args)
			if err != nil && !tt.wantErr {
				t.Errorf("QueryCreateCardResult() error = %v", err)
			}
			fmt.Printf("%+v\n", got)
		})
	}
}

func TestPayndaPaySDK_CardStatusUpdate(t *testing.T) {
	tests := []struct {
		name       string
		call       func(*PayndaPaySDK, context.Context, *CardStatusUpdateRequest) error
		request    *CardStatusUpdateRequest
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "freeze card",
			call: (*PayndaPaySDK).CardFrozen,
			request: &CardStatusUpdateRequest{
				CardID:    "2007726970992115713",
				RequestID: genNonce(),
			},
		},
		{
			name: "unfreeze card",
			call: (*PayndaPaySDK).CardUnfrozen,
			request: &CardStatusUpdateRequest{
				CardID:    "1963497742025883650",
				RequestID: genNonce(),
			},
		},
		{
			name:       "freeze card ID empty",
			call:       (*PayndaPaySDK).CardFrozen,
			request:    &CardStatusUpdateRequest{RequestID: genNonce()},
			wantErr:    true,
			wantErrMsg: ErrEmptyCardID.Error(),
		},
		{
			name:       "unfreeze request ID empty",
			call:       (*PayndaPaySDK).CardUnfrozen,
			request:    &CardStatusUpdateRequest{CardID: "1963497742025883650"},
			wantErr:    true,
			wantErrMsg: ErrEmptyRequestID.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call(pp, ctx, tt.request)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				assert.Equal(t, tt.wantErrMsg, err.Error())
				return
			}
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
		})
	}
}

func TestPayndaPaySDK_QueryCardStatusUpdate(t *testing.T) {
	tests := []struct {
		name         string
		requestID    string
		wantErr      bool
		wantNotFound bool
	}{
		{
			name:         "request ID not used",
			requestID:    fmt.Sprintf("card-status-update-%d", time.Now().UnixNano()),
			wantErr:      true,
			wantNotFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := pp.QueryCardStatusUpdate(ctx, tt.requestID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				assert.Equal(t, tt.wantNotFound, IsRequestIDNotFound(err))
				return
			}
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			if result != nil {
				fmt.Printf("%+v\n", *result)
			}
		})
	}
}

func TestCardStatusUpdateRequestValidateNil(t *testing.T) {
	var request *CardStatusUpdateRequest
	if err := request.Validate(); err != ErrEmptyParams {
		t.Fatalf("Validate() error = %v, want %v", err, ErrEmptyParams)
	}
}

var _ PayndaPaySDKInterface = (*PayndaPaySDK)(nil)
