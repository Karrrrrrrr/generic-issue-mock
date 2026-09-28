package slash

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"sdk/internal/contract"

	"github.com/google/uuid"
)

type slashFixture struct {
	suite     *contract.Suite
	client    *SlashSDK
	productID string
}

func newSlashFixture(testContext *testing.T) *slashFixture {
	testContext.Helper()
	suite := contract.New(testContext, "slash")
	virtual := suite.UI(testContext, http.MethodPost, "/managed-virtual-accounts", map[string]any{
		"account_id": suite.Config.Account,
		"name":       "SDK virtual account",
		"currency":   "USD",
	})
	suite.Config.VirtualAccount = contract.Text(testContext, virtual, "id")
	suite.UI(testContext, http.MethodPost, "/funds/transfer", map[string]any{
		"account_id": suite.Config.Account,
		"source_id":  suite.Config.Wallet,
		"target_id":  contract.Text(testContext, virtual, "wallet_id"),
		"amount":     "1000",
	})
	client := New(suite.Config)
	products, err := client.ListCardProducts(suite.Context)
	if err != nil {
		testContext.Fatal(err)
	}
	if len(products) == 0 {
		testContext.Fatal("missing card products")
	}
	return &slashFixture{
		suite:     suite,
		client:    client,
		productID: products[0].ID,
	}
}

func (fixture *slashFixture) createCard(testContext *testing.T, name string) *Card {
	testContext.Helper()
	card, err := fixture.client.CreateCard(fixture.suite.Context, &CreateCardReq{
		RequestID: contract.Unique(),
		CardID:    contract.Unique(),
		Name:      name,
		CardBinID: fixture.productID,
	})
	if err != nil {
		testContext.Fatal(err)
	}
	if card == nil {
		testContext.Fatal("missing created card")
	}
	if actual, expected := card.AccountID, fixture.suite.Config.Account; actual != expected {
		testContext.Fatalf("got %v, want %v", actual, expected)
	}
	if actual, expected := card.VirtualAccountID, fixture.suite.Config.VirtualAccount; actual != expected {
		testContext.Fatalf("got %v, want %v", actual, expected)
	}
	if actual, expected := card.Status, "active"; actual != expected {
		testContext.Fatalf("got %v, want %v", actual, expected)
	}
	if card.CreatedAt.IsZero() {
		testContext.Fatal("missing creation timestamp")
	}
	return card
}

func TestSlashAccounts(testContext *testing.T) {
	fixture := newSlashFixture(testContext)
	suite := fixture.suite
	client := fixture.client
	testContext.Run("ListLegalEntity", func(testContext *testing.T) {
		items, err := client.ListLegalEntity(suite.Context)
		if err != nil {
			testContext.Fatal(err)
		}
		if len(items) == 0 {
			testContext.Fatal("missing legal entity")
		}
	})
	testContext.Run("ListAccounts", func(testContext *testing.T) {
		items, err := client.ListAccounts(suite.Context)
		if err != nil {
			testContext.Fatal(err)
		}
		if actual, expected := len(items), 1; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
		if actual, expected := items[0].ID, suite.Config.Account; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
	})
	testContext.Run("GetAccount", func(testContext *testing.T) {
		item, err := client.GetAccount(suite.Context)
		if err != nil {
			testContext.Fatal(err)
		}
		if actual, expected := item.ID, suite.Config.Account; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
	})
	testContext.Run("ListAccountBalances", func(testContext *testing.T) {
		items, err := client.ListAccountBalances(suite.Context)
		if err != nil {
			testContext.Fatal(err)
		}
		if actual, expected := len(items), 1; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
		if actual, expected := items[0].AccountID, suite.Config.Account; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
	})
	testContext.Run("ListCardProducts", func(testContext *testing.T) {
		items, err := client.ListCardProducts(suite.Context)
		if err != nil {
			testContext.Fatal(err)
		}
		if len(items) == 0 {
			testContext.Fatal("missing card products")
		}
	})
	testContext.Run("CreateVirtualAccount", func(testContext *testing.T) {
		item, err := client.CreateVirtualAccount(suite.Context, "SDK created virtual account")
		if err != nil {
			testContext.Fatal(err)
		}
		if item == nil {
			testContext.Fatal("missing virtual account")
		}
	})
	testContext.Run("UpdateVirtualAccount", func(testContext *testing.T) {
		_, err := client.UpdateVirtualAccount(suite.Context, suite.Config.VirtualAccount, "Renamed virtual account")
		if err != nil {
			testContext.Fatal(err)
		}
	})
	testContext.Run("GetVirtualAccount", func(testContext *testing.T) {
		item, err := client.GetVirtualAccount(suite.Context)
		if err != nil {
			testContext.Fatal(err)
		}
		if actual, expected := item.VirtualAccount.ID, suite.Config.VirtualAccount; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
		if actual, expected := item.VirtualAccount.Name, "Renamed virtual account"; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
	})
	testContext.Run("ListVirtualAccounts", func(testContext *testing.T) {
		items, err := client.ListVirtualAccounts(suite.Context)
		if err != nil {
			testContext.Fatal(err)
		}
		if actual, expected := len(items), 2; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
	})
	testContext.Run("VirtualAccountTransfer", func(testContext *testing.T) {
		destination := suite.UI(testContext, http.MethodPost, "/managed-virtual-accounts", map[string]any{
			"account_id": suite.Config.Account,
			"name":       "Destination",
			"currency":   "USD",
		})
		_, err := client.VirtualAccountTransfer(suite.Context, &TransferRequest{
			RequestID:   contract.Unique(),
			Source:      suite.Config.VirtualAccount,
			Destination: contract.Text(testContext, destination, "id"),
			AmountCents: 100,
		})
		if err != nil {
			testContext.Fatal(err)
		}
	})
}

