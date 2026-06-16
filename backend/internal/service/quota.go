package service

import "context"

type QuotaGate interface {
	AllowSubscribers(ctx context.Context, projectID string, adding int) error
	AllowProject(ctx context.Context, workspaceID string) error
	AllowMember(ctx context.Context, workspaceID string) error
}
