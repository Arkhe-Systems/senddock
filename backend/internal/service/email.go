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

type EmailService struct {
	queries   *db.Queries
	publicURL string
	encSecret string
	cache     *cache.Redis
	hooks     WebhookDispatcher
}

func NewEmailService(queries *db.Queries, publicURL, encSecret string, redis *cache.Redis, hooks WebhookDispatcher) *EmailService {
	return &EmailService{queries: queries, publicURL: publicURL, encSecret: encSecret, cache: redis, hooks: hooks}
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
	return fmt.Sprintf("%s/unsubscribe/%s/%s?t=%s", s.publicURL, projectID, subscriberID, s.signUnsub(projectID, subscriberID))
}

type SendResult struct {
	Sent   int `json:"sent"`
	Failed int `json:"failed"`
}

func (s *EmailService) SendToSubscriber(ctx context.Context, projectID, subscriberID, templateID string) (SendResult, error) {
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

	body := ensureUnsubscribeFooter(template.HtmlBody)
	body = replaceVariables(body, sub)
	subject := replaceVariablesSimple(template.Subject, sub)

	unsubURL := s.unsubURL(pid.String(), sid.String())
	body = strings.ReplaceAll(body, "{{unsubscribe_url}}", unsubURL)

	sendErr := s.trackAndSend(ctx, project, pid, uuid.NullUUID{UUID: sid, Valid: true}, uuid.NullUUID{UUID: tid, Valid: true}, sub.Email, subject, body, unsubURL)
	if sendErr != nil {
		return SendResult{Failed: 1}, sendErr
	}
	return SendResult{Sent: 1}, nil
}

func (s *EmailService) Broadcast(ctx context.Context, projectID, templateID, subjectOverride string, campaignVars json.RawMessage) (SendResult, error) {
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

	if !IsPublicURLReachable(s.publicURL) {
		return SendResult{}, errors.New("PUBLIC_URL is not set to a publicly reachable URL. Newsletters need a working unsubscribe link before they can be sent. Set PUBLIC_URL in your .env to your public domain and restart the server")
	}

	tid, err := uuid.Parse(templateID)
	if err != nil {
		return SendResult{}, errors.New("invalid template id")
	}

	template, err := s.queries.GetTemplateByID(ctx, db.GetTemplateByIDParams{ID: tid, ProjectID: pid})
	if err != nil {
		return SendResult{}, errors.New("template not found")
	}

	subscribers, err := s.queries.ListActiveSubscribersByProject(ctx, pid)
	if err != nil {
		return SendResult{}, err
	}

	if len(subscribers) == 0 {
		return SendResult{}, errors.New("no active subscribers")
	}

	total := len(subscribers)

	var customVars map[string]string
	if len(campaignVars) > 0 {
		_ = json.Unmarshal(campaignVars, &customVars)
	}

	templateBody := ensureUnsubscribeFooter(template.HtmlBody)
	baseSubject := template.Subject
	if subjectOverride != "" {
		baseSubject = subjectOverride
	}

	go func() {
		bgCtx := context.Background()
		for _, sub := range subscribers {
			body := templateBody
			subject := baseSubject

			for k, v := range customVars {
				body = strings.ReplaceAll(body, "{{"+k+"}}", html.EscapeString(v))
				subject = strings.ReplaceAll(subject, "{{"+k+"}}", v)
			}

			body = replaceVariables(body, sub)
			subject = replaceVariablesSimple(subject, sub)

			unsubURL := s.unsubURL(pid.String(), sub.ID.String())
			body = strings.ReplaceAll(body, "{{unsubscribe_url}}", unsubURL)

			s.trackAndSend(bgCtx, project, pid, uuid.NullUUID{UUID: sub.ID, Valid: true}, uuid.NullUUID{UUID: tid, Valid: true}, sub.Email, subject, body, unsubURL)
		}
		log.Printf("Broadcast to %d subscribers completed for project %s", total, pid.String())
	}()

	return SendResult{Sent: total}, nil
}

func (s *EmailService) SendDirect(ctx context.Context, projectID, to, subject, htmlBody string) error {
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

	return s.trackAndSend(ctx, project, pid, uuid.NullUUID{}, uuid.NullUUID{}, to, subject, htmlBody, "")
}

