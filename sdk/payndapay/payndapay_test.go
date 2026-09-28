package payndapay

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"sdk/internal/contract"

	"github.com/shopspring/decimal"
)

type payndaFixture struct {
	suite  *contract.Suite
	client *PayndaPaySDK
	holder *Cardholder
	bin    *CardBin
}

func newPayndaFixture(testContext *testing.T) *payndaFixture {
	testContext.Helper()
	suite := contract.New(testContext, "paynda")
	client := New(suite.Config)
	holder, err := client.CreateCardHolder(suite.Context, &CardHolderRequest{
		Nonce:              contract.Unique(),
		FirstName:          "Protocol",
		LastName:           "Test",
		Email:              contract.Unique() + "@example.test",
		MobilePrefix:       "1",
		Mobile:             "2025550123",
		BillingCountryCode: "USA",
	})
	if err != nil {
		testContext.Fatal(err)
	}
	if holder == nil || holder.ID == "" {
		testContext.Fatal("missing cardholder")
	}
	bins, err := client.GetCardBin(suite.Context, CreditLimitType_INDEPENDENT)
	if err != nil {
		testContext.Fatal(err)
	}
	if len(bins) == 0 {
		testContext.Fatal("missing card BINs")
	}
	return &payndaFixture{
		suite:  suite,
		client: client,
		holder: holder,
		bin:    bins[0],
	}
}

func (fixture *payndaFixture) cardRequest() *CardCreateRequest {
	return &CardCreateRequest{
		Nonce:          contract.Unique(),
		RequestID:      contract.Unique(),
		CardholderID:   fixture.holder.ID,
		CardBinID:      fixture.bin.ID,
		Currency:       "USD",
		Amount:         "10",
		ExpirationDate: time.Now().AddDate(2, 0, 0).Format("01/06"),
	}
}

func (fixture *payndaFixture) createCard(testContext *testing.T) (*CardDetail, string) {
	testContext.Helper()
	request := fixture.cardRequest()
	card, err := fixture.client.CreateCard(fixture.suite.Context, request)
	if err != nil {
		testContext.Fatal(err)
	}
	if card == nil || card.Card.ID == "" {
		testContext.Fatal("missing created card")
	}
	return card, request.RequestID
}

func TestPayndaAccounts(testContext *testing.T) {
	fixture := newPayndaFixture(testContext)
	suite := fixture.suite
	client := fixture.client
	testContext.Run("GetMerchantWallet", func(testContext *testing.T) {
		items, err := client.GetMerchantWallet(suite.Context)
		if err != nil {
			testContext.Fatal(err)
		}
		if len(items) == 0 || items[0].ID == "" || items[0].Currency != "USD" {
			testContext.Fatalf("invalid wallets: %+v", items)
		}
		if _, err := decimal.NewFromString(items[0].Amount); err != nil {
			testContext.Fatal(err)
		}
	})
	testContext.Run("CreateBalanceAccount", func(testContext *testing.T) {
		account, err := client.CreateBalanceAccount(suite.Context, "Created account", contract.Unique())
		if err != nil {
			testContext.Fatal(err)
		}
		if account == nil || account.ID == "" || account.Name != "Created account" {
			testContext.Fatalf("invalid account: %+v", account)
		}
		if _, err := time.Parse(time.DateTime, account.CreateTime); err != nil {
			testContext.Fatal(err)
		}
	})
	testContext.Run("UpdateBalanceAccount", func(testContext *testing.T) {
		if err := client.UpdateBalanceAccount(suite.Context, "Updated account", contract.Unique()); err != nil {
			testContext.Fatal(err)
		}
	})
	testContext.Run("GetBalanceAccount", func(testContext *testing.T) {
		account, err := client.GetBalanceAccount(suite.Context)
		if err != nil {
			testContext.Fatal(err)
		}
		if account == nil || account.ID != suite.Config.Account || account.Name != "Updated account" {
			testContext.Fatalf("unexpected account: %+v", account)
		}
	})
	testContext.Run("GetBalanceAccounts", func(testContext *testing.T) {
		items, err := client.GetBalanceAccounts(suite.Context)
		if err != nil {
			testContext.Fatal(err)
		}
		if len(items) != 1 || items[0].ID != suite.Config.Account {
			testContext.Fatalf("account isolation failed: %+v", items)
		}
	})
	testContext.Run("GetBalanceAccountWallets", func(testContext *testing.T) {
		wallets, err := client.GetBalanceAccountWallets(suite.Context)
		if err != nil {
			testContext.Fatal(err)
		}
		if len(wallets) == 0 || wallets[0].BalanceAccountId != suite.Config.Account {
			testContext.Fatalf("unexpected wallets: %+v", wallets)
		}
	})
	testContext.Run("BalanceAccountWalletTransfer", func(testContext *testing.T) {
		result, err := client.BalanceAccountWalletTransfer(suite.Context, &BalanceAccountWalletTransferRequest{
			Nonce:            contract.Unique(),
			RequestID:        contract.Unique(),
			BalanceAccountID: suite.Config.Account,
			Type:             TransferType_IN,
			Currency:         "USD",
			Amount:           "10",
		})
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.BalanceAccountID != suite.Config.Account || result.Amount != "10" {
			testContext.Fatalf("invalid transfer: %+v", result)
		}
	})
	testContext.Run("DeleteBalanceAccount", func(testContext *testing.T) {
		if err := client.DeleteBalanceAccount(suite.Context, contract.Unique()); err != nil {
			testContext.Fatal(err)
		}
	})
}

