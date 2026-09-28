package photonpay

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"sdk/internal/contract"
)

type photonFixture struct {
	suite    *contract.Suite
	client   *PhotonPaySDK
	token    string
	holderID string
	cardBin  string
}

func newPhotonFixture(testContext *testing.T) *photonFixture {
	testContext.Helper()
	suite := contract.New(testContext, "photonpay")
	client := New(suite.Config)
	token, err := client.GetAccessToken(suite.Context)
	if err != nil {
		testContext.Fatal(err)
	}
	if token == nil || token.Token == "" {
		testContext.Fatal("missing access token")
	}
	certificateType := "passport"
	holder, err := client.AddCardholder(suite.Context, token.Token, &AddCardholderRequest{
		FirstName:              "Protocol",
		LastName:               "Test",
		Email:                  contract.Unique() + "@example.test",
		Mobile:                 "2025550123",
		MobilePrefix:           "1",
		DateOfBirth:            "1990-01-02",
		NationalityCountryCode: "US",
		CertType:               &certificateType,
	})
	if err != nil {
		testContext.Fatal(err)
	}
	if holder == nil || holder.CardholderID == "" {
		testContext.Fatal("missing cardholder")
	}
	bins, err := client.GetCardBin(suite.Context, token.Token, &CardBinRequest{})
	if err != nil {
		testContext.Fatal(err)
	}
	if len(bins) == 0 || bins[0].CardBin == "" {
		testContext.Fatal("missing card BIN")
	}
	return &photonFixture{
		suite:    suite,
		client:   client,
		token:    token.Token,
		holderID: holder.CardholderID,
		cardBin:  bins[0].CardBin,
	}
}

func (fixture *photonFixture) cardRequest() *OpenCardRequest {
	return &OpenCardRequest{
		RequestID:      contract.Unique(),
		CardBin:        fixture.cardBin,
		CardholderID:   fixture.holderID,
		CardCurrency:   "USD",
		CardScheme:     "MasterCard",
		CardType:       "share",
		CardFormFactor: "virtual_card",
	}
}

func (fixture *photonFixture) createCard(testContext *testing.T) *OpenCardResponse {
	testContext.Helper()
	card, err := fixture.client.OpenCard(fixture.suite.Context, fixture.token, fixture.cardRequest())
	if err != nil {
		testContext.Fatal(err)
	}
	if card == nil || card.CardDetail.CardID == "" {
		testContext.Fatal("missing created card")
	}
	return card
}

func TestPhotonAccountsAndCardholders(testContext *testing.T) {
	fixture := newPhotonFixture(testContext)
	suite := fixture.suite
	client := fixture.client
	token := fixture.token
	testContext.Run("GetAccessToken", func(testContext *testing.T) {
		result, err := client.GetAccessToken(suite.Context)
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.Token != token || result.RefreshToken == "" {
			testContext.Fatalf("invalid token response: %+v", result)
		}
	})
	testContext.Run("GetAccountSingle", func(testContext *testing.T) {
		accountType := "FT10001"
		currency := "USD"
		result, err := client.GetAccountSingle(suite.Context, token, &AccountSingleRequest{
			AccountType: &accountType,
			Currency:    &currency,
		})
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.AccountNo == "" || result.Currency != "USD" || result.AccountType != accountType {
			testContext.Fatalf("invalid account response: %+v", result)
		}
		if _, err := time.Parse("2006-01-02T15:04:05", result.ReturnedAt); err != nil {
			testContext.Fatal(err)
		}
	})
	testContext.Run("AccountHistory", func(testContext *testing.T) {
		result, err := client.AccountHistory(suite.Context, token, &AccountHistoryRequest{
			TransactedAtStart: "2020-01-01T00:00:00",
			TransactedAtEnd:   "2099-01-01T00:00:00",
		})
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.AccountNo != suite.Config.Account {
			testContext.Fatalf("invalid account history: %+v", result)
		}
	})
	testContext.Run("AddCardholder", func(testContext *testing.T) {
		certificateType := "passport"
		result, err := client.AddCardholder(suite.Context, token, &AddCardholderRequest{
			FirstName:              "Second",
			LastName:               "Holder",
			Email:                  contract.Unique() + "@example.test",
			Mobile:                 "2025550123",
			MobilePrefix:           "1",
			DateOfBirth:            "1990-01-02",
			NationalityCountryCode: "US",
			CertType:               &certificateType,
		})
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.CardholderID == "" || result.CardholderID == fixture.holderID {
			testContext.Fatalf("invalid new cardholder: %+v", result)
		}
	})
	testContext.Run("EditCardholder", func(testContext *testing.T) {
		result, err := client.EditCardholder(suite.Context, token, &EditCardholderRequest{CardholderID: fixture.holderID})
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.CardholderID != fixture.holderID {
			testContext.Fatalf("wrong edited holder: %+v", result)
		}
	})
	testContext.Run("PagingVccCardholder", func(testContext *testing.T) {
		items, err := client.PagingVccCardholder(suite.Context, token, &PagingVccCardholderRequest{})
		if err != nil {
			testContext.Fatal(err)
		}
		if len(items) != 2 {
			testContext.Fatalf("got %d holders, want 2", len(items))
		}
		if _, err := time.Parse("2006-01-02T15:04:05", items[0].CreatedAt); err != nil {
			testContext.Fatal(err)
		}
	})
	testContext.Run("GetCardBin", func(testContext *testing.T) {
		items, err := client.GetCardBin(suite.Context, token, &CardBinRequest{})
		if err != nil {
			testContext.Fatal(err)
		}
		if len(items) == 0 || items[0].CardBin != fixture.cardBin {
			testContext.Fatalf("invalid BIN response: %+v", items)
		}
	})
}

