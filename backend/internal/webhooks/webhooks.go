package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/google/uuid"
)

const (
	dispatchTickInterval = 10 * time.Second
	claimBatchSize       = 20
	maxAttempts          = 5
	httpTimeout          = 15 * time.Second
)

var backoff = []time.Duration{
	30 * time.Second,
	2 * time.Minute,
	10 * time.Minute,
	30 * time.Minute,
	2 * time.Hour,
}

type Service struct {
	queries *db.Queries
	client  *http.Client
}

func NewService(queries *db.Queries) *Service {
	return &Service{
		queries: queries,
		client:  &http.Client{Timeout: httpTimeout},
	}
}

type Event struct {
	Type      string
	ProjectID uuid.UUID
	Data      any
}

func (s *Service) Enqueue(ctx context.Context, event Event) {
	hooks, err := s.queries.ListActiveWebhooksForEvent(ctx, db.ListActiveWebhooksForEventParams{
		ProjectID: event.ProjectID,
		EventType: event.Type,
	})
	if err != nil {
		log.Printf("webhooks: list active hooks failed: %v", err)
		return
	}
	if len(hooks) == 0 {
		return
	}

	payload := map[string]any{
		"id":         "evt_" + uuid.NewString(),
		"type":       event.Type,
		"created_at": time.Now().UTC().Format(time.RFC3339),
		"data":       event.Data,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("webhooks: marshal payload failed: %v", err)
		return
	}

	for _, h := range hooks {
		_, err := s.queries.CreateWebhookDelivery(ctx, db.CreateWebhookDeliveryParams{
			WebhookID: h.ID,
			EventType: event.Type,
			Payload:   body,
		})
		if err != nil {
			log.Printf("webhooks: create delivery for %s failed: %v", h.ID, err)
		}
	}
}

func (s *Service) Start(ctx context.Context) {
	go s.run(ctx)
	log.Println("webhooks: dispatcher started")
}

func (s *Service) run(ctx context.Context) {
	t := time.NewTicker(dispatchTickInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.tick(ctx)
		}
	}
}

func (s *Service) tick(ctx context.Context) {
	deliveries, err := s.queries.ClaimPendingDeliveries(ctx, claimBatchSize)
	if err != nil {
		log.Printf("webhooks: claim pending failed: %v", err)
		return
	}
	if len(deliveries) == 0 {
		return
	}

	var wg sync.WaitGroup
	for _, d := range deliveries {
		wg.Add(1)
		go func(d db.WebhookDelivery) {
			defer wg.Done()
			s.deliver(ctx, d)
		}(d)
	}
	wg.Wait()
}

func (s *Service) deliver(ctx context.Context, delivery db.WebhookDelivery) {
	hook, err := s.queries.GetWebhookByIDOnly(ctx, delivery.WebhookID)
	if err != nil {
		s.markFailed(ctx, delivery.ID, 0, "webhook lookup failed: "+err.Error())
		return
	}

	if !hook.Active {
		s.markFailed(ctx, delivery.ID, 0, "webhook is inactive")
		return
	}

	statusCode, sendErr := s.send(ctx, hook, delivery.Payload)

	if sendErr == nil && statusCode >= 200 && statusCode < 300 {
		_ = s.queries.MarkDeliverySuccess(ctx, db.MarkDeliverySuccessParams{
			ID:             delivery.ID,
			LastStatusCode: sql.NullInt32{Int32: int32(statusCode), Valid: true},
		})
		return
	}

	errMsg := "non-2xx response"
	if sendErr != nil {
		errMsg = sendErr.Error()
	}

	nextAttempt := delivery.Attempts + 1
	if nextAttempt >= maxAttempts {
		s.markFailed(ctx, delivery.ID, statusCode, errMsg)
		return
	}

	idx := int(delivery.Attempts)
	if idx >= len(backoff) {
		idx = len(backoff) - 1
	}
	nextAt := time.Now().Add(backoff[idx])

	_ = s.queries.MarkDeliveryRetry(ctx, db.MarkDeliveryRetryParams{
		ID:             delivery.ID,
		LastStatusCode: sql.NullInt32{Int32: int32(statusCode), Valid: statusCode > 0},
		LastError:      sql.NullString{String: errMsg, Valid: true},
		NextAttemptAt:  nextAt,
	})
}

func (s *Service) markFailed(ctx context.Context, id uuid.UUID, statusCode int, errMsg string) {
	_ = s.queries.MarkDeliveryFailed(ctx, db.MarkDeliveryFailedParams{
		ID:             id,
		LastStatusCode: sql.NullInt32{Int32: int32(statusCode), Valid: statusCode > 0},
		LastError:      sql.NullString{String: errMsg, Valid: true},
	})
}

func (s *Service) send(ctx context.Context, hook db.Webhook, body []byte) (int, error) {
	timestamp := time.Now().Unix()
	mac := hmac.New(sha256.New, []byte(hook.Secret))
	fmt.Fprintf(mac, "%d.", timestamp)
	mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hook.Url, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "SendDock-Webhooks/1.0")
	req.Header.Set("X-SendDock-Signature", fmt.Sprintf("t=%d,v1=%s", timestamp, sig))

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, nil
}

func GenerateSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