func TestPayndaCardholders(testContext *testing.T) {
	fixture := newPayndaFixture(testContext)
	suite := fixture.suite
	client := fixture.client
	testContext.Run("CreateCardHolder", func(testContext *testing.T) {
		holder, err := client.CreateCardHolder(suite.Context, &CardHolderRequest{
			Nonce:              contract.Unique(),
			FirstName:          "Second",
			LastName:           "Holder",
			Email:              contract.Unique() + "@example.test",
			BillingCountryCode: "USA",
		})
		if err != nil {
			testContext.Fatal(err)
		}
		if holder == nil || holder.ID == fixture.holder.ID || holder.FirstName != "Second" {
			testContext.Fatalf("invalid cardholder: %+v", holder)
		}
	})
	testContext.Run("GetCardholder", func(testContext *testing.T) {
		holder, err := client.GetCardholder(suite.Context, fixture.holder.ID)
		if err != nil {
			testContext.Fatal(err)
		}
		if holder == nil || holder.ID != fixture.holder.ID {
			testContext.Fatalf("wrong holder: %+v", holder)
		}
	})
	testContext.Run("UpdateCardholder", func(testContext *testing.T) {
		if err := client.UpdateCardholder(suite.Context, &CardHolderRequest{
			Nonce:              contract.Unique(),
			CardholderID:       fixture.holder.ID,
			FirstName:          "Updated",
			LastName:           "Holder",
			BillingCountryCode: "USA",
		}); err != nil {
			testContext.Fatal(err)
		}
	})
	testContext.Run("GetCardholders", func(testContext *testing.T) {
		holders, err := client.GetCardholders(suite.Context)
		if err != nil {
			testContext.Fatal(err)
		}
		if len(holders) != 2 {
			testContext.Fatalf("got %d holders, want 2", len(holders))
		}
	})
	testContext.Run("GetCardholderWallet", func(testContext *testing.T) {
		wallets, err := client.GetCardholderWallet(suite.Context, fixture.holder.ID)
		if err != nil {
			testContext.Fatal(err)
		}
		if len(wallets) == 0 || wallets[0].CardholderID != fixture.holder.ID {
			testContext.Fatalf("invalid holder wallets: %+v", wallets)
		}
	})
	testContext.Run("UpdateCardholderWallet", func(testContext *testing.T) {
		wallet, err := client.UpdateCardholderWallet(suite.Context, &CardHolderWalletRequest{
			Nonce:        contract.Unique(),
			CardholderID: fixture.holder.ID,
			Type:         BalanceOperationType_INC,
			Currency:     "USD",
			Amount:       "1",
		})
		if err != nil {
			testContext.Fatal(err)
		}
		if wallet == nil || wallet.CardholderID != fixture.holder.ID {
			testContext.Fatalf("wrong wallet: %+v", wallet)
		}
	})
	testContext.Run("DeleteCardholder", func(testContext *testing.T) {
		if err := client.DeleteCardholder(suite.Context, fixture.holder.ID); err != nil {
			testContext.Fatal(err)
		}
	})
}

