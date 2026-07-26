package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log"
	"net"
	"net/smtp"
	"net/textproto"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/arkhe-systems/senddock/internal/cache"
	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/arkhe-systems/senddock/internal/webhooks"
	"github.com/google/uuid"
	premailer "github.com/vanng822/go-premailer/premailer"
)

const unsubscribeFooter = `<hr style="margin:32px 0 16px;border:none;border-top:1px solid #e5e5e5"><p style="font-size:12px;color:#888;text-align:center;font-family:Arial,sans-serif;line-height:1.5">Don't want to receive these emails? <a href="{{unsubscribe_url}}" style="color:#888;text-decoration:underline">Unsubscribe</a>.</p>`

func IsPublicURLReachable(rawURL string) bool {
	if rawURL == "" {
		return false
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		return false
	}
	if host == "localhost" || strings.HasPrefix(host, "127.") || host == "::1" || host == "0.0.0.0" {
		return false
	}
	return true
}

func ensureUnsubscribeFooter(body string) string {
	if strings.Contains(body, "{{unsubscribe_url}}") {
		return body
	}
	if idx := strings.LastIndex(strings.ToLower(body), "</body>"); idx != -1 {
		return body[:idx] + unsubscribeFooter + body[idx:]
	}
	return body + unsubscribeFooter
}

type WebhookDispatcher interface {
	Enqueue(ctx context.Context, event webhooks.Event)
}

type PublicURLProvider interface {
	PublicURL() string
}

type EmailService struct {
	queries      *db.Queries
	settings     PublicURLProvider
	encSecret    string
	cache        *cache.Redis
	hooks        WebhookDispatcher
	suppressions *SuppressionService
	segments     *SegmentService
}

func (s *EmailService) SetSegmentService(seg *SegmentService) {
	s.segments = seg
}

func NewEmailService(queries *db.Queries, settings PublicURLProvider, encSecret string, redis *cache.Redis, hooks WebhookDispatcher, suppressions *SuppressionService) *EmailService {
	return &EmailService{queries: queries, settings: settings, encSecret: encSecret, cache: redis, hooks: hooks, suppressions: suppressions}
}

func (s *EmailService) publicURL() string {
	return s.settings.PublicURL()
}

func (s *EmailService) signUnsub(projectID, subscriberID string) string {
	mac := hmac.New(sha256.New, []byte(s.encSecret))
	mac.Write([]byte(projectID + ":" + subscriberID))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *EmailService) verifyUnsubToken(projectID, subscriberID, token string) bool {
	expected := s.signUnsub(projectID, subscriberID)
	return hmac.Equal([]byte(expected), []byte(token))
}

func (s *EmailService) unsubURL(projectID, subscriberID string) string {
	return fmt.Sprintf("%s/unsubscribe/%s/%s?t=%s", s.publicURL(), projectID, subscriberID, s.signUnsub(projectID, subscriberID))
}

type SendResult struct {
	Sent        int        `json:"sent"`
	Failed      int        `json:"failed"`
	Suppressed  int        `json:"suppressed,omitempty"`
	BroadcastID *uuid.UUID `json:"broadcast_id,omitempty"`
}

var ErrRecipientSuppressed = errors.New("recipient is on the project suppression list")

type BounceError struct {
	Code    int
	Message string
}

func (e *BounceError) Error() string {
	return fmt.Sprintf("smtp bounce %d: %s", e.Code, e.Message)
}

func classifyBounce(err error) *BounceError {
	if err == nil {
		return nil
	}
	var protoErr *textproto.Error
	if !errors.As(err, &protoErr) {
		return nil
	}
	if protoErr.Code < 500 || protoErr.Code >= 600 {
		return nil
	}
	return &BounceError{Code: protoErr.Code, Message: protoErr.Msg}
}

func (s *EmailService) logSuppressed(ctx context.Context, projectID uuid.UUID, subscriberID, templateID uuid.NullUUID, to, subject string, broadcastID uuid.NullUUID) {
	_, _ = s.queries.CreateEmailLog(ctx, db.CreateEmailLogParams{
		ProjectID:    projectID,
		SubscriberID: subscriberID,
		TemplateID:   templateID,
		ToEmail:      to,
		Subject:      subject,
		Status:       "suppressed",
		Error:        sql.NullString{String: "recipient on suppression list", Valid: true},
		BroadcastID:  broadcastID,
	})
}

