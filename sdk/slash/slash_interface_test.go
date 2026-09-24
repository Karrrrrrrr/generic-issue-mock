package slash_test

import (
	"context"
	"fmt"
	"sdk/slash"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

var slashSdk slash.SlashSDKInterface
var slashLog log.Logger

// func TestMain(m *testing.M) {
// 	slashLog = log.DefaultLogger
// 	slash.WithResponseMiddlewares(trace.ResponseTracking(), traffic.SlashTrafficLog(slashLog))
// 	slashSdk = slash.NewSlashSDK(true, slash.SlashSDKConf{
// 		ApiKey:           "f708833247dbc1269e22c1cea801726519f0973769f2f7f12037391821f8296a",
// 		AccountID:        "sa_group_1sl0obja8okke", //subaccount_3ptxul4vzz07w
// 		VirtualAccountID: "subaccount_kg69vtcp39hk",
// 		CardGroupID:      "card_group_17b9geat5ktux",
// 		Url:              "https://api.joinslash.com",
// 		VaultUrl:         "https://vault.joinslash.com",
// 	})
// 	m.Run()

// }

// TestSubmitCreateCard 测试开卡
func TestCardPost(t *testing.T) {
	requestID := fmt.Sprintf("test_card_%d", time.Now().UnixNano())

	result, err := slashSdk.CardPost(context.Background(), &slash.CardPostRequest{
		Name:          "testName",
		RequestID:     requestID,
		CardProductID: "card_product_2zwp1e2k7u0im",
	})
	if err != nil {
		t.Fatal("submitCreateCard err:", err.Error())
	}

	t.Logf("result, %v", result)
}

// TestGetCardByID 测试根据卡ID获取卡信息
func TestGetCardByID(t *testing.T) {
	requestID := fmt.Sprintf("test_card_%d", time.Now().UnixNano())

	cardID, err := slashSdk.CardPost(context.Background(), &slash.CardPostRequest{
		Name:          "testName",
		RequestID:     requestID,
		CardProductID: "card_product_2zwp1e2k7u0im",
	})
	if err != nil {
		t.Fatal("submitCreateCard err:", err.Error())
	}

	result, err := slashSdk.GetCardByID(context.Background(), cardID)
	if err != nil {
		t.Fatal("GetCardByID err:", err.Error())
	}

	t.Logf("result, %v", result)
}

func TestCardStatusUpdate(t *testing.T) {
	tests := []struct {
		name string
		call func(slash.SlashSDKInterface, context.Context, string) error
	}{
		{name: "freeze card ID empty", call: func(sdk slash.SlashSDKInterface, ctx context.Context, cardID string) error {
			return sdk.CardFrozen(ctx, cardID)
		}},
		{name: "unfreeze card ID empty", call: func(sdk slash.SlashSDKInterface, ctx context.Context, cardID string) error {
			return sdk.CardUnfrozen(ctx, cardID)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call(slashSdk, context.Background(), "")
			if err != slash.ErrEmptyCardID {
				t.Errorf("CardStatusUpdate() error = %v, want %v", err, slash.ErrEmptyCardID)
			}
		})
	}
}