func TestPayndaCreateCard(testContext *testing.T) {
	fixture := newPayndaFixture(testContext)
	testContext.Run("valid", func(testContext *testing.T) {
		card, err := fixture.client.CreateCard(fixture.suite.Context, fixture.cardRequest())
		if err != nil {
			testContext.Fatal(err)
		}
		if card == nil || card.Card.CardholderId != fixture.holder.ID || card.Card.BalanceAccountId != fixture.suite.Config.Account {
			testContext.Fatalf("invalid created card: %+v", card)
		}
		if id, err := strconv.ParseInt(card.Card.ID, 10, 64); err != nil || id <= 0 {
			testContext.Fatalf("invalid card ID %q", card.Card.ID)
		}
		if _, err := time.Parse(time.DateTime, card.Card.CreateTime); err != nil {
			testContext.Fatal(err)
		}
		if _, err := time.Parse("01/06", card.CardSensitive.ExpirationDate); err != nil {
			testContext.Fatal(err)
		}
		if len(card.CardSensitive.CVV) != 3 || card.CardSensitive.CardNo == "" || card.Card.Status != "ACTIVE" {
			testContext.Fatalf("invalid sensitive/status data: %+v", card)
		}
	})
	testContext.Run("missing-request-id", func(testContext *testing.T) {
		request := fixture.cardRequest()
		request.RequestID = ""
		if _, err := fixture.client.CreateCard(fixture.suite.Context, request); err != ErrEmptyRequestID {
			testContext.Fatalf("got %v, want ErrEmptyRequestID", err)
		}
	})
	testContext.Run("zero-amount", func(testContext *testing.T) {
		request := fixture.cardRequest()
		request.Amount = "0"
		if _, err := fixture.client.CreateCard(fixture.suite.Context, request); err == nil {
			testContext.Fatal("zero amount accepted")
		}
	})
	testContext.Run("invalid-holder-id", func(testContext *testing.T) {
		request := fixture.cardRequest()
		request.CardholderID = "not-an-id"
		if _, err := fixture.client.CreateCard(fixture.suite.Context, request); err == nil {
			testContext.Fatal("invalid holder accepted")
		}
	})
}