func (s *EmailService) SendToSubscriber(ctx context.Context, projectID, subscriberID, templateID string) (SendResult, error) {
	if err := requireAuthorizedProject(ctx, projectID); err != nil {
		return SendResult{}, err
	}
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return SendResult{}, errors.New("invalid project id")
	}

	project, err := s.queries.GetProjectByIDOnly(ctx, pid)
	if err != nil {
		return SendResult{}, errors.New("project not found")
	}

	if !project.SmtpHost.Valid || !project.SmtpUser.Valid || !project.SmtpPasswordEncrypted.Valid {
		return SendResult{}, errors.New("smtp not configured")
	}

	tid, err := uuid.Parse(templateID)
	if err != nil {
		return SendResult{}, errors.New("invalid template id")
	}

	template, err := s.queries.GetTemplateByID(ctx, db.GetTemplateByIDParams{ID: tid, ProjectID: pid})
	if err != nil {
		return SendResult{}, errors.New("template not found")
	}

	sid, err := uuid.Parse(subscriberID)
	if err != nil {
		return SendResult{}, errors.New("invalid subscriber id")
	}

	sub, err := s.queries.GetSubscriberByID(ctx, db.GetSubscriberByIDParams{ID: sid, ProjectID: pid})
	if err != nil {
		return SendResult{}, errors.New("subscriber not found")
	}

	if sub.Status != "active" {
		return SendResult{}, errors.New("subscriber is not active")
	}

	if s.suppressions != nil && s.suppressions.IsSuppressed(ctx, pid, sub.Email) {
		s.logSuppressed(ctx, pid, uuid.NullUUID{UUID: sid, Valid: true}, uuid.NullUUID{UUID: tid, Valid: true}, sub.Email, template.Subject, uuid.NullUUID{})
		return SendResult{Suppressed: 1}, nil
	}

	body := ensureUnsubscribeFooter(template.HtmlBody)
	body = replaceVariables(body, sub)
	subject := replaceVariablesSimple(template.Subject, sub)

	unsubURL := s.unsubURL(pid.String(), sid.String())
	body = strings.ReplaceAll(body, "{{unsubscribe_url}}", unsubURL)

	sendErr := s.trackAndSend(ctx, project, pid, uuid.NullUUID{UUID: sid, Valid: true}, uuid.NullUUID{UUID: tid, Valid: true}, sub.Email, subject, body, unsubURL, uuid.NullUUID{})
	if sendErr != nil {
		return SendResult{Failed: 1}, sendErr
	}
	return SendResult{Sent: 1}, nil
}

func (s *EmailService) Broadcast(ctx context.Context, projectID, templateID, subjectOverride string, campaignVars json.RawMessage, segmentID string) (SendResult, error) {
	if err := requireAuthorizedProject(ctx, projectID); err != nil {
		return SendResult{}, err
	}
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return SendResult{}, errors.New("invalid project id")
	}

	project, err := s.queries.GetProjectByIDOnly(ctx, pid)
	if err != nil {
		return SendResult{}, errors.New("project not found")
	}

	if !project.SmtpHost.Valid || !project.SmtpUser.Valid || !project.SmtpPasswordEncrypted.Valid {
		return SendResult{}, errors.New("smtp not configured")
	}

	if !IsPublicURLReachable(s.publicURL()) {
		return SendResult{}, errors.New("your public URL is not set to a publicly reachable address. Newsletters need a working unsubscribe link before they can be sent. Set it under Settings → Instance")
	}

	tid, err := uuid.Parse(templateID)
	if err != nil {
		return SendResult{}, errors.New("invalid template id")
	}

	template, err := s.queries.GetTemplateByID(ctx, db.GetTemplateByIDParams{ID: tid, ProjectID: pid})
	if err != nil {
		return SendResult{}, errors.New("template not found")
	}

	recipients := make([]SubscriberRef, 0)
	if segmentID != "" {
		if s.segments == nil {
			return SendResult{}, errors.New("segments not available")
		}
		refs, err := s.segments.RecipientsByID(ctx, projectID, segmentID, true)
		if err != nil {
			return SendResult{}, err
		}
		recipients = refs
	} else {
		subscribers, err := s.queries.ListActiveSubscribersByProject(ctx, pid)
		if err != nil {
			return SendResult{}, err
		}
		for _, sub := range subscribers {
			recipients = append(recipients, SubscriberRef{ID: sub.ID, Email: sub.Email})
		}
	}

	if len(recipients) == 0 {
		return SendResult{}, errors.New("no active subscribers")
	}

	total := len(recipients)

	baseSubject := template.Subject
	if subjectOverride != "" {
		baseSubject = subjectOverride
	}

	variablesJSON := campaignVars
	if len(variablesJSON) == 0 {
		variablesJSON = []byte("{}")
	}

	broadcast, err := s.queries.CreateBroadcast(ctx, db.CreateBroadcastParams{
		ProjectID:       pid,
		TemplateID:      tid,
		Subject:         baseSubject,
		Variables:       variablesJSON,
		TotalRecipients: int32(total),
	})
	if err != nil {
		return SendResult{}, fmt.Errorf("create broadcast: %w", err)
	}

	type jobRow struct {
		SubscriberID   uuid.UUID `json:"subscriber_id"`
		RecipientEmail string    `json:"recipient_email"`
	}
	jobs := make([]jobRow, 0, total)
	for _, ref := range recipients {
		jobs = append(jobs, jobRow{SubscriberID: ref.ID, RecipientEmail: ref.Email})
	}
	jobsJSON, err := json.Marshal(jobs)
	if err != nil {
		return SendResult{}, fmt.Errorf("marshal jobs: %w", err)
	}

	if err := s.queries.BulkInsertBroadcastJobs(ctx, db.BulkInsertBroadcastJobsParams{
		BroadcastID: broadcast.ID,
		ProjectID:   pid,
		Jobs:        jobsJSON,
	}); err != nil {
		return SendResult{}, fmt.Errorf("enqueue broadcast jobs: %w", err)
	}

	log.Printf("Broadcast %s enqueued: %d jobs for project %s", broadcast.ID, total, pid.String())
	bid := broadcast.ID
	return SendResult{Sent: total, BroadcastID: &bid}, nil
}

