package pingpong

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewPingPongSDKInterface 验证 biz 层可通过接口调用各方法，且业务调用不会自动获取令牌。
func TestNewPingPongSDKInterface(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privateKey := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}))

	// cases 定义各接口对应的请求、模拟响应和业务调用。
	cases := []struct {
		name     string                                            // 子测试名称。
		method   string                                            // 预期 HTTP 方法。
		path     string                                            // 预期请求路径。
		query    map[string]string                                 // 预期 URL 查询参数。
		body     map[string]any                                    // 预期 JSON 请求字段。
		response string                                            // 模拟网关响应。
		call     func(context.Context, PingPongSDKInterface) error // 通过对外接口发起调用。
	}{
		{
			name: "获取访问令牌", method: http.MethodGet, path: "/v2/token/get",
			response: `{"code":0,"data":{"access_token":"new-token","expires_in":7200}}`,
			call: func(ctx context.Context, sdk PingPongSDKInterface) error {
				_, err := sdk.GetAccessToken(ctx)
				return err
			},
		},
		{
			name: "查询账户余额", method: http.MethodPost, path: "/api/reporting/v3/account-balance",
			response: `{"code":0,"data":{"item_list":[]}}`,
			call: func(ctx context.Context, sdk PingPongSDKInterface) error {
				_, err := sdk.QueryAccountsBalances(ctx, "existing-token", &QueryAccountsBalancesRequest{PageNo: 1})
				return err
			},
		},
		{
			name: "查询卡产品", method: http.MethodGet, path: "/api/issuing/v3/card-products",
			response: `{"code":0,"data":{"product_list":[]}}`,
			call: func(ctx context.Context, sdk PingPongSDKInterface) error {
				_, err := sdk.QueryCardProducts(ctx, "existing-token")
				return err
			},
		},
		{
			name: "创建卡片", method: http.MethodPost, path: "/api/issuing/v3/cards/apply",
			body: map[string]any{
				"request_id": "request-1", "card_product_code": "product-1",
				"card_currency": "USD", "budget_id": "budget-1", "coupon_applied": false,
			},
			response: `{"code":0,"data":{"card_id":"card-1"}}`,
			call: func(ctx context.Context, sdk PingPongSDKInterface) error {
				_, err := sdk.CreateCard(ctx, "existing-token", &CreateCardRequest{
					RequestID: "request-1", CardProductCode: "product-1",
					CardCurrency: "USD", BudgetID: "budget-1", DailyLimit: &Amount{Amount: 100, Currency: "USD"},
				})
				return err
			},
		},
		{
			name: "获取卡详情", method: http.MethodGet, path: "/api/issuing/v3/cards/details",
			response: `{"code":0,"data":{"card_id":"card-1"}}`,
			call: func(ctx context.Context, sdk PingPongSDKInterface) error {
				_, err := sdk.GetCardDetails(ctx, "existing-token", "card-1")
				return err
			},
		},
		{
			name: "卡充值转出", method: http.MethodPost, path: "/api/issuing/v3/cards/funding/actions",
			response: `{"code":0,"data":{"record_id":"record-1"}}`,
			call: func(ctx context.Context, sdk PingPongSDKInterface) error {
				_, err := sdk.CardFunding(ctx, "existing-token", &CardFundingRequest{CardID: "card-1", Action: "top_up", Amount: 1, UniqueOrderID: "order-1"})
				return err
			},
		},
		{
			name: "查询卡资金订单", method: http.MethodGet, path: "/api/issuing/v3/card/funding/orders",
			response: `{"code":0,"data":{"list":[]}}`,
			call: func(ctx context.Context, sdk PingPongSDKInterface) error {
				_, err := sdk.QueryCardFundingOrders(ctx, "existing-token", &QueryCardFundingOrdersRequest{CardID: "card-1"})
				return err
			},
		},
		{
			name: "查询卡余额", method: http.MethodGet, path: "/api/issuing/v3/card/balance",
			response: `{"code":0,"data":{"currency":"USD"}}`,
			call: func(ctx context.Context, sdk PingPongSDKInterface) error {
				_, err := sdk.QueryDedicatedCardBalance(ctx, "existing-token", "card-1")
				return err
			},
		},
		{
			name: "卡操作", method: http.MethodPost, path: "/api/issuing/v3/cards/actions",
			response: `{"code":0,"data":{}}`,
			call: func(ctx context.Context, sdk PingPongSDKInterface) error {
				return sdk.CardAction(ctx, "existing-token", &CardActionRequest{CardID: "card-1", Action: "freeze"})
			},
		},
		{
			name: "查询卡交易", method: http.MethodGet, path: "/api/issuing/v3/transactions",
			response: `{"code":0,"data":{"transaction_list":[]}}`,
			call: func(ctx context.Context, sdk PingPongSDKInterface) error {
				_, err := sdk.QueryCardTransactions(ctx, "existing-token", &QueryCardTransactionsRequest{CardID: "card-1"})
				return err
			},
		},
		{
			name: "查询3DS详情", method: http.MethodGet, path: "/api/issuing/v3/cards/3ds/details",
			response: `{"code":0,"data":{"cardholder_id":"holder-1"}}`,
			call: func(ctx context.Context, sdk PingPongSDKInterface) error {
				_, err := sdk.Query3DSDetails(ctx, "existing-token", "card-1")
				return err
			},
		},
		{
			name: "创建预算账户", method: http.MethodPost, path: "/api/issuing/v3/budgets",
			body:     map[string]any{"budget_name": "test"},
			response: `{"code":0,"data":{"budget_id":"budget-1"}}`,
			call: func(ctx context.Context, sdk PingPongSDKInterface) error {
				_, err := sdk.CreateBudgetAccount(ctx, "existing-token", &CreateBudgetAccountRequest{BudgetName: "test"})
				return err
			},
		},
		{
			name: "预算账户划转", method: http.MethodPost, path: "/api/issuing/v3/budgets/funding",
			body:     map[string]any{"unique_order_id": "order-1", "budget_id": "budget-1", "action": "transfer", "amount": float64(1), "currency": "USD", "target_budget_id": "budget-2", "target_currency": "USD"},
			response: `{"code":0,"data":{"record_id":"record-1"}}`,
			call: func(ctx context.Context, sdk PingPongSDKInterface) error {
				_, err := sdk.BudgetFunding(ctx, "existing-token", &BudgetFundingRequest{UniqueOrderID: "order-1", BudgetID: "budget-1", Action: "transfer", Amount: 1, Currency: "USD", TargetBudgetID: "budget-2", TargetCurrency: "USD"})
				return err
			},
		},
		{
			name: "查询预算资金订单", method: http.MethodGet, path: "/api/issuing/v3/funding/orders",
			query:    map[string]string{"order_id": "order-1", "action": "top_up"},
			response: `{"code":0,"data":{"order_id":"order-1","action":"top_up","status":"SUCCESS"}}`,
			call: func(ctx context.Context, sdk PingPongSDKInterface) error {
				_, err := sdk.QueryBudgetFundingOrder(ctx, "existing-token", &QueryBudgetFundingOrderRequest{OrderID: "order-1", Action: "top_up"})
				return err
			},
		},
		{
			name: "查询预算账户余额", method: http.MethodGet, path: "/api/issuing/v3/budget/balance",
			query:    map[string]string{"budget_id": "budget-1"},
			response: `{"code":0,"data":{"balance_list":[{"budget_id":"budget-1","balance":10,"currency":"USD"}]}}`,
			call: func(ctx context.Context, sdk PingPongSDKInterface) error {
				_, err := sdk.QueryBudgetAccountBalance(ctx, "existing-token", "budget-1")
				return err
			},
		},
		{
			name: "查询账户交易", method: http.MethodGet, path: "/api/issuing/v4/account/transactions",
			query:    map[string]string{"budget_id": "budget-1", "posting_start_time": "2026-09-01T00:00:00+08:00", "posting_end_time": "2026-09-02T00:00:00+08:00"},
			response: `{"code":0,"data":{"total_num":1,"transaction_list":[{"transaction_id":"txn-1","billing_amount":10}]}}`,
			call: func(ctx context.Context, sdk PingPongSDKInterface) error {
				_, err := sdk.QueryAccountTransactions(ctx, "existing-token", &QueryAccountTransactionsRequest{BudgetID: "budget-1", PostingStartTime: "2026-09-01T00:00:00+08:00", PostingEndTime: "2026-09-02T00:00:00+08:00"})
				return err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != tc.method || r.URL.Path != tc.path {
					t.Errorf("请求为 %s %s，期望 %s %s", r.Method, r.URL.Path, tc.method, tc.path)
				}
				if tc.path == "/v2/token/get" {
					if r.URL.Query().Get("app_id") != "app" || r.URL.Query().Get("app_secret") != "secret" || r.Header.Get("Authorization") != "" {
						t.Error("获取令牌请求参数不正确")
					}
				} else if r.Header.Get("Authorization") != "existing-token" {
					t.Error("业务请求应使用调用方传入的令牌")
				}
				for key, value := range tc.query {
					if got := r.URL.Query().Get(key); got != value {
						t.Errorf("查询参数 %s = %q，期望 %q", key, got, value)
					}
				}
				if tc.body != nil {
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Errorf("解析请求体失败: %v", err)
					} else {
						for key, value := range tc.body {
							if body[key] != value {
								t.Errorf("请求字段 %s = %v，期望 %v", key, body[key], value)
							}
						}
						if tc.path == "/api/issuing/v3/budgets" {
							if _, ok := body["account_type"]; ok || len(body) != 1 {
								t.Errorf("v3 创建预算账户仅发送 budget_name: %v", body)
							}
						}
						if tc.path == "/api/issuing/v3/cards/apply" {
							limit, ok := body["daily_limit"].(map[string]any)
							if !ok || limit["amount"] != float64(100) || limit["currency"] != "USD" || len(limit) != 2 {
								t.Errorf("v3 开卡 daily_limit 应包含 amount 和 currency: %v", body["daily_limit"])
							}
							if _, ok := body["card_type"]; ok {
								t.Error("v3 开卡请求不应包含 card_type")
							}
						}
					}
					if r.Header.Get("sign") == "" || r.Header.Get("sign-version") != "v1" {
						t.Error("POST 请求缺少签名或签名版本")
					}
				}
				_, _ = w.Write([]byte(tc.response))
			}))
			defer server.Close()
			sdk := NewPingPongSDK(New(testConf{
				url: server.URL, appID: "app", appSecret: "secret", privateKey: privateKey, signVersion: "v1",
			}))
			if err := tc.call(context.Background(), sdk); err != nil {
				t.Fatal(err)
			}
			if calls != 1 {
				t.Fatalf("接口调用次数为 %d，期望 1", calls)
			}
		})
	}
}

func TestIsInvalidToken(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "code", err: &APIError{Code: "1002"}, want: true},
		{name: "reason", err: &APIError{Reason: "Invalid Token"}, want: true},
		{name: "wrapped", err: fmt.Errorf("request failed: %w", &APIError{Code: "1002"}), want: true},
		{name: "other API error", err: &APIError{Code: "1004", Reason: "Access Denied"}},
		{name: "plain error", err: errors.New("invalid token")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsInvalidToken(tt.err); got != tt.want {
				t.Fatalf("IsInvalidToken()=%t, want %t", got, tt.want)
			}
		})
	}
}
