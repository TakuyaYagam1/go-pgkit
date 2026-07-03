package pgutil

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsNoRows(t *testing.T) {
	t.Parallel()
	assert.True(t, IsNoRows(pgx.ErrNoRows))
	assert.True(t, IsNoRows(errors.Join(pgx.ErrNoRows, errors.New("other"))))
	assert.False(t, IsNoRows(nil))
	assert.False(t, IsNoRows(errors.New("other")))
}

func TestIsPgUniqueViolation(t *testing.T) {
	t.Parallel()
	assert.True(t, IsPgUniqueViolation(&pgconn.PgError{Code: "23505"}))
	assert.True(t, IsPgUniqueViolation(fmt.Errorf("wrap: %w", &pgconn.PgError{Code: "23505"})))
	assert.False(t, IsPgUniqueViolation(&pgconn.PgError{Code: "23503"}))
	assert.False(t, IsPgUniqueViolation(nil))
	assert.False(t, IsPgUniqueViolation(errors.New("other")))
}

func TestIsForeignKeyViolation(t *testing.T) {
	t.Parallel()
	assert.True(t, IsForeignKeyViolation(&pgconn.PgError{Code: "23503"}))
	assert.True(t, IsForeignKeyViolation(fmt.Errorf("wrap: %w", &pgconn.PgError{Code: "23503"})))
	assert.False(t, IsForeignKeyViolation(&pgconn.PgError{Code: "23505"}))
	assert.False(t, IsForeignKeyViolation(nil))
	assert.False(t, IsForeignKeyViolation(errors.New("other")))
}

func TestIsNotNullViolation(t *testing.T) {
	t.Parallel()
	assert.True(t, IsNotNullViolation(&pgconn.PgError{Code: "23502"}))
	assert.True(t, IsNotNullViolation(fmt.Errorf("wrap: %w", &pgconn.PgError{Code: "23502"})))
	assert.False(t, IsNotNullViolation(&pgconn.PgError{Code: "23505"}))
	assert.False(t, IsNotNullViolation(nil))
	assert.False(t, IsNotNullViolation(errors.New("other")))
}

func TestAdditionalPgErrorHelpers(t *testing.T) {
	t.Parallel()
	checkErr := fmt.Errorf("wrap: %w", &pgconn.PgError{Code: CodeCheckViolation})
	serializationErr := fmt.Errorf("wrap: %w", &pgconn.PgError{Code: CodeSerializationFailure})
	deadlockErr := fmt.Errorf("wrap: %w", &pgconn.PgError{Code: CodeDeadlockDetected})
	uniqueErr := fmt.Errorf("wrap: %w", &pgconn.PgError{Code: CodeUniqueViolation})

	assert.True(t, IsCheckViolation(checkErr))
	assert.True(t, IsSerializationFailure(serializationErr))
	assert.True(t, IsDeadlockDetected(deadlockErr))
	assert.True(t, IsRetryableTxError(serializationErr))
	assert.True(t, IsRetryableTxError(deadlockErr))
	assert.False(t, IsRetryableTxError(uniqueErr))
	assert.False(t, IsCheckViolation(nil))
	assert.False(t, IsSerializationFailure(errors.New("other")))
	assert.False(t, IsDeadlockDetected(errors.New("other")))
}

func TestPgErrorCode(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "23505", PgErrorCode(&pgconn.PgError{Code: "23505"}))
	assert.Equal(t, "23503", PgErrorCode(fmt.Errorf("wrap: %w", &pgconn.PgError{Code: "23503"})))
	assert.Empty(t, PgErrorCode(nil))
	assert.Empty(t, PgErrorCode(errors.New("other")))
}

func TestTimestamptzToTime(t *testing.T) {
	t.Parallel()
	tt := time.Date(2025, 3, 15, 12, 0, 0, 0, time.UTC)
	out := TimestamptzToTime(pgtype.Timestamptz{Time: tt, Valid: true})
	require.NotNil(t, out)
	assert.True(t, out.Equal(tt))
	assert.Nil(t, TimestamptzToTime(pgtype.Timestamptz{Valid: false}))
}

func TestTimestamptzToTimeZero(t *testing.T) {
	t.Parallel()
	tt := time.Date(2025, 3, 15, 12, 0, 0, 0, time.UTC)
	assert.True(t, TimestamptzToTimeZero(pgtype.Timestamptz{Time: tt, Valid: true}).Equal(tt))
	assert.True(t, TimestamptzToTimeZero(pgtype.Timestamptz{Valid: false}).IsZero())
}

func TestTimeToTimestamptz(t *testing.T) {
	t.Parallel()
	tt := time.Date(2025, 3, 15, 12, 0, 0, 0, time.UTC)
	out := TimeToTimestamptz(&tt)
	assert.True(t, out.Valid)
	assert.True(t, out.Time.Equal(tt))
	out = TimeToTimestamptz(nil)
	assert.False(t, out.Valid)
}

func TestPtrTimeToTime(t *testing.T) {
	t.Parallel()
	tt := time.Date(2025, 3, 15, 12, 0, 0, 0, time.UTC)
	assert.True(t, PtrTimeToTime(&tt).Equal(tt))
	assert.True(t, PtrTimeToTime(nil).IsZero())
}
