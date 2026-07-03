//go:build integration

package migrate

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/wahrwelt-kit/go-pgkit/internal/testutil"
)

func TestRun(t *testing.T) {
	connStr := testutil.StartPostgres(t)

	migrationsPath, err := filepath.Abs("testdata")
	require.NoError(t, err)
	require.NoError(t, Run(context.Background(), connStr, migrationsPath))

	pool, err := pgxpool.New(context.Background(), connStr)
	require.NoError(t, err)
	defer pool.Close()
	var n int
	err = pool.QueryRow(context.Background(), "SELECT 1 FROM pg_tables WHERE tablename = 'pgkit_migrate_test'").Scan(&n)
	require.NoError(t, err)
	require.Equal(t, 1, n)

	require.NoError(t, Run(context.Background(), connStr, migrationsPath))
}

func TestRunOptions_StatementTimeout(t *testing.T) {
	connStr := testutil.StartPostgres(t)
	migrationsPath := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(migrationsPath, "000001_slow.up.sql"), []byte("SELECT pg_sleep(1);"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsPath, "000001_slow.down.sql"), []byte("SELECT 1;"), 0o600))

	err := RunOptions(context.Background(), connStr, migrationsPath, Options{
		StatementTimeout: 50 * time.Millisecond,
		LockTimeout:      time.Second,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "migrate.RunOptions: Up")
	require.True(t,
		strings.Contains(err.Error(), "context deadline exceeded") ||
			strings.Contains(err.Error(), "statement timeout"),
		"expected bounded statement timeout error, got %v", err,
	)
}
