package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/arkhe-systems/senddock/pkg/segments"
	"github.com/google/uuid"
)

var ErrInvalidPredicate = segments.ErrInvalidPredicate

type SubscriberRef struct {
	ID    uuid.UUID
	Email string
}

type SegmentService struct {
	queries *db.Queries
	conn    *sql.DB
}

func NewSegmentService(queries *db.Queries, conn *sql.DB) *SegmentService {
	return &SegmentService{queries: queries, conn: conn}
}

func (s *SegmentService) matching(ctx context.Context, pid uuid.UUID, pred segments.Predicate, activeOnly bool) ([]SubscriberRef, error) {
	query := "SELECT id, email FROM subscribers WHERE project_id = $1"
	args := []any{pid}
	if activeOnly {
		query += " AND status = 'active'"
	}
	where, whereArgs := segments.BuildWhere(pred, len(args)+1)
	if where != "" {
		query += " AND (" + where + ")"
		args = append(args, whereArgs...)
	}
	query += " ORDER BY created_at DESC"

	rows, err := s.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	refs := []SubscriberRef{}
	for rows.Next() {
		var ref SubscriberRef
		if err := rows.Scan(&ref.ID, &ref.Email); err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	return refs, rows.Err()
}

func (s *SegmentService) Preview(ctx context.Context, projectID string, raw json.RawMessage, activeOnly bool) (int, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return 0, errors.New("invalid project id")
	}
	pred, err := segments.ParsePredicate(raw)
	if err != nil {
		return 0, err
	}
	refs, err := s.matching(ctx, pid, pred, activeOnly)
	if err != nil {
		return 0, err
	}
	return len(refs), nil
}

func (s *SegmentService) RecipientsByID(ctx context.Context, projectID, segmentID string, activeOnly bool) ([]SubscriberRef, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, errors.New("invalid project id")
	}
	sid, err := uuid.Parse(segmentID)
	if err != nil {
		return nil, errors.New("invalid segment id")
	}
	segment, err := s.queries.GetSegment(ctx, db.GetSegmentParams{ID: sid, ProjectID: pid})
	if err != nil {
		return nil, err
	}
	pred, err := segments.ParsePredicate(segment.Predicate)
	if err != nil {
		return nil, err
	}
	return s.matching(ctx, pid, pred, activeOnly)
}

func (s *SegmentService) Create(ctx context.Context, projectID, name string, raw json.RawMessage) (db.Segment, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Segment{}, errors.New("invalid project id")
	}
	if strings.TrimSpace(name) == "" {
		return db.Segment{}, errors.New("name is required")
	}
	if _, err := segments.ParsePredicate(raw); err != nil {
		return db.Segment{}, err
	}
	if len(raw) == 0 {
		raw = json.RawMessage("{}")
	}
	return s.queries.CreateSegment(ctx, db.CreateSegmentParams{
		ProjectID: pid,
		Name:      name,
		Predicate: raw,
	})
}

func (s *SegmentService) List(ctx context.Context, projectID string) ([]db.Segment, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, errors.New("invalid project id")
	}
	return s.queries.ListSegmentsByProject(ctx, pid)
}

func (s *SegmentService) Update(ctx context.Context, projectID, segmentID, name string, raw json.RawMessage) (db.Segment, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.Segment{}, errors.New("invalid project id")
	}
	sid, err := uuid.Parse(segmentID)
	if err != nil {
		return db.Segment{}, errors.New("invalid segment id")
	}
	if strings.TrimSpace(name) == "" {
		return db.Segment{}, errors.New("name is required")
	}
	if _, err := segments.ParsePredicate(raw); err != nil {
		return db.Segment{}, err
	}
	if len(raw) == 0 {
		raw = json.RawMessage("{}")
	}
	return s.queries.UpdateSegment(ctx, db.UpdateSegmentParams{
		ID:        sid,
		ProjectID: pid,
		Name:      name,
		Predicate: raw,
	})
}

func (s *SegmentService) Delete(ctx context.Context, projectID, segmentID string) error {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return errors.New("invalid project id")
	}
	sid, err := uuid.Parse(segmentID)
	if err != nil {
		return errors.New("invalid segment id")
	}
	return s.queries.DeleteSegment(ctx, db.DeleteSegmentParams{ID: sid, ProjectID: pid})
}