func TestPhotonOpenCard(testContext *testing.T) {
	fixture := newPhotonFixture(testContext)
	testContext.Run("share", func(testContext *testing.T) {
		request := fixture.cardRequest()
		result, err := fixture.client.OpenCard(fixture.suite.Context, fixture.token, request)
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.RequestID != request.RequestID || result.Status != "succeed" {
			testContext.Fatalf("invalid open-card result: %+v", result)
		}
		if id, err := strconv.ParseInt(result.CardDetail.CardID, 10, 64); err != nil || id <= 0 {
			testContext.Fatalf("invalid card ID: %q", result.CardDetail.CardID)
		}
		if result.CardDetail.CardType != "share" || result.CardDetail.CardCurrency != "USD" || result.CardDetail.CardStatus != "normal" {
			testContext.Fatalf("invalid card fields: %+v", result.CardDetail)
		}
		if _, err := time.Parse("01/06", result.CardDetail.ExpirationDate); err != nil {
			testContext.Fatal(err)
		}
	})
	testContext.Run("recharge", func(testContext *testing.T) {
		request := fixture.cardRequest()
		request.CardType = "recharge"
		result, err := fixture.client.OpenCard(fixture.suite.Context, fixture.token, request)
		if err != nil || result == nil || result.CardDetail.CardType != "recharge" {
			testContext.Fatalf("recharge card failed: %+v %v", result, err)
		}
	})
	testContext.Run("missing-request-id", func(testContext *testing.T) {
		request := fixture.cardRequest()
		request.RequestID = ""
		if _, err := fixture.client.OpenCard(fixture.suite.Context, fixture.token, request); err == nil {
			testContext.Fatal("empty request ID accepted")
		}
	})
	testContext.Run("invalid-holder-id", func(testContext *testing.T) {
		request := fixture.cardRequest()
		request.CardholderID = "bad-id"
		if _, err := fixture.client.OpenCard(fixture.suite.Context, fixture.token, request); err == nil {
			testContext.Fatal("invalid holder ID accepted")
		}
	})
}

