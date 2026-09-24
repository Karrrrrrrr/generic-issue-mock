package pingpong

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"tman/pkg/dealer/middleware/trace"
	"tman/pkg/dealer/middleware/traffic"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

/***
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4WrYNuamUVtWtmd118/D
ZJShyn2HtMGkF8Y+NoWc+EPa6IoLvmz1xMZtelAG2bF8zLLFCxHGqp81wDmJ1ilw
gkUfYwFzNwehj2BrGI2fmQosTLIHYSJQei0SJe+8DE8bAuoN5eI9sglNZuAxgI1a
K9HaWPiQPT+xLoC9NBuiFDb8m4Yb49M+jBhToZdYyyzOFaRkJDpiWNAwJpB+GQ2b
DWngPvdO2/9qfTfZMG8W7hn1bi12MJ925zOdekIo2i/vZov5XW8OLjonapkbtJUx
HM4NGTRDnkRdJ6sQWJzc7iKAFjRNwBTL2lyVwKZhroSDO++1UTdw9Wzq/QQ4Cc5v
JQIDAQAB
-----END PUBLIC KEY-----

-----BEGIN PRIVATE KEY-----
MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQDhatg25qZRW1a2
Z3XXz8NklKHKfYe0waQXxj42hZz4Q9roigu+bPXExm16UAbZsXzMssULEcaqnzXA
OYnWKXCCRR9jAXM3B6GPYGsYjZ+ZCixMsgdhIlB6LRIl77wMTxsC6g3l4j2yCU1m
4DGAjVor0dpY+JA9P7EugL00G6IUNvybhhvj0z6MGFOhl1jLLM4VpGQkOmJY0DAm
kH4ZDZsNaeA+907b/2p9N9kwbxbuGfVuLXYwn3bnM516QijaL+9mi/ldbw4uOidq
mRu0lTEczg0ZNEOeRF0nqxBYnNzuIoAWNE3AFMvaXJXApmGuhIM777VRN3D1bOr9
BDgJzm8lAgMBAAECggEBAIWgUt/oxvs/jB3BIyh17zx2p5pj48iRafb1+/dSKYU6
pFBpVSDjcqXdgxSY0BbIklS+PPSc6wpGKxTyhvU/x4RR+ZM1Ttl2Wp2l6Ja7jbqp
Py2P87PvJYnnofR/MxiQ5FBL80UtYqlhvlKX4IB2StfjJO7NGqRUV3Jbus1i/CfC
fAck/AWGujwkn8mxxhvgoBMhzNKJE8ivOcPu+fmFfu8m8mWgE6RYq13f25NXlwHK
FHFw0mEOPi9wkFRXJxdfHX96KxVgvSKdRUc4bJ+eJg3fysfily/7b6gBUtHk2mV5
qi+6YreiIZ1UDD5uNFmhdLEutttrtEWg1WBWg0b1XZkCgYEA9aickHO02hkvG3YH
CmoGJo+L7pQVvVkuNcEQkSrrNiv2nzaigjgJUp4R0aVE66u6NwesvCzwn+X37djv
CxhLZuEuBI8azRYclYdrhGND0g3wgNAOoSN5LLfFgFB2GkzT5RlcYwnBZJJKqeCe
sBHfuLteLGst/MyhujXCd6Uijn8CgYEA6ugZUtuMUbCHpLI935z6c+P4OVVTjy0H
ASmtAtfrEOnmDu+QounOsLQmo6qCdxue8ryoebBjZUymqSCEgFUB4H/VhlEu+yG4
Yg7XezGQl7bHWbUPV5SKiTNcD9HNxkPrx3UOuZTHqM9/oBSExP/G0RnDjGPdhHCP
BbO2mR+mOFsCgYAKmMJgLM2RVuLEUXwOQ/KN+UU0/mhNqaonoXNgf7RzusPBrG6o
JVipmq30GCf37oly1D7sQxgCHb5rIR92oA6omnAMvEuQqzKCdLv7kvia+AT22YK4
CrqwZiD73vypN8UwLb7hess/1luoJktSFwNKibKPQfRS4lTbnnQMCzCJawKBgQCh
5p/1hI3ci4+hipusb/QKRdgCI/X4Wy9VtNSifhBsUtkV+DU2o3CqRy/OY6mRz/6o
DDEN1e1blw3SyS+ph21IvrJ65Z88xMvhAZuwM8QVXItfH7RYR2+IClbsLEzn1k49
5Ublz04g4gpzWVD8udDcsyYcr4OwUSex5V/3f2G/uwKBgQCakNuk0pko7aYAWNev
t53mIx34pE/XBxFoW/6CxiSir4SF89/th/dsbqvKLcdc3VBWMFkCMzkLAkVJqI0A
jRhk7eOSeQEElZ4AznxA7oDNF3NHnB+Kw5nYqdIbGf9iYTTgRJDY+xnXylRNzW4U
jPiLfeeUfns8AR3TjId6HerZ0g==
-----END PRIVATE KEY-----
***/

