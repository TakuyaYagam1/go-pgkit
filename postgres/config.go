package postgres

import (
	"fmt"
	"net/url"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds connection settings for New. String and GoString mask the password; do not log raw Config
// MaxConns and MinConns override any pool parameters in the URL (e.g. pool_max_conns); when 0, defaults are used
// Duration fields use zero for library defaults (RetryTimeout 30s, MaxConnLifetime 1h, MaxConnIdleTime 30m, HealthCheckPeriod 15s, ConnectTimeout 5s). Negative durations are rejected.
type Config struct {
	URL               string        // PostgreSQL connection URL (required). Password is masked in String/GoString
	MaxConns          int           // Max connections in pool; 0 uses default (10). When set must be 1..10000. Overrides URL
	MinConns          int           // Min idle connections; 0 uses default (0). Must be 0..10000 and <= MaxConns. Overrides URL
	RetryTimeout      time.Duration // Max time for connection retry; 0 uses default (30s)
	MaxConnLifetime   time.Duration // Max lifetime of a connection; 0 uses default (1h)
	MaxConnIdleTime   time.Duration // Max idle time of a connection; 0 uses default (30m)
	HealthCheckPeriod time.Duration // How often to check connection health; 0 uses default (15s)
	ConnectTimeout    time.Duration // Timeout for establishing a connection; 0 uses default (5s)
	Configure         func(*pgxpool.Config) error
}

var (
	kvPasswordRE    = regexp.MustCompile(`(?i)\bpassword\s*=\s*(?:'[^']*'|"[^"]*"|\S+)`)
	userInfoBlockRE = regexp.MustCompile(`(://)([^@]*)(@)([^/]*)(/|$)`)
)

// MaskURL masks the password in a PostgreSQL connection URL for safe logging. On parse error uses regex fallback so that userinfo (including passwords containing @) is fully masked
func MaskURL(s string) string {
	masked := maskURLKeyValue(s)
	u, err := url.Parse(masked)
	if err != nil {
		return userInfoBlockRE.ReplaceAllString(masked, "$1***$3$4$5")
	}
	if u.User != nil {
		if _, hasPass := u.User.Password(); hasPass {
			u.User = url.UserPassword(u.User.Username(), "***")
		}
	}
	return maskURLKeyValue(u.String())
}

func maskURLKeyValue(s string) string {
	return kvPasswordRE.ReplaceAllString(s, "password=***")
}

// String returns a string representation of the config with the URL password masked. Use for logging
func (c Config) String() string {
	return MaskURL(c.URL)
}

// GoString implements fmt.GoStringer with the URL password masked. Use for %#v in logs
func (c Config) GoString() string {
	return fmt.Sprintf("postgres.Config{URL:%q, MaxConns:%d, MinConns:%d, RetryTimeout:%v, MaxConnLifetime:%v, MaxConnIdleTime:%v, HealthCheckPeriod:%v, ConnectTimeout:%v}",
		MaskURL(c.URL), c.MaxConns, c.MinConns, c.RetryTimeout, c.MaxConnLifetime, c.MaxConnIdleTime, c.HealthCheckPeriod, c.ConnectTimeout)
}
