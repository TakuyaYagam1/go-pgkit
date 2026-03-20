// Package pgutil provides helpers for working with pgx (jackc/pgx).
//
// # Error checks
//
// IsNoRows reports whether err is or wraps pgx.ErrNoRows (use after QueryRow when a missing row is expected).
// IsPgErrorCode reports whether err is a PgError with the given SQLSTATE code.
// IsPgUniqueViolation, IsForeignKeyViolation, IsNotNullViolation are conveniences for common codes (23505, 23503, 23502).
//
// # Timestamp conversion
//
// pgx uses pgtype.Timestamptz for timestamp with time zone. TimestamptzToTime returns *time.Time or nil when invalid.
// TimestamptzToTimeZero returns time.Time or the zero value. TimeToTimestamptz converts *time.Time to pgtype.Timestamptz (invalid if nil).
// PtrTimeToTime dereferences a *time.Time or returns time.Time{} if nil.
package pgutil
