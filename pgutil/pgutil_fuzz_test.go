package pgutil

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func FuzzPgErrorCode(f *testing.F) {
	for _, seed := range []string{
		CodeUniqueViolation,
		CodeForeignKeyViolation,
		CodeNotNullViolation,
		CodeCheckViolation,
		CodeSerializationFailure,
		CodeDeadlockDetected,
		"",
		"XXXXX",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, code string) {
		if len(code) > 32 {
			t.Skip()
		}
		err := fmt.Errorf("wrap: %w", &pgconn.PgError{Code: code})
		if got := PgErrorCode(err); got != code {
			t.Fatalf("PgErrorCode() = %q, want %q", got, code)
		}
		if !IsPgErrorCode(err, code) {
			t.Fatalf("IsPgErrorCode() returned false for code %q", code)
		}
		if IsPgErrorCode(errors.New("other"), code) {
			t.Fatalf("IsPgErrorCode() returned true for non-pg error and code %q", code)
		}
	})
}
