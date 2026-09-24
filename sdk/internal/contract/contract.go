package contract

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type Config struct {
	URL            string
	Account        string
	VirtualAccount string
	Wallet         string
	PrivateKey     string
}

func (config *Config) GetUrl() string                 { return config.URL }
func (config *Config) GetVaultUrl() string            { return config.URL }
func (config *Config) GetApiKey() string              { return config.Account }
func (config *Config) GetAccount() string             { return config.Account }
func (config *Config) GetVirtualAccount() string      { return config.VirtualAccount }
func (config *Config) GetCardGroup() string           { return "" }
func (config *Config) GetDebug() bool                 { return false }
func (config *Config) GetRestriction() string         { return "" }
func (config *Config) GetCountryList() string         { return "" }
func (config *Config) GetRetryTimes() int32           { return 1 }
func (config *Config) GetLegalEntity() string         { return "" }
func (config *Config) GetAppId() string               { return config.Account }
func (config *Config) GetAppSecret() string           { return "local-contract-test" }
func (config *Config) GetBalanceAccount() string      { return config.Account }
func (config *Config) GetDisabledBins() string        { return "" }
func (config *Config) GetPrivateKey() string          { return config.PrivateKey }
func (config *Config) GetPublicKey() string           { return "" }
func (config *Config) GetThirdPartyPublicKey() string { return "" }

type Exchange struct {
	Method      string
	Path        string
	Status      int
	ContentType string
	Body        json.RawMessage
}

type Suite struct {
	Config    *Config
	Context   context.Context
	Channel   string
	BaseURL   string
	mu        sync.Mutex
	exchanges []Exchange
}

func New(testContext *testing.T, channel string) *Suite {
	testContext.Helper()
	base := os.Getenv("MOCK_BASE_URL")
	if base == "" {
		base = "http://127.0.0.1:18000"
	}
	target, err := url.Parse(base)
	if err != nil {
		testContext.Fatal(err)
	}
	if target.Hostname() != "localhost" && target.Hostname() != "127.0.0.1" && target.Hostname() != "::1" {
		testContext.Fatal("contract tests mutate data; MOCK_BASE_URL must be loopback")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	testContext.Cleanup(cancel)
	suite := &Suite{
		Context: ctx,
		Channel: channel,
		BaseURL: strings.TrimRight(base, "/") + "/" + channel,
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ModifyResponse = func(response *http.Response) error {
		body, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			return err
		}
		response.Body = io.NopCloser(bytes.NewReader(body))
		suite.mu.Lock()
		suite.exchanges = append(suite.exchanges, Exchange{
			Method:      response.Request.Method,
			Path:        response.Request.URL.Path,
			Status:      response.StatusCode,
			ContentType: response.Header.Get("Content-Type"),
			Body:        body,
		})
		suite.mu.Unlock()
		return nil
	}
	server := httptest.NewServer(proxy)
	testContext.Cleanup(server.Close)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		testContext.Fatal(err)
	}
	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		testContext.Fatal(err)
	}
	suite.Config = &Config{
		URL: server.URL + "/" + channel,
		PrivateKey: string(pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: keyBytes,
		})),
	}
	account := suite.UI(testContext, http.MethodPost, "/accounts", map[string]any{"name": "SDK contract " + Unique()})
	suite.Config.Account = Text(testContext, account, "id")
	suite.Config.Wallet = Text(testContext, account, "wallet_id")
	suite.UI(testContext, http.MethodPost, "/funds/transfer", map[string]any{
		"account_id": suite.Config.Account,
		"target_id":  Text(testContext, account, "wallet_id"),
		"amount":     "10000",
	})

	return suite
}

func Unique() string { return strconv.FormatInt(time.Now().UnixNano(), 10) }

func (suite *Suite) UI(testContext *testing.T, method, path string, input any) map[string]any {
	testContext.Helper()
	body, err := json.Marshal(input)
	if err != nil {
		testContext.Fatal(err)
	}
	request, err := http.NewRequest(method, suite.BaseURL+"/ui"+path, bytes.NewReader(body))
	if err != nil {
		testContext.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		testContext.Fatalf("local server unavailable: %v", err)
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		testContext.Fatal(err)
	}
	if response.StatusCode != http.StatusOK {
		testContext.Fatalf("fixture %s %s: %d %s", method, path, response.StatusCode, raw)
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		testContext.Fatalf("fixture %s: %v: %s", path, err, raw)
	}
	return result
}

func Text(testContext *testing.T, object any, path string) string {
	testContext.Helper()
	value := At(object, path)
	text, ok := value.(string)
	if !ok || text == "" {
		testContext.Fatalf("%s: expected nonempty string, got %#v", path, value)
	}
	return text
}

func At(object any, path string) any {
	for _, part := range strings.Split(path, ".") {
		switch value := object.(type) {
		case map[string]any:
			object = value[part]
		case []any:
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(value) {
				return nil
			}
			object = value[index]
		default:
			return nil
		}
	}
	return object
}

func (suite *Suite) Responses() []Exchange {
	suite.mu.Lock()
	defer suite.mu.Unlock()
	exchanges := suite.exchanges
	suite.exchanges = nil
	return exchanges
}
