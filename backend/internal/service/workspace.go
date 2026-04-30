package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/google/uuid"
)

const (
	WorkspaceRoleOwner  = "owner"
	WorkspaceRoleMember = "member"
)

var (
	ErrWorkspaceNotFound      = errors.New("workspace not found")
	ErrWorkspaceForbidden     = errors.New("forbidden")
	ErrWorkspaceOwnerRequired = errors.New("workspace owner required")
	ErrWorkspaceHasProjects   = errors.New("workspace still has projects")
	ErrInvalidRole            = errors.New("invalid role")
	ErrUserNotFound           = errors.New("user not found")
	ErrLastOwner              = errors.New("cannot remove the last owner")
)

type WorkspaceService struct {
	queries *db.Queries
}

func NewWorkspaceService(queries *db.Queries) *WorkspaceService {
	return &WorkspaceService{queries: queries}
}

func (s *WorkspaceService) Create(ctx context.Context, userID uuid.UUID, name string) (db.Workspace, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return db.Workspace{}, errors.New("name is required")
	}
	ws, err := s.queries.CreateWorkspace(ctx, db.CreateWorkspaceParams{
		Name:      name,
		CreatedBy: userID,
	})
	if err != nil {
		return db.Workspace{}, err
	}
	if _, err := s.queries.AddWorkspaceMember(ctx, db.AddWorkspaceMemberParams{
		WorkspaceID: ws.ID,
		UserID:      userID,
		Role:        WorkspaceRoleOwner,
		InvitedBy:   uuid.NullUUID{UUID: userID, Valid: true},
	}); err != nil {
		return db.Workspace{}, err
	}
	return ws, nil
}

func (s *WorkspaceService) ListByUser(ctx context.Context, userID uuid.UUID) ([]db.Workspace, error) {
	return s.queries.ListWorkspacesByUser(ctx, userID)
}

func (s *WorkspaceService) Get(ctx context.Context, workspaceID, userID uuid.UUID) (db.Workspace, error) {
	if err := s.requireMember(ctx, workspaceID, userID); err != nil {
		return db.Workspace{}, err
	}
	ws, err := s.queries.GetWorkspaceByID(ctx, workspaceID)
	if errors.Is(err, sql.ErrNoRows) {
		return db.Workspace{}, ErrWorkspaceNotFound
	}
	return ws, err
}

func (s *WorkspaceService) Rename(ctx context.Context, workspaceID, userID uuid.UUID, name string) (db.Workspace, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return db.Workspace{}, errors.New("name is required")
	}
	if err := s.requireOwner(ctx, workspaceID, userID); err != nil {
		return db.Workspace{}, err
	}
	return s.queries.RenameWorkspace(ctx, db.RenameWorkspaceParams{ID: workspaceID, Name: name})
}

func (s *WorkspaceService) Delete(ctx context.Context, workspaceID, userID uuid.UUID) error {
	if err := s.requireOwner(ctx, workspaceID, userID); err != nil {
		return err
	}
	count, err := s.queries.CountProjectsInWorkspace(ctx, workspaceID)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrWorkspaceHasProjects
	}
	return s.queries.DeleteWorkspace(ctx, workspaceID)
}

type MemberView struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt string    `json:"joined_at"`
}

func (s *WorkspaceService) ListMembers(ctx context.Context, workspaceID, userID uuid.UUID) ([]MemberView, error) {
	if err := s.requireMember(ctx, workspaceID, userID); err != nil {
		return nil, err
	}
	rows, err := s.queries.ListWorkspaceMembers(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	out := make([]MemberView, 0, len(rows))
	for _, r := range rows {
		out = append(out, MemberView{
			UserID:    r.UserID,
			Email:     r.UserEmail,
			Name:      r.UserName,
			Role:      r.Role,
			CreatedAt: r.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return out, nil
}

func (s *WorkspaceService) AddMember(ctx context.Context, workspaceID, actorID uuid.UUID, email, role string) (MemberView, error) {
	role = normalizeRole(role)
	if role == "" {
		return MemberView{}, ErrInvalidRole
	}
	if err := s.requireOwner(ctx, workspaceID, actorID); err != nil {
		return MemberView{}, err
	}
	user, err := s.queries.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if errors.Is(err, sql.ErrNoRows) {
		return MemberView{}, ErrUserNotFound
	}
	if err != nil {
		return MemberView{}, err
	}
	if _, err := s.queries.AddWorkspaceMember(ctx, db.AddWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      user.ID,
		Role:        role,
		InvitedBy:   uuid.NullUUID{UUID: actorID, Valid: true},
	}); err != nil {
		return MemberView{}, err
	}
	return MemberView{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		Role:   role,
	}, nil
}

func (s *WorkspaceService) UpdateMemberRole(ctx context.Context, workspaceID, actorID, targetID uuid.UUID, role string) error {
	role = normalizeRole(role)
	if role == "" {
		return ErrInvalidRole
	}
	if err := s.requireOwner(ctx, workspaceID, actorID); err != nil {
		return err
	}
	if role == WorkspaceRoleMember {
		if err := s.guardLastOwner(ctx, workspaceID, targetID); err != nil {
			return err
		}
	}
	_, err := s.queries.UpdateWorkspaceMemberRole(ctx, db.UpdateWorkspaceMemberRoleParams{
		WorkspaceID: workspaceID,
		UserID:      targetID,
		Role:        role,
	})
	return err
}

func (s *WorkspaceService) RemoveMember(ctx context.Context, workspaceID, actorID, targetID uuid.UUID) error {
	if actorID != targetID {
		if err := s.requireOwner(ctx, workspaceID, actorID); err != nil {
			return err
		}
	}
	if err := s.guardLastOwner(ctx, workspaceID, targetID); err != nil {
		return err
	}
	return s.queries.RemoveWorkspaceMember(ctx, db.RemoveWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      targetID,
	})
}

func (s *WorkspaceService) IsMember(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
	return s.queries.IsWorkspaceMember(ctx, db.IsWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
}

func (s *WorkspaceService) requireMember(ctx context.Context, workspaceID, userID uuid.UUID) error {
	ok, err := s.IsMember(ctx, workspaceID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrWorkspaceForbidden
	}
	return nil
}

func (s *WorkspaceService) requireOwner(ctx context.Context, workspaceID, userID uuid.UUID) error {
	m, err := s.queries.GetWorkspaceMember(ctx, db.GetWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return ErrWorkspaceForbidden
	}
	if err != nil {
		return err
	}
	if m.Role != WorkspaceRoleOwner {
		return ErrWorkspaceOwnerRequired
	}
	return nil
}

func (s *WorkspaceService) guardLastOwner(ctx context.Context, workspaceID, targetID uuid.UUID) error {
	target, err := s.queries.GetWorkspaceMember(ctx, db.GetWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      targetID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if target.Role != WorkspaceRoleOwner {
		return nil
	}
	owners, err := s.queries.CountWorkspaceOwners(ctx, workspaceID)
	if err != nil {
		return err
	}
	if owners <= 1 {
		return ErrLastOwner
	}
	return nil
}

func normalizeRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case WorkspaceRoleOwner:
		return WorkspaceRoleOwner
	case "", WorkspaceRoleMember:
		return WorkspaceRoleMember
	default:
		return ""
	}
}
