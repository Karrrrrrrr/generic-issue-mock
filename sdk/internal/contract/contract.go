package contract

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"reflect"
	"regexp"
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
	Channel   string
	BaseURL   string
	Defaults  map[string]any
	mu        sync.Mutex
	exchanges []Exchange
	covered   map[string]bool
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
	suite := &Suite{
		Channel: channel,
		BaseURL: strings.TrimRight(base, "/") + "/" + channel,
		covered: map[string]bool{},
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
	suite.Defaults = map[string]any{
		"AccountID":              suite.Config.Account,
		"BalanceAccountID":       suite.Config.Account,
		"Name":                   "Protocol Test",
		"FirstName":              "Protocol",
		"LastName":               "Test",
		"Email":                  "contract-" + Unique() + "@example.test",
		"Mobile":                 "2025550123",
		"MobilePrefix":           "1",
		"Currency":               "USD",
		"CardCurrency":           "USD",
		"Amount":                 "10",
		"AmountCents":            int64(100),
		"PageSize":               100,
		"PageIndex":              1,
		"Current":                1,
		"DateOfBirth":            "1990-01-02",
		"CertType":               "passport",
		"CardID":                 Unique(),
		"NationalityCountryCode": "US",
		"ResidentialCountryCode": "US",
		"BillingCountryCode":     "USA",
		"ResidentialAddress":     "1 Test Street",
		"ResidentialCity":        "Boston",
		"ResidentialState":       "MA",
		"ResidentialPostalCode":  "02101",
		"CardScheme":             "MasterCard",
		"CardType":               "share",
		"CardFormFactor":         "virtual_card",
		"Period":                 "DAY",
		"ExpirationDate":         time.Now().AddDate(2, 0, 0).Format("01/06"),
	}
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

type Call struct {
	Name           string
	Strings        []string
	Fields         map[string]any
	Schema         any
	EmbeddedSchema any
	Select         string
	Required       []string
	Expect         map[string]any
	Error          func(error) bool
	ErrorCode      any
	Local          bool
}

func (suite *Suite) Invoke(testContext *testing.T, client any, call Call) any {
	testContext.Helper()
	method := reflect.ValueOf(client).MethodByName(call.Name)
	if !method.IsValid() {
		testContext.Fatalf("SDK method %s not found", call.Name)
	}
	suite.covered[reflect.TypeOf(client).String()+"."+call.Name] = true
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	fields := make(map[string]any, len(suite.Defaults))
	for name, value := range suite.Defaults {
		fields[name] = value
	}
	fields["RequestID"] = Unique()
	fields["Nonce"] = Unique()
	for name, value := range call.Fields {
		fields[name] = value
	}
	arguments := make([]reflect.Value, method.Type().NumIn())
	stringIndex := 0
	for index := range arguments {
		argumentType := method.Type().In(index)
		if argumentType == reflect.TypeFor[context.Context]() {
			arguments[index] = reflect.ValueOf(ctx)
		} else if argumentType.Kind() == reflect.String {
			if stringIndex >= len(call.Strings) {
				testContext.Fatalf("%s: missing string argument %d", call.Name, stringIndex)
			}
			arguments[index] = reflect.ValueOf(call.Strings[stringIndex]).Convert(argumentType)
			stringIndex++
		} else {
			arguments[index] = makeArgument(testContext, argumentType, fields)
		}
	}
	suite.mu.Lock()
	start := len(suite.exchanges)
	suite.mu.Unlock()
	results := method.Call(arguments)
	var result any
	var callError error
	for _, value := range results {
		if value.Type().Implements(reflect.TypeFor[error]()) {
			if !value.IsNil() {
				callError = value.Interface().(error)
			}
		} else {
			result = value.Interface()
		}
	}
	suite.mu.Lock()
	exchanges := append([]Exchange(nil), suite.exchanges[start:]...)
	suite.mu.Unlock()
	if call.Error != nil {
		if !call.Error(callError) {
			testContext.Fatalf("%s: expected designated SDK error, got %v", call.Name, callError)
		}
	} else if callError != nil {
		for _, exchange := range exchanges {
			testContext.Logf("%s %s -> %d %s", exchange.Method, exchange.Path, exchange.Status, exchange.Body)
		}
		testContext.Fatalf("%s: %v", call.Name, callError)
	}
	if call.Local {
		if len(exchanges) != 0 {
			testContext.Fatal("local SDK helper unexpectedly made an HTTP request")
		}
		if raw, ok := result.(json.RawMessage); ok && !json.Valid(raw) {
			testContext.Fatal("invalid preflight JSON")
		}
		return result
	}
	if len(exchanges) == 0 {
		testContext.Fatalf("%s did not call the server", call.Name)
	}
	var raw any
	for _, exchange := range exchanges {
		if !strings.Contains(exchange.ContentType, "application/json") {
			testContext.Errorf("%s: not JSON: %s", exchange.Path, exchange.ContentType)
		}
		decoder := json.NewDecoder(bytes.NewReader(exchange.Body))
		decoder.UseNumber()
		if err := decoder.Decode(&raw); err != nil {
			testContext.Fatalf("%s: invalid JSON: %v", exchange.Path, err)
		}
		if call.Error != nil {
			if call.ErrorCode != nil && At(raw, "code") != call.ErrorCode {
				testContext.Errorf("%s: wrong error code: %v", call.Name, At(raw, "code"))
			}
			if suite.Channel == "photonpay" {
				Check(testContext, raw, reflect.TypeOf(struct {
					Code string `json:"code"`
					Msg  string `json:"msg"`
				}{}), "error")
			} else if suite.Channel == "paynda" {
				Check(testContext, raw, reflect.TypeOf(struct {
					Code    int64  `json:"code"`
					Message string `json:"message"`
					Success bool   `json:"success"`
				}{}), "error")
			}
			continue
		}
		if exchange.Status != http.StatusOK {
			testContext.Errorf("%s: status=%d body=%s", exchange.Path, exchange.Status, exchange.Body)
		}
		switch suite.Channel {
		case "photonpay":
			if At(raw, "code") != "0000" {
				testContext.Fatalf("invalid PhotonPay success envelope: %s", exchange.Body)
			}
			if _, ok := At(raw, "msg").(string); !ok {
				testContext.Error("msg must be a string")
			}
			raw = At(raw, "data")
		case "paynda":
			if At(raw, "code") != json.Number("200") || At(raw, "success") != true {
				testContext.Fatalf("invalid Paynda success envelope: %s", exchange.Body)
			}
			if _, ok := At(raw, "message").(string); !ok {
				testContext.Error("message must be a string")
			}
			raw = At(raw, "data")
		}
		if raw == nil {
			testContext.Fatalf("%s: null/missing success data", exchange.Path)
		}
		checkFormats(testContext, suite.Channel, raw, "data")
	}
	if call.Error != nil {
		return nil
	}
	for _, path := range call.Required {
		if value := At(raw, path); value == nil || value == "" {
			testContext.Errorf("%s: missing/empty required field %s", call.Name, path)
		}
	}
	for path, expected := range call.Expect {
		if actual := At(raw, path); !reflect.DeepEqual(actual, expected) {
			testContext.Errorf("%s: %s=%v, want %v", call.Name, path, actual, expected)
		}
	}
	if call.EmbeddedSchema != nil {
		encoded := Text(testContext, raw, "result")
		var nested any
		decoder := json.NewDecoder(strings.NewReader(encoded))
		decoder.UseNumber()
		if err := decoder.Decode(&nested); err != nil {
			testContext.Fatalf("invalid nested request result: %v", err)
		}
		if At(nested, "code") != json.Number("200") || At(nested, "success") != true {
			testContext.Fatalf("invalid nested request result envelope: %s", encoded)
		}
		Check(testContext, At(nested, "data"), reflect.TypeOf(call.EmbeddedSchema), "result.data")
		checkFormats(testContext, suite.Channel, At(nested, "data"), "result.data")
	}
	if call.Select != "" {
		raw = At(raw, call.Select)
	}
	schema := reflect.TypeOf(result)
	if schema == nil {
		schema = reflect.TypeFor[struct{}]()
	}
	if call.Schema != nil {
		schema = reflect.TypeOf(call.Schema)
	}
	if schema != nil {
		if suite.Channel == "paynda" && schema.Kind() == reflect.Slice {
			if records := At(raw, "records"); records != nil {
				raw = records
			}
		}
		Check(testContext, raw, schema, call.Name)
	}
	return result
}

func makeArgument(testContext *testing.T, valueType reflect.Type, fields map[string]any) reflect.Value {
	testContext.Helper()
	if valueType.Kind() == reflect.Pointer {
		value := reflect.New(valueType.Elem())
		value.Elem().Set(makeArgument(testContext, valueType.Elem(), fields))
		return value
	}
	value := reflect.New(valueType).Elem()
	if valueType.Kind() != reflect.Struct {
		return value
	}
	for index := 0; index < value.NumField(); index++ {
		field := value.Field(index)
		if !field.CanSet() {
			continue
		}
		fieldType := valueType.Field(index)
		if fieldType.Anonymous {
			field.Set(makeArgument(testContext, field.Type(), fields))
			continue
		}
		input, ok := fields[fieldType.Name]
		if !ok || input == nil {
			continue
		}
		inputValue := reflect.ValueOf(input)
		if inputValue.Type().AssignableTo(field.Type()) {
			field.Set(inputValue)
			continue
		}
		encoded, err := json.Marshal(input)
		if err != nil {
			testContext.Fatal(err)
		}
		if err := json.Unmarshal(encoded, field.Addr().Interface()); err != nil {
			testContext.Fatalf("argument %s.%s: %v", valueType, fieldType.Name, err)
		}
	}
	return value
}

type Reporter interface {
	Helper()
	Errorf(string, ...any)
}

func Check(testContext Reporter, raw any, schema reflect.Type, path string) {
	testContext.Helper()
	if schema.Kind() == reflect.Pointer {
		if raw == nil {
			testContext.Errorf("%s: unexpected null object", path)
			return
		}
		Check(testContext, raw, schema.Elem(), path)
		return
	}
	if schema == reflect.TypeFor[json.RawMessage]() || schema.Kind() == reflect.Interface {
		return
	}
	if schema == reflect.TypeFor[time.Time]() {
		text, ok := raw.(string)
		if !ok {
			testContext.Errorf("%s: time must be string, got %T", path, raw)
			return
		}
		if _, err := time.Parse(time.RFC3339Nano, text); err != nil {
			testContext.Errorf("%s: not RFC3339: %q", path, text)
		}
		return
	}
	switch schema.Kind() {
	case reflect.Struct:
		object, ok := raw.(map[string]any)
		if !ok {
			testContext.Errorf("%s: want object, got %T", path, raw)
			return
		}
		for index := 0; index < schema.NumField(); index++ {
			field := schema.Field(index)
			if field.Anonymous {
				Check(testContext, raw, field.Type, path)
				continue
			}
			name := strings.Split(field.Tag.Get("json"), ",")[0]
			if name == "" || name == "-" {
				continue
			}
			if value, exists := object[name]; exists {
				if value == nil && field.Type.Kind() == reflect.Pointer {
					continue
				}
				if strings.Contains(field.Tag.Get("json"), ",string") {
					text, ok := value.(string)
					if !ok {
						testContext.Errorf("%s.%s: expected JSON-encoded string", path, name)
						continue
					}
					if _, err := strconv.ParseFloat(text, 64); err != nil {
						testContext.Errorf("%s.%s: invalid numeric string", path, name)
					}
					continue
				}
				Check(testContext, value, field.Type, path+"."+name)
			}
		}
	case reflect.Slice, reflect.Array:
		items, ok := raw.([]any)
		if !ok {
			testContext.Errorf("%s: want array (not null), got %T", path, raw)
			return
		}
		for index, item := range items {
			Check(testContext, item, schema.Elem(), fmt.Sprintf("%s[%d]", path, index))
		}
	case reflect.String:
		if _, ok := raw.(string); !ok {
			testContext.Errorf("%s: want string, got %T (%v)", path, raw, raw)
		}
	case reflect.Bool:
		if _, ok := raw.(bool); !ok {
			testContext.Errorf("%s: want boolean, got %T", path, raw)
		}
	case reflect.Int, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint64:
		number, ok := raw.(json.Number)
		if !ok {
			testContext.Errorf("%s: want JSON integer, got %T", path, raw)
			return
		}
		if _, err := number.Int64(); err != nil {
			testContext.Errorf("%s: not an integer: %s", path, number)
		}
	case reflect.Float32, reflect.Float64:
		if _, ok := raw.(json.Number); !ok {
			testContext.Errorf("%s: want JSON number, got %T", path, raw)
		}
	}
}

func checkFormats(testContext Reporter, channel string, raw any, path string) {
	testContext.Helper()
	switch value := raw.(type) {
	case []any:
		for index, item := range value {
			checkFormats(testContext, channel, item, fmt.Sprintf("%s[%d]", path, index))
		}
	case map[string]any:
		for name, item := range value {
			fieldPath := path + "." + name
			text, ok := item.(string)
			if ok && text != "" {
				layout := ""
				switch name {
				case "createdAt", "updatedAt", "createAt", "updateAt", "timestamp", "authorizedAt", "date", "transactedAt", "createTime", "updateTime", "transactionTime", "authorizationTime", "quotedAt", "returnedAt", "txnDate":
					layout = time.RFC3339Nano
					if channel == "paynda" {
						layout = time.DateTime
					}
					if channel == "photonpay" {
						layout = "2006-01-02T15:04:05"
					}
				case "dateOfBirth":
					layout = time.DateOnly
				case "expirationDate":
					layout = "01/06"
				}
				if layout != "" {
					if _, err := time.Parse(layout, text); err != nil {
						testContext.Errorf("%s: expected %s, got %q", fieldPath, layout, text)
					}
				}
				if name == "id" || name == "cardId" || name == "cardholderId" || name == "accountId" || name == "balanceAccountId" || name == "virtualAccountId" || name == "cardProductId" {
					pattern := `^[1-9][0-9]*$`
					if channel == "slash" {
						pattern = `^00000000-0000-0000-[0-9a-f]{4}-[0-9a-f]{12}$`
					}
					if !regexp.MustCompile(pattern).MatchString(text) {
						testContext.Errorf("%s: invalid channel ID %q", fieldPath, text)
					}
					if channel != "slash" {
						if _, err := strconv.ParseInt(text, 10, 64); err != nil {
							testContext.Errorf("%s: ID overflows int64", fieldPath)
						}
					}
				}
				if name == "cvv" && !regexp.MustCompile(`^[0-9]{3,4}$`).MatchString(text) {
					testContext.Errorf("%s: invalid CVV", fieldPath)
				}
				if name == "expiryMonth" && !regexp.MustCompile(`^(0[1-9]|1[0-2])$`).MatchString(text) {
					testContext.Errorf("%s: invalid month", fieldPath)
				}
				if name == "expiryYear" && !regexp.MustCompile(`^[0-9]{4}$`).MatchString(text) {
					testContext.Errorf("%s: invalid year", fieldPath)
				}
			}
			checkFormats(testContext, channel, item, fieldPath)
		}
	}
}

func (suite *Suite) Coverage(testContext *testing.T, client any, excluded ...string) {
	testContext.Helper()
	skip := map[string]bool{"RestyRequest": true}
	for _, name := range excluded {
		skip[name] = true
	}
	clientType := reflect.TypeOf(client)
	for index := 0; index < clientType.NumMethod(); index++ {
		name := clientType.Method(index).Name
		if !skip[name] && !suite.covered[clientType.String()+"."+name] {
			testContext.Errorf("SDK method has no test: %s.%s", clientType, name)
		}
	}
	count := 0
	for index := 0; index < clientType.NumMethod(); index++ {
		if !skip[clientType.Method(index).Name] {
			count++
		}
	}
	testContext.Logf("covered SDK %s: %d public methods (excluding request transport)", clientType, count)
}