func TestPayndaCardQueries(testContext *testing.T) {
	fixture := newPayndaFixture(testContext)
	suite := fixture.suite
	client := fixture.client
	testContext.Run("GetCards/empty", func(testContext *testing.T) {
		cards, err := client.GetCards(suite.Context)
		if err != nil {
			testContext.Fatal(err)
		}
		if cards == nil || len(cards) != 0 {
			testContext.Fatalf("want empty array, got %+v", cards)
		}
	})
	card, _ := fixture.createCard(testContext)
	fixture.createCard(testContext)
	testContext.Run("GetCardBin", func(testContext *testing.T) {
		bins, err := client.GetCardBin(suite.Context, CreditLimitType_INDEPENDENT)
		if err != nil {
			testContext.Fatal(err)
		}
		if len(bins) == 0 || bins[0].CardBin == "" {
			testContext.Fatal("missing BIN")
		}
	})
	testContext.Run("ValidCardBin", func(testContext *testing.T) {
		if !client.ValidCardBin(suite.Context, fixture.bin.CardBin) {
			testContext.Fatal("enabled BIN rejected")
		}
	})
	testContext.Run("GetCard", func(testContext *testing.T) {
		result, err := client.GetCard(suite.Context, card.Card.ID)
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.ID != card.Card.ID || result.Currency != "USD" {
			testContext.Fatalf("wrong card: %+v", result)
		}
	})
	testContext.Run("GetCardByID", func(testContext *testing.T) {
		result, err := client.GetCardByID(suite.Context, card.Card.ID)
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.ID != card.Card.ID {
			testContext.Fatalf("wrong card: %+v", result)
		}
	})
	testContext.Run("GetCardBalance", func(testContext *testing.T) {
		result, err := client.GetCardBalance(suite.Context, card.Card.ID)
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil {
			testContext.Fatal("missing balance")
		}
		if _, err := decimal.NewFromString(result.Amount); err != nil {
			testContext.Fatal(err)
		}
	})
	testContext.Run("GetCardSensitive", func(testContext *testing.T) {
		result, err := client.GetCardSensitive(suite.Context, card.Card.ID)
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.CardID != card.Card.ID || result.CardNo != card.CardSensitive.CardNo || len(result.CVV) != 3 {
			testContext.Fatalf("invalid card sensitive info: %+v", result)
		}
	})
	testContext.Run("GetCards/populated", func(testContext *testing.T) {
		cards, err := client.GetCards(suite.Context)
		if err != nil {
			testContext.Fatal(err)
		}
		if len(cards) != 2 {
			testContext.Fatalf("got %d cards, want 2", len(cards))
		}
	})
	testContext.Run("ListCardByPaginate/first-page", func(testContext *testing.T) {
		page, err := client.ListCardByPaginate(suite.Context, &CardPaginateRequest{
			Current:  1,
			PageSize: 1,
		})
		if err != nil {
			testContext.Fatal(err)
		}
		if page == nil || page.Total < 1 || page.Current != 1 || page.Size != 1 || len(page.Records) != 1 {
			testContext.Fatalf("invalid first page: %+v", page)
		}
	})
	testContext.Run("ListCardByPaginate/empty-page", func(testContext *testing.T) {
		page, err := client.ListCardByPaginate(suite.Context, &CardPaginateRequest{
			Current:  3,
			PageSize: 1,
		})
		if err != nil {
			testContext.Fatal(err)
		}
		if page == nil || page.Total < 0 || page.Current != 3 || page.Size != 1 || page.Records == nil || len(page.Records) != 0 {
			testContext.Fatalf("invalid empty page: %+v", page)
		}
	})
	testContext.Run("GetCard/invalid-id", func(testContext *testing.T) {
		if _, err := client.GetCard(suite.Context, "bad-id"); err == nil {
			testContext.Fatal("invalid ID accepted")
		}
	})
	testContext.Run("GetCard/other-account", func(testContext *testing.T) {
		other := newPayndaFixture(testContext)
		if _, err := other.client.GetCard(other.suite.Context, card.Card.ID); err == nil {
			testContext.Fatal("cross-account card access accepted")
		}
	})
	testContext.Run("GetCardControls", func(testContext *testing.T) {
		controls, err := client.GetCardControls(suite.Context, card.Card.ID)
		if err != nil {
			testContext.Fatal(err)
		}
		if controls == nil {
			testContext.Fatal("controls must be an array")
		}
	})
	testContext.Run("UpdateCardControls", func(testContext *testing.T) {
		if err := client.UpdateCardControls(suite.Context, &CardControlRequest{
			Nonce:            contract.Unique(),
			CardID:           card.Card.ID,
			Period:           CardControlPeriod_DAY,
			TransactionCount: 5,
			Amount:           "50",
		}); err != nil {
			testContext.Fatal(err)
		}
	})
}

