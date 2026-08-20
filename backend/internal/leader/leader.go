package leader

import (
	"context"
	"database/sql"
)

const (
	KeyBounceIMAPPoller int64 = 1
)

func TryRun(ctx context.Context, db *sql.DB, key int64, fn func(context.Context) error) (bool, error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	var locked bool
	if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", key).Scan(&locked); err != nil {
		return false, err
	}
	if !locked {
		return false, nil
	}
	defer func() {
		_, _ = conn.ExecContext(context.WithoutCancel(ctx), "SELECT pg_advisory_unlock($1)", key)
	}()

	return true, fn(ctx)
}