type BroadcastJobOutcome int

const (
	BroadcastJobOutcomeSent BroadcastJobOutcome = iota
	BroadcastJobOutcomeSuppressed
	BroadcastJobOutcomeBounced
	BroadcastJobOutcomeTransientError
)

func (s *EmailService) SendBroadcastJob(ctx context.Context, job db.BroadcastJob, broadcast db.Broadcast) (BroadcastJobOutcome, error) {
	project, err := s.queries.GetProjectByIDOnly(ctx, job.ProjectID)
	if err != nil {
		return BroadcastJobOutcomeTransientError, fmt.Errorf("project lookup: %w", err)
	}
	if !project.SmtpHost.Valid || !project.SmtpUser.Valid || !project.SmtpPasswordEncrypted.Valid {
		return BroadcastJobOutcomeTransientError, errors.New("smtp not configured")
	}

	template, err := s.queries.GetTemplateByID(ctx, db.GetTemplateByIDParams{ID: broadcast.TemplateID, ProjectID: job.ProjectID})
	if err != nil {
		return BroadcastJobOutcomeTransientError, fmt.Errorf("template lookup: %w", err)
	}

	sub, err := s.queries.GetSubscriberByID(ctx, db.GetSubscriberByIDParams{ID: job.SubscriberID, ProjectID: job.ProjectID})
	if err != nil {
		return BroadcastJobOutcomeBounced, fmt.Errorf("subscriber not found: %w", err)
	}

	if s.suppressions != nil && s.suppressions.IsSuppressed(ctx, job.ProjectID, sub.Email) {
		s.logSuppressed(ctx, job.ProjectID, uuid.NullUUID{UUID: sub.ID, Valid: true}, uuid.NullUUID{UUID: broadcast.TemplateID, Valid: true}, sub.Email, broadcast.Subject, uuid.NullUUID{UUID: broadcast.ID, Valid: true})
		return BroadcastJobOutcomeSuppressed, nil
	}

	var customVars map[string]string
	if len(broadcast.Variables) > 0 {
		_ = json.Unmarshal(broadcast.Variables, &customVars)
	}

	body := ensureUnsubscribeFooter(template.HtmlBody)
	subject := broadcast.Subject
	if subject == "" {
		subject = template.Subject
	}

	for k, v := range customVars {
		body = strings.ReplaceAll(body, "{{"+k+"}}", html.EscapeString(v))
		subject = strings.ReplaceAll(subject, "{{"+k+"}}", v)
	}

	body = replaceCustomVariables(body, sub, true)
	subject = replaceCustomVariables(subject, sub, false)

	body = replaceVariables(body, sub)
	subject = replaceVariablesSimple(subject, sub)

	unsubURL := s.unsubURL(job.ProjectID.String(), sub.ID.String())
	body = strings.ReplaceAll(body, "{{unsubscribe_url}}", unsubURL)

	sendErr := s.trackAndSend(ctx, project, job.ProjectID, uuid.NullUUID{UUID: sub.ID, Valid: true}, uuid.NullUUID{UUID: broadcast.TemplateID, Valid: true}, sub.Email, subject, body, unsubURL, uuid.NullUUID{UUID: broadcast.ID, Valid: true})
	if sendErr == nil {
		return BroadcastJobOutcomeSent, nil
	}

	var bounce *BounceError
	if errors.As(sendErr, &bounce) {
		return BroadcastJobOutcomeBounced, sendErr
	}

	return BroadcastJobOutcomeTransientError, sendErr
}

func (s *EmailService) RecoverInProgressBroadcasts(ctx context.Context) error {
	return s.queries.ReconcileStuckBroadcasts(ctx)
}

func (s *EmailService) ListBroadcasts(ctx context.Context, projectID string, limit, offset int32) ([]db.Broadcast, int64, error) {
	if err := requireAuthorizedProject(ctx, projectID); err != nil {
		return nil, 0, err
	}
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, 0, errors.New("invalid project id")
	}

	rows, err := s.queries.ListBroadcastsByProject(ctx, db.ListBroadcastsByProjectParams{
		ProjectID: pid,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return nil, 0, err
	}

	count, _ := s.queries.CountBroadcastsByProject(ctx, pid)
	return rows, count, nil
}