// 测试环境配置可直接填写在此处；留空时读取对应的 PINGPONG_* 环境变量。
// 获取 token 不需要私钥，POST 业务接口才需要填写 testPrivateKey。
var (
	token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJQaW5nUG9uZyIsImV4cCI6MTc5MDIyNTAyOSwiYXBwX2lkIjoiODA1NTAwODExNTA2MTA0IiwiY2x1IjoiY24ifQ.6QsEXDZMkt0I3ugoUpp4FMaxSo3LwDfYowkMKUL4PHc" // 已有令牌；填写后不会重新获取。

	testGatewayURL = "http://test-pingpongx.ptmlink.top" // PingPong 测试网关地址。
	testAppID      = "805500811506104"                   // 测试应用 ID。
	testAppSecret  = "FC7FE48730A646928263F916"          // 测试应用密钥。
	testPrivateKey = `-----BEGIN PRIVATE KEY-----
MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQDhatg25qZRW1a2
Z3XXz8NklKHKfYe0waQXxj42hZz4Q9roigu+bPXExm16UAbZsXzMssULEcaqnzXA
OYnWKXCCRR9jAXM3B6GPYGsYjZ+ZCixMsgdhIlB6LRIl77wMTxsC6g3l4j2yCU1m
4DGAjVor0dpY+JA9P7EugL00G6IUNvybhhvj0z6MGFOhl1jLLM4VpGQkOmJY0DAm
kH4ZDZsNaeA+907b/2p9N9kwbxbuGfVuLXYwn3bnM516QijaL+9mi/ldbw4uOidq
mRu0lTEczg0ZNEOeRF0nqxBYnNzuIoAWNE3AFMvaXJXApmGuhIM777VRN3D1bOr9
BDgJzm8lAgMBAAECggEBAIWgUt/oxvs/jB3BIyh17zx2p5pj48iRafb1+/dSKYU6
pFBpVSDjcqXdgxSY0BbIklS+PPSc6wpGKxTyhvU/x4RR+ZM1Ttl2Wp2l6Ja7jbqp
Py2P87PvJYnnofR/MxiQ5FBL80UtYqlhvlKX4IB2StfjJO7NGqRUV3Jbus1i/CfC
fAck/AWGujwkn8mxxhvgoBMhzNKJE8ivOcPu+fmFfu8m8mWgE6RYq13f25NXlwHK
FHFw0mEOPi9wkFRXJxdfHX96KxVgvSKdRUc4bJ+eJg3fysfily/7b6gBUtHk2mV5
qi+6YreiIZ1UDD5uNFmhdLEutttrtEWg1WBWg0b1XZkCgYEA9aickHO02hkvG3YH
CmoGJo+L7pQVvVkuNcEQkSrrNiv2nzaigjgJUp4R0aVE66u6NwesvCzwn+X37djv
CxhLZuEuBI8azRYclYdrhGND0g3wgNAOoSN5LLfFgFB2GkzT5RlcYwnBZJJKqeCe
sBHfuLteLGst/MyhujXCd6Uijn8CgYEA6ugZUtuMUbCHpLI935z6c+P4OVVTjy0H
ASmtAtfrEOnmDu+QounOsLQmo6qCdxue8ryoebBjZUymqSCEgFUB4H/VhlEu+yG4
Yg7XezGQl7bHWbUPV5SKiTNcD9HNxkPrx3UOuZTHqM9/oBSExP/G0RnDjGPdhHCP
BbO2mR+mOFsCgYAKmMJgLM2RVuLEUXwOQ/KN+UU0/mhNqaonoXNgf7RzusPBrG6o
JVipmq30GCf37oly1D7sQxgCHb5rIR92oA6omnAMvEuQqzKCdLv7kvia+AT22YK4
CrqwZiD73vypN8UwLb7hess/1luoJktSFwNKibKPQfRS4lTbnnQMCzCJawKBgQCh
5p/1hI3ci4+hipusb/QKRdgCI/X4Wy9VtNSifhBsUtkV+DU2o3CqRy/OY6mRz/6o
DDEN1e1blw3SyS+ph21IvrJ65Z88xMvhAZuwM8QVXItfH7RYR2+IClbsLEzn1k49
5Ublz04g4gpzWVD8udDcsyYcr4OwUSex5V/3f2G/uwKBgQCakNuk0pko7aYAWNev
t53mIx34pE/XBxFoW/6CxiSir4SF89/th/dsbqvKLcdc3VBWMFkCMzkLAkVJqI0A
jRhk7eOSeQEElZ4AznxA7oDNF3NHnB+Kw5nYqdIbGf9iYTTgRJDY+xnXylRNzW4U
jPiLfeeUfns8AR3TjId6HerZ0g==
-----END PRIVATE KEY-----` // 测试请求签名私钥。
	testSignVersion = "v1" // 非对称密钥版本号，由 PingPong 配置后提供。
	testFetchToken  = true // 仅在需要获取新令牌时手动改为 true。
)

