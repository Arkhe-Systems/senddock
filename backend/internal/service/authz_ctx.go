package service

import (
	"context"
	"errors"
)

type ctxKey int

const (
	authorizedProjectKey ctxKey = iota
	systemContextKey
)

// ErrProjectNotAuthorized is returned by user-facing service methods when the caller
// did not wrap the context with WithAuthorizedProject before invoking the method.
// It is a fail-safe defense-in-depth check, not the primary authorization gate —
// handlers must still run their proper authz (workspace membership, capability check).
var ErrProjectNotAuthorized = errors.New("project access not authorized in this context")

// WithAuthorizedProject marks the context as having verified ownership of projectID.
// Handlers MUST call this after running their authz check (workspace membership,
// capability) before invoking EmailService methods that take a projectID.
func WithAuthorizedProject(ctx context.Context, projectID string) context.Context {
	return context.WithValue(ctx, authorizedProjectKey, projectID)
}

// WithSystemContext marks the context as a system/worker context that bypasses the
// per-request authz check. Use this from background workers, scheduled jobs, and
// explicitly-public endpoints where the project_id comes from a verified path
// (e.g. waitlist signup, unsubscribe links, tracking pixels).
func WithSystemContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, systemContextKey, true)
}

// requireAuthorizedProject is the defense-in-depth check inside user-facing service
// methods. Returns nil if the context was marked as authorized for this project or
// as system, and ErrProjectNotAuthorized otherwise.
func requireAuthorizedProject(ctx context.Context, projectID string) error {
	if v, _ := ctx.Value(systemContextKey).(bool); v {
		return nil
	}
	if v, _ := ctx.Value(authorizedProjectKey).(string); v == projectID {
		return nil
	}
	return ErrProjectNotAuthorized
}
