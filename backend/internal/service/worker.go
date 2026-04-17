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
		return
	}

	for _, campaign := range campaigns {
		w.executeCampaign(ctx, campaign)
	}
}

func (w *CampaignWorker) executeCampaign(ctx context.Context, campaign db.Campaign) {
	log.Printf("Executing campaign %s: %s", campaign.ID.String(), campaign.Name)

	w.queries.UpdateCampaignStatus(ctx, db.UpdateCampaignStatusParams{
		ID:     campaign.ID,
		Status: "sending",
	})

	result, err := w.emailService.Broadcast(ctx, campaign.ProjectID.String(), campaign.TemplateID.String())

	status := "sent"
	if err != nil {
		status = "failed"
		log.Printf("Campaign %s failed: %v", campaign.ID.String(), err)
	}

	w.queries.UpdateCampaignStatus(ctx, db.UpdateCampaignStatusParams{
		ID:          campaign.ID,
		Status:      status,
		SentCount:   int32(result.Sent),
		FailedCount: int32(result.Failed),
	})

	log.Printf("Campaign %s completed: %d sent, %d failed", campaign.ID.String(), result.Sent, result.Failed)
}