// testConf 提供真实网关测试所需的配置。
type testConf struct {
	debug       bool   // 是否开启调试模式。
	url         string // PingPong 网关地址。
	appID       string // 取令牌使用的应用 ID。
	appSecret   string // 取令牌使用的应用密钥。
	privateKey  string // POST 请求签名使用的 RSA 私钥。
	signVersion string // 非对称密钥版本号，由 PingPong 提供。
}

// GetUrl 返回测试网关地址。
func (c testConf) GetUrl() string { return c.url }

// GetAppId 返回测试应用 ID。
func (c testConf) GetAppId() string { return c.appID }

// GetAppSecret 返回测试应用密钥。
func (c testConf) GetAppSecret() string { return c.appSecret }

// GetPrivateKey 返回请求签名私钥。
func (c testConf) GetPrivateKey() string { return c.privateKey }

// GetSignVersion 返回非对称密钥版本号。
func (c testConf) GetSignVersion() string { return c.signVersion }

// GetDebug 返回是否开启调试模式。
func (c testConf) GetDebug() bool { return c.debug }

// testConfig 优先使用文件顶部的测试配置，未填写时读取环境变量。
func testConfig(value, envKey string) string {
	if value != "" {
		return value
	}
	return os.Getenv(envKey)
}

// newTestSDK 创建真实网关测试客户端；缺少配置时跳过测试。
func newTestSDK(t *testing.T) *PingPongSDK {
	t.Helper()
	conf := testConf{
		debug:       true,
		url:         testConfig(testGatewayURL, "PINGPONG_URL"),
		appID:       testConfig(testAppID, "PINGPONG_APP_ID"),
		appSecret:   testConfig(testAppSecret, "PINGPONG_APP_SECRET"),
		privateKey:  testConfig(testPrivateKey, "PINGPONG_PRIVATE_KEY"),
		signVersion: testConfig(testSignVersion, "PINGPONG_SIGN_VERSION"),
	}
	if conf.url == "" {
		t.Skip("请填写 testGatewayURL 或设置 PINGPONG_URL 后运行真实接口测试")
	}
	l := log.DefaultLogger
	return New(conf, WithResponseMiddlewares(trace.ResponseTracking(), traffic.PhotonTrafficLog(l)))
}

// testToken 优先使用文件顶部的 token，其次读取 PINGPONG_TOKEN；均未设置时跳过业务测试，不自动获取令牌。
func testToken(t *testing.T) string {
	t.Helper()
	value := token
	if value == "" {
		value = os.Getenv("PINGPONG_TOKEN")
	}
	if value == "" {
		t.Skip("请手动设置文件顶部的 token 或 PINGPONG_TOKEN；不会自动获取令牌")
	}
	return value
}

// requireTestValue 检查测试函数中手动填写的必需参数。
func requireTestValue(t *testing.T, name, value string) string {
	t.Helper()
	if value == "" {
		t.Skipf("请在当前测试函数中填写 %s", name)
	}
	return value
}

// logResponse 将接口返回值完整输出为 JSON；仅用于受控测试环境，日志可能包含敏感信息。
func logResponse(t *testing.T, response any) {
	t.Helper()
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("接口返回：%s", data)
}

