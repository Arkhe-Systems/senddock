package service

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/google/uuid"
)

type CampaignWorker struct {
	queries      *db.Queries
	emailService *EmailService
}

func NewCampaignWorker(queries *db.Queries, emailService *EmailService) *CampaignWorker {
	return &CampaignWorker{queries: queries, emailService: emailService}
}

func (w *CampaignWorker) Start() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			w.tick()
		}
	}()
	slog.Info("campaign worker started", "interval", "30s")
}

func (w *CampaignWorker) tick() {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("campaign worker tick panic", "panic", r, "stack", string(debug.Stack()))
		}
	}()

	ctx := context.Background()
	campaigns, err := w.queries.GetPendingCampaigns(ctx)
	if err != nil {
		slog.Error("campaign worker: list pending failed", "error", err)
		return
	}

	for _, campaign := range campaigns {
		w.ExecuteCampaign(ctx, campaign)
	}
}

func (w *CampaignWorker) ExecuteCampaign(ctx context.Context, campaign db.Campaign) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("campaign execute panic", "campaign_id", campaign.ID, "panic", r, "stack", string(debug.Stack()))
		}
	}()

	claimed, err := w.queries.ClaimCampaignForExecution(ctx, campaign.ID)
	if err != nil {
		slog.Error("campaign claim failed", "campaign_id", campaign.ID, "error", err)
		return
	}
	if claimed == 0 {
		return
	}

	slog.Info("executing campaign", "campaign_id", campaign.ID, "name", campaign.Name)

	result, runErr := w.emailService.Broadcast(
		WithSystemContext(ctx),
		campaign.ProjectID.String(),
		campaign.TemplateID.String(),
		campaign.Subject,
		campaign.Variables,
		nil,
		"",
	)

	if runErr != nil {
		slog.Error("campaign broadcast failed", "campaign_id", campaign.ID, "error", runErr)
		if err := w.queries.UpdateCampaignStatus(ctx, db.UpdateCampaignStatusParams{
			ID:          campaign.ID,
			Status:      "failed",
			SentCount:   0,
			FailedCount: 0,
		}); err != nil {
			slog.Error("campaign status update failed", "campaign_id", campaign.ID, "error", err)
		}
		return
	}

	if result.BroadcastID != nil {
		if err := w.queries.SetCampaignBroadcast(ctx, db.SetCampaignBroadcastParams{
			ID:          campaign.ID,
			BroadcastID: uuid.NullUUID{UUID: *result.BroadcastID, Valid: true},
		}); err != nil {
			slog.Error("campaign link to broadcast failed", "campaign_id", campaign.ID, "broadcast_id", *result.BroadcastID, "error", err)
		}
	}
	slog.Info("campaign linked to broadcast", "campaign_id", campaign.ID, "broadcast_id", result.BroadcastID, "recipients_enqueued", result.Sent)
}
