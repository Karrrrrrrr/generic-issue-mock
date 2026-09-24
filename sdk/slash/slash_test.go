package slash

import (
	"context"
	"errors"
	"fmt"
	"sdk/middleware/trace"
	"sdk/middleware/traffic"
	"testing"
	"time"
	"tman/enums"

	"github.com/avast/retry-go/v4"
	"github.com/go-kratos/kratos/v2/log"
)

var sdk *SlashSDK

// 模拟配置实现
type mockConf struct {
	apiKey         string
	account        string
	virtualAccount string
	// signingSecret  string
	cardGroup   string
	url         string
	vaultUrl    string
	restriction string
	countryList string
	debug       bool
	retryTimes  int32
	legalEntity string
}

func (m *mockConf) GetApiKey() string         { return m.apiKey }
func (m *mockConf) GetAccount() string        { return m.account }
func (m *mockConf) GetVirtualAccount() string { return m.virtualAccount }
func (m *mockConf) GetCardGroup() string      { return m.cardGroup }
func (m *mockConf) GetUrl() string            { return m.url }
func (m *mockConf) GetVaultUrl() string       { return m.vaultUrl }
func (m *mockConf) GetRestriction() string    { return m.restriction }
func (m *mockConf) GetCountryList() string    { return m.countryList }
func (m *mockConf) GetDebug() bool            { return m.debug }
func (m *mockConf) GetRetryTimes() int32      { return m.retryTimes }
func (m *mockConf) GetLegalEntity() string    { return m.legalEntity }

// 创建默认的模拟配置
func newMockConf() *mockConf {
	return &mockConf{
		apiKey:         "71b66af811dcdd060eefdcee0fb5ffa8a2ec22e84444332d881a7e4bf763b22d",
		account:        "sa_group_3cf9aqy5lhrvw",
		virtualAccount: "",
		cardGroup:      "",
		//url:            "http://localhost:8000",
		//vaultUrl:       "http://localhost:8000",
		url:         "https://api.joinslash.com",
		vaultUrl:    "https://vault.joinslash.com",
		countryList: "AX,AL,DZ,AS,AD,AO,AI,AQ,AG,AR,AM,AW,AU,AT,AZ,BS,BH,BD,BB,BY,BE,BZ,BJ,BM,BT,BO,BA,BW,BV,BR,IO,BN,BG,BF,BI,KH,CM,CA,CV,KY,CF,TD,CL,CN,CX,CC,CO,KM,CG,CD,CK,CR,CI,HR,CY,CZ,DK,DJ,DM,DO,EC,EG,SV,GQ,ER,EE,ET,FK,FO,FJ,FI,FR,GF,PF,TF,GA,GM,GE,DE,GH,GI,GR,GL,GD,GP,GU,GT,GG,GN,GW,GY,HT,HM,VA,HN,HK,HU,IS,IN,ID,IQ,IE,IM,IL,IT,JM,JP,JE,JO,KZ,KE,KI,KR,KW,KG,LA,LV,LB,LS,LR,LY,LI,LT,LU,MO,MK,MG,MW,MY,MV,ML,MT,MH,MQ,MR,MU,YT,MX,FM,MD,MC,MN,ME,MS,MA,MZ,NA,NR,NP,NL,NC,NZ,NI,NE,NG,NU,NF,MP,NO,OM,PK,PW,PS,PA,PG,PY,PE,PH,PN,PL,PT,PR,QA,RE,RO,RW,SH,KN,LC,PM,VC,WS,SM,ST,SA,SN,RS,SC,SL,SG,SK,SI,SB,SO,ZA,GS,ES,LK,SD,SR,SJ,SZ,SE,CH,TW,TJ,TZ,TH,TL,TG,TK,TO,TT,TN,TR,TM,TC,TV,UG,UA,AE,GB,US,UM,UY,UZ,VU,VE,VN,VG,VI,WF,EH,YE,ZM,ZW",
		restriction: RESTRICTION_ALLOWLIST,
		debug:       true,
	}

}

func TestMain(m *testing.M) {

	l := log.DefaultLogger
	sdk = New(newMockConf(), WithResponseMiddlewares(trace.ResponseTracking(), traffic.SlashTrafficLog(l)))

	m.Run()
}

