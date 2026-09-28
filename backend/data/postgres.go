package data

import (
	"context"

	"generic-mock/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const DefaultPostgresDSN = "host=localhost port=5432 user=postgres password=root dbname=generic_mock sslmode=disable"

const uuidV7FunctionName = "uuidv7"

const createUUIDV7FunctionSQL = `
CREATE FUNCTION uuidv7()
RETURNS uuid
LANGUAGE plpgsql
VOLATILE
AS $$
DECLARE
    unix_ts_ms bigint;
    random_uuid text;
BEGIN
    unix_ts_ms := floor(extract(epoch FROM clock_timestamp()) * 1000);
    random_uuid := replace(gen_random_uuid()::text, '-', '');

    RETURN (
        lpad(to_hex(unix_ts_ms), 12, '0') ||
        '7' ||
        substr(random_uuid, 14, 3) ||
        substr('89ab', floor(random() * 4)::integer + 1, 1) ||
        substr(random_uuid, 18, 15)
    )::uuid;
END;
$$;
`

func NewPostgresDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

func ResetPostgresDB(ctx context.Context, dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
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
	if err := ensureUUIDV7Function(db); err != nil {
		return err
	}

	if err := db.AutoMigrate(genericModels()...); err != nil {
		return err
	}
	if db.Migrator().HasColumn(&model.CardProduct{}, "is_default") {
		if err := db.Migrator().DropColumn(&model.CardProduct{}, "is_default"); err != nil {
			return err
		}
	}
	for _, index := range []string{
		"idx_cards_card_number",
		"idx_cards_request_id",
		"idx_cards_last_operation_request_id",
	} {
		if db.Migrator().HasIndex(&model.Card{}, index) {
			if err := db.Migrator().DropIndex(&model.Card{}, index); err != nil {
				return err
			}
		}
	}
	return nil
}

func ensureUUIDV7Function(db *gorm.DB) error {
	var exists bool
	if err := db.Raw(
		"SELECT EXISTS (SELECT 1 FROM pg_proc WHERE proname = ? AND pg_function_is_visible(oid))",
		uuidV7FunctionName,
	).Scan(&exists).Error; err != nil {
		return err
	}

	if exists {
		return nil
	}

	// UUIDv7 is not available on all PostgreSQL versions. pgcrypto provides entropy.
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto").Error; err != nil {
		return err
	}

	return db.Exec(createUUIDV7FunctionSQL).Error
}

func genericModels() []any {
	return []any{
		&model.Card{},
		&model.CardProduct{},
		&model.VirtualCard{},
		&model.PhysicalCard{},
		&model.Wallet{},
		&model.WalletTransfer{},
		&model.VirtualAccount{},
		&model.Account{},
		&model.CardTransaction{},
		&model.Authorization{},
		&model.CardHolder{},
		&model.WebhookConfig{},
		&model.AuthorizationConfig{},
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
