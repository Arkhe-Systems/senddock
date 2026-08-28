package service

import (
	"context"
	"errors"
	"strings"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/google/uuid"
)

type NewsletterService struct {
	queries *db.Queries
}

func NewNewsletterService(queries *db.Queries) *NewsletterService {
	return &NewsletterService{queries: queries}
}

func (s *NewsletterService) Create(ctx context.Context, projectID, name, description string) (db.Newsletter, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Newsletter{}, errors.New("invalid project id")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return db.Newsletter{}, errors.New("name is required")
	}
	return s.queries.CreateNewsletter(ctx, db.CreateNewsletterParams{
		ProjectID:   pid,
		Name:        name,
		Description: strings.TrimSpace(description),
	})
}

func (s *NewsletterService) List(ctx context.Context, projectID string) ([]db.ListNewslettersByProjectRow, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, errors.New("invalid project id")
	}
	return s.queries.ListNewslettersByProject(ctx, pid)
}

func (s *NewsletterService) Get(ctx context.Context, newsletterID, projectID string) (db.Newsletter, error) {
	nid, err := uuid.Parse(newsletterID)
	if err != nil {
		return db.Newsletter{}, errors.New("invalid newsletter id")
	}
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Newsletter{}, errors.New("invalid project id")
	}
	return s.queries.GetNewsletterByID(ctx, db.GetNewsletterByIDParams{ID: nid, ProjectID: pid})
}

func (s *NewsletterService) Update(ctx context.Context, newsletterID, projectID, name, description string) (db.Newsletter, error) {
	nid, err := uuid.Parse(newsletterID)
	if err != nil {
		return db.Newsletter{}, errors.New("invalid newsletter id")
	}
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Newsletter{}, errors.New("invalid project id")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return db.Newsletter{}, errors.New("name is required")
	}
	return s.queries.UpdateNewsletter(ctx, db.UpdateNewsletterParams{
		ID:          nid,
		ProjectID:   pid,
		Name:        name,
		Description: strings.TrimSpace(description),
	})
}

func (s *NewsletterService) Delete(ctx context.Context, newsletterID, projectID string) error {
	nid, err := uuid.Parse(newsletterID)
	if err != nil {
		return errors.New("invalid newsletter id")
	}
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return errors.New("invalid project id")
	}
	return s.queries.DeleteNewsletter(ctx, db.DeleteNewsletterParams{ID: nid, ProjectID: pid})
}

func (s *NewsletterService) ListForSubscriber(ctx context.Context, projectID, subscriberID string) ([]db.ListSubscriberNewslettersRow, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, errors.New("invalid project id")
	}
	sid, err := uuid.Parse(subscriberID)
	if err != nil {
		return nil, errors.New("invalid subscriber id")
	}
	return s.queries.ListSubscriberNewsletters(ctx, db.ListSubscriberNewslettersParams{SubscriberID: sid, ProjectID: pid})
}

func (s *NewsletterService) SetSubscriberNewsletters(ctx context.Context, projectID, subscriberID string, newsletterIDs []string) error {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return errors.New("invalid project id")
	}
	sid, err := uuid.Parse(subscriberID)
	if err != nil {
		return errors.New("invalid subscriber id")
	}
	parsed := make([]uuid.UUID, 0, len(newsletterIDs))
	for _, raw := range newsletterIDs {
		nid, err := uuid.Parse(raw)
		if err != nil {
			return errors.New("invalid newsletter id")
		}
		parsed = append(parsed, nid)
	}
	if err := s.queries.DeleteSubscriberNewsletterSubscriptions(ctx, db.DeleteSubscriberNewsletterSubscriptionsParams{
		SubscriberID: sid,
		ProjectID:    pid,
	}); err != nil {
		return err
	}
	for _, nid := range parsed {
		if err := s.queries.AddNewsletterSubscription(ctx, db.AddNewsletterSubscriptionParams{
			SubscriberID: sid,
			NewsletterID: nid,
			ProjectID:    pid,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *NewsletterService) BulkAdd(ctx context.Context, projectID, newsletterID string, subscriberIDs []string) error {
	pid, nid, ids, err := parseBulkNewsletterArgs(projectID, newsletterID, subscriberIDs)
	if err != nil {
		return err
	}
	return s.queries.BulkAddNewsletterSubscriptions(ctx, db.BulkAddNewsletterSubscriptionsParams{
		NewsletterID:  nid,
		ProjectID:     pid,
		SubscriberIds: ids,
	})
}

func (s *NewsletterService) BulkRemove(ctx context.Context, projectID, newsletterID string, subscriberIDs []string) error {
	pid, nid, ids, err := parseBulkNewsletterArgs(projectID, newsletterID, subscriberIDs)
	if err != nil {
		return err
	}
	return s.queries.BulkRemoveNewsletterSubscriptions(ctx, db.BulkRemoveNewsletterSubscriptionsParams{
		NewsletterID:  nid,
		ProjectID:     pid,
		SubscriberIds: ids,
	})
}

func parseBulkNewsletterArgs(projectID, newsletterID string, subscriberIDs []string) (uuid.UUID, uuid.UUID, []uuid.UUID, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return uuid.Nil, uuid.Nil, nil, errors.New("invalid project id")
	}
	nid, err := uuid.Parse(newsletterID)
	if err != nil {
		return uuid.Nil, uuid.Nil, nil, errors.New("invalid newsletter id")
	}
	if len(subscriberIDs) == 0 {
		return uuid.Nil, uuid.Nil, nil, errors.New("subscriber_ids is required")
	}
	ids := make([]uuid.UUID, 0, len(subscriberIDs))
	for _, raw := range subscriberIDs {
		sid, err := uuid.Parse(raw)
		if err != nil {
			return uuid.Nil, uuid.Nil, nil, errors.New("invalid subscriber id")
		}
		ids = append(ids, sid)
	}
	return pid, nid, ids, nil
}
