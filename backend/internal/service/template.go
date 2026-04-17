package service

import (
	"context"
	"errors"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/google/uuid"
)

type TemplateService struct {
	queries *db.Queries
}

func NewTemplateService(queries *db.Queries) *TemplateService {
	return &TemplateService{queries: queries}
}

func (s *TemplateService) Create(ctx context.Context, projectID, name, subject, htmlBody, textBody string) (db.Template, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Template{}, errors.New("invalid project id")
	}

	return s.queries.CreateTemplate(ctx, db.CreateTemplateParams{
		ProjectID: pid,
		Name:      name,
		Subject:   subject,
		HtmlBody:  htmlBody,
		TextBody:  textBody,
	})
}

func (s *TemplateService) ListByProject(ctx context.Context, projectID string) ([]db.Template, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, errors.New("invalid project id")
	}

	return s.queries.ListTemplatesByProject(ctx, pid)
}

func (s *TemplateService) GetByID(ctx context.Context, templateID, projectID string) (db.Template, error) {
	tid, err := uuid.Parse(templateID)
	if err != nil {
		return db.Template{}, errors.New("invalid template id")
	}

	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Template{}, errors.New("invalid project id")
	}

	return s.queries.GetTemplateByID(ctx, db.GetTemplateByIDParams{
		ID:        tid,
		ProjectID: pid,
	})
}

func (s *TemplateService) Update(ctx context.Context, templateID, projectID, name, subject, htmlBody, textBody string) (db.Template, error) {
	tid, err := uuid.Parse(templateID)
	if err != nil {
		return db.Template{}, errors.New("invalid template id")
	}

	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Template{}, errors.New("invalid project id")
	}

	return s.queries.UpdateTemplate(ctx, db.UpdateTemplateParams{
		ID:        tid,
		ProjectID: pid,
		Name:      name,
		Subject:   subject,
		HtmlBody:  htmlBody,
		TextBody:  textBody,
	})
}

func (s *TemplateService) Delete(ctx context.Context, templateID, projectID string) error {
	tid, err := uuid.Parse(templateID)
	if err != nil {
		return errors.New("invalid template id")
	}

	pid, err := uuid.Parse(projectID)
	if err != nil {
		return errors.New("invalid project id")
	}

	return s.queries.DeleteTemplate(ctx, db.DeleteTemplateParams{
		ID:        tid,
		ProjectID: pid,
	})
}