func (s *EmailService) SendDirect(ctx context.Context, projectID, to, subject, htmlBody string) error {
	if err := requireAuthorizedProject(ctx, projectID); err != nil {
		return err
	}
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return errors.New("invalid project id")
	}

	project, err := s.queries.GetProjectByIDOnly(ctx, pid)
	if err != nil {
		return errors.New("project not found")
	}

	if !project.SmtpHost.Valid || !project.SmtpUser.Valid || !project.SmtpPasswordEncrypted.Valid {
		return errors.New("smtp not configured")
	}

	if s.suppressions != nil && s.suppressions.IsSuppressed(ctx, pid, to) {
		s.logSuppressed(ctx, pid, uuid.NullUUID{}, uuid.NullUUID{}, to, subject, uuid.NullUUID{})
		return ErrRecipientSuppressed
	}

	return s.trackAndSend(ctx, project, pid, uuid.NullUUID{}, uuid.NullUUID{}, to, subject, htmlBody, "", uuid.NullUUID{})
}

func (s *EmailService) SendWithTemplate(ctx context.Context, projectID, templateID, to, subjectOverride string, variables map[string]string) error {
	if err := requireAuthorizedProject(ctx, projectID); err != nil {
		return err
	}
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return errors.New("invalid project id")
	}

	project, err := s.queries.GetProjectByIDOnly(ctx, pid)
	if err != nil {
		return errors.New("project not found")
	}

	if !project.SmtpHost.Valid || !project.SmtpUser.Valid || !project.SmtpPasswordEncrypted.Valid {
		return errors.New("smtp not configured")
	}

	tid, err := uuid.Parse(templateID)
	if err != nil {
		return errors.New("invalid template id")
	}

	template, err := s.queries.GetTemplateByID(ctx, db.GetTemplateByIDParams{ID: tid, ProjectID: pid})
	if err != nil {
		return errors.New("template not found")
	}

	if s.suppressions != nil && s.suppressions.IsSuppressed(ctx, pid, to) {
		s.logSuppressed(ctx, pid, uuid.NullUUID{}, uuid.NullUUID{UUID: tid, Valid: true}, to, template.Subject, uuid.NullUUID{})
		return ErrRecipientSuppressed
	}

	body := template.HtmlBody
	subject := template.Subject
	if subjectOverride != "" {
		subject = subjectOverride
	}
	for key, val := range variables {
		body = strings.ReplaceAll(body, "{{"+key+"}}", html.EscapeString(val))
		subject = strings.ReplaceAll(subject, "{{"+key+"}}", val)
	}

	var unsubscribeURL string
	needsSubscriber := strings.Contains(body, "{{unsubscribe_url}}") ||
		strings.Contains(body, "{{custom.") || strings.Contains(subject, "{{custom.")
	if needsSubscriber {
		sub, err := s.queries.GetSubscriberByEmail(ctx, db.GetSubscriberByEmailParams{Email: to, ProjectID: pid})
		if err == nil {
			unsubscribeURL = s.unsubURL(pid.String(), sub.ID.String())
			body = replaceCustomVariables(body, sub, true)
			subject = replaceCustomVariables(subject, sub, false)
			body = strings.ReplaceAll(body, "{{name}}", html.EscapeString(sub.Name))
			body = strings.ReplaceAll(body, "{{email}}", html.EscapeString(sub.Email))
			body = strings.ReplaceAll(body, "{{subscriber_id}}", sub.ID.String())
		}
		body = strings.ReplaceAll(body, "{{unsubscribe_url}}", unsubscribeURL)
	}

	return s.trackAndSend(ctx, project, pid, uuid.NullUUID{}, uuid.NullUUID{UUID: tid, Valid: true}, to, subject, body, unsubscribeURL, uuid.NullUUID{})
}

type LogFilters struct {
	Status     string
	From       string
	To         string
	Search     string
	TemplateID string
}

func (f LogFilters) parse() (string, time.Time, time.Time, string, uuid.UUID) {
	var fromTime, toTime time.Time
	if f.From != "" {
		if t, err := time.Parse(time.RFC3339, f.From); err == nil {
			fromTime = t
		}
	}
	if f.To != "" {
		if t, err := time.Parse(time.RFC3339, f.To); err == nil {
			toTime = t
		}
	}
	tid := uuid.Nil
	if f.TemplateID != "" {
		if u, err := uuid.Parse(f.TemplateID); err == nil {
			tid = u
		}
	}
	return strings.TrimSpace(f.Status), fromTime, toTime, strings.TrimSpace(f.Search), tid
}

func (f LogFilters) isEmpty() bool {
	status, fromT, toT, search, tid := f.parse()
	return status == "" && fromT.IsZero() && toT.IsZero() && search == "" && tid == uuid.Nil
}

