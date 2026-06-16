package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/google/uuid"
)

type ProjectService struct {
	queries   *db.Queries
	encSecret string
	quota     QuotaGate
}

func NewProjectService(queries *db.Queries, encSecret string) *ProjectService {
	return &ProjectService{queries: queries, encSecret: encSecret}
}

func (s *ProjectService) SetQuotaGate(g QuotaGate) {
	s.quota = g
}

func (s *ProjectService) Create(ctx context.Context, userID, workspaceID, name, description string) (db.Project, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return db.Project{}, errors.New("invalid user id")
	}

	wid, err := uuid.Parse(workspaceID)
	if err != nil {
		return db.Project{}, errors.New("invalid workspace id")
	}

	member, err := s.queries.IsWorkspaceMember(ctx, db.IsWorkspaceMemberParams{
		WorkspaceID: wid,
		UserID:      uid,
	})
	if err != nil {
		return db.Project{}, err
	}
	if !member {
		return db.Project{}, ErrWorkspaceForbidden
	}

	if s.quota != nil {
		if err := s.quota.AllowProject(ctx, workspaceID); err != nil {
			return db.Project{}, err
		}
	}

	project, err := s.queries.CreateProject(ctx, db.CreateProjectParams{
		WorkspaceID: wid,
		UserID:      uid,
		Name:        name,
		Description: sql.NullString{String: description, Valid: description != ""},
	})

	if err != nil {
		return db.Project{}, err
	}

	return project, nil
}

func (s *ProjectService) ListByWorkspace(ctx context.Context, workspaceID, userID string) ([]db.Project, error) {
	wid, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, errors.New("invalid workspace id")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}
	return s.queries.GetProjectsByWorkspaceForUser(ctx, db.GetProjectsByWorkspaceForUserParams{
		WorkspaceID: wid,
		UserID:      uid,
	})
}

func (s *ProjectService) Update(ctx context.Context, projectID, userID, name, description string) (db.Project, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Project{}, errors.New("invalid project id")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return db.Project{}, errors.New("invalid user id")
	}

	return s.queries.UpdateProject(ctx, db.UpdateProjectParams{
		ID:          pid,
		UserID:      uid,
		Name:        name,
		Description: sql.NullString{String: description, Valid: description != ""},
	})
}

func (s *ProjectService) ListByUser(ctx context.Context, userID string) ([]db.Project, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	return s.queries.GetProjectsByUserID(ctx, uid)
}

func (s *ProjectService) GetByID(ctx context.Context, projectID, userID string) (db.Project, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Project{}, errors.New("invalid project id")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return db.Project{}, errors.New("invalid user id")
	}

	return s.queries.GetProjectByID(ctx, db.GetProjectByIDParams{
		ID:     pid,
		UserID: uid,
	})
}

func (s *ProjectService) UpdateSMTP(ctx context.Context, projectID, userID, smtpHost string, smtpPort int32, smtpUser, smtpPassword, fromName, fromEmail string) (db.Project, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Project{}, errors.New("invalid project id")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return db.Project{}, errors.New("invalid user id")
	}

	encryptedPass := smtpPassword
	if smtpPassword != "" {
		enc, err := Encrypt(smtpPassword, s.encSecret)
		if err != nil {
			return db.Project{}, errors.New("failed to encrypt smtp password")
		}
		encryptedPass = enc
	}

	return s.queries.UpdateProjectSMTP(ctx, db.UpdateProjectSMTPParams{
		ID:                    pid,
		UserID:                uid,
		SmtpHost:              sql.NullString{String: smtpHost, Valid: smtpHost != ""},
		SmtpPort:              sql.NullInt32{Int32: smtpPort, Valid: smtpPort != 0},
		SmtpUser:              sql.NullString{String: smtpUser, Valid: smtpUser != ""},
		SmtpPasswordEncrypted: sql.NullString{String: encryptedPass, Valid: encryptedPass != ""},
		FromName:              sql.NullString{String: fromName, Valid: fromName != ""},
		FromEmail:             sql.NullString{String: fromEmail, Valid: fromEmail != ""},
	})
}

func (s *ProjectService) Delete(ctx context.Context, projectID, userID string) error {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return errors.New("invalid project id")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid user id")
	}

	return s.queries.DeleteProject(ctx, db.DeleteProjectParams{
		ID:     pid,
		UserID: uid,
	})
}

func (s *ProjectService) UpdateBounceIMAP(ctx context.Context, projectID, userID, host string, port int32, user, password, folder string, enabled bool) (db.Project, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Project{}, errors.New("invalid project id")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return db.Project{}, errors.New("invalid user id")
	}

	current, err := s.queries.GetProjectByID(ctx, db.GetProjectByIDParams{ID: pid, UserID: uid})
	if err != nil {
		return db.Project{}, errors.New("project not found")
	}

	encrypted := current.BounceImapPasswordEncrypted
	if password != "" {
		enc, err := Encrypt(password, s.encSecret)
		if err != nil {
			return db.Project{}, err
		}
		encrypted = sql.NullString{String: enc, Valid: true}
	}

	hostNS := sql.NullString{}
	if host != "" {
		hostNS = sql.NullString{String: host, Valid: true}
	}
	portNS := sql.NullInt32{}
	if port != 0 {
		portNS = sql.NullInt32{Int32: port, Valid: true}
	}
	userNS := sql.NullString{}
	if user != "" {
		userNS = sql.NullString{String: user, Valid: true}
	}
	if folder == "" {
		folder = "INBOX"
	}

	return s.queries.UpdateBounceIMAP(ctx, db.UpdateBounceIMAPParams{
		ID:                          pid,
		UserID:                      uid,
		BounceImapHost:              hostNS,
		BounceImapPort:              portNS,
		BounceImapUser:               userNS,
		BounceImapPasswordEncrypted: encrypted,
		BounceImapFolder:            folder,
		BounceImapEnabled:           enabled,
	})
}

func (s *ProjectService) RotateBounceToken(ctx context.Context, projectID, userID string) (db.Project, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Project{}, errors.New("invalid project id")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return db.Project{}, errors.New("invalid user id")
	}
	return s.queries.RotateBounceToken(ctx, db.RotateBounceTokenParams{ID: pid, UserID: uid})
}