func TestPayndaCardStatus(testContext *testing.T) {
	fixture := newPayndaFixture(testContext)
	suite := fixture.suite
	client := fixture.client
	card, _ := fixture.createCard(testContext)
	testContext.Run("FrozenCard", func(testContext *testing.T) {
		if err := client.FrozenCard(suite.Context, card.Card.ID); err != nil {
			testContext.Fatal(err)
		}
		result, err := client.GetCard(suite.Context, card.Card.ID)
		if err != nil || result == nil || result.Status != "FROZEN" {
			testContext.Fatalf("freeze not visible: %+v %v", result, err)
		}
	})
	testContext.Run("UnfrozenCard", func(testContext *testing.T) {
		if err := client.UnfrozenCard(suite.Context, card.Card.ID); err != nil {
			testContext.Fatal(err)
		}
		result, err := client.GetCard(suite.Context, card.Card.ID)
		if err != nil || result == nil || result.Status != "ACTIVE" {
			testContext.Fatalf("unfreeze not visible: %+v %v", result, err)
		}
	})
	freezeID := contract.Unique()
	testContext.Run("CardFrozen", func(testContext *testing.T) {
		if err := client.CardFrozen(suite.Context, &CardStatusUpdateRequest{
			CardID:    card.Card.ID,
			RequestID: freezeID,
		}); err != nil {
			testContext.Fatal(err)
		}
	})
	testContext.Run("QueryCardStatusUpdate", func(testContext *testing.T) {
		result, err := client.QueryCardStatusUpdate(suite.Context, freezeID)
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || !result.Success || result.Record == nil || result.Record.ID != card.Card.ID {
			testContext.Fatalf("invalid status result: %+v", result)
		}
	})
	testContext.Run("CardUnfrozen", func(testContext *testing.T) {
		if err := client.CardUnfrozen(suite.Context, &CardStatusUpdateRequest{
			CardID:    card.Card.ID,
			RequestID: contract.Unique(),
		}); err != nil {
			testContext.Fatal(err)
		}
	})
	testContext.Run("ReleaseCard", func(testContext *testing.T) {
		if err := client.ReleaseCard(suite.Context, contract.Unique(), card.Card.ID); err != nil {
			testContext.Fatal(err)
		}
		result, err := client.GetCard(suite.Context, card.Card.ID)
		if err != nil || result == nil || result.Status != "DELETED" {
			testContext.Fatalf("release not visible: %+v %v", result, err)
		}
	})
	testContext.Run("CardFrozen/missing-request-id", func(testContext *testing.T) {
		if err := client.CardFrozen(suite.Context, &CardStatusUpdateRequest{CardID: card.Card.ID}); err != ErrEmptyRequestID {
			testContext.Fatalf("got %v, want ErrEmptyRequestID", err)
		}
	})
}

