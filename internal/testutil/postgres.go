// Package testutil provides internal test helpers for go-pgkit integration tests
package testutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// StartPostgres starts a PostgreSQL container (postgres:18-alpine) and returns a connection string with sslmode=disable. The container is terminated via t.Cleanup. Use in integration tests that need a real database. Do not log the returned connection string without masking it (e.g. postgres.MaskURL)
func StartPostgres(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	c, err := postgres.Run(ctx, "postgres:18-alpine",
		postgres.WithDatabase("pgkit_test"),
		postgres.WithUsername("u"),
		postgres.WithPassword("p"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2)),
	)
	require.NoError(t, err)
	connStr, err := c.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Terminate(ctx) })
	return connStr
}