func TestSlashCreateCard(testContext *testing.T) {
	fixture := newSlashFixture(testContext)
	testContext.Run("valid", func(testContext *testing.T) {
		fixture.suite.Responses()
		first, err := fixture.client.CreateCard(fixture.suite.Context, &CreateCardReq{
			RequestID: contract.Unique(),
			CardID:    contract.Unique(),
			Name:      "First card",
			CardBinID: fixture.productID,
		})
		if err != nil {
			testContext.Fatal(err)
		}
		if first == nil || first.AccountID != fixture.suite.Config.Account || first.Status != "active" {
			testContext.Fatalf("invalid created card: %+v", first)
		}
		if parsed, err := uuid.Parse(first.ID); err != nil || parsed.String() != first.ID || parsed == uuid.Nil {
			testContext.Fatalf("invalid canonical UUID: %q", first.ID)
		}
		if first.CreatedAt.IsZero() {
			testContext.Fatal("missing RFC3339 creation time")
		}
		responses := fixture.suite.Responses()
		if len(responses) != 2 || responses[0].Method != http.MethodPost || responses[1].Method != http.MethodGet {
			testContext.Fatalf("expected create and detail responses: %+v", responses)
		}
		if responses[0].Status != http.StatusOK || !strings.HasPrefix(responses[0].ContentType, "application/json") {
			testContext.Fatalf("invalid card HTTP response: %+v", responses[0])
		}
		var wire struct {
			ID         string `json:"id"`
			AccountID  string `json:"accountId"`
			CreatedAt  string `json:"createdAt"`
			ExpiryYear string `json:"expiryYear"`
			IsPhysical *bool  `json:"isPhysical"`
		}
		if err := json.Unmarshal(responses[0].Body, &wire); err != nil {
			testContext.Fatalf("invalid card field types: %v", err)
		}
		if wire.ID != first.ID || wire.AccountID != first.AccountID || wire.ExpiryYear != first.ExpiryYear || wire.IsPhysical == nil || *wire.IsPhysical {
			testContext.Fatalf("invalid card wire fields: %+v", wire)
		}
		if _, err := time.Parse(time.RFC3339, wire.CreatedAt); err != nil {
			testContext.Fatal(err)
		}
		if month, err := strconv.Atoi(first.ExpiryMonth); err != nil || month < 1 || month > 12 {
			testContext.Fatalf("invalid expiry month: %q", first.ExpiryMonth)
		}
		if _, err := time.Parse("2006", first.ExpiryYear); err != nil {
			testContext.Fatal(err)
		}
		if _, err := strconv.Atoi(first.Cvv); err != nil || len(first.Cvv) != 3 {
			testContext.Fatalf("invalid CVV: %q", first.Cvv)
		}
		second, err := fixture.client.CreateCard(fixture.suite.Context, &CreateCardReq{
			RequestID: contract.Unique(),
			CardID:    contract.Unique(),
			Name:      "Second card",
			CardBinID: fixture.productID,
		})
		if err != nil {
			testContext.Fatal(err)
		}
		if second == nil || first.ID == second.ID {
			testContext.Fatal("distinct requests returned the same card")
		}
	})
	testContext.Run("missing-request-id", func(testContext *testing.T) {
		_, err := fixture.client.CreateCard(fixture.suite.Context, &CreateCardReq{
			Name:      "Invalid card",
			CardBinID: fixture.productID,
		})
		if err == nil {
			testContext.Fatal("expected request rejection")
		}
		if actual, expected := err, ErrEmptyRequestID; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
	})
	testContext.Run("invalid-product-id", func(testContext *testing.T) {
		_, err := fixture.client.CreateCard(fixture.suite.Context, &CreateCardReq{
			RequestID: contract.Unique(),
			Name:      "Invalid product",
			CardBinID: "not-a-uuid",
		})
		if err == nil {
			testContext.Fatal("expected request rejection")
		}
	})
}

