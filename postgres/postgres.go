package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxConns        = 10
	defaultMinConns        = 0
	maxConnsLimit          = 10000
	defaultMaxConnLifetime = time.Hour
	defaultMaxConnIdleTime = 30 * time.Minute
	defaultHealthCheck     = 15 * time.Second
	defaultConnectTimeout  = 5 * time.Second
	pgConnRetryTimeout     = 30 * time.Second
)

// New creates a pgxpool with exponential backoff until the database is reachable. ctx can cancel the retry. Returns an error if ctx or cfg is nil, the URL is invalid, or pool limits are out of range
func New(ctx context.Context, cfg *Config) (*pgxpool.Pool, error) {
	if ctx == nil {
		return nil, errors.New("postgres - New: context is nil")
	}
	if cfg == nil {
		return nil, errors.New("postgres - New: config is nil")
	}
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}
	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("postgres - New: invalid DSN format: %w", err)
	}
	applyPoolConfig(poolCfg, cfg)
	if cfg.Configure != nil {
		if err := cfg.Configure(poolCfg); err != nil {
			return nil, fmt.Errorf("postgres - New: configure: %w", err)
		}
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("postgres - New: pool creation failed: %w", err)
	}

	if err := pingWithRetry(ctx, pool, cfg.RetryTimeout); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func validateConfig(cfg *Config) error {
	if strings.TrimSpace(cfg.URL) == "" {
		return errors.New("postgres - New: URL is empty")
	}
	if err := validateLimits(cfg); err != nil {
		return err
	}
	return validateDurations(cfg)
}

func validateLimits(cfg *Config) error {
	if cfg.MaxConns != 0 && (cfg.MaxConns < 1 || cfg.MaxConns > maxConnsLimit) {
		return fmt.Errorf("postgres - New: MaxConns must be 0 (default) or 1..%d", maxConnsLimit)
	}
	if cfg.MinConns < 0 || cfg.MinConns > maxConnsLimit {
		return fmt.Errorf("postgres - New: MinConns must be 0..%d", maxConnsLimit)
	}
	maxConns := defaultMaxConns
	if cfg.MaxConns > 0 {
		maxConns = cfg.MaxConns
	}
	minConns := defaultMinConns
	if cfg.MinConns > 0 {
		minConns = cfg.MinConns
	}
	if minConns > maxConns {
		return fmt.Errorf("postgres - New: MinConns (%d) must be <= MaxConns (%d)", minConns, maxConns)
	}
	return nil
}

func validateDurations(cfg *Config) error {
	durations := []struct {
		name  string
		value time.Duration
	}{
		{name: "RetryTimeout", value: cfg.RetryTimeout},
		{name: "MaxConnLifetime", value: cfg.MaxConnLifetime},
		{name: "MaxConnIdleTime", value: cfg.MaxConnIdleTime},
		{name: "HealthCheckPeriod", value: cfg.HealthCheckPeriod},
		{name: "ConnectTimeout", value: cfg.ConnectTimeout},
	}
	for _, d := range durations {
		if d.value < 0 {
			return fmt.Errorf("postgres - New: %s must be >= 0", d.name)
		}
	}
	return nil
}

func durationOrDefault(v, def time.Duration) time.Duration {
	if v > 0 {
		return v
	}
	return def
}

func applyPoolConfig(poolCfg *pgxpool.Config, cfg *Config) {
	maxConns := defaultMaxConns
	if cfg.MaxConns > 0 {
		maxConns = cfg.MaxConns
	}
	minConns := defaultMinConns
	if cfg.MinConns > 0 {
		minConns = cfg.MinConns
	}
	poolCfg.MaxConns = int32(maxConns)
	poolCfg.MinConns = int32(minConns)
	poolCfg.MaxConnLifetime = durationOrDefault(cfg.MaxConnLifetime, defaultMaxConnLifetime)
	poolCfg.MaxConnIdleTime = durationOrDefault(cfg.MaxConnIdleTime, defaultMaxConnIdleTime)
	poolCfg.HealthCheckPeriod = durationOrDefault(cfg.HealthCheckPeriod, defaultHealthCheck)
	poolCfg.ConnConfig.ConnectTimeout = durationOrDefault(cfg.ConnectTimeout, defaultConnectTimeout)
}

func pingWithRetry(ctx context.Context, pool *pgxpool.Pool, retryTimeout time.Duration) error {
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = pgConnRetryTimeout
	if retryTimeout > 0 {
		bo.MaxElapsedTime = retryTimeout
	}
	if err := backoff.Retry(func() error { return pool.Ping(ctx) }, backoff.WithContext(bo, ctx)); err != nil {
		return fmt.Errorf("postgres - New: ping failed after retries: %w", err)
	}
	return nil
}
