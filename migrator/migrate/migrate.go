// Package migrate provides PostgreSQL migrations using golang-migrate
package migrate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres database driver
	_ "github.com/golang-migrate/migrate/v4/source/file"       // file source driver

	"github.com/wahrwelt-kit/go-pgkit/postgres"
)

// Run runs golang-migrate "up" from file://migrationsPath using connURL. ErrNoChange is ignored. ctx is checked before starting; if cancelled, returns ctx.Err(). The library's Up() does not accept context, so a migration in progress cannot be cancelled-on context cancellation Run returns only after the current Up() call completes. For graceful shutdown, run migrations in a separate one-off process (e.g. init container or CI job); otherwise an in-progress migration will block process exit until it finishes
// connURL and migrationsPath must be non-empty. migrationsPath is cleaned and should be under application control (not user input)
func Run(ctx context.Context, connURL, migrationsPath string) (err error) {
	if connURL == "" {
		return errors.New("migrate.Run: connection URL is empty")
	}
	if migrationsPath == "" {
		return errors.New("migrate.Run: migrations path is empty")
	}
	cleanPath := filepath.Clean(migrationsPath)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("migrate.Run: migrations path: %w", err)
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("migrate.Run: migrations path: %w", err)
	}
	if !info.IsDir() {
		return errors.New("migrate.Run: migrations path is not a directory")
	}
	if ctx.Err() != nil {
		return fmt.Errorf("migrate.Run: %w", ctx.Err())
	}
	m, err := migrate.New("file://"+absPath, connURL)
	if err != nil {
		return fmt.Errorf("migrate.Run: New failed for %s: %w", postgres.MaskURL(connURL), err)
	}
	defer func() {
		if se, de := m.Close(); se != nil || de != nil {
			closeErr := errors.Join(se, de)
			wrapClose := fmt.Errorf("migrate.Run: Close: %w", closeErr)
			if err != nil {
				err = errors.Join(err, wrapClose)
			} else {
				err = wrapClose
			}
		}
	}()

	if ctx.Err() != nil {
		return fmt.Errorf("migrate.Run: %w", ctx.Err())
	}
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate.Run: Up: %w", err)
	}
	if ctx.Err() != nil {
		return fmt.Errorf("migrate.Run: %w", ctx.Err())
	}
	return nil
}
