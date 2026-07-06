package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/google/uuid"
)

type WebhookService struct {
	queries *db.Queries
}

func NewWebhookService(queries *db.Queries) *WebhookService {
	return &WebhookService{queries: queries}
}

func generateWebhookSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *WebhookService) Create(ctx context.Context, projectID, url string, events []string) (db.Webhook, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Webhook{}, errors.New("invalid project id")
	}
	secret, err := generateWebhookSecret()
	if err != nil {
		return db.Webhook{}, err
	}
	return s.queries.CreateWebhook(ctx, db.CreateWebhookParams{
		ProjectID: pid,
		Url:       url,
		Secret:    secret,
		Events:    events,
	})
}

func (s *WebhookService) List(ctx context.Context, projectID string) ([]db.Webhook, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, errors.New("invalid project id")
	}
	return s.queries.ListWebhooksByProject(ctx, pid)
}

func (s *WebhookService) Get(ctx context.Context, projectID, webhookID string) (db.Webhook, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Webhook{}, errors.New("invalid project id")
	}
	wid, err := uuid.Parse(webhookID)
	if err != nil {
		return db.Webhook{}, errors.New("invalid webhook id")
	}
	return s.queries.GetWebhook(ctx, db.GetWebhookParams{ID: wid, ProjectID: pid})
}

func (s *WebhookService) UpdateActive(ctx context.Context, projectID, webhookID string, active bool) (db.Webhook, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Webhook{}, errors.New("invalid project id")
	}
	wid, err := uuid.Parse(webhookID)
	if err != nil {
		return db.Webhook{}, errors.New("invalid webhook id")
	}
	return s.queries.UpdateWebhookActive(ctx, db.UpdateWebhookActiveParams{
		ID:        wid,
		ProjectID: pid,
		Active:    active,
	})
}

func (s *WebhookService) Delete(ctx context.Context, projectID, webhookID string) error {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return errors.New("invalid project id")
	}
	wid, err := uuid.Parse(webhookID)
	if err != nil {
		return errors.New("invalid webhook id")
	}
	return s.queries.DeleteWebhook(ctx, db.DeleteWebhookParams{ID: wid, ProjectID: pid})
}

func (s *WebhookService) ListDeliveries(ctx context.Context, projectID, webhookID string, limit int32) ([]db.WebhookDelivery, error) {
	if _, err := s.Get(ctx, projectID, webhookID); err != nil {
		return nil, err
	}
	wid, err := uuid.Parse(webhookID)
	if err != nil {
		return nil, errors.New("invalid webhook id")
	}
	return s.queries.ListWebhookDeliveries(ctx, db.ListWebhookDeliveriesParams{
		WebhookID: wid,
		Limit:     limit,
		Offset:    0,
	})
}
