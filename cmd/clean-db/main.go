package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	ctx := context.Background()
	connStr := "postgresql://postgres.arhvidzxwvgmwoegrvre:pF8xvppeKUk5N6tI@aws-1-us-east-1.pooler.supabase.com:5432/postgres?sslmode=require"

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	fmt.Println("Cleaning database...")

	// Drop all data (in reverse dependency order)
	tables := []string{
		"refresh_tokens",
		"access_tokens",
		"authorization_codes",
		"user_project_roles",
		"oauth2_clients",
		"projects",
		"roles",
		"sessions",
		"credentials",
		"users",
	}

	for _, table := range tables {
		_, err := db.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			fmt.Printf("Warning: failed to truncate %s: %v\n", table, err)
		} else {
			fmt.Printf("✓ Cleared %s\n", table)
		}
	}

	// Re-insert default roles
	fmt.Println("\nRe-inserting default roles...")
	roles := []struct {
		id, name, desc string
		perms          []string
	}{
		{"role_admin", "admin", "Administrator with full access", []string{"*"}},
		{"role_developer", "developer", "Developer with read/write access", []string{"read", "write", "deploy"}},
		{"role_viewer", "viewer", "Viewer with read-only access", []string{"read"}},
	}

	for _, r := range roles {
		permsJSON := `["` + r.perms[0]
		for i := 1; i < len(r.perms); i++ {
			permsJSON += `","` + r.perms[i]
		}
		permsJSON += `"]`

		_, err := db.ExecContext(ctx, `
			INSERT INTO roles (id, name, description, permissions)
			VALUES ($1, $2, $3, $4::jsonb)
		`, r.id, r.name, r.desc, permsJSON)

		if err != nil {
			log.Printf("Failed to insert role %s: %v", r.name, err)
		} else {
			fmt.Printf("✓ Created role: %s\n", r.name)
		}
	}

	fmt.Println("\n✓✓✓ Database cleaned and reset successfully!")
}