func TestPayndaCardFundingAndRequestResults(testContext *testing.T) {
	fixture := newPayndaFixture(testContext)
	suite := fixture.suite
	client := fixture.client
	card, createID := fixture.createCard(testContext)
	testContext.Run("QueryRequestResult", func(testContext *testing.T) {
		result, err := client.QueryRequestResult(suite.Context, createID)
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.RequestID != createID || result.Result == "" {
			testContext.Fatalf("invalid request result: %+v", result)
		}
		var envelope struct {
			Code    int64      `json:"code"`
			Success bool       `json:"success"`
			Data    CardDetail `json:"data"`
		}
		if err := json.Unmarshal([]byte(result.Result), &envelope); err != nil {
			testContext.Fatal(err)
		}
		if envelope.Code != 200 || !envelope.Success || envelope.Data.Card.ID != card.Card.ID {
			testContext.Fatalf("invalid embedded JSON envelope: %+v", envelope)
		}
	})
	testContext.Run("QueryCreateCardResult", func(testContext *testing.T) {
		result, err := client.QueryCreateCardResult(suite.Context, createID)
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || !result.Success || result.Result == nil || result.Result.Card.ID != card.Card.ID {
			testContext.Fatalf("invalid create result: %+v", result)
		}
	})
	testContext.Run("QueryRequestResult/missing", func(testContext *testing.T) {
		if _, err := client.QueryRequestResult(suite.Context, contract.Unique()); !IsRequestIDNotFound(err) {
			testContext.Fatalf("want missing-request error, got %v", err)
		}
	})
	testContext.Run("UpdateCardBalance", func(testContext *testing.T) {
		result, err := client.UpdateCardBalance(suite.Context, &CardBalanceUpdateRequest{
			Nonce:     contract.Unique(),
			RequestID: contract.Unique(),
			CardID:    card.Card.ID,
			Amount:    "2",
		})
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil {
			testContext.Fatal("missing card balance")
		}
		if _, err := decimal.NewFromString(result.Amount); err != nil {
			testContext.Fatal(err)
		}
	})
	transferID := contract.Unique()
	testContext.Run("CardTransfer/in", func(testContext *testing.T) {
		result, err := client.CardTransfer(suite.Context, &CardBalanceTransferRequest{
			Nonce:     contract.Unique(),
			RequestID: transferID,
			CardID:    card.Card.ID,
			Amount:    "3",
			Type:      TransferType_IN,
		})
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.CardID != card.Card.ID || result.Amount != "3" || result.Type != "IN" {
			testContext.Fatalf("invalid transfer result: %+v", result)
		}
	})
	testContext.Run("CardTransfer/out", func(testContext *testing.T) {
		result, err := client.CardTransfer(suite.Context, &CardBalanceTransferRequest{
			Nonce:     contract.Unique(),
			RequestID: contract.Unique(),
			CardID:    card.Card.ID,
			Amount:    "1",
			Type:      TransferType_OUT,
		})
		if err != nil || result == nil || result.Type != "OUT" {
			testContext.Fatalf("out transfer failed: %+v %v", result, err)
		}
	})
	testContext.Run("CardTransfer/zero-amount", func(testContext *testing.T) {
		if _, err := client.CardTransfer(suite.Context, &CardBalanceTransferRequest{
			Nonce:     contract.Unique(),
			RequestID: contract.Unique(),
			CardID:    card.Card.ID,
			Amount:    "0",
			Type:      TransferType_IN,
		}); err == nil {
			testContext.Fatal("zero transfer accepted")
		}
	})
	testContext.Run("QueryTransfer", func(testContext *testing.T) {
		result, err := client.QueryTransfer(suite.Context, transferID)
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || !result.Success || result.TransferRecord == nil || result.TransferRecord.CardID != card.Card.ID {
			testContext.Fatalf("invalid transfer record: %+v", result)
		}
	})
	testContext.Run("QueryCardTransfer", func(testContext *testing.T) {
		result, err := client.QueryCardTransfer(suite.Context, transferID)
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.Status != CardTransferStatusSuccess || !result.Amount.Equal(decimal.NewFromInt(3)) || result.TransactionTime.IsZero() {
			testContext.Fatalf("invalid mapped transfer: %+v", result)
		}
	})
	testContext.Run("QueryCardBalanceUpdate", func(testContext *testing.T) {
		items, err := client.QueryCardBalanceUpdate(suite.Context)
		if err != nil {
			testContext.Fatal(err)
		}
		if len(items) == 0 {
			testContext.Fatal("missing balance history")
		}
	})
	testContext.Run("CardTransferIn", func(testContext *testing.T) {
		result, err := client.CardTransferIn(suite.Context, &CardTransferRequest{
			CardID:    card.Card.ID,
			CardBin:   fixture.bin.CardBin,
			RequestID: contract.Unique(),
			Amount:    decimal.NewFromInt(2),
		})
		if err != nil || result == nil || result.NeedRetry {
			testContext.Fatalf("transfer in failed: %+v %v", result, err)
		}
	})
	testContext.Run("CardTransferOut", func(testContext *testing.T) {
		result, err := client.CardTransferOut(suite.Context, &CardTransferRequest{
			CardID:    card.Card.ID,
			CardBin:   fixture.bin.CardBin,
			RequestID: contract.Unique(),
			Amount:    decimal.NewFromInt(1),
		})
		if err != nil || result == nil || result.NeedRetry {
			testContext.Fatalf("transfer out failed: %+v %v", result, err)
		}
	})
}