func (s *EmailService) SendWithTemplate(ctx context.Context, projectID, templateID, to, subjectOverride string, variables map[string]string) error {
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
	if strings.Contains(body, "{{unsubscribe_url}}") {
		sub, err := s.queries.GetSubscriberByEmail(ctx, db.GetSubscriberByEmailParams{Email: to, ProjectID: pid})
		if err == nil {
			unsubscribeURL = s.unsubURL(pid.String(), sub.ID.String())
			body = strings.ReplaceAll(body, "{{name}}", html.EscapeString(sub.Name))
			body = strings.ReplaceAll(body, "{{email}}", html.EscapeString(sub.Email))
			body = strings.ReplaceAll(body, "{{subscriber_id}}", sub.ID.String())
		}
		body = strings.ReplaceAll(body, "{{unsubscribe_url}}", unsubscribeURL)
	}

	return s.trackAndSend(ctx, project, pid, uuid.NullUUID{}, uuid.NullUUID{UUID: tid, Valid: true}, to, subject, body, unsubscribeURL)
}

func (s *EmailService) GetLogs(ctx context.Context, projectID string, limit, offset int32, status, from, to string) ([]db.EmailLog, int64, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, 0, errors.New("invalid project id")
	}

	var fromTime, toTime time.Time
	if from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			fromTime = t
		}
	}
	if to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			toTime = t
		}
	}

	if status != "" || !fromTime.IsZero() || !toTime.IsZero() {
		logs, err := s.queries.ListEmailLogsByProjectFiltered(ctx, db.ListEmailLogsByProjectFilteredParams{
			ProjectID: pid,
			Limit:     limit,
			Offset:    offset,
			Column4:   status,
			Column5:   fromTime,
			Column6:   toTime,
		})
		if err != nil {
			return nil, 0, err
		}

		count, _ := s.queries.CountEmailLogsByProjectFiltered(ctx, db.CountEmailLogsByProjectFilteredParams{
			ProjectID: pid,
			Column2:   status,
			Column3:   fromTime,
			Column4:   toTime,
		})

		return logs, count, nil
	}

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

func (s *EmailService) GetStats(ctx context.Context, projectID string) (map[string]int64, error) {
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
	opened, _ := s.queries.CountEmailLogsOpened(ctx, pid)

	stats := map[string]int64{
		"total":  total,
		"sent":   sent,
		"failed": failed,
		"opened": opened,
	}

	s.cache.Set(ctx, cacheKey, stats, 30*time.Second)
	return stats, nil
}

func (s *EmailService) logPending(ctx context.Context, projectID uuid.UUID, subscriberID, templateID uuid.NullUUID, toEmail, subject string) uuid.UUID {
	logEntry, _ := s.queries.CreateEmailLog(ctx, db.CreateEmailLogParams{
		ProjectID:    projectID,
		SubscriberID: subscriberID,
		TemplateID:   templateID,
		ToEmail:      toEmail,
		Subject:      subject,
		Status:       "sent",
	})
	s.cache.Delete(ctx, "stats:"+projectID.String())
	return logEntry.ID
}

func (s *EmailService) markLogFailed(ctx context.Context, projectID, logID uuid.UUID, sendErr error) {
	if sendErr == nil || logID == uuid.Nil {
		return
	}
	s.queries.UpdateEmailLogStatus(ctx, db.UpdateEmailLogStatusParams{
		ID:     logID,
		Status: "failed",
		Error:  sql.NullString{String: sendErr.Error(), Valid: true},
	})
	s.cache.Delete(ctx, "stats:"+projectID.String())
}

func (s *EmailService) trackAndSend(ctx context.Context, project db.Project, projectID uuid.UUID, subscriberID, templateID uuid.NullUUID, to, subject, body, unsubscribeURL string) error {
	logID := s.logPending(ctx, projectID, subscriberID, templateID, to, subject)
	body = s.injectTrackingPixel(body, logID)
	body = s.rewriteLinksForTracking(body, logID)
	sendErr := s.sendSMTP(project, to, subject, body, unsubscribeURL)
	if sendErr != nil {
		s.markLogFailed(ctx, projectID, logID, sendErr)
		s.dispatchEmail(ctx, "email.failed", projectID, logID, to, subject, sendErr.Error())
	} else {
		s.dispatchEmail(ctx, "email.sent", projectID, logID, to, subject, "")
	}
	return sendErr
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
	pixel := fmt.Sprintf(`<img src="%s/t/%s.gif" width="1" height="1" style="display:none" />`, s.publicURL, logID.String())
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
	return fmt.Sprintf("%s/c/%s/%s.%s", s.publicURL, logID.String(), encoded, sig)
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
		if strings.HasPrefix(trimmed, s.publicURL+"/unsubscribe/") || strings.HasPrefix(trimmed, s.publicURL+"/c/") || strings.HasPrefix(trimmed, s.publicURL+"/t/") {
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