// TestPingPongTrafficLog 验证中间件不记录令牌、密钥、查询参数和敏感响应。
func TestPingPongTrafficLog(t *testing.T) {
	var output bytes.Buffer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"data":{"card_number":"4111111111111111","cvc":"secret-cvc-value"}}`))
	}))
	defer server.Close()

	p := New(testConf{url: server.URL}, WithResponseMiddlewares(traffic.PingPongTrafficLog(log.NewStdLogger(&output))))
	_, err := p.GetCardDetails(context.Background(), "sensitive-token", "sensitive-card-id")
	if err != nil {
		t.Fatal(err)
	}
	logged := output.String()
	for _, secret := range []string{"sensitive-token", "sensitive-card-id", "4111111111111111", "secret-cvc-value", "app_secret", "card_id="} {
		if strings.Contains(logged, secret) {
			t.Errorf("日志包含敏感信息 %q", secret)
		}
	}
	if !strings.Contains(logged, "/api/issuing/v3/cards/details") || !strings.Contains(logged, "pingpong") {
		t.Errorf("缺少必要的请求路径或渠道信息: %s", logged)
	}
}

// TestPingPongSDK_GetAccessToken 仅在无手动令牌且显式允许时获取新令牌。
func TestPingPongSDK_GetAccessToken(t *testing.T) {
	if !testFetchToken && os.Getenv("PINGPONG_FETCH_TOKEN") != "1" {
		t.Skip("获取新令牌可能使现有令牌失效；请将 testFetchToken 设为 true 或设置 PINGPONG_FETCH_TOKEN=1")
	}
	p := newTestSDK(t)
	if p.conf.GetAppId() == "" || p.conf.GetAppSecret() == "" {
		t.Skip("请填写 testAppID、testAppSecret 或设置 PINGPONG_APP_ID、PINGPONG_APP_SECRET")
	}
	got, err := p.GetAccessToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.AccessToken == "" {
		t.Fatal("响应中没有 access_token")
	}
	logResponse(t, got)
}

// TestPingPongSDK_QueryAccountsBalances 查询真实账户余额。
func TestPingPongSDK_QueryAccountsBalances(t *testing.T) {
	accessToken := testToken(t)
	p := newTestSDK(t)
	if p.conf.GetPrivateKey() == "" {
		t.Skip("请填写 testPrivateKey 或设置 PINGPONG_PRIVATE_KEY")
	}
	got, err := p.QueryAccountsBalances(context.Background(), accessToken, &QueryAccountsBalancesRequest{PageNo: 1, PageSize: 20})
	if err != nil {
		t.Fatal(err)
	}
	logResponse(t, got)
}

// TestPingPongSDK_QueryCardProducts 查询真实卡产品。
func TestPingPongSDK_QueryCardProducts(t *testing.T) {
	accessToken := testToken(t)
	p := newTestSDK(t)
	got, err := p.QueryCardProducts(context.Background(), accessToken)
	if err != nil {
		t.Fatal(err)
	}
	logResponse(t, got)
}

// TestPingPongSDK_CreateCard 使用当前账户的卡产品与预算申请卡片。
func TestPingPongSDK_CreateCard(t *testing.T) {
	accessToken := testToken(t)
	p := newTestSDK(t)
	if p.conf.GetPrivateKey() == "" {
		t.Skip("请填写 testPrivateKey 或设置 PINGPONG_PRIVATE_KEY")
	}
	got, err := p.CreateCard(context.Background(), accessToken, &CreateCardRequest{
		RequestID:       uuid.NewString(),
		CardProductCode: "TLKSAQ",
		CardCurrency:    "USD",
		BudgetID:        "ci2026091510281379531",
		CouponApplied:   false,
		Remark:          "测试卡片",
	})
	if err != nil {
		t.Fatal(err)
	}
	logResponse(t, got)
}

// TestPingPongSDK_GetCardDetails 查询卡详情并输出完整响应；仅在受控测试环境使用。
func TestPingPongSDK_GetCardDetails(t *testing.T) {
	cardID := "card6420260924104905198107" // 填写当前账户真实的 card_id。
	cardID = requireTestValue(t, "cardID", cardID)
	accessToken := testToken(t)
	p := newTestSDK(t)
	got, err := p.GetCardDetails(context.Background(), accessToken, cardID)
	if err != nil {
		t.Fatal(err)
	}
	logResponse(t, got)
}

// TestPingPongSDK_CardFunding 对真实卡片执行资金操作。
func TestPingPongSDK_CardFunding(t *testing.T) {
	cardID := "card6420260924104905198107" // 填写当前账户真实的 card_id。
	action := CardFundingTopUp.ToString()  // top_up 为充值，withdraw 为转出。
	amount := 10.0                         // 填写本次操作金额。
	cardID = requireTestValue(t, "cardID", cardID)
	accessToken := testToken(t)
	p := newTestSDK(t)
	if p.conf.GetPrivateKey() == "" {
		t.Skip("请填写 testPrivateKey 或设置 PINGPONG_PRIVATE_KEY")
	}
	got, err := p.CardFunding(context.Background(), accessToken, &CardFundingRequest{
		CardID:        cardID,
		Action:        action,
		Amount:        amount,
		UniqueOrderID: "202609211937418133",
	})
	if err != nil {
		assert.NoError(t, err)
		fmt.Printf("v: %+v\n", got)
	}
	logResponse(t, got)
}

// TestPingPongSDK_QueryCardFundingOrders 查询真实卡资金订单。
func TestPingPongSDK_QueryCardFundingOrders(t *testing.T) {
	// cardID := "card201202609181722314762340" // 填写当前账户真实的 card_id。
	// cardID = requireTestValue(t, "cardID", cardID)
	accessToken := testToken(t)
	p := newTestSDK(t)
	got, err := p.QueryCardFundingOrders(context.Background(), accessToken, &QueryCardFundingOrdersRequest{
		CardID: "card24202609201937418132",
		// Status: FundingOrderStatusSuccess.ToString(),
		UniqueOrderID: "202609211937418133",
		PageNo:        1,
		PageSize:      20,
	})
	if err != nil {
		t.Fatal(err)
	}
	logResponse(t, got)
}

// TestPingPongSDK_QueryDedicatedCardBalance 查询真实卡片余额并输出完整响应；仅在受控测试环境使用。
func TestPingPongSDK_QueryDedicatedCardBalance(t *testing.T) {
	cardID := "card642026092311222550131" // 填写当前账户真实的 card_id。
	cardID = requireTestValue(t, "cardID", cardID)
	accessToken := testToken(t)
	p := newTestSDK(t)
	got, err := p.QueryDedicatedCardBalance(context.Background(), accessToken, cardID)
	if err != nil {
		t.Fatal(err)
	}
	logResponse(t, got)
}

// TestPingPongSDK_CardAction 对真实卡片执行指定操作。
func TestPingPongSDK_CardAction(t *testing.T) {
	cardID := "card642026092311222550131" // 填写当前账户真实的 card_id。
	action := "freeze"                    // 填写 freeze、unfreeze、close 或 update_remark。
	remark := "测试备注"                      // close 或 update_remark 时填写备注。
	cardID = requireTestValue(t, "cardID", cardID)
	action = requireTestValue(t, "action", action)
	if action != "freeze" && action != "unfreeze" && action != "close" && action != "update_remark" {
		t.Fatal("action 必须是 freeze、unfreeze、close 或 update_remark")
	}
	accessToken := testToken(t)
	p := newTestSDK(t)
	if p.conf.GetPrivateKey() == "" {
		t.Skip("请填写 testPrivateKey 或设置 PINGPONG_PRIVATE_KEY")
	}
	params := &CardActionRequest{CardID: cardID, Action: action}
	if action == "close" || action == "update_remark" {
		params.Remark = requireTestValue(t, "remark", remark)
	}
	if err := p.CardAction(context.Background(), accessToken, params); err != nil {
		t.Fatal(err)
	}
	t.Log("卡片操作调用成功")
}

// TestPingPongSDK_QueryCardTransactions 查询真实卡交易。
func TestPingPongSDK_QueryCardTransactions(t *testing.T) {
	cardID := "card201202609181722314762340" // 填写当前账户真实的 card_id。
	cardID = requireTestValue(t, "cardID", cardID)
	accessToken := testToken(t)
	p := newTestSDK(t)
	got, err := p.QueryCardTransactions(context.Background(), accessToken, &QueryCardTransactionsRequest{CardID: cardID, PageNo: 1, PageSize: 20})
	if err != nil {
		t.Fatal(err)
	}
	logResponse(t, got)
}

// TestPingPongSDK_Query3DSDetails 查询 3DS 资料并输出完整响应；仅在受控测试环境使用。
func TestPingPongSDK_Query3DSDetails(t *testing.T) {
	cardID := "" // 填写当前账户真实的 card_id。
	cardID = requireTestValue(t, "cardID", cardID)
	accessToken := testToken(t)
	p := newTestSDK(t)
	got, err := p.Query3DSDetails(context.Background(), accessToken, cardID)
	if err != nil {
		t.Fatal(err)
	}
	logResponse(t, got)
}

// TestPingPongSDK_CreateBudgetAccount 创建真实预算账户。
func TestPingPongSDK_CreateBudgetAccount(t *testing.T) {
	budgetName := "" // 填写本次新建预算账户的名称。
	budgetName = requireTestValue(t, "budgetName", budgetName)
	accessToken := testToken(t)
	p := newTestSDK(t)
	got, err := p.CreateBudgetAccount(context.Background(), accessToken, &CreateBudgetAccountRequest{
		BudgetName: budgetName,
	})
	if err != nil {
		t.Fatal(err)
	}
	logResponse(t, got)
}

// TestPingPongSDK_BudgetFunding 对真实预算账户充值或划转。
func TestPingPongSDK_BudgetFunding(t *testing.T) {
	budgetID := "ci2026091510281379531" // 填写来源预算账户 ID。
	action := "top_up"                  // top_up 为充值，transfer 为预算账户间划转。
	amount := 1.0                       // 填写实际金额。
	currency := ""                      // 填写来源币种，如 USD。
	targetBudgetID := ""                // transfer 时填写目标预算账户 ID。
	targetCurrency := ""                // transfer 时可填写目标币种。
	budgetID = requireTestValue(t, "budgetID", budgetID)
	currency = requireTestValue(t, "currency", currency)
	if action == "transfer" {
		targetBudgetID = requireTestValue(t, "targetBudgetID", targetBudgetID)
	}
	accessToken := testToken(t)
	p := newTestSDK(t)
	got, err := p.BudgetFunding(context.Background(), accessToken, &BudgetFundingRequest{
		UniqueOrderID: uuid.NewString(), BudgetID: budgetID, Action: action,
		Amount: amount, Currency: currency, TargetBudgetID: targetBudgetID, TargetCurrency: targetCurrency,
	})
	if err != nil {
		t.Fatal(err)
	}
	logResponse(t, got)
}

// TestPingPongSDK_QueryBudgetFundingOrder 查询真实预算资金订单。
func TestPingPongSDK_QueryBudgetFundingOrder(t *testing.T) {
	orderID := ""      // 填写资金订单 ID。
	action := "top_up" // top_up 或 transfer，需与订单操作一致。
	orderID = requireTestValue(t, "orderID", orderID)
	accessToken := testToken(t)
	p := newTestSDK(t)
	got, err := p.QueryBudgetFundingOrder(context.Background(), accessToken, &QueryBudgetFundingOrderRequest{
		OrderID: orderID, Action: action,
	})
	if err != nil {
		t.Fatal(err)
	}
	logResponse(t, got)
}

// TestPingPongSDK_QueryBudgetAccountBalance 查询指定或全部预算账户余额。
func TestPingPongSDK_QueryBudgetAccountBalance(t *testing.T) {
	budgetID := "" // 留空查询全部；填写真实 budget_id 查询指定预算账户。
	accessToken := testToken(t)
	p := newTestSDK(t)
	got, err := p.QueryBudgetAccountBalance(context.Background(), accessToken, budgetID)
	if err != nil {
		t.Fatal(err)
	}
	logResponse(t, got)
}

// TestPingPongSDK_QueryAccountTransactions 查询真实账户的已入账交易。
func TestPingPongSDK_QueryAccountTransactions(t *testing.T) {
	budgetID := "ci202608270955472524"              // 与 cardID 至少填写一个真实 ID。
	cardID := ""                                    // 同时填写时仅查询同时匹配的记录。
	postingStartTime := "2026-09-01T00:00:00+08:00" // 填写 ISO 8601 时间，如 2026-09-01T00:00:00+08:00。
	postingEndTime := "2026-09-19T00:00:00+08:00"   // 填写结束时间，范围不得超过 31 天。
	if budgetID == "" && cardID == "" {
		t.Skip("请在当前测试函数中填写 budgetID 或 cardID")
	}
	postingStartTime = requireTestValue(t, "postingStartTime", postingStartTime)
	postingEndTime = requireTestValue(t, "postingEndTime", postingEndTime)
	accessToken := testToken(t)
	p := newTestSDK(t)
	got, err := p.QueryAccountTransactions(context.Background(), accessToken, &QueryAccountTransactionsRequest{
		PageNo: 1, PageSize: 20, BudgetID: budgetID, CardID: cardID,
		PostingStartTime: postingStartTime, PostingEndTime: postingEndTime,
	})
	if err != nil {
		t.Fatal(err)
	}
	logResponse(t, got)
}
