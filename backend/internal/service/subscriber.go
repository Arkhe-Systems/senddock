package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/arkhe-systems/senddock/internal/webhooks"
	"github.com/google/uuid"
)

type SubscriberService struct {
	queries      *db.Queries
	hooks        WebhookDispatcher
	validator    *EmailValidator
	suppressions *SuppressionService
	quota        QuotaGate
	fields       *FieldDefinitionService
}

func NewSubscriberService(queries *db.Queries, hooks WebhookDispatcher, validator *EmailValidator, suppressions *SuppressionService) *SubscriberService {
	return &SubscriberService{queries: queries, hooks: hooks, validator: validator, suppressions: suppressions}
}

func (s *SubscriberService) SetQuotaGate(g QuotaGate) {
	s.quota = g
}

func (s *SubscriberService) SetFieldService(f *FieldDefinitionService) {
	s.fields = f
}

func (s *SubscriberService) validateFields(ctx context.Context, projectID string, fields map[string]any) (json.RawMessage, error) {
	if s.fields != nil {
		return s.fields.ValidateFields(ctx, projectID, fields)
	}
	if len(fields) == 0 {
		return json.RawMessage("{}"), nil
	}
	return json.Marshal(fields)
}

func (s *SubscriberService) Create(ctx context.Context, projectID, email, name, status string, fields map[string]any, tags []string) (db.Subscriber, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Subscriber{}, errors.New("invalid project id")
	}

	if status == "" {
		status = "active"
	}

	metadata, err := s.validateFields(ctx, projectID, fields)
	if err != nil {
		return db.Subscriber{}, err
	}

	if s.quota != nil {
		if err := s.quota.AllowSubscribers(ctx, projectID, 1); err != nil {
			return db.Subscriber{}, err
		}
	}

	sub, err := s.queries.CreateSubscriber(ctx, db.CreateSubscriberParams{
		ProjectID: pid,
		Email:     email,
		Name:      name,
		Status:    status,
		Metadata:  metadata,
		Tags:      normalizeTags(tags),
	})
	if err != nil {
		return sub, err
	}

	s.dispatch(ctx, "subscriber.created", sub)
	return sub, nil
}

func normalizeTags(tags []string) []string {
	result := make([]string, 0, len(tags))
	seen := make(map[string]bool, len(tags))
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		result = append(result, trimmed)
	}
	return result
}

func (s *SubscriberService) SetTags(ctx context.Context, subscriberID, projectID string, tags []string) (db.Subscriber, error) {
	sid, err := uuid.Parse(subscriberID)
	if err != nil {
		return db.Subscriber{}, errors.New("invalid subscriber id")
	}
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Subscriber{}, errors.New("invalid project id")
	}
	return s.queries.SetSubscriberTags(ctx, db.SetSubscriberTagsParams{
		ID:        sid,
		ProjectID: pid,
		Tags:      normalizeTags(tags),
	})
}

func (s *SubscriberService) ListTags(ctx context.Context, projectID string) ([]string, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, errors.New("invalid project id")
	}
	tags, err := s.queries.ListDistinctTagsByProject(ctx, pid)
	if err != nil {
		return nil, err
	}
	if tags == nil {
		return []string{}, nil
	}
	return tags, nil
}

func (s *SubscriberService) parseIDs(ids []string) []uuid.UUID {
	uuids := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if uid, err := uuid.Parse(id); err == nil {
			uuids = append(uuids, uid)
		}
	}
	return uuids
}

func (s *SubscriberService) BulkAddTags(ctx context.Context, projectID string, ids, tags []string) error {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return errors.New("invalid project id")
	}
	uuids := s.parseIDs(ids)
	if len(uuids) == 0 {
		return nil
	}
	return s.queries.BulkAddSubscriberTags(ctx, db.BulkAddSubscriberTagsParams{
		ProjectID: pid,
		Column2:   uuids,
		Column3:   normalizeTags(tags),
	})
}

