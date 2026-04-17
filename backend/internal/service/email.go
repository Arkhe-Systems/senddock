package service

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/google/uuid"
	premailer "github.com/vanng822/go-premailer/premailer"
)

type EmailService struct {
	queries  *db.Queries
	baseURL  string
}

func NewEmailService(queries *db.Queries, baseURL string) *EmailService {
	return &EmailService{queries: queries, baseURL: baseURL}
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

	body := replaceVariables(template.HtmlBody, sub)
	subject := replaceVariablesSimple(template.Subject, sub)

	unsubURL := fmt.Sprintf("%s/unsubscribe/%s/%s", s.baseURL, pid.String(), sid.String())
	body = strings.ReplaceAll(body, "{{unsubscribe_url}}", unsubURL)

	sendErr := sendSMTP(project, sub.Email, subject, body)

	s.logEmail(ctx, pid, uuid.NullUUID{UUID: sid, Valid: true}, uuid.NullUUID{UUID: tid, Valid: true}, sub.Email, subject, sendErr)

	if sendErr != nil {
		return SendResult{Failed: 1}, sendErr
	}
	return SendResult{Sent: 1}, nil
}

func (s *EmailService) Broadcast(ctx context.Context, projectID, templateID string) (SendResult, error) {
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

	subscribers, err := s.queries.ListActiveSubscribersByProject(ctx, pid)
	if err != nil {
		return SendResult{}, err
	}

	if len(subscribers) == 0 {
		return SendResult{}, errors.New("no active subscribers")
	}

	result := SendResult{}
	for _, sub := range subscribers {
		body := replaceVariables(template.HtmlBody, sub)
		subject := replaceVariablesSimple(template.Subject, sub)

		unsubURL := fmt.Sprintf("%s/unsubscribe/%s/%s", s.baseURL, pid.String(), sub.ID.String())
		body = strings.ReplaceAll(body, "{{unsubscribe_url}}", unsubURL)

		sendErr := sendSMTP(project, sub.Email, subject, body)

		s.logEmail(ctx, pid, uuid.NullUUID{UUID: sub.ID, Valid: true}, uuid.NullUUID{UUID: tid, Valid: true}, sub.Email, subject, sendErr)

		if sendErr != nil {
			result.Failed++
		} else {
			result.Sent++
		}
	}

	return result, nil
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

	sendErr := sendSMTP(project, to, subject, htmlBody)

	s.logEmail(ctx, pid, uuid.NullUUID{}, uuid.NullUUID{}, to, subject, sendErr)

	return sendErr
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
		body = strings.ReplaceAll(body, "{{"+key+"}}", val)
		subject = strings.ReplaceAll(subject, "{{"+key+"}}", val)
	}

	sendErr := sendSMTP(project, to, subject, body)

	s.logEmail(ctx, pid, uuid.NullUUID{}, uuid.NullUUID{UUID: tid, Valid: true}, to, subject, sendErr)

	return sendErr
}

func (s *EmailService) GetLogs(ctx context.Context, projectID string, limit, offset int32) ([]db.EmailLog, int64, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, 0, errors.New("invalid project id")
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

	total, _ := s.queries.CountEmailLogsByProject(ctx, pid)
	sent, _ := s.queries.CountEmailLogsByStatus(ctx, db.CountEmailLogsByStatusParams{ProjectID: pid, Status: "sent"})
	failed, _ := s.queries.CountEmailLogsByStatus(ctx, db.CountEmailLogsByStatusParams{ProjectID: pid, Status: "failed"})

	return map[string]int64{
		"total":  total,
		"sent":   sent,
		"failed": failed,
	}, nil
}

func (s *EmailService) logEmail(ctx context.Context, projectID uuid.UUID, subscriberID, templateID uuid.NullUUID, toEmail, subject string, sendErr error) {
	status := "sent"
	var errMsg sql.NullString
	if sendErr != nil {
		status = "failed"
		errMsg = sql.NullString{String: sendErr.Error(), Valid: true}
	}

	s.queries.CreateEmailLog(ctx, db.CreateEmailLogParams{
		ProjectID:    projectID,
		SubscriberID: subscriberID,
		TemplateID:   templateID,
		ToEmail:      toEmail,
		Subject:      subject,
		Status:       status,
		Error:        errMsg,
	})
}

func (s *EmailService) Unsubscribe(ctx context.Context, projectID, subscriberID string) error {
	sid, err := uuid.Parse(subscriberID)
	if err != nil {
		return errors.New("invalid subscriber id")
	}

	pid, err := uuid.Parse(projectID)
	if err != nil {
		return errors.New("invalid project id")
	}

	_, err = s.queries.UpdateSubscriberStatus(ctx, db.UpdateSubscriberStatusParams{
		ID:        sid,
		ProjectID: pid,
		Status:    "unsubscribed",
	})
	return err
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

	return sendSMTP(project, fromEmail, "SendDock SMTP Test", "<h2>SMTP is working!</h2><p>Your SendDock SMTP configuration is correct.</p>")
}

func sendSMTP(project db.Project, to, subject, htmlBody string) error {
	host := project.SmtpHost.String
	port := project.SmtpPort.Int32
	user := project.SmtpUser.String
	pass := project.SmtpPasswordEncrypted.String

	fromEmail := user
	if project.FromEmail.Valid && project.FromEmail.String != "" {
		fromEmail = project.FromEmail.String
	}

	from := fromEmail
	if project.FromName.Valid && project.FromName.String != "" {
		from = fmt.Sprintf("%s <%s>", project.FromName.String, fromEmail)
	}

	inlinedBody := inlineCSS(htmlBody)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		from, to, subject, inlinedBody)

	addr := fmt.Sprintf("%s:%d", host, port)

	if port == 465 {
		return sendSMTPImplicitTLS(host, addr, user, pass, fromEmail, to, []byte(msg))
	}

	auth := smtp.PlainAuth("", user, pass, host)
	return smtp.SendMail(addr, auth, fromEmail, []string{to}, []byte(msg))
}

func sendSMTPImplicitTLS(host, addr, user, pass, from, to string, msg []byte) error {
	tlsConfig := &tls.Config{ServerName: host}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("tls connection failed: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp client failed: %w", err)
	}
	defer client.Close()

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

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("smtp write failed: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("smtp close failed: %w", err)
	}

	return client.Quit()
}

func replaceVariables(body string, sub db.Subscriber) string {
	r := strings.NewReplacer(
		"{{name}}", sub.Name,
		"{{email}}", sub.Email,
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
