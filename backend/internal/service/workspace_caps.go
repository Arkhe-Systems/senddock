package service

import (
	"context"
	"errors"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/google/uuid"
)

type Capability string

const (
	CapMembersManage      Capability = "members.manage"
	CapWorkspaceDelete    Capability = "workspace.delete"
	CapProjectSettings    Capability = "project.settings"
	CapTemplatesWrite     Capability = "templates.write"
	CapSubscribersWrite   Capability = "subscribers.write"
	CapSendTransactional  Capability = "send.transactional"
	CapBroadcast          Capability = "broadcast"
	CapCampaignsWrite     Capability = "campaigns.write"
	CapAPIKeysManage      Capability = "api_keys.manage"
	CapSuppressionsWrite  Capability = "suppressions.write"
	CapWebhooksWrite      Capability = "webhooks.write"
)

const (
	WorkspaceRoleAdmin     = "admin"
	WorkspaceRoleDeveloper = "developer"
	WorkspaceRoleViewer    = "viewer"
)

var ErrCapabilityForbidden = errors.New("forbidden for this role")

var capabilities = map[string]map[Capability]bool{
	WorkspaceRoleOwner: {
		CapMembersManage: true, CapWorkspaceDelete: true,
		CapProjectSettings: true, CapTemplatesWrite: true, CapSubscribersWrite: true,
		CapSendTransactional: true, CapBroadcast: true, CapCampaignsWrite: true,
		CapAPIKeysManage: true, CapSuppressionsWrite: true, CapWebhooksWrite: true,
	},
	WorkspaceRoleAdmin: {
		CapProjectSettings: true, CapTemplatesWrite: true, CapSubscribersWrite: true,
		CapSendTransactional: true, CapBroadcast: true, CapCampaignsWrite: true,
		CapAPIKeysManage: true, CapSuppressionsWrite: true, CapWebhooksWrite: true,
	},
	WorkspaceRoleDeveloper: {
		CapSendTransactional: true,
	},
	WorkspaceRoleMember: {
		CapProjectSettings: true, CapTemplatesWrite: true, CapSubscribersWrite: true,
		CapSendTransactional: true, CapBroadcast: true, CapCampaignsWrite: true,
		CapAPIKeysManage: true, CapSuppressionsWrite: true, CapWebhooksWrite: true,
	},
	WorkspaceRoleViewer: {},
}

func (s *ProjectService) roleForProject(ctx context.Context, projectID, userID uuid.UUID) (string, error) {
	return s.queries.GetProjectMemberRole(ctx, db.GetProjectMemberRoleParams{
		ID:     projectID,
		UserID: userID,
	})
}

func (s *ProjectService) RequireCapability(ctx context.Context, projectID, userID, cap string) error {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return ErrWorkspaceForbidden
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return ErrWorkspaceForbidden
	}
	role, err := s.roleForProject(ctx, pid, uid)
	if err != nil {
		return ErrWorkspaceForbidden
	}
	if !RoleHasCapability(role, Capability(cap)) {
		return ErrCapabilityForbidden
	}
	return nil
}

func RoleHasCapability(role string, cap Capability) bool {
	caps, ok := capabilities[role]
	if !ok {
		return false
	}
	return caps[cap]
}