func TestPayndaTransactions(testContext *testing.T) {
	fixture := newPayndaFixture(testContext)
	suite := fixture.suite
	client := fixture.client
	card, _ := fixture.createCard(testContext)
	testContext.Run("GetCardTransactions/empty", func(testContext *testing.T) {
		result, err := client.GetCardTransactions(suite.Context, &CardTransactionsRequest{
			Current:  1,
			PageSize: 10,
			CardID:   card.Card.ID,
		})
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.Total != 0 || result.Records == nil || len(result.Records) != 0 {
			testContext.Fatalf("want empty transaction page, got %+v", result)
		}
	})
	first := suite.UI(testContext, http.MethodPost, "/simulate/refunds", map[string]any{
		"card_id":                card.Card.ID,
		"amount":                 2,
		"currency":               "USD",
		"merchant_name":          "First refund",
		"merchant_country":       "US",
		"merchant_category_code": "5411",
	})
	transactionID := contract.Text(testContext, first, "id")
	suite.UI(testContext, http.MethodPost, "/simulate/refunds", map[string]any{
		"card_id":                card.Card.ID,
		"amount":                 3,
		"currency":               "USD",
		"merchant_name":          "Second refund",
		"merchant_country":       "US",
		"merchant_category_code": "5411",
	})
	testContext.Run("GetCardTransactions/populated", func(testContext *testing.T) {
		result, err := client.GetCardTransactions(suite.Context, &CardTransactionsRequest{
			Current:  1,
			PageSize: 10,
			CardID:   card.Card.ID,
		})
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.Total != 2 || len(result.Records) != 2 {
			testContext.Fatalf("invalid transaction page: %+v", result)
		}
		for _, transaction := range result.Records {
			if transaction.CardID.String() != card.Card.ID || transaction.ID == "" {
				testContext.Fatalf("wrong transaction: %+v", transaction)
			}
			if _, err := time.Parse(time.DateTime, transaction.TransactionTime); err != nil {
				testContext.Fatal(err)
			}
		}
	})
	testContext.Run("GetCardTransactions/pagination", func(testContext *testing.T) {
		result, err := client.GetCardTransactions(suite.Context, &CardTransactionsRequest{
			Current:  2,
			PageSize: 1,
			CardID:   card.Card.ID,
		})
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.Total < 1 || result.Current != 2 || result.Size != 1 || len(result.Records) != 1 || result.Records[0].ID != transactionID {
			testContext.Fatalf("invalid second page: %+v", result)
		}
	})
	testContext.Run("GetCardTransactions/date-range", func(testContext *testing.T) {
		result, err := client.GetCardTransactions(suite.Context, &CardTransactionsRequest{
			Current:              1,
			PageSize:             10,
			CardID:               card.Card.ID,
			TransactionTimeStart: "2000-01-01 00:00:00",
			TransactionTimeEnd:   "2000-01-02 00:00:00",
		})
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.Total != 0 || len(result.Records) != 0 {
			testContext.Fatalf("date filter ignored: %+v", result)
		}
	})
	testContext.Run("GetCardTransactions/invalid-date", func(testContext *testing.T) {
		if _, err := client.GetCardTransactions(suite.Context, &CardTransactionsRequest{
			Current:              1,
			PageSize:             10,
			TransactionTimeStart: "not-a-date",
		}); err == nil {
			testContext.Fatal("invalid date accepted")
		}
	})
	testContext.Run("QueryCardTransaction/refund", func(testContext *testing.T) {
		result, err := client.QueryCardTransaction(suite.Context, transactionID)
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.ID != transactionID || result.CardID.String() != card.Card.ID {
			testContext.Fatalf("wrong transaction: %+v", result)
		}
		if _, err := time.Parse(time.DateTime, result.CreateTime); err != nil {
			testContext.Fatal(err)
		}
	})
	testContext.Run("QueryCardTransaction/invalid-id", func(testContext *testing.T) {
		if _, err := client.QueryCardTransaction(suite.Context, "bad-id"); err == nil {
			testContext.Fatal("invalid transaction ID accepted")
		}
	})
	testContext.Run("QueryCardTransaction/other-account", func(testContext *testing.T) {
		other := newPayndaFixture(testContext)
		if _, err := other.client.QueryCardTransaction(other.suite.Context, transactionID); err == nil {
			testContext.Fatal("cross-account transaction accepted")
		}
	})
}