func TestSlashCardQueries(testContext *testing.T) {
	fixture := newSlashFixture(testContext)
	suite := fixture.suite
	client := fixture.client
	testContext.Run("ListCards/empty", func(testContext *testing.T) {
		items, err := client.ListCards(suite.Context, &QueryCardsParams{})
		if err != nil {
			testContext.Fatal(err)
		}
		if items == nil || items.Items == nil || len(items.Items) != 0 {
			testContext.Fatalf("expected empty cards array: %+v", items)
		}
		if actual, expected := items.Metadata.Count, 0; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
	})
	card := fixture.createCard(testContext, "Query card")
	testContext.Run("GetCard/valid", func(testContext *testing.T) {
		item, err := client.GetCard(suite.Context, card.ID)
		if err != nil {
			testContext.Fatal(err)
		}
		if actual, expected := item.ID, card.ID; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
		if actual, expected := item.Pan, card.Pan; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
	})
	testContext.Run("QueryCard/valid", func(testContext *testing.T) {
		item, err := client.QueryCard(suite.Context, card.ID)
		if err != nil {
			testContext.Fatal(err)
		}
		if actual, expected := item.ID, card.ID; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
	})
	testContext.Run("GetCard/empty-id", func(testContext *testing.T) {
		_, err := client.GetCard(suite.Context, "")
		if err == nil {
			testContext.Fatal("expected request rejection")
		}
	})
	testContext.Run("QueryCard/invalid-id", func(testContext *testing.T) {
		_, err := client.QueryCard(suite.Context, "not-a-uuid")
		if err == nil {
			testContext.Fatal("expected request rejection")
		}
	})
	testContext.Run("GetCard/other-account", func(testContext *testing.T) {
		other := newSlashFixture(testContext)
		_, err := other.client.GetCard(other.suite.Context, card.ID)
		if err == nil {
			testContext.Fatal("expected request rejection")
		}
	})
	testContext.Run("ListCards/populated", func(testContext *testing.T) {
		items, err := client.ListCards(suite.Context, &QueryCardsParams{})
		if err != nil {
			testContext.Fatal(err)
		}
		if actual, expected := len(items.Items), 1; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
		if actual, expected := items.Items[0].ID, card.ID; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
	})
	testContext.Run("ListCards/empty-status-filter", func(testContext *testing.T) {
		items, err := client.ListCards(suite.Context, &QueryCardsParams{Status: "closed"})
		if err != nil {
			testContext.Fatal(err)
		}
		if actual, expected := len(items.Items), 0; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
	})
}

func TestSlashCardStatus(testContext *testing.T) {
	fixture := newSlashFixture(testContext)
	suite := fixture.suite
	client := fixture.client
	card := fixture.createCard(testContext, "Status card")
	testContext.Run("FrozenCard", func(testContext *testing.T) {
		err := client.FrozenCard(suite.Context, card.ID)
		if err != nil {
			testContext.Fatal(err)
		}
		item, err := client.GetCard(suite.Context, card.ID)
		if err != nil {
			testContext.Fatal(err)
		}
		if actual, expected := item.Status, "paused"; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
	})
	testContext.Run("UnfrozenCard", func(testContext *testing.T) {
		err := client.UnfrozenCard(suite.Context, card.ID)
		if err != nil {
			testContext.Fatal(err)
		}
		item, err := client.GetCard(suite.Context, card.ID)
		if err != nil {
			testContext.Fatal(err)
		}
		if actual, expected := item.Status, "active"; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
	})
	testContext.Run("ReleaseCard", func(testContext *testing.T) {
		err := client.ReleaseCard(suite.Context, card.ID)
		if err != nil {
			testContext.Fatal(err)
		}
		item, err := client.GetCard(suite.Context, card.ID)
		if err != nil {
			testContext.Fatal(err)
		}
		if actual, expected := item.Status, "closed"; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
	})
	testContext.Run("FrozenCard/empty-id", func(testContext *testing.T) {
		if err := client.FrozenCard(suite.Context, ""); err == nil {
			testContext.Fatal("expected request rejection")
		}
	})
	testContext.Run("UnfrozenCard/empty-id", func(testContext *testing.T) {
		if err := client.UnfrozenCard(suite.Context, ""); err == nil {
			testContext.Fatal("expected request rejection")
		}
	})
	testContext.Run("ReleaseCard/empty-id", func(testContext *testing.T) {
		if err := client.ReleaseCard(suite.Context, ""); err == nil {
			testContext.Fatal("expected request rejection")
		}
	})
}

