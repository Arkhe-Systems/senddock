package leader

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"hash/fnv"
)

var KeyBounceIMAPPoller = Key("senddock:bounce_imap_poller")

func Key(name string) int64 {
	h := fnv.New64a()
	h.Write([]byte(name))
	return int64(h.Sum64())
}

func TryRun(ctx context.Context, db *sql.DB, key int64, fn func(context.Context)) (bool, error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return false, err
	}

	var locked bool
	if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", key).Scan(&locked); err != nil {
		conn.Close()
		return false, err
	}
	if !locked {
		conn.Close()
		return false, nil
	}

	defer func() {
		if _, err := conn.ExecContext(context.WithoutCancel(ctx), "SELECT pg_advisory_unlock($1)", key); err != nil {
			_ = conn.Raw(func(any) error { return driver.ErrBadConn })
		}
		conn.Close()
	}()

	fn(ctx)
	return true, nil
}
