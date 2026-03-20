package goose

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/takuya-go-kit/go-pgkit/postgres"
)

// Run runs pressly/goose "up" migrations from migrationsPath using the given PostgreSQL connection string. ctx is used for cancellation. connStr and migrationsPath must be non-empty. migrationsPath is cleaned with filepath.Clean and should be under application control (not user input). Uses a single connection (SetMaxOpenConns(1)).
func Run(ctx context.Context, connStr, migrationsPath string) error {
	if connStr == "" {
		return fmt.Errorf("goose.Run: connection string is empty")
	}
	if migrationsPath == "" {
		return fmt.Errorf("goose.Run: migrations path is empty")
	}
	cleanPath := filepath.Clean(migrationsPath)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("goose.Run: migrations path: %w", err)
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("goose.Run: migrations path: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("goose.Run: migrations path is not a directory")
	}
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("goose.Run: sql.Open failed for %s: %w", postgres.MaskURL(connStr), err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	provider, err := goose.NewProvider(goose.DialectPostgres, db, os.DirFS(absPath))
	if err != nil {
		return fmt.Errorf("goose.Run: NewProvider: %w", err)
	}
	defer provider.Close()

	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("goose.Run: Up: %w", err)
	}
	return nil
}
