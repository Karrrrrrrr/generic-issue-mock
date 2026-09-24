package slash

import (
	"net/http"
	"testing"

	"sdk/internal/contract"
)

func TestSDKProtocol(testContext *testing.T) {
	suite := contract.New(testContext, "slash")
	virtual := suite.UI(testContext, http.MethodPost, "/managed-virtual-accounts", map[string]any{
		"account_id": suite.Config.Account,
		"name":       "Protocol virtual account",
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
	testContext.Run("empty-cards", func(testContext *testing.T) {
		cards := suite.Invoke(testContext, client, contract.Call{
			Name:     "ListCards",
			Required: []string{"items", "metadata.count"},
		}).(*ListCardsResponse)
		if len(cards.Items) != 0 || cards.Metadata.Count != 0 {
			testContext.Fatal("fresh account has cards")
		}
	})
	products := suite.Invoke(testContext, client, contract.Call{
		Name:     "ListCardProducts",
		Select:   "items",
		Required: []string{"items.0.id", "items.0.prefix", "metadata.count"},
	}).([]CardProduct)
	suite.Defaults["CardBinID"] = products[0].ID
	suite.Defaults["CardProductID"] = products[0].ID
	card := suite.Invoke(testContext, client, contract.Call{
		Name:     "CreateCard",
		Required: []string{"id", "accountId", "virtualAccountId", "createdAt", "expiryMonth", "expiryYear", "status"},
		Expect: map[string]any{
			"accountId":        suite.Config.Account,
			"virtualAccountId": suite.Config.VirtualAccount,
			"status":           "active",
		},
	}).(*Card)
	suite.Defaults["CardID"] = card.ID
	if len(card.ExpiryMonth) != 2 || len(card.ExpiryYear) != 4 {
		testContext.Fatal("invalid card expiry widths")
	}
	transaction := suite.UI(testContext, http.MethodPost, "/simulate/refunds", map[string]any{
		"card_id":                card.ID,
		"amount":                 2,
		"currency":               "USD",
		"merchant_name":          "Protocol merchant",
		"merchant_country":       "US",
		"merchant_category_code": "5411",
	})
	transactionID := contract.Text(testContext, transaction, "id")
	otherVirtual := suite.UI(testContext, http.MethodPost, "/managed-virtual-accounts", map[string]any{
		"account_id": suite.Config.Account,
		"name":       "Protocol destination",
		"currency":   "USD",
	})
	calls := []contract.Call{
		{
			Name:   "ListLegalEntity",
			Select: "items",
		},
		{
			Name:     "ListAccounts",
			Select:   "items",
			Required: []string{"items.0.id", "items.0.createdAt"},
		},
		{
			Name:     "GetAccount",
			Required: []string{"id", "createdAt", "balances"},
		},
		{
			Name:     "ListAccountBalances",
			Select:   "balances",
			Required: []string{"balances.0.timestamp", "balances.0.available.amountCents"},
		},
		{
			Name:     "CreateVirtualAccount",
			Strings:  []string{"SDK virtual"},
			Required: []string{"virtualAccount.id"},
		},
		{
			Name:    "UpdateVirtualAccount",
			Strings: []string{suite.Config.VirtualAccount, "Renamed SDK virtual"},
		},
		{
			Name:     "GetVirtualAccount",
			Required: []string{"virtualAccount.id", "balance.amountCents"},
		},
		{
			Name:     "ListVirtualAccounts",
			Select:   "items",
			Required: []string{"items.0.virtualAccount.id"},
		},
		{
			Name: "VirtualAccountTransfer",
			Fields: map[string]any{
				"Source":      suite.Config.VirtualAccount,
				"Destination": contract.Text(testContext, otherVirtual, "id"),
			},
			Required: []string{"transferId"},
		},
		{
			Name:     "GetCard",
			Strings:  []string{card.ID},
			Required: []string{"pan", "cvv", "createdAt"},
		},
		{
			Name:     "QueryCard",
			Strings:  []string{card.ID},
			Required: []string{"id", "createdAt"},
		},
		{
			Name:     "ListCards",
			Required: []string{"items.0.id", "metadata.count"},
		},
		{
			Name:    "FrozenCard",
			Strings: []string{card.ID},
			Expect:  map[string]any{"status": "paused"},
		},
		{
			Name:    "UnfrozenCard",
			Strings: []string{card.ID},
			Expect:  map[string]any{"status": "active"},
		},
		{
			Name:     "GetCardUtilization",
			Strings:  []string{card.ID},
			Required: []string{"spend.amountCents", "availableBalance.amountCents"},
		},
		{
			Name:    "UpdateSpendingConstraint",
			Strings: []string{card.ID},
		},
		{
			Name:    "SetSpendingConstraint",
			Strings: []string{card.ID},
		},
		{
			Name:    "GetCardModifiers",
			Strings: []string{card.ID},
			Select:  "modifiers",
		},
		{
			Name:    "SetCardModifiers",
			Strings: []string{card.ID},
		},
		{
			Name:    "ListCardGroups",
			Strings: []string{""},
			Select:  "items",
		},
		{
			Name:     "CreateCardGroup",
			Fields:   map[string]any{"VirtualAccountID": suite.Config.VirtualAccount},
			Required: []string{"id"},
		},
		{
			Name:    "GetCardGroup",
			Strings: []string{suite.Config.Account},
		},
		{
			Name:    "UpdateCardGroup",
			Strings: []string{suite.Config.Account},
		},
		{
			Name:    "UpdateCardGroupSpendingConstraint",
			Strings: []string{suite.Config.Account},
		},
		{
			Name:    "SetCardGroupSpendingConstraint",
			Strings: []string{suite.Config.Account},
		},
		{
			Name:    "GetCardGroupUtilization",
			Strings: []string{suite.Config.Account},
		},
		{
			Name:   "ListMerchants",
			Select: "items",
		},
		{
			Name:    "GetMerchant",
			Strings: []string{suite.Config.Account},
		},
		{
			Name:   "ListMerchantCategories",
			Select: "items",
		},
		{
			Name:     "CreateWebhook",
			Fields:   map[string]any{"URL": "http://127.0.0.1:1/callback"},
			Required: []string{"id"},
		},
		{
			Name:   "ListWebhooks",
			Select: "items",
		},
		{
			Name: "UpdateWebhook",
			Fields: map[string]any{
				"WebhookID": suite.Config.Account,
				"URL":       "http://127.0.0.1:1/callback",
			},
		},
		{
			Name: "SetAuthWebhook",
			Fields: map[string]any{
				"WebhookUrl": "http://127.0.0.1:1/authorize",
				"Status":     "disabled",
			},
		},
		{
			Name:     "GetAuthWebhook",
			Required: []string{"createdAt", "updatedAt"},
		},
		{
			Name:     "ListTransactions",
			Required: []string{"items.0.id", "items.0.date"},
		},
		{
			Name:   "GetTransactionAggregations",
			Fields: map[string]any{"AccountId": suite.Config.Account},
		},
		{
			Name:     "GetTransaction",
			Strings:  []string{transactionID},
			Required: []string{"id", "date", "amountCents"},
		},
		{
			Name:    "GetTransactionFeeDetails",
			Strings: []string{transactionID},
			Select:  "items",
		},
		{
			Name:    "ReleaseCard",
			Strings: []string{card.ID},
		},
	}
	for _, call := range calls {
		testContext.Run(call.Name, func(testContext *testing.T) { suite.Invoke(testContext, client, call) })
	}
	suite.Coverage(testContext, client)
	wrapped := NewSlashSDK(false, SlashSDKConf{
		Url:              suite.Config.URL,
		VaultUrl:         suite.Config.URL,
		ApiKey:           suite.Config.Account,
		AccountID:        suite.Config.Account,
		VirtualAccountID: suite.Config.VirtualAccount,
	})
	testContext.Run("interface", func(testContext *testing.T) {
		suite.Invoke(testContext, wrapped, contract.Call{
			Name:  "CardPostPre",
			Local: true,
		})
		createdID := suite.Invoke(testContext, wrapped, contract.Call{
			Name:     "CardPost",
			Schema:   Card{},
			Required: []string{"id"},
		}).(string)
		for _, name := range []string{"GetCardByID", "CardFrozen", "CardUnfrozen"} {
			suite.Invoke(testContext, wrapped, contract.Call{
				Name:    name,
				Strings: []string{createdID},
				Schema:  Card{},
			})
		}
		suite.Invoke(testContext, wrapped, contract.Call{
			Name:  "GetVirtualAccount",
			Local: true,
		})
		suite.Coverage(testContext, wrapped)
	})
}