func TestSlashTransactions(testContext *testing.T) {
	fixture := newSlashFixture(testContext)
	suite := fixture.suite
	client := fixture.client
	card := fixture.createCard(testContext, "Transaction card")
	otherCard := fixture.createCard(testContext, "Other transaction card")
	testContext.Run("ListTransactions/empty", func(testContext *testing.T) {
		items, err := client.ListTransactions(suite.Context, QueryTransactionsParams{})
		if err != nil {
			testContext.Fatal(err)
		}
		if actual, expected := len(items.Items), 0; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
	})
	refund := suite.UI(testContext, http.MethodPost, "/simulate/refunds", map[string]any{
		"card_id":                card.ID,
		"amount":                 2,
		"currency":               "USD",
		"merchant_name":          "Refund merchant",
		"merchant_country":       "US",
		"merchant_category_code": "5411",
	})
	refundID := contract.Text(testContext, refund, "id")
	suite.UI(testContext, http.MethodPost, "/simulate/refunds", map[string]any{
		"card_id":                otherCard.ID,
		"amount":                 3,
		"currency":               "USD",
		"merchant_name":          "Other merchant",
		"merchant_country":       "US",
		"merchant_category_code": "5411",
	})
	testContext.Run("ListTransactions/populated", func(testContext *testing.T) {
		items, err := client.ListTransactions(suite.Context, QueryTransactionsParams{})
		if err != nil {
			testContext.Fatal(err)
		}
		if actual, expected := len(items.Items), 2; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
	})
	testContext.Run("ListTransactions/card-filter", func(testContext *testing.T) {
		items, err := client.ListTransactions(suite.Context, QueryTransactionsParams{FilterCardId: card.ID})
		if err != nil {
			testContext.Fatal(err)
		}
		if actual, expected := len(items.Items), 1; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
		if actual, expected := items.Items[0].ID, refundID; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
	})
	testContext.Run("GetTransaction/refund", func(testContext *testing.T) {
		item, err := client.GetTransaction(suite.Context, refundID)
		if err != nil {
			testContext.Fatal(err)
		}
		if actual, expected := item.ID, refundID; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
		if actual, expected := item.CardID, card.ID; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
		if _, err := time.Parse(time.RFC3339, item.Date); err != nil {
			testContext.Fatal(err)
		}
	})
	testContext.Run("GetTransaction/invalid-id", func(testContext *testing.T) {
		_, err := client.GetTransaction(suite.Context, "not-a-uuid")
		if err == nil {
			testContext.Fatal("expected request rejection")
		}
	})
	testContext.Run("GetTransaction/other-account", func(testContext *testing.T) {
		other := newSlashFixture(testContext)
		_, err := other.client.GetTransaction(other.suite.Context, refundID)
		if err == nil {
			testContext.Fatal("expected request rejection")
		}
	})
	testContext.Run("ListTransactions/invalid-card-filter", func(testContext *testing.T) {
		_, err := client.ListTransactions(suite.Context, QueryTransactionsParams{FilterCardId: "bad-id"})
		if err == nil {
			testContext.Fatal("expected request rejection")
		}
	})
}

