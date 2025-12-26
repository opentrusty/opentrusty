package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	url := "postgres://opentrusty:opentrusty@localhost:5432/opentrusty_test?sslmode=disable"
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), "DROP TABLE IF EXISTS sessions CASCADE; DROP TABLE IF EXISTS schema_migrations CASCADE; DROP TABLE IF EXISTS rbac_assignments CASCADE;")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Drop table failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Dropped sessions, rbac_assignments and schema_migrations tables successfully.")
}
