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
	items := []*model.CardProduct{
		{
			Channel:   enums.Channel_Slash,
			Prefix:    slashDefaultCardProductPrefix,
			IsDefault: true,
		},
		{
			Channel:   enums.Channel_PhotonPay,
			Prefix:    photon.DefaultCardBin,
			IsDefault: true,
		},
		{
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
	if err := seedChannelCards(ctx, db, items); err != nil {
		return err
	}

	var wallet model.Wallet
	result := db.WithContext(ctx).Where(&model.Wallet{
		Type:     enums.WalletType_VirtualAccount,
		Currency: enums.Currency_USD,
	}).First(&wallet)
	if result.Error == gorm.ErrRecordNotFound {
		wallet = model.Wallet{
			Amount:   decimal.NewFromInt(1_000_000),
			Type:     enums.WalletType_VirtualAccount,
			Currency: enums.Currency_USD,
		}
		if err := db.WithContext(ctx).Create(&wallet).Error; err != nil {
			return err
		}
	} else if result.Error != nil {
		return result.Error
	}

	var account model.VirtualAccount
	result = db.WithContext(ctx).Where(&model.VirtualAccount{Name: "Slash Primary"}).First(&account)
	if result.Error == gorm.ErrRecordNotFound {
		if err := db.WithContext(ctx).Create(&model.VirtualAccount{
			WalletID: wallet.ID,
			Name:     "Slash Primary",
		}).Error; err != nil {
			return err
		}
	} else if result.Error != nil {
		return result.Error
	}

	return nil
}

type channelCardSeed struct {
	channel    enums.Channel
	firstName  string
	lastName   string
	email      string
	cardScheme enums.CardScheme
	withWallet bool
}

func seedChannelCards(ctx context.Context, db *gorm.DB, products []*model.CardProduct) error {
	seeds := []channelCardSeed{
		{channel: enums.Channel_Slash, firstName: "Slash", lastName: "Demo", email: "slash.demo@example.test", cardScheme: enums.CardScheme_Visa},
		{channel: enums.Channel_PhotonPay, firstName: "PhotonPay", lastName: "Demo", email: "photonpay.demo@example.test", cardScheme: enums.CardScheme_Visa},
		{channel: enums.Channel_Paynda, firstName: "Paynda", lastName: "Demo", email: "paynda.demo@example.test", cardScheme: enums.CardScheme_MasterCard, withWallet: true},
	}

	for index, seed := range seeds {
		var holder model.CardHolder
		if err := db.WithContext(ctx).Where(&model.CardHolder{Channel: seed.channel, Email: seed.email}).FirstOrCreate(&holder, &model.CardHolder{
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
		if err := db.WithContext(ctx).Where(&model.CardProduct{Channel: product.Channel, Prefix: product.Prefix}).First(&persistedProduct).Error; err != nil {
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
		result := db.WithContext(ctx).Where(&model.Card{Channel: seed.channel, CardNumber: cardNumber}).First(&existingCard)
		if result.Error == nil {
			continue
		}
		if result.Error != gorm.ErrRecordNotFound {
			return result.Error
		}

		card := &model.Card{
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
			wallet := &model.Wallet{Amount: decimal.NewFromInt(1_000), Type: enums.WalletType_Card, Currency: enums.Currency_USD}
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
