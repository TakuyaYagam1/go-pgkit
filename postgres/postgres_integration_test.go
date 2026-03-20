//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/takuya-go-kit/go-pgkit/migrator/testutil"
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