func TestSlashAuthorizationTransactions(testContext *testing.T) {
	fixture := newSlashFixture(testContext)
	suite := fixture.suite
	card := fixture.createCard(testContext, "Authorization card")
	authorization := suite.UI(testContext, http.MethodPost, "/simulate/authorizations", map[string]any{
		"card_id":                card.ID,
		"transaction_amount":     2,
		"transaction_currency":   "USD",
		"merchant_name":          "Authorization merchant",
		"merchant_country":       "US",
		"merchant_category_code": "5411",
	})
	authorizationID := contract.Text(testContext, authorization, "authorization.id")
	transactionID := contract.Text(testContext, authorization, "transaction.id")
	testContext.Run("GetTransaction/authorization", func(testContext *testing.T) {
		result, err := fixture.client.GetTransaction(suite.Context, transactionID)
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.ID != transactionID || result.CardID != card.ID || result.AmountCents != 200 {
			testContext.Fatalf("invalid authorization transaction: %+v", result)
		}
		if _, err := time.Parse(time.RFC3339, result.AuthorizedAt); err != nil {
			testContext.Fatal(err)
		}
	})
	clearing := suite.UI(testContext, http.MethodPost, "/authorizations/"+authorizationID+"/clear", map[string]any{
		"amount": "3",
	})
	clearingID := contract.Text(testContext, clearing, "id")
	testContext.Run("GetTransaction/over-clearing", func(testContext *testing.T) {
		result, err := fixture.client.GetTransaction(suite.Context, clearingID)
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.ID != clearingID || result.CardID != card.ID || result.AmountCents != 300 {
			testContext.Fatalf("invalid clearing transaction: %+v", result)
		}
		if _, err := time.Parse(time.RFC3339, result.Date); err != nil {
			testContext.Fatal(err)
		}
		if _, err := time.Parse(time.RFC3339, result.AuthorizedAt); err != nil {
			testContext.Fatal(err)
		}
	})
	testContext.Run("ListTransactions/authorization-and-clearing", func(testContext *testing.T) {
		result, err := fixture.client.ListTransactions(suite.Context, QueryTransactionsParams{FilterCardId: card.ID})
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || len(result.Items) != 2 || result.Items[0].ID != clearingID || result.Items[1].ID != transactionID {
			testContext.Fatalf("invalid transaction list: %+v", result)
		}
	})
}

func TestSlashInterface(testContext *testing.T) {
	fixture := newSlashFixture(testContext)
	suite := fixture.suite
	client := NewSlashSDK(false, SlashSDKConf{
		Url:              suite.Config.URL,
		VaultUrl:         suite.Config.URL,
		ApiKey:           suite.Config.Account,
		AccountID:        suite.Config.Account,
		VirtualAccountID: suite.Config.VirtualAccount,
	})
	request := &CardPostRequest{
		Name:          "Interface card",
		CardProductID: fixture.productID,
		RequestID:     contract.Unique(),
		CardID:        contract.Unique(),
	}
	testContext.Run("CardPostPre", func(testContext *testing.T) {
		raw, err := client.CardPostPre(request)
		if err != nil {
			testContext.Fatal(err)
		}
		var decoded CreateCardRequest
		if err := json.Unmarshal(raw, &decoded); err != nil {
			testContext.Fatal(err)
		}
		if actual, expected := decoded.CardProductID, fixture.productID; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
	})
	var cardID string
	testContext.Run("CardPost", func(testContext *testing.T) {
		var err error
		cardID, err = client.CardPost(suite.Context, request)
		if err != nil {
			testContext.Fatal(err)
		}
		if cardID == "" {
			testContext.Fatal("missing card ID")
		}
	})
	if cardID == "" {
		testContext.Fatal("card creation failed")
	}
	testContext.Run("GetCardByID", func(testContext *testing.T) {
		card, err := client.GetCardByID(suite.Context, cardID)
		if err != nil {
			testContext.Fatal(err)
		}
		if card.Pan != "" && card.Cvv == "" {
			testContext.Fatal("missing sensitive card data")
		}
	})
	testContext.Run("CardFrozen", func(testContext *testing.T) {
		if err := client.CardFrozen(suite.Context, cardID); err != nil {
			testContext.Fatal(err)
		}
		card, err := client.GetCardByID(suite.Context, cardID)
		if err != nil {
			testContext.Fatal(err)
		}
		if actual, expected := card.Status, CardStatusEnumPaused; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
	})
	testContext.Run("CardUnfrozen", func(testContext *testing.T) {
		if err := client.CardUnfrozen(suite.Context, cardID); err != nil {
			testContext.Fatal(err)
		}
		card, err := client.GetCardByID(suite.Context, cardID)
		if err != nil {
			testContext.Fatal(err)
		}
		if actual, expected := card.Status, CardStatusEnumActive; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
	})
	testContext.Run("GetVirtualAccount", func(testContext *testing.T) {
		if actual, expected := client.GetVirtualAccount(), suite.Config.VirtualAccount; actual != expected {
			testContext.Fatalf("got %v, want %v", actual, expected)
		}
	})
}
