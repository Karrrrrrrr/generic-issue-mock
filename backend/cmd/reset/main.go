package main

import (
	"context"
	"log"
	"os"

	"generic-mock/data"
)

func main() {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = data.DefaultPostgresDSN
	}

	if _, err := data.ResetPostgresDB(context.Background(), dsn); err != nil {
		log.Fatalf("reset generic mock schema: %v", err)
	}
}
