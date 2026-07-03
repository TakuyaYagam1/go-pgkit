package migrate

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testConnURL = "postgres://localhost/db"

func TestRun_EmptyParams(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dir := t.TempDir()
	absDir, err := filepath.Abs(dir)
	require.NoError(t, err)

	tests := []struct {
		name    string
		ctx     context.Context
		connURL string
		path    string
		opts    Options
		want    string
	}{
		{"nil context", nil, testConnURL, absDir, Options{}, "context is nil"},
		{"empty connURL", ctx, "", absDir, Options{}, "connection URL is empty"},
		{"empty path", ctx, testConnURL, "", Options{}, "migrations path is empty"},
		{"negative statement timeout", ctx, testConnURL, absDir, Options{StatementTimeout: -time.Nanosecond}, "StatementTimeout must be >= 0"},
		{"negative lock timeout", ctx, testConnURL, absDir, Options{LockTimeout: -time.Nanosecond}, "LockTimeout must be >= 0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RunOptions(tt.ctx, tt.connURL, tt.path, tt.opts)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestRun_NotDirectory(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "migration.sql")
	require.NoError(t, os.WriteFile(path, []byte("SELECT 1;"), 0o600))

	err := Run(context.Background(), testConnURL, path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "migrations path is not a directory")
}

func TestRun_CancelledContext(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	dir := t.TempDir()
	absDir, err := filepath.Abs(dir)
	require.NoError(t, err)

	err = Run(ctx, "postgres://user:pass@localhost:5432/db?sslmode=disable", absDir)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}
