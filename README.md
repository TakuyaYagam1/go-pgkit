# go-pgkit

[![CI](https://github.com/wahrwelt-kit/go-pgkit/actions/workflows/ci.yml/badge.svg)](https://github.com/wahrwelt-kit/go-pgkit/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/wahrwelt-kit/go-pgkit.svg)](https://pkg.go.dev/github.com/wahrwelt-kit/go-pgkit)
[![Go Report Card](https://goreportcard.com/badge/github.com/wahrwelt-kit/go-pgkit)](https://goreportcard.com/report/github.com/wahrwelt-kit/go-pgkit)

PostgreSQL helpers for pgx: SQLSTATE checks, timestamptz converters, pool creation with retry, transaction helper, and migration runners (goose and golang-migrate).

## Install

```bash
go get github.com/wahrwelt-kit/go-pgkit
```

```go
import "github.com/wahrwelt-kit/go-pgkit/pgutil"
import "github.com/wahrwelt-kit/go-pgkit/postgres"
import "github.com/wahrwelt-kit/go-pgkit/migrator/goose"
import "github.com/wahrwelt-kit/go-pgkit/migrator/migrate"
```

## Subpackages

### pgutil

- **IsNoRows(err)** - true if err is or wraps pgx.ErrNoRows
- **IsPgErrorCode(err, code)** - true if err is a PgError with the given SQLSTATE code
- **IsPgUniqueViolation(err)** - true if PostgreSQL unique violation (23505)
- **IsForeignKeyViolation(err)** - true if PostgreSQL foreign key violation (23503)
- **IsNotNullViolation(err)** - true if PostgreSQL not null violation (23502)
- **IsCheckViolation(err)** - true if PostgreSQL check violation (23514)
- **IsSerializationFailure(err)** - true if PostgreSQL serialization failure (40001)
- **IsDeadlockDetected(err)** - true if PostgreSQL deadlock detected (40P01)
- **IsRetryableTxError(err)** - true for retryable transaction errors (serialization failure or deadlock)
- **PgErrorCode(err)** - SQLSTATE code or ""
- **TimestamptzToTime(t)** - \*time.Time or nil if invalid
- **TimestamptzToTimeZero(t)** - time.Time or zero if invalid
- **TimeToTimestamptz(t)** - pgtype.Timestamptz (invalid if t is nil)
- **PtrTimeToTime(t)** - dereference or time.Time{}

### postgres

- **Config** - URL, MaxConns, MinConns, RetryTimeout; optional MaxConnLifetime, MaxConnIdleTime, HealthCheckPeriod, ConnectTimeout (0 = defaults; negative durations are rejected), and Configure hook for pgxpool customization
- **New(ctx, cfg)** - create pgxpool with exponential backoff retry until Ping succeeds
- **WithinTx(ctx, pool, fn)** - run a function inside a transaction, commit on nil error, rollback on error
- **WithinTxOptions(ctx, pool, opts, fn)** - transaction helper with explicit pgx TxOptions

### migrator

Two runners in subpackages; use the one that matches your migration layout.

- **goose.Run(ctx, connStr, migrationsPath)** - pressly/goose: SQL files with `-- +goose Up` / `-- +goose Down`
- **goose.RunFS(ctx, connStr, migrationsFS)** - pressly/goose migrations from `fs.FS`, useful with `go:embed`
- **migrate.Run(ctx, connURL, migrationsPath)** - golang-migrate: `NNNNNN_name.up.sql` / `NNNNNN_name.down.sql`; treats ErrNoChange as success and uses bounded lock/statement timeouts
- **migrate.RunOptions(ctx, connURL, migrationsPath, opts)** - golang-migrate with explicit StatementTimeout and LockTimeout

## Example

```go
pool, err := postgres.New(ctx, &postgres.Config{
    URL:     os.Getenv("DATABASE_URL"),
    MaxConns: 20,
})
if err != nil {
    log.Fatal(err)
}
defer pool.Close()

if err := goose.Run(ctx, connStr, "./migrations"); err != nil {
    log.Fatal(err)
}

if err := migrate.RunOptions(ctx, connStr, "./migrate", migrate.Options{
    StatementTimeout: 10 * time.Minute,
    LockTimeout:      30 * time.Second,
}); err != nil {
    log.Fatal(err)
}

err = postgres.WithinTx(ctx, pool, func(tx pgx.Tx) error {
    _, err := tx.Exec(ctx, "INSERT INTO audit_log (event) VALUES ($1)", "user.created")
    return err
})
if err != nil {
    return err
}

var t pgtype.Timestamptz
err := pool.QueryRow(ctx, "SELECT created_at FROM users WHERE id = $1", id).Scan(&t)
if pgutil.IsNoRows(err) {
    return ErrNotFound
}
createdAt := pgutil.TimestamptzToTime(t)
```