func (s *EmailService) GetLogs(ctx context.Context, projectID string, limit, offset int32, filters LogFilters) ([]db.EmailLog, int64, error) {
	if err := requireAuthorizedProject(ctx, projectID); err != nil {
		return nil, 0, err
	}
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, 0, errors.New("invalid project id")
	}

	if filters.isEmpty() {
		logs, err := s.queries.ListEmailLogsByProject(ctx, db.ListEmailLogsByProjectParams{
			ProjectID: pid,
			Limit:     limit,
			Offset:    offset,
		})
		if err != nil {
			return nil, 0, err
		}
		count, _ := s.queries.CountEmailLogsByProject(ctx, pid)
		return logs, count, nil
	}

	status, fromT, toT, search, tid := filters.parse()

	logs, err := s.queries.ListEmailLogsByProjectFiltered(ctx, db.ListEmailLogsByProjectFilteredParams{
		ProjectID: pid,
		Limit:     limit,
		Offset:    offset,
		Column4:   status,
		Column5:   fromT,
		Column6:   toT,
		Column7:   search,
		Column8:   tid,
	})
	if err != nil {
		return nil, 0, err
	}

	count, _ := s.queries.CountEmailLogsByProjectFiltered(ctx, db.CountEmailLogsByProjectFilteredParams{
		ProjectID: pid,
		Column2:   status,
		Column3:   fromT,
		Column4:   toT,
		Column5:   search,
		Column6:   tid,
	})

	return logs, count, nil
}

func (s *EmailService) GetLogDetail(ctx context.Context, projectID, logID string) (db.EmailLog, []db.ListEmailClicksByLogRow, error) {
	if err := requireAuthorizedProject(ctx, projectID); err != nil {
		return db.EmailLog{}, nil, err
	}
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.EmailLog{}, nil, errors.New("invalid project id")
	}
	lid, err := uuid.Parse(logID)
	if err != nil {
		return db.EmailLog{}, nil, errors.New("invalid log id")
	}

	log, err := s.queries.GetEmailLog(ctx, db.GetEmailLogParams{ID: lid, ProjectID: pid})
	if err != nil {
		return db.EmailLog{}, nil, errors.New("log not found")
	}

	clicks, err := s.queries.ListEmailClicksByLog(ctx, lid)
	if err != nil {
		return log, nil, nil
	}
	return log, clicks, nil
}

func (s *EmailService) ExportLogs(ctx context.Context, projectID string, filters LogFilters) ([]db.EmailLog, error) {
	if err := requireAuthorizedProject(ctx, projectID); err != nil {
		return nil, err
	}
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, errors.New("invalid project id")
	}
	status, fromT, toT, search, tid := filters.parse()
	return s.queries.ListEmailLogsByProjectExport(ctx, db.ListEmailLogsByProjectExportParams{
		ProjectID: pid,
		Column2:   status,
		Column3:   fromT,
		Column4:   toT,
		Column5:   search,
		Column6:   tid,
	})
}

func (s *EmailService) GetStats(ctx context.Context, projectID string) (map[string]int64, error) {
	if err := requireAuthorizedProject(ctx, projectID); err != nil {
		return nil, err
	}
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, errors.New("invalid project id")
	}

	cacheKey := "stats:" + projectID
	var cached map[string]int64
	if s.cache.Get(ctx, cacheKey, &cached) {
		return cached, nil
	}

	total, _ := s.queries.CountEmailLogsByProject(ctx, pid)
	sent, _ := s.queries.CountEmailLogsByStatus(ctx, db.CountEmailLogsByStatusParams{ProjectID: pid, Status: "sent"})
	failed, _ := s.queries.CountEmailLogsByStatus(ctx, db.CountEmailLogsByStatusParams{ProjectID: pid, Status: "failed"})
	bounced, _ := s.queries.CountEmailLogsByStatus(ctx, db.CountEmailLogsByStatusParams{ProjectID: pid, Status: "bounced"})
	suppressed, _ := s.queries.CountEmailLogsByStatus(ctx, db.CountEmailLogsByStatusParams{ProjectID: pid, Status: "suppressed"})
	opened, _ := s.queries.CountEmailLogsOpened(ctx, pid)

	stats := map[string]int64{
		"total":      total,
		"sent":       sent,
		"failed":     failed,
		"bounced":    bounced,
		"suppressed": suppressed,
		"opened":     opened,
	}

	s.cache.Set(ctx, cacheKey, stats, 30*time.Second)
	return stats, nil
}

func (s *EmailService) logPending(ctx context.Context, projectID uuid.UUID, subscriberID, templateID uuid.NullUUID, toEmail, subject string, broadcastID uuid.NullUUID) uuid.UUID {
	logEntry, _ := s.queries.CreateEmailLog(ctx, db.CreateEmailLogParams{
		ProjectID:    projectID,
		SubscriberID: subscriberID,
		TemplateID:   templateID,
		ToEmail:      toEmail,
		Subject:      subject,
		Status:       "sent",
		BroadcastID:  broadcastID,
	})
	s.cache.Delete(ctx, "stats:"+projectID.String())
	return logEntry.ID
}

