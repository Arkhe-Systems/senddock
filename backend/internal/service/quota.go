package service

import "context"

// QuotaGate is an optional hook the embedding binary (cloud) can install to
// enforce per-plan quantity caps. A nil gate means no caps (self-hosted).
// Each method returns a non-nil error only when the action exceeds the plan
// limit; it should fail open (return nil) on internal lookup errors.
type QuotaGate interface {
	AllowSubscribers(ctx context.Context, projectID string, adding int) error
	AllowProject(ctx context.Context, workspaceID string) error
}