func TestPhotonCardQueries(testContext *testing.T) {
	fixture := newPhotonFixture(testContext)
	suite := fixture.suite
	client := fixture.client
	token := fixture.token
	testContext.Run("PagingVccCard/empty", func(testContext *testing.T) {
		items, err := client.PagingVccCard(suite.Context, token, &PagingVccCardRequest{})
		if err != nil {
			testContext.Fatal(err)
		}
		if items == nil || len(items) != 0 {
			testContext.Fatalf("want empty cards array, got %+v", items)
		}
	})
	card := fixture.createCard(testContext)
	testContext.Run("GetCardDetail", func(testContext *testing.T) {
		suite.Responses()
		result, err := client.GetCardDetail(suite.Context, token, &GetCardDetailRequest{CardID: card.CardDetail.CardID})
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.CardID != card.CardDetail.CardID || result.CardholderID != fixture.holderID {
			testContext.Fatalf("wrong card: %+v", result)
		}
		if _, err := strconv.ParseFloat(result.CardBalance, 64); err != nil {
			testContext.Fatal(err)
		}
		if _, err := time.Parse("2006-01-02T15:04:05", result.CreatedAt); err != nil {
			testContext.Fatal(err)
		}
		responses := suite.Responses()
		if len(responses) != 1 || responses[0].Status != http.StatusOK || !strings.HasPrefix(responses[0].ContentType, "application/json") {
			testContext.Fatalf("invalid card HTTP response: %+v", responses)
		}
		var wire struct {
			Code string `json:"code"`
			Msg  string `json:"msg"`
			Data struct {
				CardID      string `json:"cardId"`
				CardBalance string `json:"cardBalance"`
				CreatedAt   string `json:"createdAt"`
			} `json:"data"`
		}
		if err := json.Unmarshal(responses[0].Body, &wire); err != nil {
			testContext.Fatalf("invalid card field types: %v", err)
		}
		if wire.Code != "0000" || wire.Msg == "" || wire.Data.CardID != result.CardID || wire.Data.CardBalance != result.CardBalance || wire.Data.CreatedAt != result.CreatedAt {
			testContext.Fatalf("invalid card envelope: %+v", wire)
		}
	})
	testContext.Run("GetCvv", func(testContext *testing.T) {
		result, err := client.GetCvv(suite.Context, token, &GetCvvRequest{CardID: card.CardDetail.CardID})
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.CardID != card.CardDetail.CardID || len(result.Cvv) != 3 {
			testContext.Fatalf("invalid CVV response: %+v", result)
		}
	})
	testContext.Run("PagingVccCard/populated", func(testContext *testing.T) {
		items, err := client.PagingVccCard(suite.Context, token, &PagingVccCardRequest{})
		if err != nil {
			testContext.Fatal(err)
		}
		if len(items) != 1 || items[0].CardID != card.CardDetail.CardID {
			testContext.Fatalf("unexpected card list: %+v", items)
		}
	})
	testContext.Run("GetRequestResult/found", func(testContext *testing.T) {
		result, err := client.GetRequestResult(suite.Context, token, &GetRequestResultRequest{
			RequestID: card.RequestID,
			Type:      "apply_card",
		})
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.CardDetail.CardID != card.CardDetail.CardID {
			testContext.Fatalf("wrong request result: %+v", result)
		}
	})
	testContext.Run("GetRequestResult/VCC1039", func(testContext *testing.T) {
		suite.Responses()
		_, err := client.GetRequestResult(suite.Context, token, &GetRequestResultRequest{
			RequestID: contract.Unique(),
			Type:      "apply_card",
		})
		if !IsVCC1039(err) {
			testContext.Fatalf("want VCC1039, got %v", err)
		}
		responses := suite.Responses()
		if len(responses) == 0 {
			testContext.Fatal("no HTTP response for idempotency probe")
		}
		var envelope struct {
			Code    string `json:"code"`
			Message string `json:"msg"`
		}
		if err := json.Unmarshal(responses[len(responses)-1].Body, &envelope); err != nil {
			testContext.Fatal(err)
		}
		if envelope.Code != "VCC1039" || envelope.Message == "" {
			testContext.Fatalf("wrong missing-request envelope: %+v", envelope)
		}
	})
	testContext.Run("GetCardDetail/invalid-id", func(testContext *testing.T) {
		if _, err := client.GetCardDetail(suite.Context, token, &GetCardDetailRequest{CardID: "bad-id"}); err == nil {
			testContext.Fatal("invalid card ID accepted")
		}
	})
	testContext.Run("GetCardDetail/other-account", func(testContext *testing.T) {
		other := newPhotonFixture(testContext)
		if _, err := other.client.GetCardDetail(other.suite.Context, other.token, &GetCardDetailRequest{CardID: card.CardDetail.CardID}); err == nil {
			testContext.Fatal("cross-account card accepted")
		}
	})
}

