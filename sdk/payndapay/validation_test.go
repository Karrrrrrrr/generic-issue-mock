package payndapay

import (
	"context"
	"testing"

	"sdk/crypto"
	"sdk/internal/contract"

	"github.com/shopspring/decimal"
)

type disabledBinConfig struct {
	contract.Config
	disabled string
}

func (config *disabledBinConfig) GetDisabledBins() string { return config.disabled }

func TestGenerateSignatureFunc(testContext *testing.T) {
	request := &crypto.SignatureParams{
		AppID:     "local-test",
		AppSecret: "test-secret",
		Path:      "/test/path",
		Nonce:     "test-nonce",
		Timestamp: "1234567890",
	}
	signature := crypto.GenerateSignature(request)
	if signature == "" || signature != crypto.GenerateSignature(request) {
		testContext.Fatal("signature must be nonempty and deterministic")
	}
	request.Nonce = "another-nonce"
	if signature == crypto.GenerateSignature(request) {
		testContext.Fatal("signature must include nonce")
	}
}

func TestPayndaPaySDKValidCardBin(testContext *testing.T) {
	client := New(&disabledBinConfig{disabled: "123456,654321"})
	if client.ValidCardBin(context.Background(), "123456") {
		testContext.Fatal("disabled BIN accepted")
	}
	if !client.ValidCardBin(context.Background(), "999999") {
		testContext.Fatal("enabled BIN rejected")
	}
}

func TestPayndaPaySDKCardTransferRejectsDisabledBin(testContext *testing.T) {
	client := New(&disabledBinConfig{disabled: "123456"})
	request := &CardTransferRequest{
		CardID:    "1",
		CardBin:   "123456",
		Amount:    decimal.NewFromInt(1),
		RequestID: "request-id",
	}
	for _, transfer := range []func(context.Context, *CardTransferRequest) (*CardTransferResult, error){client.CardTransferIn, client.CardTransferOut} {
		result, err := transfer(context.Background(), request)
		if err == nil || result == nil || result.Status != CardTransferStatusFailed {
			testContext.Fatalf("disabled BIN: result=%+v error=%v", result, err)
		}
	}
}

func TestCardStatusUpdateRequestValidateNil(testContext *testing.T) {
	var request *CardStatusUpdateRequest
	if err := request.Validate(); err != ErrEmptyParams {
		testContext.Fatalf("got %v, want %v", err, ErrEmptyParams)
	}
}

var _ PayndaPaySDKInterface = (*PayndaPaySDK)(nil)