// 测试CreateVirtualAccount方法
func TestSlashSDK_CreateVirtualAccount(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name       string
		args       args
		want       *VirtualAccountResponse
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "成功创建",
			args: args{
				name: "fluxo test VA",
			},
			want:       nil,
			wantErr:    false,
			wantErrMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.CreateVirtualAccount(context.Background(), tt.args.name)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("CreateVirtualAccount() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("CreateVirtualAccount(),got %v", err)
				return
			}
			if got != nil {
				fmt.Printf("CreateVirtualAccount() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// {
//    "virtualAccount": {
//       "id": "subaccount_kg69vtcp39hk",
//       "name": "Test VA",
//       "accountNumber": "606966150338460",
//       "routingNumber": "121145307",
//       "slashAccountGroupId": "sa_group_1sl0obja8okke",
//       "accountType": "default",
//       "accountId": "sa_group_1sl0obja8okke"
//    },
//    "commissionRule": {
//       "id": "subaccount_commission_rule_1s8pqhh3bkam1",
//       "virtualAccountId": "subaccount_kg69vtcp39hk",
//       "commissionDetails": {
//          "type": "flatFee",
//          "amount": {
//             "amountCents": 0
//          },
//          "frequency": "monthly",
//          "startDate": "2025-09-24T12:29:24.310926+08:00"
//       }
//    }
// }

func TestSlashSDK_GetVirtualAccount(t *testing.T) {
	tests := []struct {
		name       string
		want       *VirtualAccount
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "成功获取虚拟账户",
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.GetVirtualAccount(context.Background())
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("GetVirtualAccount() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("GetVirtualAccount() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("GetVirtualAccount() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试ListVirtualAccounts方法
func TestSlashSDK_ListVirtualAccounts(t *testing.T) {
	tests := []struct {
		name       string
		want       []*VirtualAccount
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "成功获取虚拟账户列表",
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.ListVirtualAccounts(context.Background())
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("ListVirtualAccounts() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("ListVirtualAccounts() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("ListVirtualAccounts() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试VirtualAccountTransfer方法
func TestSlashSDK_VirtualAccountTransfer(t *testing.T) {
	type args struct {
		param *TransferRequest
	}
	tests := []struct {
		name       string
		args       args
		want       *TransferResponse
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "成功转账",
			args: args{
				param: &TransferRequest{
					RequestID:   "20260112000123",
					Source:      "subaccount_kg69vtcp39hk",
					Destination: "subaccount_3ptxul4vzz07w",
					AmountCents: 6146,
				},
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.VirtualAccountTransfer(context.Background(), tt.args.param)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("VirtualAccountTransfer() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("VirtualAccountTransfer() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("VirtualAccountTransfer() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试ListCardProducts方法
func TestSlashSDK_ListCardProducts(t *testing.T) {
	tests := []struct {
		name       string
		want       []CardProduct
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "成功获取卡产品列表",
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.ListCardProducts(context.Background())
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("ListCardProducts() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("ListCardProducts() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("ListCardProducts() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

//	{
//	   "id": "c_1mqpgzfy9srn7",
//	   "name": "test_card",
//	   "last4": "",
//	   "accountId": "sa_group_1sl0obja8okke",
//	   "virtualAccountId": "subaccount_kg69vtcp39hk",
//	   "expiryYear": "",
//	   "expiryMonth": "",
//	   "cardGroupId": "card_group_17b9geat5ktux",
//	   "createdAt": "2025-09-24T07:02:31.606Z",
//	   "isPhysical": false,
//	   "isSingleUse": false,
//	   "status": "inactive",
//	   "userData": {
//	      "cardId": "1234",
//	      "requestId": "123456"
//	   }
//	}

//	{
//	   "id": "c_2599mnei13lkc",
//	   "name": "test_card1",
//	   "last4": "",
//	   "accountId": "sa_group_1sl0obja8okke",
//	   "virtualAccountId": "subaccount_kg69vtcp39hk",
//	   "expiryYear": "",
//	   "expiryMonth": "",
//	   "cardGroupId": "card_group_17b9geat5ktux",
//	   "createdAt": "2025-09-24T07:16:50.647Z",
//	   "cvv": null,
//	   "isPhysical": false,
//	   "isSingleUse": false,
//	   "pan": null,
//	   "status": "inactive",
//	   "userData": {
//	      "cardId": "12344555",
//	      "requestId": "123431313431"
//	   }
//
// 测试CreateCard方法
func TestSlashSDK_CreateCard(t *testing.T) {
	type args struct {
		param *CreateCardReq
	}
	tests := []struct {
		name       string
		args       args
		want       *Card
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "成功创建卡片",
			args: args{
				param: &CreateCardReq{
					RequestID: "642422224111",
					CardID:    "422234511",
					Name:      "test_card8",
					CardBinID: "card_product_2zwp1e2k7u0im",
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "空参数",
			args: args{
				param: nil,
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "参数不能为空",
		},
		{
			name: "缺少必要字段",
			args: args{
				param: &CreateCardReq{
					Name: "tttt",
				},
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "requestID 不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.CreateCard(context.Background(), tt.args.param)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("CreateCard() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("CreateCard() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("CreateCard() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

//	{
//	   "id": "c_2599mnei13lkc",
//	   "name": "test_card1",
//	   "last4": "6735",
//	   "accountId": "sa_group_1sl0obja8okke",
//	   "virtualAccountId": "subaccount_kg69vtcp39hk",
//	   "expiryYear": "2029",
//	   "expiryMonth": "03",
//	   "cardGroupId": "card_group_17b9geat5ktux",
//	   "createdAt": "2025-09-24T07:16:50.647Z",
//	   "cvv": "791",
//	   "isPhysical": false,
//	   "isSingleUse": false,
//	   "pan": "4361207805516735",
//	   "status": "active",
//	   "userData": {
//	      "cardId": "12344555",
//	      "requestId": "123431313431"
//	   },
//	   "cardProductId": "card_product_2zwp1e2k7u0im"
//	}
//
// 测试GetCard方法
func TestSlashSDK_GetCard(t *testing.T) {
	type args struct {
		cardID string
	}
	tests := []struct {
		name       string
		args       args
		want       *Card
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "空卡ID",
			args: args{
				cardID: "",
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "cardID不能为空",
		},
		{
			name: "成功获取卡片信息",
			args: args{
				cardID: "c_s0lj5jqnhwwi",
			},
			want:       nil,
			wantErr:    false,
			wantErrMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.GetCard(context.Background(), tt.args.cardID)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("GetCard() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("GetCard() got %v", err)
				return
			}
			if got != nil {
				fmt.Printf("GetCard() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试ListCards方法
func TestSlashSDK_ListCards(t *testing.T) {
	type args struct {
		params *QueryCardsParams
	}
	tests := []struct {
		name       string
		args       args
		want       []*Card
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "成功获取卡片列表",
			args: args{
				params: &QueryCardsParams{
					AccountId: "sa_group_3cf9aqy5lhrvw",
					Status:    "inactive",
				},
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.ListCards(context.Background(), tt.args.params)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("ListCards() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("ListCards() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("ListCards() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试FrozenCard方法
func TestSlashSDK_FrozenCard(t *testing.T) {
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
			name: "成功冻结卡片",
			args: args{
				cardID: "c_2599mnei13lkc",
			},
			wantErr: false,
		},
		{
			name: "空卡ID",
			args: args{
				cardID: "",
			},
			wantErr:    true,
			wantErrMsg: "cardID不能为空",
		},
		{
			name: "卡片不存在",
			args: args{
				cardID: "invalid-card",
			},
			wantErr:    true,
			wantErrMsg: "卡片不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sdk.FrozenCard(context.Background(), tt.args.cardID)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("FrozenCard() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("FrozenCard() err = %v", err)
			}
		})
	}
}

// 测试UnfrozenCard方法
func TestSlashSDK_UnfrozenCard(t *testing.T) {
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
			name: "成功解冻卡片",
			args: args{
				cardID: "c_s0lj5jqnhwwi",
			},
			wantErr: false,
		},
		{
			name: "空卡ID",
			args: args{
				cardID: "",
			},
			wantErr:    true,
			wantErrMsg: "cardID不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sdk.UnfrozenCard(context.Background(), tt.args.cardID)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("UnfrozenCard() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("UnfrozenCard() err = %v", err)
			}
		})
	}
}

// 测试ReleaseCard方法
func TestSlashSDK_ReleaseCard(t *testing.T) {
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
			name: "成功删除卡片",
			args: args{
				cardID: "test_card_id",
			},
			wantErr: false,
		},
		{
			name: "空卡ID",
			args: args{
				cardID: "",
			},
			wantErr:    true,
			wantErrMsg: "cardID不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sdk.ReleaseCard(context.Background(), tt.args.cardID)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("ReleaseCard() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("ReleaseCard() err = %v", err)
			}
		})
	}
}

// 测试GetCardUtilization方法
func TestSlashSDK_GetCardUtilization(t *testing.T) {
	type args struct {
		cardID string
	}
	tests := []struct {
		name       string
		args       args
		want       *Utilization
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "成功获取CardUtilization",
			args: args{
				cardID: "c_2599mnei13lkc",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "空卡ID",
			args: args{
				cardID: "",
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "cardID不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.GetCardUtilization(context.Background(), tt.args.cardID)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("GetCardUtilization() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("GetCardUtilization() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("GetCardUtilization() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试UpdateSpendingConstraint方法
func TestSlashSDK_UpdateSpendingConstraint(t *testing.T) {
	type args struct {
		cardID string
		req    *UpdateCardSpendingConstraintRequest
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "成功更新SpendingConstraint",
			args: args{
				cardID: "c_2599mnei13lkc",
				req:    nil,
			},
			wantErr: false,
		},
		{
			name: "空参数",
			args: args{
				cardID: "c_2599mnei13lkc",
				req:    nil,
			},
			wantErr:    true,
			wantErrMsg: "参数不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sdk.UpdateSpendingConstraint(context.Background(), tt.args.cardID, tt.args.req)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("UpdateSpendingConstraint() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateSpendingConstraint() err = %v", err)
			}
		})
	}
}

// 测试SetSpendingConstraint方法
func TestSlashSDK_SetSpendingConstraint(t *testing.T) {
	type args struct {
		cardID string
		req    *SetCardSpendingConstraintRequest
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "成功设置SpendingConstraint",
			args: args{
				cardID: "c_3ft8avhygck8m",
				req: &SetCardSpendingConstraintRequest{
					CountryRule: &CountryRule{
						// Countries:   []string{"KR", "IR"}, //  []string{"US", "CN"},
						// Restriction: RESTRICTION_BLACKLIST,
						Restriction: RESTRICTION_ALLOWLIST,
						Countries:   []string{"AX", "AL", "DZ", "AS", "AD", "AO", "AI", "AQ", "AG", "AR", "AM", "AW", "AU", "AT", "AZ", "BS", "BH", "BD", "BB", "BY", "BE", "BZ", "BJ", "BM", "BT", "BO", "BA", "BW", "BV", "BR", "IO", "BN", "BG", "BF", "BI", "KH", "CM", "CA", "CV", "KY", "CF", "TD", "CL", "CN", "CX", "CC", "CO", "KM", "CG", "CD", "CK", "CR", "CI", "HR", "CY", "CZ", "DK", "DJ", "DM", "DO", "EC", "EG", "SV", "GQ", "ER", "EE", "ET", "FK", "FO", "FJ", "FI", "FR", "GF", "PF", "TF", "GA", "GM", "GE", "DE", "GH", "GI", "GR", "GL", "GD", "GP", "GU", "GT", "GG", "GN", "GW", "GY", "HT", "HM", "VA", "HN", "HK", "HU", "IS", "IN", "ID", "IQ", "IE", "IM", "IL", "IT", "JM", "JP", "JE", "JO", "KZ", "KE", "KI", "KR", "KW", "KG", "LA", "LV", "LB", "LS", "LR", "LY", "LI", "LT", "LU", "MO", "MK", "MG", "MW", "MY", "MV", "ML", "MT", "MH", "MQ", "MR", "MU", "YT", "MX", "FM", "MD", "MC", "MN", "ME", "MS", "MA", "MZ", "NA", "NR", "NP", "NL", "NC", "NZ", "NI", "NE", "NG", "NU", "NF", "MP", "NO", "OM", "PK", "PW", "PS", "PA", "PG", "PY", "PE", "PH", "PN", "PL", "PT", "PR", "QA", "RE", "RO", "RW", "SH", "KN", "LC", "PM", "VC", "WS", "SM", "ST", "SA", "SN", "RS", "SC", "SL", "SG", "SK", "SI", "SB", "SO", "ZA", "GS", "ES", "LK", "SD", "SR", "SJ", "SZ", "SE", "CH", "TW", "TJ", "TZ", "TH", "TL", "TG", "TK", "TO", "TT", "TN", "TR", "TM", "TC", "TV", "UG", "UA", "AE", "GB", "US", "UM", "UY", "UZ", "VU", "VE", "VN", "VG", "VI", "WF", "EH", "YE", "ZM", "ZW"},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sdk.SetSpendingConstraint(context.Background(), tt.args.cardID, tt.args.req)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("SetSpendingConstraint() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("SetSpendingConstraint() err = %v", err)
			}
		})
	}
}

// 测试GetCardModifiers方法 (无权限)
func TestSlashSDK_GetCardModifiers(t *testing.T) {
	type args struct {
		cardID string
	}
	tests := []struct {
		name       string
		args       args
		want       []CardModifier
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "成功获取CardModifier",
			args: args{
				cardID: "c_2599mnei13lkc",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "空卡ID",
			args: args{
				cardID: "",
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "cardID不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.GetCardModifiers(context.Background(), tt.args.cardID)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("GetCardModifiers() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("GetCardModifiers() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("GetCardModifiers() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试SetCardModifiers方法
func TestSlashSDK_SetCardModifiers(t *testing.T) {
	type args struct {
		cardID    string
		modifiers SetCardModifierRequest
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "成功设置CardModifier",
			args: args{
				cardID: "c_2599mnei13lkc",
				modifiers: SetCardModifierRequest{
					Name:  "international",
					Value: true,
				},
			},
			wantErr: false,
		},
		{
			name: "空卡ID",
			args: args{
				cardID: "",
			},
			wantErr:    true,
			wantErrMsg: "cardID不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sdk.SetCardModifiers(context.Background(), tt.args.cardID, tt.args.modifiers)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("SetCardModifiers() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("SetCardModifiers() err = %v", err)
			}
		})
	}
}

// 测试ListCardGroups方法
func TestSlashSDK_ListCardGroups(t *testing.T) {
	tests := []struct {
		name       string
		want       []CardGroup
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "成功获取CardGroup列表",
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.ListCardGroups(context.Background(), "")
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("ListCardGroups() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("ListCardGroups() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("ListCardGroups() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试GetCardGroup方法
func TestSlashSDK_GetCardGroup(t *testing.T) {
	type args struct {
		cardGroupID string
	}
	tests := []struct {
		name       string
		args       args
		want       *CardGroup
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "成功获取CardGroup",
			args: args{
				cardGroupID: "card_group_17b9geat5ktux",
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.GetCardGroup(context.Background(), tt.args.cardGroupID)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("GetCardGroup() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("GetCardGroup() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("GetCardGroup() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试CreateCardGroup方法
func TestSlashSDK_CreateCardGroup(t *testing.T) {
	type args struct {
		req CreateCardGroupRequest
	}
	tests := []struct {
		name       string
		args       args
		want       *CardGroup
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "成功创建CardGroup",
			args: args{
				req: CreateCardGroupRequest{
					Name: "fluxo Test Group",
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "空名称",
			args: args{
				req: CreateCardGroupRequest{
					Name: "",
				},
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "name不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.CreateCardGroup(context.Background(), tt.args.req)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("CreateCardGroup() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("CreateCardGroup() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("CreateCardGroup() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试UpdateCardGroup方法
func TestSlashSDK_UpdateCardGroup(t *testing.T) {
	type args struct {
		cardGroupID string
		req         UpdateCardGroupRequest
	}
	tests := []struct {
		name       string
		args       args
		want       *CardGroup
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "成功更新CardGroup",
			args: args{
				cardGroupID: "card_group_jkqfczb6ceox",
				req: UpdateCardGroupRequest{
					Name: "Product Card Group",
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "空CardGroupID",
			args: args{
				cardGroupID: "",
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "cardGroupID不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.UpdateCardGroup(context.Background(), tt.args.cardGroupID, tt.args.req)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("UpdateCardGroup() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateCardGroup() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("UpdateCardGroup() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试UpdateCardGroupSpendingConstraint方法
func TestSlashSDK_UpdateCardGroupSpendingConstraint(t *testing.T) {
	type args struct {
		cardGroupID string
		req         UpdateCardGroupSpendingConstraintRequest
	}
	tests := []struct {
		name       string
		args       args
		want       *SpendingConstraint
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "成功更新CardGroup SpendingConstraint",
			args: args{
				cardGroupID: "card_group_17b9geat5ktux",
				req: UpdateCardGroupSpendingConstraintRequest{
					CountryRule: &CountryRule{
						Restriction: RESTRICTION_ALLOWLIST,
						Countries:   []string{"AX", "AL", "DZ", "AS", "AD", "AO", "AI", "AQ", "AG", "AR", "AM", "AW", "AU", "AT", "AZ", "BS", "BH", "BD", "BB", "BY", "BE", "BZ", "BJ", "BM", "BT", "BO", "BA", "BW", "BV", "BR", "IO", "BN", "BG", "BF", "BI", "KH", "CM", "CA", "CV", "KY", "CF", "TD", "CL", "CN", "CX", "CC", "CO", "KM", "CG", "CD", "CK", "CR", "CI", "HR", "CY", "CZ", "DK", "DJ", "DM", "DO", "EC", "EG", "SV", "GQ", "ER", "EE", "ET", "FK", "FO", "FJ", "FI", "FR", "GF", "PF", "TF", "GA", "GM", "GE", "DE", "GH", "GI", "GR", "GL", "GD", "GP", "GU", "GT", "GG", "GN", "GW", "GY", "HT", "HM", "VA", "HN", "HK", "HU", "IS", "IN", "ID", "IQ", "IE", "IM", "IL", "IT", "JM", "JP", "JE", "JO", "KZ", "KE", "KI", "KR", "KW", "KG", "LA", "LV", "LB", "LS", "LR", "LY", "LI", "LT", "LU", "MO", "MK", "MG", "MW", "MY", "MV", "ML", "MT", "MH", "MQ", "MR", "MU", "YT", "MX", "FM", "MD", "MC", "MN", "ME", "MS", "MA", "MZ", "NA", "NR", "NP", "NL", "NC", "NZ", "NI", "NE", "NG", "NU", "NF", "MP", "NO", "OM", "PK", "PW", "PS", "PA", "PG", "PY", "PE", "PH", "PN", "PL", "PT", "PR", "QA", "RE", "RO", "RW", "SH", "KN", "LC", "PM", "VC", "WS", "SM", "ST", "SA", "SN", "RS", "SC", "SL", "SG", "SK", "SI", "SB", "SO", "ZA", "GS", "ES", "LK", "SD", "SR", "SJ", "SZ", "SE", "CH", "TW", "TJ", "TZ", "TH", "TL", "TG", "TK", "TO", "TT", "TN", "TR", "TM", "TC", "TV", "UG", "UA", "AE", "GB", "US", "UM", "UY", "UZ", "VU", "VE", "VN", "VG", "VI", "WF", "EH", "YE", "ZM", "ZW"},
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
		// {
		// 	name: "空CardGroupID",
		// 	args: args{
		// 		cardGroupID: "",
		// 	},
		// 	want:       nil,
		// 	wantErr:    true,
		// 	wantErrMsg: "cardGroupID不能为空",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.UpdateCardGroupSpendingConstraint(context.Background(), tt.args.cardGroupID, tt.args.req)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("UpdateCardGroupSpendingConstraint() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateCardGroupSpendingConstraint() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("UpdateCardGroupSpendingConstraint() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试SetCardGroupSpendingConstraint方法
func TestSlashSDK_SetCardGroupSpendingConstraint(t *testing.T) {
	type args struct {
		cardGroupID string
		req         SetCardGroupSpendingConstraintRequest
	}
	tests := []struct {
		name       string
		args       args
		want       *SpendingConstraint
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "成功设置CardGroup SpendingConstraint",
			args: args{
				cardGroupID: "card_group_17b9geat5ktux",
				req:         SetCardGroupSpendingConstraintRequest{},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "空CardGroupID",
			args: args{
				cardGroupID: "",
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "cardGroupID不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.SetCardGroupSpendingConstraint(context.Background(), tt.args.cardGroupID, tt.args.req)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("SetCardGroupSpendingConstraint() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("SetCardGroupSpendingConstraint() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("SetCardGroupSpendingConstraint() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试GetCardGroupUtilization方法
func TestSlashSDK_GetCardGroupUtilization(t *testing.T) {
	type args struct {
		cardGroupID string
	}
	tests := []struct {
		name       string
		args       args
		want       *Utilization
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "成功获取CardGroup Utilization",
			args: args{
				cardGroupID: "card_group_17b9geat5ktux",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "空CardGroupID",
			args: args{
				cardGroupID: "",
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "cardGroupID不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.GetCardGroupUtilization(context.Background(), tt.args.cardGroupID)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("GetCardGroupUtilization() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("GetCardGroupUtilization() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("GetCardGroupUtilization() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试ListMerchants方法
func TestSlashSDK_ListMerchants(t *testing.T) {
	tests := []struct {
		name       string
		want       []Merchant
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "成功获取商户列表",
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.ListMerchants(context.Background())
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("ListMerchants() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("ListMerchants() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("ListMerchants() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试GetMerchant方法
func TestSlashSDK_GetMerchant(t *testing.T) {
	type args struct {
		merchantID string
	}
	tests := []struct {
		name       string
		args       args
		want       *Merchant
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "成功获取商户信息",
			args: args{
				merchantID: "merchant_v2_v4vw8zhdiaro",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "空商户ID",
			args: args{
				merchantID: "",
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "merchantID不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.GetMerchant(context.Background(), tt.args.merchantID)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("GetMerchant() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("GetMerchant() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("GetMerchant() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试ListMerchantCategories方法
func TestSlashSDK_ListMerchantCategories(t *testing.T) {
	tests := []struct {
		name       string
		want       []MerchantCategory
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "成功获取商户分类列表",
			want:    nil,
			wantErr: false,
		},
		{
			name:       "API返回错误",
			want:       nil,
			wantErr:    true,
			wantErrMsg: "获取商户分类列表失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.ListMerchantCategories(context.Background())
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("ListMerchantCategories() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("ListMerchantCategories() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("ListMerchantCategories() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试ListWebhooks方法
func TestSlashSDK_ListWebhooks(t *testing.T) {
	tests := []struct {
		name       string
		want       []Webhook
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "成功获取Webhook列表",
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.ListWebhooks(context.Background())
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("ListWebhooks() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("ListWebhooks() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("ListWebhooks() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试CreateWebhook方法
func TestSlashSDK_CreateWebhook(t *testing.T) {
	type args struct {
		req CreateWebhookRequest
	}
	tests := []struct {
		name       string
		args       args
		want       *Webhook
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "成功创建Webhook",
			args: args{
				req: CreateWebhookRequest{
					Name: "Product Webhook",
					URL:  "https://x-test-api.ptmlink.com/api/v1/notify/xz-event",
				},
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.CreateWebhook(context.Background(), tt.args.req)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("CreateWebhook() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("CreateWebhook() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("CreateWebhook() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试UpdateWebhook方法
func TestSlashSDK_UpdateWebhook(t *testing.T) {
	type args struct {
		req UpdateWebhookRequest
	}
	tests := []struct {
		name       string
		args       args
		want       *Webhook
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "成功更新Webhook",
			args: args{
				req: UpdateWebhookRequest{
					WebhookID: "public_webhook_endpoint_22txyz3547pba",
					Status:    "archived",
					Reason:    "remove test webhook",
				},
			},
			want:    nil,
			wantErr: false,
		},
		// {
		// 	name: "空Webhook ID",
		// 	args: args{
		// 		req: UpdateWebhookRequest{
		// 			WebhookID: "",
		// 		},
		// 	},
		// 	want:       nil,
		// 	wantErr:    true,
		// 	wantErrMsg: "webhookID不能为空",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.UpdateWebhook(context.Background(), tt.args.req)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("UpdateWebhook() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateWebhook() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("UpdateWebhook() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试GetAuthWebhook方法
func TestSlashSDK_GetAuthWebhook(t *testing.T) {
	tests := []struct {
		name       string
		want       *AuthWebhook
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "成功获取授权Webhook",
			want:    nil,
			wantErr: false,
		},
		{
			name:       "API返回错误",
			want:       nil,
			wantErr:    true,
			wantErrMsg: "获取授权Webhook失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.GetAuthWebhook(context.Background())
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("GetAuthWebhook() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("GetAuthWebhook() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("GetAuthWebhook() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试SetAuthWebhook方法
func TestSlashSDK_SetAuthWebhook(t *testing.T) {
	type args struct {
		req UpdateAuthWebhookRequest
	}
	tests := []struct {
		name       string
		args       args
		want       *AuthWebhook
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "成功设置授权Webhook",
			args: args{
				req: UpdateAuthWebhookRequest{
					// WebhookUrl: "https://vcc-sl.ptmlink.com/api/v1/notify/xz-authorization",
					// WebhookUrl: "https://vcc-sl2.ptmlink.com/api/v1/notify/xz-authorization", // 生产
					WebhookUrl: "https://x-test-api.ptmlink1.com/api/v1/notify/xz-authorization",
					Status:     "active",
					Config: AuthWebhookConfig{
						FallbackBehavior: "reject",
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.SetAuthWebhook(context.Background(), tt.args.req)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("SetAuthWebhook() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("SetAuthWebhook() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("SetAuthWebhook() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试ListTransactions方法
func TestSlashSDK_ListTransactions(t *testing.T) {
	type args struct {
		params QueryTransactionsParams
	}
	tests := []struct {
		name       string
		args       args
		want       []Transaction
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "成功获取交易列表",
			args: args{
				params: QueryTransactionsParams{
					ProviderAuthorizationId: "00010348ed29516e-21c2-4f16-8b00-472f9552263a",
					// FilterDetailedStatus: "refund",
				},
			},
			want:    nil,
			wantErr: false,
		},
		// {
		// 	name: "带参数查询",
		// 	args: args{
		// 		params: QueryTransactionsParams{
		// 			FilterStatus: "pending",
		// 		},
		// 	},
		// 	want:    nil,
		// 	wantErr: false,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.ListTransactions(context.Background(), tt.args.params)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("ListTransactions() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("ListTransactions() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("ListTransactions() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试GetTransactionAggregations方法
func TestSlashSDK_GetTransactionAggregations(t *testing.T) {
	type args struct {
		params QueryTransactionAggregationsParams
	}
	tests := []struct {
		name       string
		args       args
		want       *TransactionAggregation
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "成功获取交易聚Aggregations",
			args: args{
				params: QueryTransactionAggregationsParams{},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "带过滤条件查询",
			args: args{
				params: QueryTransactionAggregationsParams{
					FilterStatus: "pending",
				},
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.GetTransactionAggregations(context.Background(), tt.args.params)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("GetTransactionAggregations() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("GetTransactionAggregations() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("GetTransactionAggregations() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试GetTransaction方法
func TestSlashSDK_GetTransaction(t *testing.T) {
	type args struct {
		transactionID string
	}
	tests := []struct {
		name       string
		args       args
		want       *Transaction
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "成功获取交易详情",
			args: args{
				transactionID: "122",
			},
			want:    nil,
			wantErr: false,
		},
		// {
		// 	name: "空交易ID",
		// 	args: args{
		// 		transactionID: "",
		// 	},
		// 	want:       nil,
		// 	wantErr:    true,
		// 	wantErrMsg: "transactionID不能为空",
		// },
		// {
		// 	name: "API返回错误",
		// 	args: args{
		// 		transactionID: "txn-123",
		// 	},
		// 	want:       nil,
		// 	wantErr:    true,
		// 	wantErrMsg: "获取交易详情失败",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.GetTransaction(context.Background(), tt.args.transactionID)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("GetTransaction() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("GetTransaction() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("GetTransaction() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

// 测试GetTransactionFeeDetails方法
func TestSlashSDK_GetTransactionFeeDetails(t *testing.T) {
	type args struct {
		transactionID string
	}
	tests := []struct {
		name       string
		args       args
		want       []Fee
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "成功获取交易费用详情",
			args: args{
				transactionID: "txn-123",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "空交易ID",
			args: args{
				transactionID: "",
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "transactionID不能为空",
		},
		{
			name: "API返回错误",
			args: args{
				transactionID: "txn-123",
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "获取交易费用详情失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.GetTransactionFeeDetails(context.Background(), tt.args.transactionID)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("GetTransactionFeeDetails() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("GetTransactionFeeDetails() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("GetTransactionFeeDetails() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

func TestSlashSDK_ListAccounts(t *testing.T) {
	tests := []struct {
		name       string
		want       []Account
		wantErr    bool
		wantErrMsg string
	}{
		// TODO: Add test cases.
		{
			name:    "成功获取账户列表",
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.ListAccounts(context.Background())
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("ListAccounts() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("ListAccounts() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("ListAccounts() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

func TestSlashSDK_GetAccount(t *testing.T) {
	tests := []struct {
		name       string
		want       []Account
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "成功获取账户",
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.GetAccount(context.Background())
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("GetAccount() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("GetAccount() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("GetAccount() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

func TestSlashSDK_ListAccountBalances(t *testing.T) {
	tests := []struct {
		name       string
		want       []Account
		wantErr    bool
		wantErrMsg string
	}{
		// TODO: Add test cases.
		{
			name:    "成功获取账户资金",
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.ListAccountBalances(context.Background())
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("ListAccountBalances() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("ListAccountBalances() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("ListAccountBalances() = %v, Want %v\n", got, tt.want)
			}
		})
	}
}

func Test_slashSDKRepo_Retry(t *testing.T) {
	ctx := context.Background()

	retryOptions := []retry.Option{
		retry.RetryIf(func(err error) bool {
			return err.Error() == CARD_INACTIVE_ERR
		}),
		retry.Attempts(DEFAULT_RETRY_TIMES),
		retry.Delay(time.Second), // 查询间隔
		retry.DelayType(retry.FixedDelay),
		retry.LastErrorOnly(true),
		retry.Context(ctx),
	}

	// 卡信息查询
	qryFunc := func() (*Card, error) {
		card, getErr := sdk.GetCard(ctx, "c_2wh99n4u1yvs2")
		if getErr != nil {
			return nil, getErr
		}
		if card.Status != CARD_STATUS_INACTIVE {
			return card, errors.New(CARD_INACTIVE_ERR)
		}
		return card, nil
	}

	card, err := retry.DoWithData(qryFunc, retryOptions...)

	fmt.Println(card, err)
}

func Test_BatchSetCardSpendingConstraint(t *testing.T) {
	next := ""
	for {
		res, err := sdk.ListCards(context.Background(), &QueryCardsParams{Cursor: next})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		cards := res.Items
		for _, card := range cards {
			err = sdk.SetSpendingConstraint(context.Background(), card.ID, &SetCardSpendingConstraintRequest{
				CountryRule: &CountryRule{
					Countries:   []string{string(enums.COUNTRY_TWO_ENUM_KP), string(enums.COUNTRY_TWO_ENUM_IR)},
					Restriction: RESTRICTION_BLACKLIST,
				},
			})
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		}
		if res.Metadata.NextCursor == "" {
			break
		}
		next = res.Metadata.NextCursor
	}
}

func TestSlashSDK_UpdateVirtualAccount(t *testing.T) {
	type args struct {
		vaId string
		name string
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		// TODO: Add test cases.
		{
			name: "成功更新虚拟账户",
			args: args{
				vaId: "subaccount_2an46anvbusw",
				name: "Test VA2",
			},
			wantErrMsg: "",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdk.UpdateVirtualAccount(context.Background(), tt.args.vaId, tt.args.name)
			if tt.wantErr {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("UpdateVirtualAccount() err = %v, WantErr %v", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateVirtualAccount() err = %v", err)
				return
			}
			if got != nil {
				fmt.Printf("UpdateVirtualAccount() = %v\n", got)
			}
		})
	}
}

func Test_ListCards(t *testing.T) {
	aclist := []string{}
	next := ""
	for {
		res, err := sdk.ListCards(context.Background(),
			&QueryCardsParams{
				Cursor:           next,
				VirtualAccountId: "subaccount_301xh9119kjjn",
				Status:           "inactive",
			})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		cards := res.Items
		for _, card := range cards {
			aclist = append(aclist, card.ID)
		}
		if res.Metadata.NextCursor == "" {
			break
		}
		next = res.Metadata.NextCursor
	}
	fmt.Println(aclist)
}

func TestSlashSDK_ListLegalEntities(t *testing.T) {

	got, err := sdk.ListLegalEntity(context.Background())

	if err != nil {
		t.Errorf("ListLegalEntities() err = %v", err)
		return
	}
	if got != nil {
		fmt.Printf("ListLegalEntities() = %v\n", got)
	}
}
