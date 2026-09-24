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
		log.Fatal("DATABASE_DSN is required for database initialization")
	}
	db, err := data.NewPostgresDB(dsn)
	if err != nil {
		log.Fatalf("initialize generic mock schema: %v", err)
	}
	if err := data.SeedInitialData(context.Background(), db); err != nil {
		log.Fatalf("seed generic mock database: %v", err)
	}
	log.Print("database schema and initial data are ready")
}
