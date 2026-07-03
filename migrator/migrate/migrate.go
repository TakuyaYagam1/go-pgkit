// Package migrate provides PostgreSQL migrations using golang-migrate
package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	gomigrate "github.com/golang-migrate/migrate/v4"
	pgx5migrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file" // file source driver

	"github.com/wahrwelt-kit/go-pgkit/postgres"
)

const (
	DefaultStatementTimeout = 5 * time.Minute
	DefaultLockTimeout      = 15 * time.Second
)

// Options controls migrate.RunOptions.
type Options struct {
	// StatementTimeout bounds each migration SQL statement. Zero uses DefaultStatementTimeout.
	StatementTimeout time.Duration
	// LockTimeout bounds golang-migrate advisory lock acquisition. Zero uses DefaultLockTimeout.
	LockTimeout time.Duration
}

// Run runs golang-migrate "up" from file://migrationsPath using connURL.
// ErrNoChange is ignored. connURL and migrationsPath must be non-empty.
// migrationsPath is cleaned and should be under application control (not user input).
func Run(ctx context.Context, connURL, migrationsPath string) error {
	return RunOptions(ctx, connURL, migrationsPath, Options{})
}

// RunOptions runs golang-migrate "up" with explicit execution bounds.
// Zero timeouts use DefaultStatementTimeout and DefaultLockTimeout; negative values are rejected.
func RunOptions(ctx context.Context, connURL, migrationsPath string, opts Options) (err error) {
	absPath, err := validateRunArgs(ctx, connURL, migrationsPath, opts)
	if err != nil {
		return err
	}
	m, err := newMigrate(ctx, connURL, absPath, opts)
	if err != nil {
		return err
	}
	stopContextWatch := watchContext(ctx, m)
	defer stopContextWatch()
	defer func() {
		if se, de := m.Close(); se != nil || de != nil {
			closeErr := errors.Join(se, de)
			wrapClose := fmt.Errorf("migrate.RunOptions: Close: %w", closeErr)
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
	if err = m.Up(); err != nil && !errors.Is(err, gomigrate.ErrNoChange) {
		return fmt.Errorf("migrate.RunOptions: Up: %w", err)
	}
	if ctx.Err() != nil {
		return fmt.Errorf("migrate.Run: %w", ctx.Err())
	}
	return nil
}

func validateRunArgs(ctx context.Context, connURL, migrationsPath string, opts Options) (string, error) {
	if ctx == nil {
		return "", errors.New("migrate.Run: context is nil")
	}
	if connURL == "" {
		return "", errors.New("migrate.Run: connection URL is empty")
	}
	if migrationsPath == "" {
		return "", errors.New("migrate.Run: migrations path is empty")
	}
	cleanPath := filepath.Clean(migrationsPath)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("migrate.Run: migrations path: %w", err)
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("migrate.Run: migrations path: %w", err)
	}
	if !info.IsDir() {
		return "", errors.New("migrate.Run: migrations path is not a directory")
	}
	if opts.StatementTimeout < 0 {
		return "", errors.New("migrate.Run: StatementTimeout must be >= 0")
	}
	if opts.LockTimeout < 0 {
		return "", errors.New("migrate.Run: LockTimeout must be >= 0")
	}
	if ctx.Err() != nil {
		return "", fmt.Errorf("migrate.Run: %w", ctx.Err())
	}
	return absPath, nil
}

func newMigrate(ctx context.Context, connURL, migrationsPath string, opts Options) (*gomigrate.Migrate, error) {
	db, err := sql.Open("pgx/v5", connURL)
	if err != nil {
		return nil, fmt.Errorf("migrate.RunOptions: sql.Open failed for %s: %w", postgres.MaskURL(connURL), err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	ownsDB := true
	defer func() {
		if ownsDB {
			_ = db.Close()
		}
	}()

	statementTimeout := durationOrDefault(opts.StatementTimeout, DefaultStatementTimeout)
	if err := setStatementTimeout(ctx, db, statementTimeout); err != nil {
		return nil, err
	}
	driver, err := pgx5migrate.WithInstance(db, &pgx5migrate.Config{
		StatementTimeout: statementTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("migrate.RunOptions: database driver for %s: %w", postgres.MaskURL(connURL), err)
	}
	m, err := gomigrate.NewWithDatabaseInstance("file://"+migrationsPath, "pgx5", driver)
	if err != nil {
		return nil, fmt.Errorf("migrate.RunOptions: NewWithDatabaseInstance: %w", err)
	}
	m.LockTimeout = durationOrDefault(opts.LockTimeout, DefaultLockTimeout)
	ownsDB = false
	return m, nil
}

func setStatementTimeout(ctx context.Context, db *sql.DB, timeout time.Duration) error {
	ms := max(timeout.Milliseconds(), int64(1))
	if _, err := db.ExecContext(ctx, `SELECT set_config('statement_timeout', $1, false)`, fmt.Sprintf("%dms", ms)); err != nil {
		return fmt.Errorf("migrate.RunOptions: set statement_timeout: %w", err)
	}
	return nil
}

func durationOrDefault(v, def time.Duration) time.Duration {
	if v > 0 {
		return v
	}
	return def
}

func watchContext(ctx context.Context, m *gomigrate.Migrate) func() {
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			select {
			case m.GracefulStop <- true:
			default:
			}
		case <-done:
		}
	}()
	return func() { close(done) }
}