func TestPhotonCardUpdates(testContext *testing.T) {
	fixture := newPhotonFixture(testContext)
	suite := fixture.suite
	client := fixture.client
	token := fixture.token
	card := fixture.createCard(testContext)
	cardID := card.CardDetail.CardID
	testContext.Run("UpdateCard", func(testContext *testing.T) {
		nickname := "Updated nickname"
		result, err := client.UpdateCard(suite.Context, token, &UpdateCardRequest{
			CardID:    cardID,
			RequestID: contract.Unique(),
			Nickname:  &nickname,
		})
		if err != nil || result == nil || result.CardDetail.CardID != cardID {
			testContext.Fatalf("invalid update response: %+v %v", result, err)
		}
	})
	testContext.Run("EditCardBillingAddress", func(testContext *testing.T) {
		address := "1 Test Street"
		if err := client.EditCardBillingAddress(suite.Context, token, &EditCardBillingAddressRequest{
			CardID:         cardID,
			BillingAddress: &address,
		}); err != nil {
			testContext.Fatal(err)
		}
	})
	testContext.Run("FreezeCard/freeze", func(testContext *testing.T) {
		if err := client.FreezeCard(suite.Context, token, &FreezeCardRequest{
			CardID:    cardID,
			RequestID: contract.Unique(),
			Status:    "freeze",
		}); err != nil {
			testContext.Fatal(err)
		}
		result, err := client.GetCardDetail(suite.Context, token, &GetCardDetailRequest{CardID: cardID})
		if err != nil || result == nil || result.CardStatus != "frozen" {
			testContext.Fatalf("freeze not visible: %+v %v", result, err)
		}
	})
	testContext.Run("FreezeCard/unfreeze", func(testContext *testing.T) {
		if err := client.FreezeCard(suite.Context, token, &FreezeCardRequest{
			CardID:    cardID,
			RequestID: contract.Unique(),
			Status:    "unfreeze",
		}); err != nil {
			testContext.Fatal(err)
		}
		result, err := client.GetCardDetail(suite.Context, token, &GetCardDetailRequest{CardID: cardID})
		if err != nil || result == nil || result.CardStatus != "normal" {
			testContext.Fatalf("unfreeze not visible: %+v %v", result, err)
		}
	})
	testContext.Run("FreezeCard/invalid-status", func(testContext *testing.T) {
		if err := client.FreezeCard(suite.Context, token, &FreezeCardRequest{
			CardID:    cardID,
			RequestID: contract.Unique(),
			Status:    "invalid",
		}); err == nil {
			testContext.Fatal("invalid status accepted")
		}
	})
	testContext.Run("CancelCard", func(testContext *testing.T) {
		if err := client.CancelCard(suite.Context, token, &CancelCardRequest{CardID: cardID}); err != nil {
			testContext.Fatal(err)
		}
		result, err := client.GetCardDetail(suite.Context, token, &GetCardDetailRequest{CardID: cardID})
		if err != nil || result == nil || result.CardStatus != "cancelled" {
			testContext.Fatalf("cancel not visible: %+v %v", result, err)
		}
	})
}

