package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	ctx := context.Background()

	// Read connection string from .env or args
	connStr := "postgresql://postgres.arhvidzxwvgmwoegrvre:pF8xvppeKUk5N6tI@aws-1-us-east-1.pooler.supabase.com:5432/postgres?sslmode=require"

	if len(os.Args) > 1 {
		connStr = os.Args[1]
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping: %v", err)
	}

	fmt.Println("✓ Connected to database")

	// Run migrations
	migrations := []string{
		"internal/store/postgres/migrations/001_initial_schema.up.sql",
		"internal/store/postgres/migrations/002_oauth2_support.up.sql",
	}

	for _, migFile := range migrations {
		fmt.Printf("Running %s...\n", migFile)

		content, err := os.ReadFile(migFile)
		if err != nil {
			log.Fatalf("Failed to read %s: %v", migFile, err)
		}

		if _, err := db.ExecContext(ctx, string(content)); err != nil {
			log.Fatalf("Failed to execute %s: %v", migFile, err)
		}

		fmt.Printf("✓ %s completed\n", migFile)
	}

	fmt.Println("\n✓✓✓ All migrations completed successfully!")
}