func TestPayndaAuthorizationTransactions(testContext *testing.T) {
	fixture := newPayndaFixture(testContext)
	suite := fixture.suite
	card, _ := fixture.createCard(testContext)
	authorization := suite.UI(testContext, http.MethodPost, "/simulate/authorizations", map[string]any{
		"card_id":                card.Card.ID,
		"transaction_amount":     "2",
		"transaction_currency":   "USD",
		"merchant_name":          "Authorization merchant",
		"merchant_country":       "US",
		"merchant_category_code": "5411",
	})
	authorizationID := contract.Text(testContext, authorization, "authorization.id")
	transactionID := contract.Text(testContext, authorization, "transaction.id")
	testContext.Run("QueryCardTransaction/authorization", func(testContext *testing.T) {
		suite.Responses()
		result, err := fixture.client.QueryCardTransaction(suite.Context, transactionID)
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.ID != transactionID || result.CardID.String() != card.Card.ID || result.Type != "transaction.authentication.approved" {
			testContext.Fatalf("invalid authorization transaction: %+v", result)
		}
		if _, err := time.Parse(time.DateTime, result.AuthorizationTime); err != nil {
			testContext.Fatal(err)
		}
		responses := suite.Responses()
		if len(responses) != 1 || responses[0].Status != http.StatusOK || !strings.HasPrefix(responses[0].ContentType, "application/json") {
			testContext.Fatalf("invalid HTTP responses: %+v", responses)
		}
		var wire struct {
			Code    int64 `json:"code"`
			Success bool  `json:"success"`
			Data    struct {
				ID                string `json:"id"`
				CardID            string `json:"cardId"`
				PreAuthAmount     string `json:"preAuthAmount"`
				AuthorizationTime string `json:"authorizationTime"`
			} `json:"data"`
		}
		if err := json.Unmarshal(responses[0].Body, &wire); err != nil {
			testContext.Fatalf("invalid wire field types: %v", err)
		}
		if wire.Code != 200 || !wire.Success || wire.Data.ID != transactionID || wire.Data.CardID != card.Card.ID || wire.Data.PreAuthAmount == "" || wire.Data.AuthorizationTime != result.AuthorizationTime {
			testContext.Fatalf("invalid transaction envelope: %+v", wire)
		}
	})
	clearing := suite.UI(testContext, http.MethodPost, "/authorizations/"+authorizationID+"/clear", map[string]any{
		"amount": "3",
	})
	clearingID := contract.Text(testContext, clearing, "id")
	testContext.Run("QueryCardTransaction/over-clearing", func(testContext *testing.T) {
		result, err := fixture.client.QueryCardTransaction(suite.Context, clearingID)
		if err != nil {
			testContext.Fatal(err)
		}
		if result == nil || result.ID != clearingID || result.Type != "transaction.authentication.settled" || result.CardID.String() != card.Card.ID {
			testContext.Fatalf("invalid clearing transaction: %+v", result)
		}
		if _, err := decimal.NewFromString(result.PostedAmount); err != nil {
			testContext.Fatal(err)
		}
		if _, err := time.Parse(time.DateTime, result.AuthorizationTime); err != nil {
			testContext.Fatal(err)
		}
		if _, err := time.Parse(time.DateTime, result.TransactionTime); err != nil {
			testContext.Fatal(err)
		}
	})
}
