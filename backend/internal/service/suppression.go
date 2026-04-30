package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/google/uuid"
)

type SuppressionService struct {
	queries *db.Queries
}

func NewSuppressionService(queries *db.Queries) *SuppressionService {
	return &SuppressionService{queries: queries}
}

const (
	SuppressionReasonUnsubscribe = "unsubscribe"
	SuppressionReasonBounce      = "bounce"
	SuppressionReasonComplaint   = "complaint"
	SuppressionReasonManual      = "manual"
)

func (s *SuppressionService) Add(ctx context.Context, projectID uuid.UUID, email, reason, source string) (db.Suppression, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return db.Suppression{}, errors.New("email is required")
	}
	return s.queries.UpsertSuppression(ctx, db.UpsertSuppressionParams{
		ProjectID:       projectID,
		EmailNormalized: normalized,
		Reason:          reason,
		Source:          sql.NullString{String: source, Valid: source != ""},
	})
}

func (s *SuppressionService) IsSuppressed(ctx context.Context, projectID uuid.UUID, email string) bool {
	if strings.TrimSpace(email) == "" {
		return false
	}
	suppressed, err := s.queries.IsSuppressed(ctx, db.IsSuppressedParams{
		ProjectID: projectID,
		Lower:     email,
	})
	if err != nil {
		return false
	}
	return suppressed
}

func (s *SuppressionService) List(ctx context.Context, projectID uuid.UUID, reason string, limit, offset int32) ([]db.Suppression, int64, error) {
	rows, err := s.queries.ListSuppressionsByProject(ctx, db.ListSuppressionsByProjectParams{
		ProjectID: projectID,
		Column2:   reason,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return nil, 0, err
	}
	count, err := s.queries.CountSuppressionsByProject(ctx, db.CountSuppressionsByProjectParams{
		ProjectID: projectID,
		Column2:   reason,
	})
	if err != nil {
		return nil, 0, err
	}
	return rows, count, nil
}

func (s *SuppressionService) Remove(ctx context.Context, projectID, suppressionID uuid.UUID) error {
	return s.queries.DeleteSuppression(ctx, db.DeleteSuppressionParams{
		ID:        suppressionID,
		ProjectID: projectID,
	})
}

func (s *SuppressionService) RemoveByEmail(ctx context.Context, projectID uuid.UUID, email string) error {
	return s.queries.DeleteSuppressionByEmail(ctx, db.DeleteSuppressionByEmailParams{
		ProjectID: projectID,
		Lower:     email,
	})
}
