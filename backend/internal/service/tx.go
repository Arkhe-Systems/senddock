package service

import (
	"context"
	"database/sql"

	"github.com/arkhe-systems/senddock/internal/db"
)

// withTransaction runs fn inside a single database transaction and commits it only when fn
// returns nil.
//
// It exists for the flows that write more than one row to mean one thing: an account and
// its workspace and its membership, a workspace and its owner. Written as separate
// statements, a failure partway through leaves half of it behind — an account with no
// workspace, or a workspace no member can see or delete. The caller sees one error and
// assumes nothing happened; the database disagrees.
//
// The queries handed to fn are bound to that transaction. The ones the caller already
// holds keep pointing at the pool, so a statement issued through them would run outside
// it — every write in the flow has to go through the parameter.
func withTransaction(ctx context.Context, conn *sql.DB, queries *db.Queries, fn func(*db.Queries) error) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// A rollback after a successful commit reports sql.ErrTxDone, so the error is dropped:
	// what matters is the result of the commit, returned below.
	defer func() { _ = tx.Rollback() }()

	if err := fn(queries.WithTx(tx)); err != nil {
		return err
	}

	return tx.Commit()
}
