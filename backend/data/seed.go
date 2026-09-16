package data

import (
	"context"

	photon "generic-mock/channel/photonpay/enums"
	"generic-mock/enums"
	"generic-mock/model"

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
