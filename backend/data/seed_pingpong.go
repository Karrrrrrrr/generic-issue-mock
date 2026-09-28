package data

import (
	"context"
	"time"

	"generic-mock/enums"
	"generic-mock/internal/query"
	"generic-mock/model"
	"generic-mock/pkg/cardnumber"
	"generic-mock/pkg/cardwallet"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func seedPingPongData(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tables := query.Use(tx)
		productTable := tables.CardProduct
		product, err := productTable.WithContext(ctx).
			Clauses(clause.Locking{
				Strength: "UPDATE",
				Table:    clause.Table{Name: clause.CurrentTable},
			}).
			Where(
				productTable.Channel.Eq(string(enums.Channel_PingPong)),
				productTable.Prefix.Eq(pingpongSeedCardProductPrefix),
			).
			First()
		if err != nil {
			return err
		}

		accountTable := tables.Account
		accountQuery := accountTable.WithContext(ctx).Where(
			accountTable.Channel.Eq(string(enums.Channel_PingPong)),
			accountTable.Name.Eq("pingpong Primary"),
		)
		accountCount, err := accountQuery.Count()
		if err != nil {
			return err
		}
		account := &model.Account{
			Channel: enums.Channel_PingPong,
			Name:    "pingpong Primary",
		}
		if accountCount == 0 {
			if err := accountTable.WithContext(ctx).Create(account); err != nil {
				return err
			}
		} else {
			account, err = accountQuery.First()
			if err != nil {
				return err
			}
		}
		if account.WalletID == 0 {
			wallet := &model.Wallet{
				AccountID: account.ID,
				Channel:   enums.Channel_PingPong,
				Available: decimal.NewFromInt(1_000_000),
				Type:      enums.WalletType_Account,
				Currency:  enums.Currency_USD,
			}
			if err := tables.Wallet.WithContext(ctx).Create(wallet); err != nil {
				return err
			}
			account.WalletID = wallet.ID
			if _, err := accountTable.WithContext(ctx).
				Where(
					accountTable.ID.Eq(account.ID),
					accountTable.Channel.Eq(string(enums.Channel_PingPong)),
				).
				UpdateSimple(accountTable.WalletID.Value(wallet.ID)); err != nil {
				return err
			}
		}

		holderTable := tables.CardHolder
		holderQuery := holderTable.WithContext(ctx).Where(
			holderTable.AccountID.Eq(account.ID),
			holderTable.Channel.Eq(string(enums.Channel_PingPong)),
			holderTable.Email.Eq("pingpong.demo@example.test"),
		)
		holderCount, err := holderQuery.Count()
		if err != nil {
			return err
		}
		holder := &model.CardHolder{
			AccountID:    account.ID,
			Channel:      enums.Channel_PingPong,
			FirstName:    "PingPong",
			LastName:     "Demo",
			Email:        "pingpong.demo@example.test",
			Status:       enums.CardHolderStatus_Normal,
			ReviewStatus: enums.CardHolderReviewStatus_Approved,
			Shared:       true,
		}
		if holderCount == 0 {
			if err := holderTable.WithContext(ctx).Create(holder); err != nil {
				return err
			}
		} else {
			holder, err = holderQuery.First()
			if err != nil {
				return err
			}
		}

		cardTable := tables.Card
		cardCount, err := cardTable.WithContext(ctx).
			Where(
				cardTable.AccountID.Eq(account.ID),
				cardTable.Channel.Eq(string(enums.Channel_PingPong)),
				cardTable.RequestID.Eq("seed-pingpong-card"),
			).
			Count()
		if err != nil {
			return err
		}
		if cardCount != 0 {
			return nil
		}

		virtualAccountTable := tables.VirtualAccount
		virtualAccountQuery := virtualAccountTable.WithContext(ctx).Where(
			virtualAccountTable.AccountID.Eq(account.ID),
			virtualAccountTable.Channel.Eq(string(enums.Channel_PingPong)),
			virtualAccountTable.Name.Eq("PingPong Primary"),
		)
		virtualAccountCount, err := virtualAccountQuery.Count()
		if err != nil {
			return err
		}
		virtualAccount := &model.VirtualAccount{
			AccountID: account.ID,
			Channel:   enums.Channel_PingPong,
			Name:      "PingPong Primary",
		}
		if virtualAccountCount == 0 {
			wallet := &model.Wallet{
				AccountID: account.ID,
				Channel:   enums.Channel_PingPong,
				Available: decimal.NewFromInt(1_000_000),
				Type:      enums.WalletType_VirtualAccount,
				Currency:  enums.Currency_USD,
			}
			if err := tables.Wallet.WithContext(ctx).Create(wallet); err != nil {
				return err
			}
			virtualAccount.WalletID = wallet.ID
			virtualAccount.Wallet = wallet
			if err := virtualAccountTable.WithContext(ctx).Create(virtualAccount); err != nil {
				return err
			}
		} else {
			virtualAccount, err = virtualAccountQuery.Preload(virtualAccountTable.Wallet).First()
			if err != nil {
				return err
			}
		}

		assignment, ok := cardwallet.Prepare(cardwallet.PrepareRequest{
			AccountID:      account.ID,
			Channel:        enums.Channel_PingPong,
			CardType:       enums.CardType_VirtualAccountSingle,
			Currency:       enums.Currency_USD,
			VirtualAccount: virtualAccount,
		})
		if !ok {
			return gorm.ErrInvalidData
		}
		product.NextCardNumber++
		generatedCard, ok := cardnumber.Generate(cardnumber.GenerateRequest{
			Channel:  enums.Channel_PingPong,
			Prefix:   product.Prefix,
			Sequence: product.NextCardNumber,
		})
		if !ok {
			return gorm.ErrInvalidData
		}
		if _, err := productTable.WithContext(ctx).
			Where(
				productTable.ID.Eq(product.ID),
				productTable.Channel.Eq(string(enums.Channel_PingPong)),
			).
			UpdateSimple(productTable.NextCardNumber.Value(product.NextCardNumber)); err != nil {
			return err
		}
		assignment.Wallet.Available = decimal.NewFromInt(1_000)
		if err := tables.Wallet.WithContext(ctx).Create(assignment.Wallet); err != nil {
			return err
		}
		return cardTable.WithContext(ctx).Create(&model.Card{
			AccountID:              account.ID,
			Channel:                enums.Channel_PingPong,
			CardProductID:          product.ID,
			CardHolderID:           holder.ID,
			VirtualAccountID:       assignment.VirtualAccountID,
			WalletID:               assignment.Wallet.ID,
			CardBin:                generatedCard.Bin,
			CardNumber:             generatedCard.Number,
			Cvv:                    "123",
			ExpireAt:               time.Now().UTC().AddDate(2, 0, 0),
			Status:                 enums.CardStatus_Active,
			FormType:               enums.CardFormType_Virtual,
			RequestID:              "seed-pingpong-card",
			LastOperationRequestID: "seed-pingpong-card",
			LastOperationType:      enums.OperationType_OpenCard,
			LastOperationStatus:    enums.OperationStatus_Succeed,
			CardCurrency:           enums.Currency_USD,
			CardScheme:             enums.CardScheme_Visa,
			CardType:               enums.CardType_VirtualAccountSingle,
		})
	})
}
