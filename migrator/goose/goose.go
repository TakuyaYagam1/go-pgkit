// Package goose provides PostgreSQL migrations using pressly/goose
package goose

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx postgres driver registration
	"github.com/pressly/goose/v3"

	"github.com/wahrwelt-kit/go-pgkit/postgres"
)

// Run runs pressly/goose "up" migrations from migrationsPath using the given PostgreSQL connection string. ctx is used for cancellation. connStr and migrationsPath must be non-empty. migrationsPath is cleaned with filepath.Clean and should be under application control (not user input). Uses a single connection (SetMaxOpenConns(1))
func Run(ctx context.Context, connStr, migrationsPath string) error {
	if ctx == nil {
		return errors.New("goose.Run: context is nil")
	}
	if connStr == "" {
		return errors.New("goose.Run: connection string is empty")
	}
	if migrationsPath == "" {
		return errors.New("goose.Run: migrations path is empty")
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
		return errors.New("goose.Run: migrations path is not a directory")
	}
	return runFS(ctx, "goose.Run", connStr, os.DirFS(absPath))
}

// RunFS runs pressly/goose "up" migrations from migrationsFS using the given PostgreSQL connection string.
// Use with go:embed for application-owned migration files. ctx is used for cancellation.
func RunFS(ctx context.Context, connStr string, migrationsFS fs.FS) error {
	return runFS(ctx, "goose.RunFS", connStr, migrationsFS)
}

func runFS(ctx context.Context, op, connStr string, migrationsFS fs.FS) error {
	if ctx == nil {
		return fmt.Errorf("%s: context is nil", op)
	}
	if connStr == "" {
		return fmt.Errorf("%s: connection string is empty", op)
	}
	if migrationsFS == nil {
		return fmt.Errorf("%s: migrations fs is nil", op)
	}
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("%s: sql.Open failed for %s: %w", op, postgres.MaskURL(connStr), err)
	}
	defer func() { _ = db.Close() }()
	db.SetMaxOpenConns(1)

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrationsFS)
	if err != nil {
		return fmt.Errorf("%s: NewProvider: %w", op, err)
	}
	defer func() { _ = provider.Close() }()

	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("%s: Up: %w", op, err)
	}
	return nil
}
