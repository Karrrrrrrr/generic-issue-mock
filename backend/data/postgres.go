package data

import (
	"context"

	"generic-mock/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const DefaultPostgresDSN = "host=localhost port=5432 user=postgres password=root dbname=generic_mock sslmode=disable"

func NewPostgresDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

func ResetPostgresDB(ctx context.Context, dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.WithContext(ctx).Migrator().DropTable(genericModels()...); err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

func migrate(db *gorm.DB) error {
	return db.AutoMigrate(genericModels()...)
}

func genericModels() []any {
	return []any{
		&model.Card{},
		&model.VirtualCard{},
		&model.PhysicalCard{},
		&model.Wallet{},
		&model.VirtualAccount{},
		&model.Account{},
		&model.CardTransaction{},
		&model.Authorization{},
		&model.CardHolder{},
		&model.WebhookConfig{},
		&model.WebhookRecord{},
	}
}

func Ping(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.PingContext(ctx)
}