func (s *SubscriberService) BulkRemoveTags(ctx context.Context, projectID string, ids, tags []string) error {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return errors.New("invalid project id")
	}
	uuids := s.parseIDs(ids)
	if len(uuids) == 0 {
		return nil
	}
	return s.queries.BulkRemoveSubscriberTags(ctx, db.BulkRemoveSubscriberTagsParams{
		ProjectID: pid,
		Column2:   uuids,
		Column3:   normalizeTags(tags),
	})
}

func (s *SubscriberService) UpdateFields(ctx context.Context, subscriberID, projectID string, fields map[string]any) (db.Subscriber, error) {
	sid, err := uuid.Parse(subscriberID)
	if err != nil {
		return db.Subscriber{}, errors.New("invalid subscriber id")
	}
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Subscriber{}, errors.New("invalid project id")
	}

	metadata, err := s.validateFields(ctx, projectID, fields)
	if err != nil {
		return db.Subscriber{}, err
	}

	return s.queries.UpdateSubscriberMetadata(ctx, db.UpdateSubscriberMetadataParams{
		ID:        sid,
		ProjectID: pid,
		Metadata:  metadata,
	})
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
	Suppressed    int           `json:"suppressed"`
	Rejected      []RejectedRow `json:"rejected"`
}

type RejectedRow struct {
	Email  string `json:"email"`
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

type ImportSubscriber struct {
	Email  string         `json:"email"`
	Name   string         `json:"name"`
	Status string         `json:"status"`
	Fields map[string]any `json:"fields"`
	Tags   []string       `json:"tags"`
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

		if s.suppressions != nil && s.suppressions.IsSuppressed(ctx, pid, normalized) {
			result.Suppressed++
			result.Rejected = append(result.Rejected, RejectedRow{Email: normalized, Name: sub.Name, Reason: "suppressed"})
			continue
		}

		status := sub.Status
		if status == "" {
			status = "active"
		}

		metadata, err := s.validateFields(ctx, projectID, sub.Fields)
		if err != nil {
			result.Rejected = append(result.Rejected, RejectedRow{Email: normalized, Name: sub.Name, Reason: err.Error()})
			continue
		}

		_, err = s.queries.CreateSubscriber(ctx, db.CreateSubscriberParams{
			ProjectID: pid,
			Email:     normalized,
			Name:      sub.Name,
			Status:    status,
			Metadata:  metadata,
			Tags:      normalizeTags(sub.Tags),
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

func (s *SubscriberService) ListByProject(ctx context.Context, projectID string, limit, offset int32, statusFilter, tagFilter, newsletterFilter string) ([]db.Subscriber, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, errors.New("invalid project id")
	}

	return s.queries.ListSubscribersByProject(ctx, db.ListSubscribersByProjectParams{
		ProjectID: pid,
		Limit:     limit,
		Offset:    offset,
		Column4:   strings.TrimSpace(statusFilter),
		Column5:   strings.TrimSpace(tagFilter),
		Column6:   parseFilterUUID(newsletterFilter),
	})
}

func parseFilterUUID(raw string) uuid.UUID {
	if raw == "" {
		return uuid.Nil
	}
	if u, err := uuid.Parse(raw); err == nil {
		return u
	}
	return uuid.Nil
}

func (s *SubscriberService) CountByProject(ctx context.Context, projectID string, statusFilter, tagFilter, newsletterFilter string) (int64, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return 0, errors.New("invalid project id")
	}

	return s.queries.CountSubscribersByProject(ctx, db.CountSubscribersByProjectParams{
		ProjectID: pid,
		Column2:   strings.TrimSpace(statusFilter),
		Column3:   strings.TrimSpace(tagFilter),
		Column4:   parseFilterUUID(newsletterFilter),
	})
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
		if s.suppressions != nil {
			_, _ = s.suppressions.Add(ctx, pid, sub.Email, SuppressionReasonUnsubscribe, "manual subscriber status change")
		}
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
