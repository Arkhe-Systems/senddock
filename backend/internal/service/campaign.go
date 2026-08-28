package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/google/uuid"
)

type CampaignService struct {
	queries *db.Queries
}

func NewCampaignService(queries *db.Queries) *CampaignService {
	return &CampaignService{queries: queries}
}

func parseOptionalNewsletter(newsletterID string) (uuid.NullUUID, error) {
	if newsletterID == "" {
		return uuid.NullUUID{}, nil
	}
	nid, err := uuid.Parse(newsletterID)
	if err != nil {
		return uuid.NullUUID{}, errors.New("invalid newsletter id")
	}
	return uuid.NullUUID{UUID: nid, Valid: true}, nil
}

func (s *CampaignService) Create(ctx context.Context, projectID, templateID, name, subject string, scheduledAt time.Time, variables json.RawMessage, newsletterID string) (db.Campaign, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Campaign{}, errors.New("invalid project id")
	}

	tid, err := uuid.Parse(templateID)
	if err != nil {
		return db.Campaign{}, errors.New("invalid template id")
	}

	if scheduledAt.Before(time.Now().Add(-1 * time.Minute)) {
		return db.Campaign{}, errors.New("scheduled time must be in the future")
	}

	nlid, err := parseOptionalNewsletter(newsletterID)
	if err != nil {
		return db.Campaign{}, err
	}

	return s.queries.CreateCampaign(ctx, db.CreateCampaignParams{
		ProjectID:    pid,
		TemplateID:   tid,
		Name:         name,
		Subject:      subject,
		ScheduledAt:  scheduledAt,
		Variables:    variables,
		NewsletterID: nlid,
	})
}

func (s *CampaignService) Update(ctx context.Context, campaignID, projectID, templateID, name, subject string, scheduledAt time.Time, variables json.RawMessage, newsletterID string) (db.Campaign, error) {
	cid, err := uuid.Parse(campaignID)
	if err != nil {
		return db.Campaign{}, errors.New("invalid campaign id")
	}

	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Campaign{}, errors.New("invalid project id")
	}

	tid, err := uuid.Parse(templateID)
	if err != nil {
		return db.Campaign{}, errors.New("invalid template id")
	}

	if scheduledAt.Before(time.Now().Add(-1 * time.Minute)) {
		return db.Campaign{}, errors.New("scheduled time must be in the future")
	}

	nlid, err := parseOptionalNewsletter(newsletterID)
	if err != nil {
		return db.Campaign{}, err
	}

	return s.queries.UpdateCampaign(ctx, db.UpdateCampaignParams{
		ID:           cid,
		ProjectID:    pid,
		TemplateID:   tid,
		Name:         name,
		Subject:      subject,
		ScheduledAt:  scheduledAt,
		Variables:    variables,
		NewsletterID: nlid,
	})
}

func (s *CampaignService) ListByProject(ctx context.Context, projectID string) ([]db.Campaign, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, errors.New("invalid project id")
	}
	rows, err := s.queries.ListCampaignsByProject(ctx, pid)
	if err != nil {
		return nil, err
	}
	out := make([]db.Campaign, len(rows))
	for i, r := range rows {
		out[i] = db.Campaign{
			ID:          r.ID,
			ProjectID:   r.ProjectID,
			TemplateID:  r.TemplateID,
			Name:        r.Name,
			Subject:     r.Subject,
			ScheduledAt: r.ScheduledAt,
			SentAt:      r.SentAt,
			CreatedAt:   r.CreatedAt,
			Status:      r.Status,
			SentCount:   r.SentCount,
			FailedCount: r.FailedCount,
			Variables:   r.Variables,
			BroadcastID: r.BroadcastID,
		}
	}
	return out, nil
}

func (s *CampaignService) Delete(ctx context.Context, campaignID, projectID string) error {
	cid, err := uuid.Parse(campaignID)
	if err != nil {
		return errors.New("invalid campaign id")
	}
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return errors.New("invalid project id")
	}
	rows, err := s.queries.DeleteCampaign(ctx, db.DeleteCampaignParams{
		ID:        cid,
		ProjectID: pid,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("campaign not found")
	}
	return nil
}

func (s *CampaignService) GetPending(ctx context.Context) ([]db.Campaign, error) {
	return s.queries.GetPendingCampaigns(ctx)
}

func (s *CampaignService) MarkCompleted(ctx context.Context, campaignID uuid.UUID, sent, failed int) error {
	return s.queries.UpdateCampaignStatus(ctx, db.UpdateCampaignStatusParams{
		ID:          campaignID,
		Status:      "sent",
		SentCount:   int32(sent),
		FailedCount: int32(failed),
	})
}
