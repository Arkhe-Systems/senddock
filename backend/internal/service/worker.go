package service

import (
	"context"
	"log"
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
	log.Println("Campaign worker started (checking every 30s)")
}

func (w *CampaignWorker) tick() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("campaign worker tick panic: %v\n%s", r, debug.Stack())
		}
	}()

	ctx := context.Background()
	campaigns, err := w.queries.GetPendingCampaigns(ctx)
	if err != nil {
		log.Printf("campaign worker: list pending failed: %v", err)
		return
	}

	for _, campaign := range campaigns {
		w.ExecuteCampaign(ctx, campaign)
	}
}

func (w *CampaignWorker) ExecuteCampaign(ctx context.Context, campaign db.Campaign) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("campaign %s: execute panic: %v\n%s", campaign.ID, r, debug.Stack())
		}
	}()

	claimed, err := w.queries.ClaimCampaignForExecution(ctx, campaign.ID)
	if err != nil {
		log.Printf("campaign %s: claim failed: %v", campaign.ID, err)
		return
	}
	if claimed == 0 {
		return
	}

	log.Printf("Executing campaign %s: %s", campaign.ID, campaign.Name)

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
		log.Printf("campaign %s: broadcast failed: %v", campaign.ID, runErr)
		if err := w.queries.UpdateCampaignStatus(ctx, db.UpdateCampaignStatusParams{
			ID:          campaign.ID,
			Status:      "failed",
			SentCount:   0,
			FailedCount: 0,
		}); err != nil {
			log.Printf("campaign %s: marking failed: %v", campaign.ID, err)
		}
		return
	}

	if result.BroadcastID != nil {
		if err := w.queries.SetCampaignBroadcast(ctx, db.SetCampaignBroadcastParams{
			ID:          campaign.ID,
			BroadcastID: uuid.NullUUID{UUID: *result.BroadcastID, Valid: true},
		}); err != nil {
			log.Printf("campaign %s: link to broadcast %s failed: %v", campaign.ID, *result.BroadcastID, err)
		}
	}
	log.Printf("Campaign %s linked to broadcast %v (%d recipients enqueued)", campaign.ID, result.BroadcastID, result.Sent)
}