func TestPhotonFundingAndUtilities(testContext *testing.T) {
	fixture := newPhotonFixture(testContext)
	suite := fixture.suite
	client := fixture.client
	token := fixture.token
	card := fixture.createCard(testContext)
	cardID := card.CardDetail.CardID
	testContext.Run("PreRecharge", func(testContext *testing.T) {
		amount := float64(10)
		result, err := client.PreRecharge(suite.Context, token, &PreRechargeRequest{
			RequestID:      contract.Unique(),
			AccountID:      suite.Config.Account,
			CardID:         cardID,
			RechargeAmount: &amount,
		})
		if err != nil || result == nil || result.RechargeAmount != amount {
			testContext.Fatalf("invalid quotation: %+v %v", result, err)
		}
		if _, err := time.Parse("2006-01-02T15:04:05", result.QuotedAt); err != nil {
			testContext.Fatal(err)
		}
	})
	testContext.Run("Recharge", func(testContext *testing.T) {
		result, err := client.Recharge(suite.Context, token, &RechargeRequest{RequestID: contract.Unique()})
		if err != nil || result == nil {
			testContext.Fatalf("invalid recharge result: %+v %v", result, err)
		}
	})
	testContext.Run("RechargeReturn", func(testContext *testing.T) {
		result, err := client.RechargeReturn(suite.Context, token, &RechargeReturnRequest{
			RequestID:    contract.Unique(),
			CardID:       cardID,
			ReturnAmount: 1,
		})
		if err != nil || result == nil || result.CardID != cardID {
			testContext.Fatalf("invalid return result: %+v %v", result, err)
		}
	})
	testContext.Run("PagingIssuingHistory", func(testContext *testing.T) {
		items, err := client.PagingIssuingHistory(suite.Context, token, &PagingIssuingHistoryRequest{})
		if err != nil || len(items) != 1 || items[0].CardID != cardID {
			testContext.Fatalf("invalid issuing history: %+v %v", items, err)
		}
	})
	testContext.Run("PagingRechargeCardFundsDetail", func(testContext *testing.T) {
		items, err := client.PagingRechargeCardFundsDetail(suite.Context, token, &PagingRechargeCardFundsDetailRequest{})
		if err != nil || len(items) != 1 || items[0].CardID != cardID {
			testContext.Fatalf("invalid funding history: %+v %v", items, err)
		}
	})
	testContext.Run("PagingShareCardTxnLimitDetail", func(testContext *testing.T) {
		items, err := client.PagingShareCardTxnLimitDetail(suite.Context, token, &PagingShareCardTxnLimitDetailRequest{})
		if err != nil || len(items) != 1 || items[0].CardID != cardID {
			testContext.Fatalf("invalid limit history: %+v %v", items, err)
		}
	})
	testContext.Run("ApiUploadFile", func(testContext *testing.T) {
		url, err := client.ApiUploadFile(suite.Context, token, &ApiUploadFileRequest{
			FileReader:  io.NopCloser(strings.NewReader("local test document")),
			FileName:    "test.txt",
			BusinessKey: "cardholder",
		})
		if err != nil || url == "" {
			testContext.Fatalf("missing file URL: %q %v", url, err)
		}
	})
	testContext.Run("SetWebhookNotification", func(testContext *testing.T) {
		if err := client.SetWebhookNotification(suite.Context, token, &WebhookNotificationRequest{
			TopicCode:    "vcc",
			TemplateCode: "transaction",
		}); err != nil {
			testContext.Fatal(err)
		}
	})
	testContext.Run("GetWebhookNotification", func(testContext *testing.T) {
		result, err := client.GetWebhookNotification(suite.Context, token)
		if err != nil || result == nil || len(result.Categories) == 0 {
			testContext.Fatalf("invalid subscriptions: %+v %v", result, err)
		}
	})
	testContext.Run("DelWebhookNotification", func(testContext *testing.T) {
		if err := client.DelWebhookNotification(suite.Context, token, &WebhookNotificationRequest{
			TopicCode:    "vcc",
			TemplateCode: "transaction",
		}); err != nil {
			testContext.Fatal(err)
		}
	})
}

