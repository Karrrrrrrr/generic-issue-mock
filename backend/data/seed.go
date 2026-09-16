package data

import (
	"context"

	photon "generic-mock/channel/photonpay/enums"
	"generic-mock/enums"
	"generic-mock/model"

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

	return nil
}
