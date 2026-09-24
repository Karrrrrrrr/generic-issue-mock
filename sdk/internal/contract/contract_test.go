package contract

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

type report struct{ errors []string }

func (collector *report) Helper() {}
func (collector *report) Errorf(format string, arguments ...any) {
	collector.errors = append(collector.errors, fmt.Sprintf(format, arguments...))
}

func TestWireTypeValidation(testContext *testing.T) {
	type payload struct {
		ID        string    `json:"id"`
		Count     int64     `json:"count"`
		Amount    float64   `json:"amount,string"`
		Active    bool      `json:"active"`
		CreatedAt time.Time `json:"createdAt"`
		Items     []string  `json:"items"`
	}
	tests := []struct {
		name  string
		body  string
		valid bool
	}{
		{
			name:  "valid",
			body:  `{"id":"9223372036854775807","count":2,"amount":"1.25","active":true,"createdAt":"2026-09-24T10:00:00+08:00","items":[]}`,
			valid: true,
		},
		{
			name: "numeric ID instead of string",
			body: `{"id":12}`,
		},
		{
			name: "null string",
			body: `{"id":null}`,
		},
		{
			name: "quoted count",
			body: `{"count":"2"}`,
		},
		{
			name: "fractional integer",
			body: `{"count":1.5}`,
		},
		{
			name: "overflow integer",
			body: `{"count":9223372036854775808}`,
		},
		{
			name: "number instead of encoded string",
			body: `{"amount":1.25}`,
		},
		{
			name: "invalid numeric string",
			body: `{"amount":"money"}`,
		},
		{
			name: "quoted boolean",
			body: `{"active":"true"}`,
		},
		{
			name: "timestamp without timezone",
			body: `{"createdAt":"2026-09-24T10:00:00"}`,
		},
		{
			name: "numeric timestamp",
			body: `{"createdAt":123}`,
		},
		{
			name: "null array",
			body: `{"items":null}`,
		},
		{
			name: "object instead of array",
			body: `{"items":{}}`,
		},
		{
			name: "array element mismatch",
			body: `{"items":[1]}`,
		},
	}
	for _, test := range tests {
		testContext.Run(test.name, func(testContext *testing.T) {
			var raw any
			decoder := json.NewDecoder(strings.NewReader(test.body))
			decoder.UseNumber()
			if err := decoder.Decode(&raw); err != nil {
				testContext.Fatal(err)
			}
			collector := &report{}
			Check(collector, raw, reflect.TypeFor[payload](), "response")
			if (len(collector.errors) == 0) != test.valid {
				testContext.Fatalf("valid=%v errors=%v", test.valid, collector.errors)
			}
		})
	}
}

func TestChannelFormats(testContext *testing.T) {
	tests := []struct {
		channel string
		field   string
		value   string
		valid   bool
	}{
		{
			channel: "slash",
			field:   "id",
			value:   "00000000-0000-0000-0000-000000000001",
			valid:   true,
		},
		{
			channel: "slash",
			field:   "id",
			value:   "1",
		},
		{
			channel: "paynda",
			field:   "id",
			value:   "42",
			valid:   true,
		},
		{
			channel: "paynda",
			field:   "id",
			value:   "042",
		},
		{
			channel: "paynda",
			field:   "id",
			value:   "9223372036854775808",
		},
		{
			channel: "paynda",
			field:   "createTime",
			value:   "2026-09-24 10:00:00",
			valid:   true,
		},
		{
			channel: "paynda",
			field:   "createTime",
			value:   "2026-09-24T10:00:00Z",
		},
		{
			channel: "photonpay",
			field:   "createdAt",
			value:   "2026-09-24T10:00:00",
			valid:   true,
		},
		{
			channel: "photonpay",
			field:   "createdAt",
			value:   "2026-09-24 10:00:00",
		},
		{
			channel: "slash",
			field:   "createdAt",
			value:   "2026-09-24T10:00:00.123456789Z",
			valid:   true,
		},
		{
			channel: "slash",
			field:   "createdAt",
			value:   "2026-09-24T10:00:00",
		},
		{
			channel: "slash",
			field:   "expiryMonth",
			value:   "13",
		},
		{
			channel: "slash",
			field:   "expiryYear",
			value:   "28",
		},
		{
			channel: "photonpay",
			field:   "expirationDate",
			value:   "09/28",
			valid:   true,
		},
		{
			channel: "photonpay",
			field:   "expirationDate",
			value:   "2028-09",
		},
		{
			channel: "paynda",
			field:   "cvv",
			value:   "123",
			valid:   true,
		},
		{
			channel: "paynda",
			field:   "cvv",
			value:   "abc",
		},
	}
	for index, test := range tests {
		testContext.Run(fmt.Sprintf("%s/%s/%d", test.channel, test.field, index), func(testContext *testing.T) {
			collector := &report{}
			checkFormats(collector, test.channel, map[string]any{test.field: test.value}, "response")
			if (len(collector.errors) == 0) != test.valid {
				testContext.Fatalf("valid=%v errors=%v", test.valid, collector.errors)
			}
		})
	}
}
