package goose

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testConnStr = "postgres://localhost/db"

func TestRun_EmptyParams(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dir := t.TempDir()
	absDir, err := filepath.Abs(dir)
	require.NoError(t, err)

	tests := []struct {
		name    string
		ctx     context.Context
		connStr string
		path    string
		want    string
	}{
		{"nil context", nil, testConnStr, absDir, "context is nil"},
		{"empty connStr", ctx, "", absDir, "connection string is empty"},
		{"empty path", ctx, testConnStr, "", "migrations path is empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Run(tt.ctx, tt.connStr, tt.path)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestRunFS_EmptyParams(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name    string
		ctx     context.Context
		connStr string
		fsys    fs.FS
		want    string
	}{
		{"nil context", nil, testConnStr, os.DirFS("."), "context is nil"},
		{"empty connStr", ctx, "", os.DirFS("."), "connection string is empty"},
		{"nil fs", ctx, testConnStr, nil, "migrations fs is nil"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RunFS(tt.ctx, tt.connStr, tt.fsys)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}
