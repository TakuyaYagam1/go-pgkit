package postgres

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func contains(s, sub string) bool { return strings.Contains(s, sub) }

func TestMaskURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		in     string
		noShow []string
	}{
		{
			name:   "standard URL with password",
			in:     "postgres://user:secret@localhost:5432/db",
			noShow: []string{"secret"},
		},
		{
			name:   "password contains @",
			in:     "postgres://user:p@ss@host:5432/db",
			noShow: []string{"p@ss"},
		},
		{
			name:   "no password",
			in:     "postgres://user@localhost:5432/db",
			noShow: nil,
		},
		{
			name:   "keyword password",
			in:     "host=localhost user=u password=secret dbname=db",
			noShow: []string{"secret"},
		},
		{
			name:   "quoted keyword password",
			in:     "host=localhost password='quoted secret' dbname=db",
			noShow: []string{"quoted secret"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskURL(tt.in)
			for _, s := range tt.noShow {
				assert.NotContains(t, got, s, "masked output must not contain password")
			}
			if tt.noShow != nil {
				assert.True(t, got != "" && (contains(got, "***") || contains(got, "%2A%2A%2A")), "masked output should contain placeholder")
			}
		})
	}
}

func TestConfig_String(t *testing.T) {
	t.Parallel()
	cfg := Config{URL: "postgres://u:secret@h/db", MaxConns: 10}
	assert.NotContains(t, cfg.String(), "secret")
	assert.True(t, contains(cfg.String(), "***") || contains(cfg.String(), "%2A%2A%2A"))
}

func TestConfig_GoString(t *testing.T) {
	t.Parallel()
	cfg := Config{URL: "postgres://u:secret@h/db", MaxConns: 10}
	assert.NotContains(t, cfg.GoString(), "secret")
	assert.True(t, contains(cfg.GoString(), "***") || contains(cfg.GoString(), "%2A%2A%2A"))
}
