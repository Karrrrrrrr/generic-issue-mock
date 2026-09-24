package main

import (
	"log"
	"os"

	"generic-mock/data"
	"generic-mock/model"

	"gorm.io/gen"
)

func main() {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = data.DefaultPostgresDSN
	}

	db, err := data.NewPostgresDB(dsn)
	if err != nil {
		log.Fatalf("initialize database schema: %v", err)
	}

	g := gen.NewGenerator(gen.Config{
		OutPath: "internal/query",
		Mode:    gen.WithDefaultQuery | gen.WithQueryInterface | gen.WithGeneric,
	})
	g.UseDB(db)
	g.ApplyBasic(
		&model.Card{},
		&model.CardProduct{},
		&model.VirtualCard{},
		&model.PhysicalCard{},
		&model.Wallet{},
		&model.VirtualAccount{},
		&model.Account{},
		&model.CardTransaction{},
		&model.Authorization{},
		&model.CardHolder{},
		&model.WebhookConfig{},
		&model.AuthorizationConfig{},
		&model.WebhookRecord{},
	)
	g.Execute()
}
