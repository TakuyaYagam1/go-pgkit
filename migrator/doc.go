// Package migrator provides migration runners in subpackages. Use the one that matches your migration layout
//
// # goose (migrator/goose)
//
// Run(ctx, connStr, migrationsPath) runs pressly/goose "up" migrations. SQL files use +goose Up/Down directives
// connStr and migrationsPath must be non-empty. migrationsPath is cleaned with filepath.Clean and should be
// under application control (not user input). ctx is used for cancellation. RunFS(ctx, connStr, migrationsFS)
// runs the same goose provider from fs.FS, which is useful for go:embed migration files
//
// # migrate (migrator/migrate)
//
// Run(ctx, connURL, migrationsPath) runs golang-migrate "up" from file://migrationsPath. Expects separate .up.sql
// and .down.sql files. connURL and migrationsPath must be non-empty. migrationsPath is cleaned and should be under
// application control. ErrNoChange is ignored. RunOptions accepts explicit statement and lock timeouts
package migrator
