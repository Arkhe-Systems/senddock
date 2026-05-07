package service

import (
	"context"
	"log"
	"runtime/debug"
	"time"

	"github.com/arkhe-systems/senddock/internal/db"
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
		ctx,
		campaign.ProjectID.String(),
		campaign.TemplateID.String(),
		campaign.Subject,
		campaign.Variables,
	)

	status := "sent"
	if runErr != nil {
		status = "failed"
		log.Printf("campaign %s: broadcast failed: %v", campaign.ID, runErr)
	}

	if err := w.queries.UpdateCampaignStatus(ctx, db.UpdateCampaignStatusParams{
		ID:          campaign.ID,
		Status:      status,
		SentCount:   int32(result.Sent),
		FailedCount: int32(result.Failed),
	}); err != nil {
		log.Printf("campaign %s: final status update failed (status=%s, sent=%d, failed=%d): %v",
			campaign.ID, status, result.Sent, result.Failed, err)
		return
	}

	log.Printf("Campaign %s completed: %d sent, %d failed", campaign.ID, result.Sent, result.Failed)
}