func (s *EmailService) markLogStatus(ctx context.Context, projectID, logID uuid.UUID, status string, sendErr error) {
	if logID == uuid.Nil {
		return
	}
	errStr := sql.NullString{}
	if sendErr != nil {
		errStr = sql.NullString{String: sendErr.Error(), Valid: true}
	}
	s.queries.UpdateEmailLogStatus(ctx, db.UpdateEmailLogStatusParams{
		ID:     logID,
		Status: status,
		Error:  errStr,
	})
	s.cache.Delete(ctx, "stats:"+projectID.String())
}

func (s *EmailService) markLogFailed(ctx context.Context, projectID, logID uuid.UUID, sendErr error) {
	if sendErr == nil {
		return
	}
	s.markLogStatus(ctx, projectID, logID, "failed", sendErr)
}

func (s *EmailService) trackAndSend(ctx context.Context, project db.Project, projectID uuid.UUID, subscriberID, templateID uuid.NullUUID, to, subject, body, unsubscribeURL string, broadcastID uuid.NullUUID) error {
	logID := s.logPending(ctx, projectID, subscriberID, templateID, to, subject, broadcastID)
	body = s.injectTrackingPixel(body, logID)
	body = s.rewriteLinksForTracking(body, logID)
	sendErr := s.sendSMTP(project, to, subject, body, unsubscribeURL)

	var bounce *BounceError
	if errors.As(sendErr, &bounce) {
		s.markLogStatus(ctx, projectID, logID, "bounced", sendErr)
		if s.suppressions != nil {
			source := fmt.Sprintf("smtp %d during rcpt: %s", bounce.Code, bounce.Message)
			_, _ = s.suppressions.Add(ctx, projectID, to, SuppressionReasonBounce, source)
		}
		s.dispatchBounceEvent(ctx, projectID, logID, to, subject, bounce)
		return sendErr
	}

	if sendErr != nil {
		s.markLogFailed(ctx, projectID, logID, sendErr)
		s.dispatchEmail(ctx, "email.failed", projectID, logID, to, subject, sendErr.Error())
	} else {
		s.dispatchEmail(ctx, "email.sent", projectID, logID, to, subject, "")
	}
	return sendErr
}

func (s *EmailService) dispatchBounceEvent(ctx context.Context, projectID, logID uuid.UUID, to, subject string, bounce *BounceError) {
	if s.hooks == nil {
		return
	}
	s.hooks.Enqueue(ctx, webhooks.Event{
		Type:      "email.bounced",
		ProjectID: projectID,
		Data: map[string]any{
			"log_id":     logID.String(),
			"project_id": projectID.String(),
			"to_email":   to,
			"subject":    subject,
			"smtp_code":  bounce.Code,
			"reason":     bounce.Message,
		},
	})
}

func (s *EmailService) dispatchEmail(ctx context.Context, eventType string, projectID, logID uuid.UUID, to, subject, errMsg string) {
	if s.hooks == nil {
		return
	}
	data := map[string]any{
		"log_id":     logID.String(),
		"project_id": projectID.String(),
		"to_email":   to,
		"subject":    subject,
	}
	if errMsg != "" {
		data["error"] = errMsg
	}
	s.hooks.Enqueue(ctx, webhooks.Event{
		Type:      eventType,
		ProjectID: projectID,
		Data:      data,
	})
}

func (s *EmailService) injectTrackingPixel(body string, logID uuid.UUID) string {
	pixel := fmt.Sprintf(`<img src="%s/t/%s.gif" width="1" height="1" style="display:none" />`, s.publicURL(), logID.String())
	if strings.Contains(body, "</body>") {
		return strings.Replace(body, "</body>", pixel+"</body>", 1)
	}
	return body + pixel
}

func (s *EmailService) signClickToken(logID, rawURL string) string {
	mac := hmac.New(sha256.New, []byte(s.encSecret))
	mac.Write([]byte(logID + ":" + rawURL))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))[:22]
}

func (s *EmailService) verifyClickToken(logID, rawURL, token string) bool {
	expected := s.signClickToken(logID, rawURL)
	return hmac.Equal([]byte(expected), []byte(token))
}

func (s *EmailService) clickURL(logID uuid.UUID, rawURL string) string {
	encoded := base64.RawURLEncoding.EncodeToString([]byte(rawURL))
	sig := s.signClickToken(logID.String(), rawURL)
	return fmt.Sprintf("%s/c/%s/%s.%s", s.publicURL(), logID.String(), encoded, sig)
}

func shortURLHash(rawURL string) string {
	h := sha256.Sum256([]byte(rawURL))
	return base64.RawURLEncoding.EncodeToString(h[:])[:16]
}

