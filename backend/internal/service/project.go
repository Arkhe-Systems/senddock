package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/google/uuid"
)

type ProjectService struct {
	queries *db.Queries
}

func NewProjectService(queries *db.Queries) *ProjectService {
	return &ProjectService{queries: queries}
}

func (s *ProjectService) Create(ctx context.Context, userID, name, description string) (db.Project, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return db.Project{}, errors.New("invalid user id")
	}

	project, err := s.queries.CreateProject(ctx, db.CreateProjectParams{
		UserID:      uid,
		Name:        name,
		Description: sql.NullString{String: description, Valid: description != ""},
	})

	if err != nil {
		return db.Project{}, err
	}

	return project, nil
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

	return s.queries.UpdateProjectSMTP(ctx, db.UpdateProjectSMTPParams{
		ID:                    pid,
		UserID:                uid,
		SmtpHost:              sql.NullString{String: smtpHost, Valid: smtpHost != ""},
		SmtpPort:              sql.NullInt32{Int32: smtpPort, Valid: smtpPort != 0},
		SmtpUser:              sql.NullString{String: smtpUser, Valid: smtpUser != ""},
		SmtpPasswordEncrypted: sql.NullString{String: smtpPassword, Valid: smtpPassword != ""},
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
