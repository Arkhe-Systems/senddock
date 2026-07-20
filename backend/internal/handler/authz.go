package handler

import (
	"context"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/google/uuid"
)

func IsWorkspaceOwner(ctx context.Context, queries *db.Queries, userID string) bool {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return false
	}
	owned, err := queries.CountOwnedWorkspacesByUser(ctx, uid)
	return err == nil && owned > 0
}
