package pgutil

import (
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// IsNoRows reports whether err is or wraps pgx.ErrNoRows. Use after QueryRow when a missing row is acceptable.
func IsNoRows(err error) bool {
	return err != nil && errors.Is(err, pgx.ErrNoRows)
}

// IsPgErrorCode reports whether err is or wraps a PgError with the given SQLSTATE code.
func IsPgErrorCode(err error, code string) bool {
	var pgErr *pgconn.PgError
	return err != nil && errors.As(err, &pgErr) && pgErr.Code == code
}

// IsPgUniqueViolation reports whether err is a PostgreSQL unique constraint violation (SQLSTATE 23505).
func IsPgUniqueViolation(err error) bool {
	return IsPgErrorCode(err, "23505")
}

// IsForeignKeyViolation reports whether err is a PostgreSQL foreign key violation (SQLSTATE 23503).
func IsForeignKeyViolation(err error) bool {
	return IsPgErrorCode(err, "23503")
}

// IsNotNullViolation reports whether err is a PostgreSQL not null violation (SQLSTATE 23502).
func IsNotNullViolation(err error) bool {
	return IsPgErrorCode(err, "23502")
}

// PgErrorCode extracts the SQLSTATE code from err, or "" if err is not a PgError.
func PgErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

// TimestamptzToTime returns a pointer to the time, or nil if t.Valid is false.
func TimestamptzToTime(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// TimestamptzToTimeZero returns t.Time, or time.Time{} if t.Valid is false.
func TimestamptzToTimeZero(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

// TimeToTimestamptz converts *time.Time to pgtype.Timestamptz. Returns an invalid Timestamptz if t is nil.
func TimeToTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// PtrTimeToTime returns *t, or time.Time{} if t is nil.
func PtrTimeToTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
