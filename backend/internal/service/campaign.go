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

func (s *CampaignService) Create(ctx context.Context, projectID, templateID, name string, scheduledAt time.Time, variables json.RawMessage) (db.Campaign, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Campaign{}, errors.New("invalid project id")
	}

	tid, err := uuid.Parse(templateID)
	if err != nil {
		return db.Campaign{}, errors.New("invalid template id")
	}

	if scheduledAt.Before(time.Now()) {
		return db.Campaign{}, errors.New("scheduled time must be in the future")
	}

	return s.queries.CreateCampaign(ctx, db.CreateCampaignParams{
		ProjectID:   pid,
		TemplateID:  tid,
		Name:        name,
		ScheduledAt: scheduledAt,
		Variables:   variables,
	})
}

func (s *CampaignService) Update(ctx context.Context, campaignID, projectID, templateID, name string, scheduledAt time.Time, variables json.RawMessage) (db.Campaign, error) {
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

	if scheduledAt.Before(time.Now()) {
		return db.Campaign{}, errors.New("scheduled time must be in the future")
	}

	return s.queries.UpdateCampaign(ctx, db.UpdateCampaignParams{
		ID:          cid,
		ProjectID:   pid,
		TemplateID:  tid,
		Name:        name,
		ScheduledAt: scheduledAt,
		Variables:   variables,
	})
}

func (s *CampaignService) ListByProject(ctx context.Context, projectID string) ([]db.Campaign, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, errors.New("invalid project id")
	}
	return s.queries.ListCampaignsByProject(ctx, pid)
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
	return s.queries.DeleteCampaign(ctx, db.DeleteCampaignParams{
		ID:        cid,
		ProjectID: pid,
	})
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
