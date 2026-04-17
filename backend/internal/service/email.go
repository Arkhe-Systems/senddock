package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/google/uuid"
)

type EmailService struct {
	queries *db.Queries
}

func NewEmailService(queries *db.Queries) *EmailService {
	return &EmailService{queries: queries}
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

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		from, to, subject, htmlBody)

	addr := fmt.Sprintf("%s:%d", host, port)
	auth := smtp.PlainAuth("", user, pass, host)

	return smtp.SendMail(addr, auth, fromEmail, []string{to}, []byte(msg))
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
