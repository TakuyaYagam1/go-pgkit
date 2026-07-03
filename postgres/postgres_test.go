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
)

const testPgURL = "postgres://localhost/db"

func TestNew_Validation(t *testing.T) {
	t.Parallel()
	validURL := "postgres://user:pass@localhost:5432/dbname?sslmode=disable"

	tests := []struct {
		name string
		ctx  context.Context
		cfg  *Config
		want string
	}{
		{
			name: "nil context",
			ctx:  nil,
			cfg:  &Config{URL: validURL},
			want: "context is nil",
		},
		{
			name: "nil config",
			ctx:  context.Background(),
			cfg:  nil,
			want: "config is nil",
		},
		{
			name: "invalid DSN format",
			ctx:  context.Background(),
			cfg:  &Config{URL: "invalid://bad"},
			want: "invalid DSN format",
		},
		{
			name: "empty URL",
			ctx:  context.Background(),
			cfg:  &Config{URL: ""},
			want: "URL is empty",
		},
		{
			name: "blank URL",
			ctx:  context.Background(),
			cfg:  &Config{URL: " \t\n "},
			want: "URL is empty",
		},
		{
			name: "MaxConns negative",
			ctx:  context.Background(),
			cfg:  &Config{URL: validURL, MaxConns: -1},
			want: "MaxConns must be 0 (default) or 1..10000",
		},
		{
			name: "MinConns negative",
			ctx:  context.Background(),
			cfg:  &Config{URL: validURL, MinConns: -1},
			want: "MinConns must be 0..10000",
		},
		{
			name: "MinConns greater than MaxConns",
			ctx:  context.Background(),
			cfg:  &Config{URL: validURL, MaxConns: 1, MinConns: 5},
			want: "MinConns (5) must be <= MaxConns (1)",
		},
		{
			name: "MaxConns too high",
			ctx:  context.Background(),
			cfg:  &Config{URL: validURL, MaxConns: 10001},
			want: "MaxConns must be 0 (default) or 1..10000",
		},
		{
			name: "MinConns too high",
			ctx:  context.Background(),
			cfg:  &Config{URL: validURL, MinConns: 10001},
			want: "MinConns must be 0..10000",
		},
		{
			name: "RetryTimeout negative",
			ctx:  context.Background(),
			cfg:  &Config{URL: validURL, RetryTimeout: -time.Second},
			want: "RetryTimeout must be >= 0",
		},
		{
			name: "MaxConnLifetime negative",
			ctx:  context.Background(),
			cfg:  &Config{URL: validURL, MaxConnLifetime: -time.Second},
			want: "MaxConnLifetime must be >= 0",
		},
		{
			name: "MaxConnIdleTime negative",
			ctx:  context.Background(),
			cfg:  &Config{URL: validURL, MaxConnIdleTime: -time.Second},
			want: "MaxConnIdleTime must be >= 0",
		},
		{
			name: "HealthCheckPeriod negative",
			ctx:  context.Background(),
			cfg:  &Config{URL: validURL, HealthCheckPeriod: -time.Second},
			want: "HealthCheckPeriod must be >= 0",
		},
		{
			name: "ConnectTimeout negative",
			ctx:  context.Background(),
			cfg:  &Config{URL: validURL, ConnectTimeout: -time.Second},
			want: "ConnectTimeout must be >= 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := New(tt.ctx, tt.cfg)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestDurationOrDefault(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 5*time.Second, durationOrDefault(5*time.Second, time.Minute))
	assert.Equal(t, time.Minute, durationOrDefault(0, time.Minute))
	assert.Equal(t, time.Minute, durationOrDefault(-1*time.Second, time.Minute))
}

func TestValidateLimits_Defaults(t *testing.T) {
	t.Parallel()
	require.NoError(t, validateLimits(&Config{URL: testPgURL}))
}

func TestValidateLimits_ValidBounds(t *testing.T) {
	t.Parallel()
	require.NoError(t, validateLimits(&Config{URL: testPgURL, MaxConns: 100, MinConns: 10}))
}

func TestValidateLimits_MinEqualsMax(t *testing.T) {
	t.Parallel()
	require.NoError(t, validateLimits(&Config{URL: testPgURL, MaxConns: 5, MinConns: 5}))
}

func TestNew_ConfigureError(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("configure failed")
	called := false
	_, err := New(context.Background(), &Config{
		URL: testPgURL,
		Configure: func(*pgxpool.Config) error {
			called = true
			return wantErr
		},
	})
	require.ErrorIs(t, err, wantErr)
	require.True(t, called)
}

func TestWithinTx_Validation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		ctx  context.Context
		fn   func() error
		want string
	}{
		{name: "nil context", ctx: nil, fn: func() error { return nil }, want: "context is nil"},
		{name: "nil pool", ctx: context.Background(), fn: func() error { return nil }, want: "pool is nil"},
		{name: "nil function", ctx: context.Background(), fn: nil, want: "function is nil"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var fn func(pgx.Tx) error
			if tt.fn != nil {
				fn = func(pgx.Tx) error { return tt.fn() }
			}
			err := WithinTx(tt.ctx, nil, fn)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}