func TestPhotonTransactions(testContext *testing.T) {
	fixture := newPhotonFixture(testContext)
	suite := fixture.suite
	client := fixture.client
	token := fixture.token
	card := fixture.createCard(testContext)
	cardID := card.CardDetail.CardID
	cards := suite.UI(testContext, http.MethodGet, "/cards?account_id="+suite.Config.Account+"&id="+cardID, nil)
	if actual := contract.Text(testContext, cards, "data.0.id"); actual != cardID {
		testContext.Fatalf("unexpected funding card: %s", actual)
	}
	suite.UI(testContext, http.MethodPost, "/funds/transfer", map[string]any{
		"account_id": suite.Config.Account,
		"source_id":  suite.Config.Wallet,
		"target_id":  contract.Text(testContext, cards, "data.0.wallet_id"),
		"amount":     "10",
	})
	testContext.Run("PagingVccTradeOrder/empty", func(testContext *testing.T) {
		items, err := client.PagingVccTradeOrder(suite.Context, token, &PagingVccTradeOrderRequest{CardID: &cardID})
		if err != nil || items == nil || len(items) != 0 {
			testContext.Fatalf("want empty trades, got %+v %v", items, err)
		}
	})
	authorization := suite.UI(testContext, http.MethodPost, "/simulate/authorizations", map[string]any{
		"card_id":                cardID,
		"transaction_amount":     2,
		"transaction_currency":   "USD",
		"merchant_name":          "SDK merchant",
		"merchant_country":       "US",
		"merchant_category_code": "5411",
	})
	authorizationID := contract.Text(testContext, authorization, "authorization.id")
	authorizationTransactionID := contract.Text(testContext, authorization, "transaction.id")
	testContext.Run("PagingVccTradeOrder/authorization", func(testContext *testing.T) {
		items, err := client.PagingVccTradeOrder(suite.Context, token, &PagingVccTradeOrderRequest{CardID: &cardID})
		if err != nil || len(items) != 1 {
			testContext.Fatalf("want one authorization, got %+v %v", items, err)
		}
		item := items[0]
		if item.CardID != cardID || item.TransactionID != authorizationTransactionID || item.TransactionType != "auth" || item.TransactionCurrency != "USD" {
			testContext.Fatalf("invalid trade: %+v", item)
		}
		if _, err := time.Parse("2006-01-02T15:04:05", item.CreatedAt); err != nil {
			testContext.Fatal(err)
		}
	})
	linkedRefund := suite.UI(testContext, http.MethodPost, "/authorizations/"+authorizationID+"/refund", map[string]any{
		"amount": 1,
	})
	linkedRefundID := contract.Text(testContext, linkedRefund, "id")
	independentRefund := suite.UI(testContext, http.MethodPost, "/simulate/refunds", map[string]any{
		"card_id":                cardID,
		"amount":                 3,
		"currency":               "USD",
		"merchant_name":          "SDK merchant",
		"merchant_country":       "US",
		"merchant_category_code": "5411",
	})
	independentRefundID := contract.Text(testContext, independentRefund, "id")
	reversal := suite.UI(testContext, http.MethodPost, "/authorizations/"+authorizationID+"/reverse", map[string]any{
		"amount": 2,
	})
	reversalID := contract.Text(testContext, reversal, "id")
	testContext.Run("PagingVccTradeOrder/refunds-and-void", func(testContext *testing.T) {
		items, err := client.PagingVccTradeOrder(suite.Context, token, &PagingVccTradeOrderRequest{CardID: &cardID})
		if err != nil || len(items) != 4 {
			testContext.Fatalf("invalid transaction list: %+v %v", items, err)
		}
		if items[0].TransactionID != reversalID || items[1].TransactionID != independentRefundID || items[2].TransactionID != linkedRefundID || items[3].TransactionID != authorizationTransactionID {
			testContext.Fatalf("wrong transaction IDs or order: %+v", items)
		}
		if items[0].TransactionType != "void" || items[1].TransactionType != "refund" || items[2].TransactionType != "refund" || items[3].TransactionType != "auth" {
			testContext.Fatalf("wrong transaction types: %+v", items)
		}
		for _, item := range items {
			if item.CardID != cardID || item.TransactionID == "" || item.Status == "" {
				testContext.Fatalf("invalid trade: %+v", item)
			}
			if _, err := time.Parse("2006-01-02T15:04:05", item.TxnDate); err != nil {
				testContext.Fatal(err)
			}
		}
	})
	testContext.Run("PagingVccTradeOrder/other-account", func(testContext *testing.T) {
		other := newPhotonFixture(testContext)
		items, err := other.client.PagingVccTradeOrder(other.suite.Context, other.token, &PagingVccTradeOrderRequest{})
		if err != nil || len(items) != 0 {
			testContext.Fatalf("cross-account trades leaked: %+v %v", items, err)
		}
	})
	testContext.Run("PagingVccTradeOrder/pagination", func(testContext *testing.T) {
		page := uint32(2)
		size := uint32(1)
		items, err := client.PagingVccTradeOrder(suite.Context, token, &PagingVccTradeOrderRequest{
			PageIndex: &page,
			PageSize:  &size,
		})
		if err != nil {
			testContext.Fatal(err)
		}
		if len(items) != 1 || items[0].TransactionType != "refund" || items[0].CardID != cardID {
			testContext.Fatalf("invalid second transaction page: %+v", items)
		}
	})
}