func (s *EmailService) ClickURLHash(rawURL string) string {
	return shortURLHash(rawURL)
}

func (s *EmailService) DecodeClickURL(encoded string) (string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func (s *EmailService) VerifyClickToken(logID, rawURL, token string) bool {
	return s.verifyClickToken(logID, rawURL, token)
}

func (s *EmailService) rewriteLinksForTracking(body string, logID uuid.UUID) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return body
	}
	doc.Find("a").Each(func(_ int, sel *goquery.Selection) {
		href, ok := sel.Attr("href")
		if !ok {
			return
		}
		trimmed := strings.TrimSpace(href)
		if trimmed == "" {
			return
		}
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "#") || strings.HasPrefix(lower, "mailto:") || strings.HasPrefix(lower, "tel:") {
			return
		}
		if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
			return
		}
		if strings.HasPrefix(trimmed, s.publicURL()+"/unsubscribe/") || strings.HasPrefix(trimmed, s.publicURL()+"/c/") || strings.HasPrefix(trimmed, s.publicURL()+"/t/") {
			return
		}
		sel.SetAttr("href", s.clickURL(logID, trimmed))
	})
	out, err := doc.Html()
	if err != nil {
		return body
	}
	return out
}

type UnsubscribeContext struct {
	ProjectName string
	Email       string
}

func (s *EmailService) ResolveUnsubscribe(ctx context.Context, projectID, subscriberID, token string) (UnsubscribeContext, error) {
	if !s.verifyUnsubToken(projectID, subscriberID, token) {
		return UnsubscribeContext{}, errors.New("invalid token")
	}

	sid, err := uuid.Parse(subscriberID)
	if err != nil {
		return UnsubscribeContext{}, errors.New("invalid subscriber id")
	}

	pid, err := uuid.Parse(projectID)
	if err != nil {
		return UnsubscribeContext{}, errors.New("invalid project id")
	}

	sub, err := s.queries.GetSubscriberByID(ctx, db.GetSubscriberByIDParams{ID: sid, ProjectID: pid})
	if err != nil {
		return UnsubscribeContext{}, errors.New("subscriber not found")
	}

	project, err := s.queries.GetProjectByIDOnly(ctx, pid)
	if err != nil {
		return UnsubscribeContext{}, errors.New("project not found")
	}

	return UnsubscribeContext{ProjectName: project.Name, Email: sub.Email}, nil
}

func (s *EmailService) Unsubscribe(ctx context.Context, projectID, subscriberID, token string) error {
	if !s.verifyUnsubToken(projectID, subscriberID, token) {
		return errors.New("invalid token")
	}

	sid, err := uuid.Parse(subscriberID)
	if err != nil {
		return errors.New("invalid subscriber id")
	}

	pid, err := uuid.Parse(projectID)
	if err != nil {
		return errors.New("invalid project id")
	}

	sub, err := s.queries.UpdateSubscriberStatus(ctx, db.UpdateSubscriberStatusParams{
		ID:        sid,
		ProjectID: pid,
		Status:    "unsubscribed",
		Column4:   "unsubscribed",
	})
	if err != nil {
		return err
	}
	if s.suppressions != nil {
		_, _ = s.suppressions.Add(ctx, pid, sub.Email, SuppressionReasonUnsubscribe, "unsubscribe link click")
	}
	if s.hooks != nil {
		s.hooks.Enqueue(ctx, webhooks.Event{
			Type:      "subscriber.unsubscribed",
			ProjectID: pid,
			Data: map[string]any{
				"subscriber_id": sub.ID.String(),
				"project_id":    pid.String(),
				"email":         sub.Email,
				"name":          sub.Name,
			},
		})
	}
	return nil
}

func (s *EmailService) TestSMTP(ctx context.Context, projectID string) error {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return errors.New("invalid project id")
	}

	project, err := s.queries.GetProjectByIDOnly(ctx, pid)
	if err != nil {
		return errors.New("project not found")
	}

	if !project.SmtpHost.Valid || !project.SmtpUser.Valid || !project.SmtpPasswordEncrypted.Valid {
		return errors.New("smtp not configured")
	}

	fromEmail := project.SmtpUser.String
	if project.FromEmail.Valid && project.FromEmail.String != "" {
		fromEmail = project.FromEmail.String
	}

	return s.sendSMTPWithTimeouts(project, fromEmail, "SendDock SMTP Test", "<h2>SMTP is working!</h2><p>Your SendDock SMTP configuration is correct.</p>", "", smtpTestConnectTimeout, smtpTestSessionTimeout)
}

func (s *EmailService) sendSMTP(project db.Project, to, subject, htmlBody, unsubscribeURL string) error {
	return s.sendSMTPWithTimeouts(project, to, subject, htmlBody, unsubscribeURL, smtpConnectTimeout, smtpSessionTimeout)
}

