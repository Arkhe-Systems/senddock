package license

import "context"

const (
	FeatureWorkspaceMembers = "workspace.members"
)

type Gate interface {
	AllowsFeature(ctx context.Context, feature string) bool
}

type denyGate struct{}

func (denyGate) AllowsFeature(_ context.Context, _ string) bool { return false }

func DenyAll() Gate { return denyGate{} }
