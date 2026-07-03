//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wahrwelt-kit/go-pgkit/internal/testutil"
)

func TestNew(t *testing.T) {
	connStr := testutil.StartPostgres(t)

	pool, err := New(context.Background(), &Config{URL: connStr})
	require.NoError(t, err)
	defer pool.Close()
	require.NoError(t, pool.Ping(context.Background()))
}

func TestNew_CustomConfig(t *testing.T) {
	connStr := testutil.StartPostgres(t)

	pool, err := New(context.Background(), &Config{
		URL:               connStr,
		MaxConns:          5,
		MinConns:          2,
		MaxConnLifetime:   2 * time.Hour,
		MaxConnIdleTime:   time.Hour,
		HealthCheckPeriod: 30 * time.Second,
		ConnectTimeout:    10 * time.Second,
		RetryTimeout:      10 * time.Second,
	})
	require.NoError(t, err)
	defer pool.Close()
	require.NoError(t, pool.Ping(context.Background()))

	stat := pool.Stat()
	assert.Equal(t, int32(5), stat.MaxConns())
}

func TestNew_Configure(t *testing.T) {
	connStr := testutil.StartPostgres(t)

	pool, err := New(context.Background(), &Config{
		URL: connStr,
		Configure: func(poolCfg *pgxpool.Config) error {
			poolCfg.MaxConns = 4
			return nil
		},
	})
	require.NoError(t, err)
	defer pool.Close()
	require.NoError(t, pool.Ping(context.Background()))
	assert.Equal(t, int32(4), pool.Stat().MaxConns())
}

func TestNew_DefaultDurations(t *testing.T) {
	connStr := testutil.StartPostgres(t)

	pool, err := New(context.Background(), &Config{URL: connStr, MaxConns: 3})
	require.NoError(t, err)
	defer pool.Close()
	require.NoError(t, pool.Ping(context.Background()))
	assert.Equal(t, int32(3), pool.Stat().MaxConns())
}

func TestNew_CancelledContext(t *testing.T) {
	connStr := testutil.StartPostgres(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := New(ctx, &Config{URL: connStr})
	require.Error(t, err)
}

func TestNew_InvalidURL(t *testing.T) {
	_, err := New(context.Background(), &Config{
		URL:          "postgres://invalid:5432/nonexistent?sslmode=disable",
		RetryTimeout: 2 * time.Second,
	})
	require.Error(t, err)
}

func TestWithinTx_CommitAndRollback(t *testing.T) {
	connStr := testutil.StartPostgres(t)

	pool, err := New(context.Background(), &Config{URL: connStr})
	require.NoError(t, err)
	defer pool.Close()

	_, err = pool.Exec(context.Background(), "CREATE TABLE tx_items (id INT PRIMARY KEY)")
	require.NoError(t, err)

	err = WithinTx(context.Background(), pool, func(tx pgx.Tx) error {
		_, execErr := tx.Exec(context.Background(), "INSERT INTO tx_items (id) VALUES (1)")
		return execErr
	})
	require.NoError(t, err)

	err = WithinTx(context.Background(), pool, func(tx pgx.Tx) error {
		_, execErr := tx.Exec(context.Background(), "INSERT INTO tx_items (id) VALUES (2)")
		if execErr != nil {
			return execErr
		}
		return errors.New("rollback")
	})
	require.Error(t, err)

	var count int
	err = pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM tx_items").Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}
