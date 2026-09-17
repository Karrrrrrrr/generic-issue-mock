package data

import (
	"context"
	"time"

	photon "generic-mock/channel/photonpay/enums"
	"generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/cardnumber"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const slashDefaultCardProductPrefix = "424242"
const payndaDefaultCardProductPrefix = "523456"

func SeedInitialData(ctx context.Context, db *gorm.DB) error {
	accounts, err := seedChannelAccounts(ctx, db)
	if err != nil {
		return err
	}
	items := []*model.CardProduct{
		{
			AccountID: accounts[enums.Channel_Slash].ID,
			Channel:   enums.Channel_Slash,
			Prefix:    slashDefaultCardProductPrefix,
			IsDefault: true,
		},
		{
			AccountID: accounts[enums.Channel_PhotonPay].ID,
			Channel:   enums.Channel_PhotonPay,
			Prefix:    photon.DefaultCardBin,
			IsDefault: true,
		},
		{
			AccountID: accounts[enums.Channel_Paynda].ID,
			Channel:   enums.Channel_Paynda,
			Prefix:    payndaDefaultCardProductPrefix,
			IsDefault: true,
		},
	}

	for _, item := range items {
		if err := db.WithContext(ctx).Where(&model.CardProduct{
			Channel: item.Channel,
			Prefix:  item.Prefix,
		}).FirstOrCreate(item).Error; err != nil {
			return err
		}
	}
	if err := seedChannelCards(ctx, db, accounts, items); err != nil {
		return err
	}

	var wallet model.Wallet
	result := db.WithContext(ctx).Where(&model.Wallet{
		AccountID: accounts[enums.Channel_Slash].ID,
		Channel:   enums.Channel_Slash,
		Type:      enums.WalletType_VirtualAccount,
		Currency:  enums.Currency_USD,
	}).First(&wallet)
	if result.Error == gorm.ErrRecordNotFound {
		wallet = model.Wallet{
			AccountID: accounts[enums.Channel_Slash].ID,
			Channel:   enums.Channel_Slash,
			Amount:    decimal.NewFromInt(1_000_000),
			Type:      enums.WalletType_VirtualAccount,
			Currency:  enums.Currency_USD,
		}
		if err := db.WithContext(ctx).Create(&wallet).Error; err != nil {
			return err
		}
	} else if result.Error != nil {
		return result.Error
	}

	var account model.VirtualAccount
	result = db.WithContext(ctx).Where(&model.VirtualAccount{
		AccountID: accounts[enums.Channel_Slash].ID,
		Channel:   enums.Channel_Slash,
		Name:      "Slash Primary",
	}).First(&account)
	if result.Error == gorm.ErrRecordNotFound {
		if err := db.WithContext(ctx).Create(&model.VirtualAccount{
			AccountID: accounts[enums.Channel_Slash].ID,
			Channel:   enums.Channel_Slash,
			WalletID:  wallet.ID,
			Name:      "Slash Primary",
		}).Error; err != nil {
			return err
		}
	} else if result.Error != nil {
		return result.Error
	}

	return nil
}

func seedChannelAccounts(ctx context.Context, db *gorm.DB) (map[enums.Channel]*model.Account, error) {
	accounts := make(map[enums.Channel]*model.Account, 3)
	for _, channel := range []enums.Channel{enums.Channel_Slash, enums.Channel_PhotonPay, enums.Channel_Paynda} {
		account := &model.Account{}
		result := db.WithContext(ctx).Where(&model.Account{Channel: channel, Name: string(channel) + " Primary"}).First(account)
		if result.Error == gorm.ErrRecordNotFound {
			account = &model.Account{Channel: channel, Name: string(channel) + " Primary"}
			if err := db.WithContext(ctx).Create(account).Error; err != nil {
				return nil, err
			}
		} else if result.Error != nil {
			return nil, result.Error
		}

		if account.WalletID == 0 {
			wallet := &model.Wallet{
				AccountID: account.ID,
				Channel:   channel,
				Amount:    decimal.NewFromInt(1_000_000),
				Type:      enums.WalletType_Account,
				Currency:  enums.Currency_USD,
			}
			if err := db.WithContext(ctx).Create(wallet).Error; err != nil {
				return nil, err
			}
			account.WalletID = wallet.ID
			if err := db.WithContext(ctx).Save(account).Error; err != nil {
				return nil, err
			}
		}
		accounts[channel] = account
	}

	return accounts, nil
}

type channelCardSeed struct {
	channel    enums.Channel
	firstName  string
	lastName   string
	email      string
	cardScheme enums.CardScheme
	withWallet bool
}

func seedChannelCards(
	ctx context.Context,
	db *gorm.DB,
	accounts map[enums.Channel]*model.Account,
	products []*model.CardProduct,
) error {
	seeds := []channelCardSeed{
		{channel: enums.Channel_Slash, firstName: "Slash", lastName: "Demo", email: "slash.demo@example.test", cardScheme: enums.CardScheme_Visa},
		{channel: enums.Channel_PhotonPay, firstName: "PhotonPay", lastName: "Demo", email: "photonpay.demo@example.test", cardScheme: enums.CardScheme_Visa},
		{channel: enums.Channel_Paynda, firstName: "Paynda", lastName: "Demo", email: "paynda.demo@example.test", cardScheme: enums.CardScheme_MasterCard, withWallet: true},
	}

	for index, seed := range seeds {
		account := accounts[seed.channel]
		var holder model.CardHolder
		if err := db.WithContext(ctx).Where(&model.CardHolder{AccountID: account.ID, Channel: seed.channel, Email: seed.email}).FirstOrCreate(&holder, &model.CardHolder{
			AccountID:    account.ID,
			Channel:      seed.channel,
			FirstName:    seed.firstName,
			LastName:     seed.lastName,
			Email:        seed.email,
			Status:       enums.CardHolderStatus_Normal,
			ReviewStatus: enums.CardHolderReviewStatus_Approved,
			Shared:       true,
		}).Error; err != nil {
			return err
		}

		product := products[index]
		var persistedProduct model.CardProduct
		if err := db.WithContext(ctx).Where(&model.CardProduct{AccountID: account.ID, Channel: product.Channel, Prefix: product.Prefix}).First(&persistedProduct).Error; err != nil {
			return err
		}
		if persistedProduct.NextCardNumber < 1 {
			persistedProduct.NextCardNumber = 1
			if err := db.WithContext(ctx).Save(&persistedProduct).Error; err != nil {
				return err
			}
		}

		cardNumber, ok := cardnumber.Generate(persistedProduct.Prefix, 1)
		if !ok {
			return gorm.ErrInvalidData
		}
		var existingCard model.Card
		result := db.WithContext(ctx).Where(&model.Card{AccountID: account.ID, Channel: seed.channel, CardNumber: cardNumber}).First(&existingCard)
		if result.Error == nil {
			continue
		}
		if result.Error != gorm.ErrRecordNotFound {
			return result.Error
		}

		card := &model.Card{
			AccountID:              account.ID,
			Channel:                seed.channel,
			CardProductID:          persistedProduct.ID,
			CardBin:                persistedProduct.Prefix,
			CardNumber:             cardNumber,
			Cvv:                    "123",
			ExpireAt:               time.Now().UTC().AddDate(2, 0, 0),
			Status:                 enums.CardStatus_Active,
			CardHolderID:           holder.ID,
			FormType:               enums.CardFormType_Virtual,
			RequestID:              "seed-" + string(seed.channel) + "-card",
			LastOperationRequestID: "seed-" + string(seed.channel) + "-card",
			LastOperationType:      enums.OperationType_OpenCard,
			LastOperationStatus:    enums.OperationStatus_Succeed,
			CardCurrency:           enums.Currency_USD,
			CardScheme:             seed.cardScheme,
			CardType:               enums.CardType_Single,
		}
		if seed.withWallet {
			wallet := &model.Wallet{AccountID: account.ID, Channel: seed.channel, Amount: decimal.NewFromInt(1_000), Type: enums.WalletType_Card, Currency: enums.Currency_USD}
			if err := db.WithContext(ctx).Create(wallet).Error; err != nil {
				return err
			}
			card.WalletID = &wallet.ID
		}
		if err := db.WithContext(ctx).Create(card).Error; err != nil {
			return err
		}
	}

	return nil
}
