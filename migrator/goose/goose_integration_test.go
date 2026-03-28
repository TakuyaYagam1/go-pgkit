//go:build integration

package goose

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/wahrwelt-kit/go-pgkit/migrator/testutil"
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
	err = pool.QueryRow(context.Background(), "SELECT 1 FROM pg_tables WHERE tablename = 'pgkit_test'").Scan(&n)
	require.NoError(t, err)
	require.Equal(t, 1, n)
}
