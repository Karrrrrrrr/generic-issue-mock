package photonpay

import (
	"io"
	"strings"
	"testing"

	"sdk/internal/contract"
)

func TestSDKProtocol(testContext *testing.T) {
	suite := contract.New(testContext, "photonpay")
	client := New(suite.Config)
	token := suite.Invoke(testContext, client, contract.Call{
		Name:     "GetAccessToken",
		Required: []string{"token", "expiresIn", "refreshExpiresIn", "refreshToken"},
	}).(*AccessTokenResponse).Token
	testContext.Run("empty-cards", func(testContext *testing.T) {
		cards := suite.Invoke(testContext, client, contract.Call{
			Name:    "PagingVccCard",
			Strings: []string{token},
			Fields: map[string]any{
				"CardType":       nil,
				"CardFormFactor": nil,
			},
		}).([]*CardDetailResponse)
		if len(cards) != 0 {
			testContext.Fatal("fresh account has cards")
		}
	})
	holder := suite.Invoke(testContext, client, contract.Call{
		Name:     "AddCardholder",
		Strings:  []string{token},
		Required: []string{"cardholderId", "status"},
	}).(*AddCardholderResponse)
	suite.Defaults["CardholderID"] = holder.CardholderID
	bins := suite.Invoke(testContext, client, contract.Call{
		Name:    "GetCardBin",
		Strings: []string{token},
	}).([]*CardBinResponse)
	if len(bins) == 0 {
		testContext.Fatal("missing card bins")
	}
	suite.Defaults["CardBin"] = bins[0].CardBin
	requestID := contract.Unique()
	card := suite.Invoke(testContext, client, contract.Call{
		Name:     "OpenCard",
		Strings:  []string{token},
		Fields:   map[string]any{"RequestID": requestID},
		Required: []string{"cardDetail.cardId", "status", "requestId"},
	}).(*OpenCardResponse)
	suite.Defaults["CardID"] = card.CardDetail.CardID
	calls := []contract.Call{
		{
			Name:     "GetAccountSingle",
			Fields:   map[string]any{"AccountType": "FT10001"},
			Required: []string{"accountNo", "realTimeBalance", "returnedAt"},
		},
		{
			Name: "AccountHistory",
			Fields: map[string]any{
				"TransactedAtStart": "2020-01-01T00:00:00",
				"TransactedAtEnd":   "2099-01-01T00:00:00",
			},
		},
		{Name: "EditCardholder"},
		{Name: "PagingVccCardholder"},
		{
			Name: "GetRequestResult",
			Fields: map[string]any{
				"RequestID": requestID,
				"Type":      "apply_card",
			},
			Required: []string{"cardDetail.cardId", "status"},
		},
		{
			Name:     "GetCardDetail",
			Required: []string{"cardId", "createdAt", "cardBalance"},
			Expect:   map[string]any{"cardId": card.CardDetail.CardID},
		},
		{
			Name: "PagingVccCard",
			Fields: map[string]any{
				"CardType":       nil,
				"CardFormFactor": nil,
			},
		},
		{
			Name:     "GetCvv",
			Required: []string{"cvv"},
		},
		{
			Name:   "UpdateCard",
			Fields: map[string]any{"TransactionLimitType": "unlimited"},
		},
		{Name: "EditCardBillingAddress"},
		{
			Name:   "FreezeCard",
			Fields: map[string]any{"Status": "freeze"},
			Expect: map[string]any{"cardStatus": "frozen"},
		},
		{
			Name:   "FreezeCard",
			Fields: map[string]any{"Status": "unfreeze"},
			Expect: map[string]any{"cardStatus": "normal"},
		},
		{
			Name:   "PreRecharge",
			Fields: map[string]any{"RechargeAmount": 10},
		},
		{Name: "Recharge"},
		{
			Name:   "RechargeReturn",
			Fields: map[string]any{"ReturnAmount": 1},
		},
		{Name: "PagingIssuingHistory"},
		{
			Name:   "PagingRechargeCardFundsDetail",
			Fields: map[string]any{"CardFormFactor": nil},
		},
		{
			Name:   "PagingShareCardTxnLimitDetail",
			Fields: map[string]any{"CardFormFactor": nil},
		},
		{
			Name: "SandBoxTransaction",
			Fields: map[string]any{
				"Cvv":              card.CardDetail.CVV,
				"ExpirationDate":   card.CardDetail.ExpirationDate,
				"TxnCurrency":      "USD",
				"TxnAmount":        1,
				"TxnType":          "auth",
				"MerchantName":     "Protocol merchant",
				"Mcc":              "5411",
				"MerchantCountry":  "US",
				"MerchantCity":     "Boston",
				"MerchantPostcode": "02101",
			},
		},
		{
			Name:   "PagingVccTradeOrder",
			Fields: map[string]any{"CardFormFactor": nil},
		},
		{
			Name: "ApiUploadFile",
			Fields: map[string]any{
				"FileReader":  io.NopCloser(strings.NewReader("test document")),
				"FileName":    "test.txt",
				"BusinessKey": "cardholder",
			},
		},
		{
			Name: "SetWebhookNotification",
			Fields: map[string]any{
				"TopicCode":    "vcc",
				"TemplateCode": "transaction",
			},
		},
		{Name: "GetWebhookNotification"},
		{
			Name: "DelWebhookNotification",
			Fields: map[string]any{
				"TopicCode":    "vcc",
				"TemplateCode": "transaction",
			},
		},
		{
			Name: "GetRequestResult",
			Fields: map[string]any{
				"RequestID": contract.Unique(),
				"Type":      "apply_card",
			},
			Error:     IsVCC1039,
			ErrorCode: "VCC1039",
		},
	}
	for _, call := range calls {
		call.Strings = []string{token}
		testContext.Run(call.Name, func(testContext *testing.T) { suite.Invoke(testContext, client, call) })
	}
	wrapped := NewPhotonpaySDK(client)
	testContext.Run("interface", func(testContext *testing.T) {
		suite.Invoke(testContext, wrapped, contract.Call{
			Name:  "CreateCardPre",
			Local: true,
		})
		suite.Invoke(testContext, wrapped, contract.Call{Name: "GetAccessToken"})
		created := suite.Invoke(testContext, wrapped, contract.Call{
			Name:    "CreateCard",
			Strings: []string{token},
		}).(*OpenCardResponse)
		suite.Invoke(testContext, wrapped, contract.Call{
			Name:    "GetCardInfo",
			Strings: []string{token, created.CardDetail.CardID},
		})
		suite.Invoke(testContext, wrapped, contract.Call{
			Name:    "GetCvv",
			Strings: []string{token, created.CardDetail.CardID},
		})
		suite.Invoke(testContext, wrapped, contract.Call{
			Name:    "GetRequestResult",
			Strings: []string{token},
			Fields: map[string]any{
				"RequestID": requestID,
				"Type":      "apply_card",
			},
			Schema: OpenCardResponse{},
		})
		suite.Invoke(testContext, wrapped, contract.Call{
			Name:    "UpdateCardStatus",
			Strings: []string{token},
			Fields:  map[string]any{"Status": "freeze"},
		})
		suite.Invoke(testContext, wrapped, contract.Call{
			Name:    "CancelCard",
			Strings: []string{token},
		})
		suite.Coverage(testContext, wrapped)
	})
	testContext.Run("CancelCard", func(testContext *testing.T) {
		suite.Invoke(testContext, client, contract.Call{
			Name:    "CancelCard",
			Strings: []string{token},
		})
	})
	suite.Coverage(testContext, client)
}