func TestPhotonInterface(testContext *testing.T) {
	fixture := newPhotonFixture(testContext)
	suite, token := fixture.suite, fixture.token
	client := NewPhotonpaySDK(fixture.client)
	request := &PhotonpayCreateCardRequest{
		RequestID:      contract.Unique(),
		CardBin:        fixture.cardBin,
		CardholderID:   fixture.holderID,
		CardCurrency:   "USD",
		CardScheme:     "MasterCard",
		CardType:       "share",
		CardFormFactor: "virtual_card",
	}
	testContext.Run("CreateCardPre", func(testContext *testing.T) {
		raw, err := client.CreateCardPre(request)
		if err != nil {
			testContext.Fatal(err)
		}
		var decoded OpenCardRequest
		if err := json.Unmarshal(raw, &decoded); err != nil {
			testContext.Fatal(err)
		}
		if decoded.CardBin != fixture.cardBin || decoded.RequestID != request.RequestID {
			testContext.Fatalf("invalid preflight request: %+v", decoded)
		}
	})
	testContext.Run("GetAccessToken", func(testContext *testing.T) {
		result, err := client.GetAccessToken(suite.Context)
		if err != nil || result == nil || result.Token != token {
			testContext.Fatalf("invalid wrapper token: %+v %v", result, err)
		}
	})
	var cardID string
	testContext.Run("CreateCard", func(testContext *testing.T) {
		result, err := client.CreateCard(suite.Context, token, request)
		if err != nil || result == nil || result.CardDetail.CardID == "" {
			testContext.Fatalf("invalid wrapper card: %+v %v", result, err)
		}
		cardID = result.CardDetail.CardID
	})
	if cardID == "" {
		testContext.Fatal("wrapper creation failed")
	}
	testContext.Run("GetCardInfo", func(testContext *testing.T) {
		result, err := client.GetCardInfo(suite.Context, token, cardID)
		if err != nil || result == nil || result.CardID != cardID {
			testContext.Fatalf("invalid wrapper detail: %+v %v", result, err)
		}
	})
	testContext.Run("GetCvv", func(testContext *testing.T) {
		result, err := client.GetCvv(suite.Context, token, cardID)
		if err != nil || result == nil || len(result.Cvv) != 3 {
			testContext.Fatalf("invalid wrapper CVV: %+v %v", result, err)
		}
	})
	testContext.Run("GetRequestResult", func(testContext *testing.T) {
		result, err := client.GetRequestResult(suite.Context, token, &PhotonpayGetRequestResultRequest{
			RequestID: request.RequestID,
			Type:      "apply_card",
		})
		if err != nil || result == nil || result.CardDetail == nil || result.CardDetail.CardID != cardID {
			testContext.Fatalf("invalid wrapper request result: %+v %v", result, err)
		}
	})
	testContext.Run("UpdateCardStatus", func(testContext *testing.T) {
		if err := client.UpdateCardStatus(suite.Context, token, &FreezeCardRequest{
			CardID:    cardID,
			RequestID: contract.Unique(),
			Status:    "freeze",
		}); err != nil {
			testContext.Fatal(err)
		}
		result, err := client.GetCardInfo(suite.Context, token, cardID)
		if err != nil || result == nil || result.CardStatus != "frozen" {
			testContext.Fatalf("wrapper freeze not visible: %+v %v", result, err)
		}
	})
	testContext.Run("CancelCard", func(testContext *testing.T) {
		if err := client.CancelCard(suite.Context, token, &CancelCardRequest{CardID: cardID}); err != nil {
			testContext.Fatal(err)
		}
	})
}
