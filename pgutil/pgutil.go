package pgutil

import (
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	CodeUniqueViolation      = "23505"
	CodeForeignKeyViolation  = "23503"
	CodeNotNullViolation     = "23502"
	CodeCheckViolation       = "23514"
	CodeSerializationFailure = "40001"
	CodeDeadlockDetected     = "40P01"
)

// IsNoRows reports whether err is or wraps pgx.ErrNoRows. Use after QueryRow when a missing row is acceptable
func IsNoRows(err error) bool {
	return err != nil && errors.Is(err, pgx.ErrNoRows)
}

// IsPgErrorCode reports whether err is or wraps a PgError with the given SQLSTATE code
func IsPgErrorCode(err error, code string) bool {
	var pgErr *pgconn.PgError
	return err != nil && errors.As(err, &pgErr) && pgErr.Code == code
}

// IsPgUniqueViolation reports whether err is a PostgreSQL unique constraint violation (SQLSTATE 23505)
func IsPgUniqueViolation(err error) bool {
	return IsPgErrorCode(err, CodeUniqueViolation)
}

// IsForeignKeyViolation reports whether err is a PostgreSQL foreign key violation (SQLSTATE 23503)
func IsForeignKeyViolation(err error) bool {
	return IsPgErrorCode(err, CodeForeignKeyViolation)
}

// IsNotNullViolation reports whether err is a PostgreSQL not null violation (SQLSTATE 23502)
func IsNotNullViolation(err error) bool {
	return IsPgErrorCode(err, CodeNotNullViolation)
}

// IsCheckViolation reports whether err is a PostgreSQL check constraint violation (SQLSTATE 23514)
func IsCheckViolation(err error) bool {
	return IsPgErrorCode(err, CodeCheckViolation)
}

// IsSerializationFailure reports whether err is a PostgreSQL serialization failure (SQLSTATE 40001)
func IsSerializationFailure(err error) bool {
	return IsPgErrorCode(err, CodeSerializationFailure)
}

// IsDeadlockDetected reports whether err is a PostgreSQL deadlock detected error (SQLSTATE 40P01)
func IsDeadlockDetected(err error) bool {
	return IsPgErrorCode(err, CodeDeadlockDetected)
}

// IsRetryableTxError reports whether err is a transaction error commonly safe to retry at the transaction boundary
func IsRetryableTxError(err error) bool {
	return IsSerializationFailure(err) || IsDeadlockDetected(err)
}

// PgErrorCode extracts the SQLSTATE code from err, or "" if err is not a PgError
func PgErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

// TimestamptzToTime returns a pointer to the time, or nil if t.Valid is false
func TimestamptzToTime(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// TimestamptzToTimeZero returns t.Time, or time.Time{} if t.Valid is false
func TimestamptzToTimeZero(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

// TimeToTimestamptz converts *time.Time to pgtype.Timestamptz. Returns an invalid Timestamptz if t is nil
func TimeToTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// PtrTimeToTime returns *t, or time.Time{} if t is nil
func PtrTimeToTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
