package payndapay

import (
	"net/http"
	"testing"

	"sdk/internal/contract"

	"github.com/shopspring/decimal"
)

func TestSDKProtocol(testContext *testing.T) {
	suite := contract.New(testContext, "paynda")
	client := New(suite.Config)
	testContext.Run("empty-cards", func(testContext *testing.T) {
		cards := suite.Invoke(testContext, client, contract.Call{Name: "GetCards"}).([]*Card)
		if len(cards) != 0 {
			testContext.Fatal("fresh account has cards")
		}
	})
	holder := suite.Invoke(testContext, client, contract.Call{
		Name:     "CreateCardHolder",
		Required: []string{"id", "createTime"},
	}).(*Cardholder)
	suite.Defaults["CardholderID"] = holder.ID
	bins := suite.Invoke(testContext, client, contract.Call{
		Name:    "GetCardBin",
		Strings: []string{"INDEPENDENT"},
	}).([]*CardBin)
	if len(bins) == 0 {
		testContext.Fatal("missing card bins")
	}
	suite.Defaults["CardBinID"] = bins[0].ID
	suite.Defaults["CardBin"] = bins[0].CardBin
	createID := contract.Unique()
	card := suite.Invoke(testContext, client, contract.Call{
		Name:     "CreateCard",
		Fields:   map[string]any{"RequestID": createID},
		Required: []string{"card.id", "card.createTime", "sensitiveInfo.cardNo", "balance.amount"},
	}).(*CardDetail)
	suite.Defaults["CardID"] = card.Card.ID
	transaction := suite.UI(testContext, http.MethodPost, "/simulate/refunds", map[string]any{
		"account_id":             suite.Config.Account,
		"card_id":                card.Card.ID,
		"amount":                 2,
		"currency":               "USD",
		"merchant_name":          "Protocol merchant",
		"merchant_country":       "US",
		"merchant_category_code": "5411",
	})
	transactionID := contract.Text(testContext, transaction, "id")
	transferID := contract.Unique()
	freezeID := contract.Unique()
	calls := []contract.Call{
		{Name: "GetMerchantWallet"},
		{
			Name:     "CreateBalanceAccount",
			Strings:  []string{"SDK account", contract.Unique()},
			Required: []string{"id", "createTime"},
		},
		{
			Name:    "UpdateBalanceAccount",
			Strings: []string{"SDK renamed", contract.Unique()},
		},
		{
			Name:     "GetBalanceAccount",
			Required: []string{"id", "name", "createTime"},
			Expect:   map[string]any{"id": suite.Config.Account},
		},
		{Name: "GetBalanceAccounts"},
		{Name: "GetBalanceAccountWallets"},
		{
			Name:   "BalanceAccountWalletTransfer",
			Fields: map[string]any{"Type": "IN"},
		},
		{
			Name:    "GetCardholder",
			Strings: []string{holder.ID},
		},
		{Name: "UpdateCardholder"},
		{Name: "GetCardholders"},
		{
			Name:    "GetCardholderWallet",
			Strings: []string{holder.ID},
		},
		{
			Name:   "UpdateCardholderWallet",
			Fields: map[string]any{"Type": "INC"},
		},
		{
			Name:     "GetCard",
			Strings:  []string{card.Card.ID},
			Required: []string{"id", "createTime", "status"},
		},
		{
			Name:    "GetCardByID",
			Strings: []string{card.Card.ID},
		},
		{
			Name:     "GetCardBalance",
			Strings:  []string{card.Card.ID},
			Required: []string{"amount", "availableAmount"},
		},
		{Name: "GetCards"},
		{
			Name:     "ListCardByPaginate",
			Required: []string{"records.0.id", "total", "size", "current"},
		},
		{
			Name:     "GetCardSensitive",
			Strings:  []string{card.Card.ID},
			Required: []string{"cardNo", "cvv", "expirationDate"},
		},
		{
			Name:    "FrozenCard",
			Strings: []string{card.Card.ID},
		},
		{
			Name:    "UnfrozenCard",
			Strings: []string{card.Card.ID},
		},
		{
			Name:   "CardFrozen",
			Fields: map[string]any{"RequestID": freezeID},
		},
		{
			Name:    "QueryCardStatusUpdate",
			Strings: []string{freezeID},
			Schema:  RequestResult{},
		},
		{Name: "CardUnfrozen"},
		{
			Name:    "GetCardControls",
			Strings: []string{card.Card.ID},
		},
		{Name: "UpdateCardControls"},
		{
			Name:   "UpdateCardBalance",
			Fields: map[string]any{"Type": "INC"},
		},
		{
			Name: "CardTransfer",
			Fields: map[string]any{
				"Type":      "IN",
				"RequestID": transferID,
			},
		},
		{Name: "QueryCardBalanceUpdate"},
		{
			Name:           "QueryRequestResult",
			Strings:        []string{createID},
			EmbeddedSchema: CardDetail{},
			Required:       []string{"id", "requestId", "createTime", "result"},
		},
		{
			Name:           "QueryCreateCardResult",
			Strings:        []string{createID},
			Schema:         RequestResult{},
			EmbeddedSchema: CardDetail{},
		},
		{
			Name:           "QueryTransfer",
			Strings:        []string{transferID},
			Schema:         RequestResult{},
			EmbeddedSchema: CardBalanceTransferRecord{},
		},
		{
			Name:           "QueryCardTransfer",
			Strings:        []string{transferID},
			Schema:         RequestResult{},
			EmbeddedSchema: CardBalanceTransferRecord{},
		},
		{
			Name:   "CardTransferIn",
			Fields: map[string]any{"Amount": decimal.NewFromInt(1)},
			Schema: CardBalanceTransfer{},
		},
		{
			Name:   "CardTransferOut",
			Fields: map[string]any{"Amount": decimal.NewFromInt(1)},
			Schema: CardBalanceTransfer{},
		},
		{
			Name:     "GetCardTransactions",
			Required: []string{"records.0.id", "records.0.transactionTime"},
		},
		{
			Name:     "QueryCardTransaction",
			Strings:  []string{transactionID},
			Required: []string{"id", "transactionTime"},
		},
		{
			Name:    "ValidCardBin",
			Strings: []string{bins[0].CardBin},
			Local:   true,
		},
		{
			Name:    "QueryRequestResult",
			Strings: []string{contract.Unique()},
			Error:   IsRequestIDNotFound,
		},
		{
			Name:    "ReleaseCard",
			Strings: []string{contract.Unique(), card.Card.ID},
		},
		{
			Name:    "DeleteCardholder",
			Strings: []string{holder.ID},
		},
		{
			Name:    "DeleteBalanceAccount",
			Strings: []string{contract.Unique()},
		},
	}
	for _, call := range calls {
		testContext.Run(call.Name, func(testContext *testing.T) { suite.Invoke(testContext, client, call) })
	}
	suite.Coverage(testContext, client)
}
