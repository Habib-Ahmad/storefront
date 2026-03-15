package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"storefront/backend/migrations"
)

// Run from backend/:
//
//	go run ./cmd/migrate up
//	go run ./cmd/migrate down
//	go run ./cmd/migrate status
//	go run ./cmd/migrate create <name>
//	go run ./cmd/migrate version
//	go run ./cmd/migrate reset
func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		log.Fatal("usage: migrate up|down|status|create <name>|reset|version")
	}

	command := args[0]

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}

	goose.SetSequential(true)

	if command == "create" {
		if len(args) < 2 {
			log.Fatal("usage: migrate create <migration_name>")
		}
		dir := migrationsDirFromEnv()
		if err := goose.CreateWithTemplate(nil, dir, nil, args[1], "sql"); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("created new migration in %s\n", dir)
		return
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	goose.SetBaseFS(migrations.FS)

	if err := goose.RunContext(context.Background(), command, db, ".", args[1:]...); err != nil {
		log.Fatalf("goose %s: %v", command, err)
	}

	fmt.Printf("migration %q applied successfully\n", command)
}

func migrationsDirFromEnv() string {
	if dir := os.Getenv("MIGRATIONS_DIR"); dir != "" {
		return dir
	}
	return "migrations"
}
