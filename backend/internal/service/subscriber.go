package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/arkhe-systems/senddock/internal/webhooks"
	"github.com/google/uuid"
)

type SubscriberService struct {
	queries   *db.Queries
	hooks     WebhookDispatcher
	validator *EmailValidator
}

func NewSubscriberService(queries *db.Queries, hooks WebhookDispatcher, validator *EmailValidator) *SubscriberService {
	return &SubscriberService{queries: queries, hooks: hooks, validator: validator}
}

func (s *SubscriberService) Create(ctx context.Context, projectID, email, name, status string) (db.Subscriber, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Subscriber{}, errors.New("invalid project id")
	}

	if status == "" {
		status = "active"
	}

	sub, err := s.queries.CreateSubscriber(ctx, db.CreateSubscriberParams{
		ProjectID: pid,
		Email:     email,
		Name:      name,
		Status:    status,
	})
	if err != nil {
		return sub, err
	}

	s.dispatch(ctx, "subscriber.created", sub)
	return sub, nil
}

func (s *SubscriberService) dispatch(ctx context.Context, eventType string, sub db.Subscriber) {
	if s.hooks == nil {
		return
	}
	s.hooks.Enqueue(ctx, webhooks.Event{
		Type:      eventType,
		ProjectID: sub.ProjectID,
		Data: map[string]any{
			"subscriber_id": sub.ID.String(),
			"project_id":    sub.ProjectID.String(),
			"email":         sub.Email,
			"name":          sub.Name,
			"status":        sub.Status,
		},
	})
}

type ImportResult struct {
	Imported      int           `json:"imported"`
	Duplicates    int           `json:"duplicates"`
	SyntaxInvalid int           `json:"syntax_invalid"`
	NoMX          int           `json:"no_mx"`
	Disposable    int           `json:"disposable"`
	Rejected      []RejectedRow `json:"rejected"`
}

type RejectedRow struct {
	Email  string `json:"email"`
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

type ImportSubscriber struct {
	Email  string `json:"email"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type ImportOptions struct {
	ValidateMX         bool
	ValidateDisposable bool
}

func (s *SubscriberService) BulkImport(ctx context.Context, projectID string, subscribers []ImportSubscriber, opts ImportOptions) (ImportResult, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return ImportResult{}, errors.New("invalid project id")
	}

	result := ImportResult{Rejected: []RejectedRow{}}
	mxCache := make(map[string]bool)

	for _, sub := range subscribers {
		raw := sub.Email
		normalized, ok := s.validator.Syntax(raw)
		if !ok {
			result.SyntaxInvalid++
			result.Rejected = append(result.Rejected, RejectedRow{Email: raw, Name: sub.Name, Reason: "syntax_invalid"})
			continue
		}

		if opts.ValidateDisposable && s.validator.IsDisposable(normalized) {
			result.Disposable++
			result.Rejected = append(result.Rejected, RejectedRow{Email: normalized, Name: sub.Name, Reason: "disposable"})
			continue
		}

		if opts.ValidateMX && !s.validator.HasMX(ctx, normalized, mxCache) {
			result.NoMX++
			result.Rejected = append(result.Rejected, RejectedRow{Email: normalized, Name: sub.Name, Reason: "no_mx"})
			continue
		}

		status := sub.Status
		if status == "" {
			status = "active"
		}
		_, err := s.queries.CreateSubscriber(ctx, db.CreateSubscriberParams{
			ProjectID: pid,
			Email:     normalized,
			Name:      sub.Name,
			Status:    status,
		})
		if err != nil {
			result.Duplicates++
			result.Rejected = append(result.Rejected, RejectedRow{Email: normalized, Name: sub.Name, Reason: "duplicate"})
			continue
		}
		result.Imported++
	}
	return result, nil
}

func (s *SubscriberService) ListByProject(ctx context.Context, projectID string, limit, offset int32) ([]db.Subscriber, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, errors.New("invalid project id")
	}

	return s.queries.ListSubscribersByProject(ctx, db.ListSubscribersByProjectParams{
		ProjectID: pid,
		Limit:     limit,
		Offset:    offset,
	})
}

func (s *SubscriberService) CountByProject(ctx context.Context, projectID string) (int64, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return 0, errors.New("invalid project id")
	}

	return s.queries.CountSubscribersByProject(ctx, pid)
}

func (s *SubscriberService) UpdateStatus(ctx context.Context, subscriberID, projectID, status string) (db.Subscriber, error) {
	if subscriberID == "" {
		return db.Subscriber{}, errors.New("subscriber id is empty")
	}

	sid, err := uuid.Parse(subscriberID)
	if err != nil {
		return db.Subscriber{}, fmt.Errorf("invalid subscriber id: %s", subscriberID)
	}

	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Subscriber{}, fmt.Errorf("invalid project id: %s", projectID)
	}

	sub, err := s.queries.UpdateSubscriberStatus(ctx, db.UpdateSubscriberStatusParams{
		ID:        sid,
		ProjectID: pid,
		Status:    status,
		Column4:   status,
	})
	if err != nil {
		return db.Subscriber{}, fmt.Errorf("update failed for %s in project %s: %w", subscriberID, projectID, err)
	}
	if status == "unsubscribed" {
		s.dispatch(ctx, "subscriber.unsubscribed", sub)
	}
	return sub, nil
}

func (s *SubscriberService) Delete(ctx context.Context, subscriberID, projectID string) error {
	sid, err := uuid.Parse(subscriberID)
	if err != nil {
		return errors.New("invalid subscriber id")
	}

	pid, err := uuid.Parse(projectID)
	if err != nil {
		return errors.New("invalid project id")
	}

	return s.queries.DeleteSubscriber(ctx, db.DeleteSubscriberParams{
		ID:        sid,
		ProjectID: pid,
	})
}

func (s *SubscriberService) BulkDelete(ctx context.Context, projectID string, ids []string) error {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return errors.New("invalid project id")
	}

	uuids := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if uid, err := uuid.Parse(id); err == nil {
			uuids = append(uuids, uid)
		}
	}

	if len(uuids) == 0 {
		return nil
	}

	return s.queries.BulkDeleteSubscribers(ctx, db.BulkDeleteSubscribersParams{
		ProjectID: pid,
		Column2:   uuids,
	})
}

func (s *SubscriberService) BulkUpdateStatus(ctx context.Context, projectID string, ids []string, status string) error {
	if status == "" {
		return errors.New("status is required")
	}

	pid, err := uuid.Parse(projectID)
	if err != nil {
		return errors.New("invalid project id")
	}

	uuids := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if uid, err := uuid.Parse(id); err == nil {
			uuids = append(uuids, uid)
		}
	}

	if len(uuids) == 0 {
		return nil
	}

	return s.queries.BulkUpdateSubscriberStatus(ctx, db.BulkUpdateSubscriberStatusParams{
		ProjectID: pid,
		Column2:   uuids,
		Status:    status,
	})
}
