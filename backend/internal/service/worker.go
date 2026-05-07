package service

import (
	"context"
	"log"
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
			w.processPending()
		}
	}()
	log.Println("Campaign worker started (checking every 30s)")
}

func (w *CampaignWorker) processPending() {
	ctx := context.Background()

	campaigns, err := w.queries.GetPendingCampaigns(ctx)
	if err != nil {
		log.Printf("campaign worker: list pending failed: %v", err)
		return
	}

	for _, campaign := range campaigns {
		w.executeCampaign(ctx, campaign)
	}
}

func (w *CampaignWorker) executeCampaign(ctx context.Context, campaign db.Campaign) {
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
		log.Printf("campaign %s failed: %v", campaign.ID, runErr)
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
