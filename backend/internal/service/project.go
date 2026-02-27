package service

import (
	"context"
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

func (s *ProjectService) Create(ctx context.Context, userID, name, fromName, fromEmail string) (db.Project, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return db.Project{}, errors.New("invalid user id")
	}

	project, err := s.queries.CreateProject(ctx, db.CreateProjectParams{
		UserID:    uid,
		Name:      name,
		FromName:  fromName,
		FromEmail: fromEmail,
	})

	if err != nil {
		return db.Project{}, err
	}

	return project, nil
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
