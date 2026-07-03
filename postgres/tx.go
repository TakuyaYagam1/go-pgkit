package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WithinTx runs fn inside a PostgreSQL transaction using default pgx options.
func WithinTx(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) error {
	return WithinTxOptions(ctx, pool, pgx.TxOptions{}, fn)
}

// WithinTxOptions runs fn inside a PostgreSQL transaction.
// It commits when fn returns nil, rolls back when fn returns an error, and re-panics after rollback when fn panics.
func WithinTxOptions(ctx context.Context, pool *pgxpool.Pool, opts pgx.TxOptions, fn func(pgx.Tx) error) (err error) {
	if ctx == nil {
		return errors.New("postgres - WithinTx: context is nil")
	}
	if fn == nil {
		return errors.New("postgres - WithinTx: function is nil")
	}
	if pool == nil {
		return errors.New("postgres - WithinTx: pool is nil")
	}

	tx, err := pool.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("postgres - WithinTx: begin: %w", err)
	}

	committed := false
	defer func() {
		if committed {
			return
		}
		if rbErr := tx.Rollback(context.WithoutCancel(ctx)); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			wrapped := fmt.Errorf("postgres - WithinTx: rollback: %w", rbErr)
			if err != nil {
				err = errors.Join(err, wrapped)
				return
			}
			err = wrapped
		}
	}()

	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("postgres - WithinTx: commit: %w", err)
	}
	committed = true
	return nil
}