func (s *EmailService) sendSMTPWithTimeouts(project db.Project, to, subject, htmlBody, unsubscribeURL string, connectTimeout, sessionTimeout time.Duration) error {
	host := project.SmtpHost.String
	port := project.SmtpPort.Int32
	user := project.SmtpUser.String

	pass := project.SmtpPasswordEncrypted.String
	if decrypted, err := Decrypt(pass, s.encSecret); err == nil {
		pass = decrypted
	}

	fromEmail := user
	if project.FromEmail.Valid && project.FromEmail.String != "" {
		fromEmail = project.FromEmail.String
	}

	from := fromEmail
	if project.FromName.Valid && project.FromName.String != "" {
		from = fmt.Sprintf("%s <%s>", project.FromName.String, fromEmail)
	}

	inlinedBody := inlineCSS(htmlBody)

	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n",
		from, to, subject)
	if unsubscribeURL != "" {
		headers += fmt.Sprintf("List-Unsubscribe: <%s>\r\nList-Unsubscribe-Post: List-Unsubscribe=One-Click\r\n", unsubscribeURL)
	}
	msg := headers + "\r\n" + inlinedBody

	addr := fmt.Sprintf("%s:%d", host, port)

	return deliverSMTP(host, addr, user, pass, fromEmail, to, []byte(msg), port == 465, connectTimeout, sessionTimeout)
}

const (
	smtpConnectTimeout = 10 * time.Second
	smtpSessionTimeout = 30 * time.Second

	smtpTestConnectTimeout = 5 * time.Second
	smtpTestSessionTimeout = 10 * time.Second
)

func wrapDialError(addr string, connectTimeout time.Duration, err error) error {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return fmt.Errorf("could not reach SMTP server at %s within %s. The host or port may be wrong, or your network is blocking outbound SMTP (residential ISPs often block ports 25/465/587). Try from a cloud server or use a local SMTP catcher like Mailpit", addr, connectTimeout)
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if strings.Contains(err.Error(), "connection refused") {
			return fmt.Errorf("connection to %s refused. The SMTP server is up but not accepting connections on this port — confirm the port number with your provider", addr)
		}
		if strings.Contains(err.Error(), "no such host") {
			return fmt.Errorf("DNS lookup failed for SMTP host. Confirm the hostname is correct (no typos, no http:// prefix)")
		}
	}
	return fmt.Errorf("smtp connection failed: %w", err)
}

func deliverSMTP(host, addr, user, pass, from, to string, msg []byte, implicitTLS bool, connectTimeout, sessionTimeout time.Duration) error {
	tlsConfig := &tls.Config{ServerName: host, InsecureSkipVerify: true}
	dialer := &net.Dialer{Timeout: connectTimeout}

	var conn net.Conn
	var err error
	if implicitTLS {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	} else {
		conn, err = dialer.Dial("tcp", addr)
	}
	if err != nil {
		return wrapDialError(addr, connectTimeout, err)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(sessionTimeout)); err != nil {
		return fmt.Errorf("smtp set deadline failed: %w", err)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp client failed: %w", err)
	}
	defer client.Close()

	if !implicitTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err = client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("smtp starttls failed: %w", err)
			}
		}
	}

	auth := smtp.PlainAuth("", user, pass, host)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth failed: %w", err)
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail from failed: %w", err)
	}

	if err = client.Rcpt(to); err != nil {
		if be := classifyBounce(err); be != nil {
			return be
		}
		return fmt.Errorf("smtp rcpt to failed: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data failed: %w", err)
	}
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("smtp write failed: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("smtp close failed: %w", err)
	}

	return client.Quit()
}

func replaceVariables(body string, sub db.Subscriber) string {
	r := strings.NewReplacer(
		"{{name}}", html.EscapeString(sub.Name),
		"{{email}}", html.EscapeString(sub.Email),
		"{{subscriber_id}}", sub.ID.String(),
	)
	return r.Replace(body)
}

func replaceVariablesSimple(text string, sub db.Subscriber) string {
	r := strings.NewReplacer(
		"{{name}}", sub.Name,
		"{{email}}", sub.Email,
	)
	return r.Replace(text)
}

func replaceCustomVariables(text string, sub db.Subscriber, escape bool) string {
	if len(sub.Metadata) == 0 || !strings.Contains(text, "{{custom.") {
		return text
	}
	var fields map[string]any
	if err := json.Unmarshal(sub.Metadata, &fields); err != nil {
		return text
	}
	for key, value := range fields {
		rendered := fmt.Sprintf("%v", value)
		if escape {
			rendered = html.EscapeString(rendered)
		}
		text = strings.ReplaceAll(text, "{{custom."+key+"}}", rendered)
	}
	return text
}

func inlineCSS(html string) string {
	prem, err := premailer.NewPremailerFromString(html, premailer.NewOptions())
	if err != nil {
		return html
	}
	result, err := prem.Transform()
	if err != nil {
		return html
	}
	return result
}
