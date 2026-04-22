package response

import (
	"database/sql"
	"time"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/google/uuid"
)

type Project struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	FromName    *string `json:"from_name"`
	FromEmail   *string `json:"from_email"`
	SmtpHost    *string `json:"smtp_host"`
	SmtpPort    *int32  `json:"smtp_port"`
	SmtpUser    *string `json:"smtp_user"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type Subscriber struct {
	ID             string  `json:"id"`
	ProjectID      string  `json:"project_id"`
	Email          string  `json:"email"`
	Name           string  `json:"name"`
	Status         string  `json:"status"`
	SubscribedAt   string  `json:"subscribed_at"`
	UnsubscribedAt *string `json:"unsubscribed_at"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

type Template struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	Subject   string `json:"subject"`
	HtmlBody  string `json:"html_body"`
	TextBody  string `json:"text_body"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type APIKey struct {
	ID         string  `json:"id"`
	ProjectID  string  `json:"project_id"`
	Name       string  `json:"name"`
	KeyPrefix  string  `json:"key_prefix"`
	LastUsedAt *string `json:"last_used_at"`
	CreatedAt  string  `json:"created_at"`
}

func nullStr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func nullInt32(ni sql.NullInt32) *int32 {
	if ni.Valid {
		return &ni.Int32
	}
	return nil
}

func nullTime(nt sql.NullTime) *string {
	if nt.Valid {
		s := nt.Time.Format(time.RFC3339)
		return &s
	}
	return nil
}

func FromProject(p db.Project) Project {
	return Project{
		ID:          p.ID.String(),
		Name:        p.Name,
		Description: nullStr(p.Description),
		FromName:    nullStr(p.FromName),
		FromEmail:   nullStr(p.FromEmail),
		SmtpHost:    nullStr(p.SmtpHost),
		SmtpPort:    nullInt32(p.SmtpPort),
		SmtpUser:    nullStr(p.SmtpUser),
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}
}

func FromProjects(projects []db.Project) []Project {
	result := make([]Project, len(projects))
	for i, p := range projects {
		result[i] = FromProject(p)
	}
	return result
}

func FromSubscriber(s db.Subscriber) Subscriber {
	return Subscriber{
		ID:             s.ID.String(),
		ProjectID:      s.ProjectID.String(),
		Email:          s.Email,
		Name:           s.Name,
		Status:         s.Status,
		SubscribedAt:   s.SubscribedAt.Format(time.RFC3339),
		UnsubscribedAt: nullTime(s.UnsubscribedAt),
		CreatedAt:      s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      s.UpdatedAt.Format(time.RFC3339),
	}
}

func FromSubscribers(subs []db.Subscriber) []Subscriber {
	result := make([]Subscriber, len(subs))
	for i, s := range subs {
		result[i] = FromSubscriber(s)
	}
	return result
}

func FromTemplate(t db.Template) Template {
	return Template{
		ID:        t.ID.String(),
		ProjectID: t.ProjectID.String(),
		Name:      t.Name,
		Subject:   t.Subject,
		HtmlBody:  t.HtmlBody,
		TextBody:  t.TextBody,
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
		UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
	}
}

func FromTemplates(templates []db.Template) []Template {
	result := make([]Template, len(templates))
	for i, t := range templates {
		result[i] = FromTemplate(t)
	}
	return result
}

type EmailLog struct {
	ID           string  `json:"id"`
	ProjectID    string  `json:"project_id"`
	SubscriberID *string `json:"subscriber_id"`
	TemplateID   *string `json:"template_id"`
	ToEmail      string  `json:"to_email"`
	Subject      string  `json:"subject"`
	Status       string  `json:"status"`
	Error        *string `json:"error"`
	SentAt       string  `json:"sent_at"`
}

func nullUUID(nu uuid.NullUUID) *string {
	if nu.Valid {
		s := nu.UUID.String()
		return &s
	}
	return nil
}

func FromEmailLog(l db.EmailLog) EmailLog {
	return EmailLog{
		ID:           l.ID.String(),
		ProjectID:    l.ProjectID.String(),
		SubscriberID: nullUUID(l.SubscriberID),
		TemplateID:   nullUUID(l.TemplateID),
		ToEmail:      l.ToEmail,
		Subject:      l.Subject,
		Status:       l.Status,
		Error:        nullStr(l.Error),
		SentAt:       l.SentAt.Format(time.RFC3339),
	}
}

func FromEmailLogs(logs []db.EmailLog) []EmailLog {
	result := make([]EmailLog, len(logs))
	for i, l := range logs {
		result[i] = FromEmailLog(l)
	}
	return result
}

func FromAPIKey(k db.ApiKey) APIKey {
	return APIKey{
		ID:         k.ID.String(),
		ProjectID:  k.ProjectID.String(),
		Name:       k.Name,
		KeyPrefix:  k.KeyPrefix,
		LastUsedAt: nullTime(k.LastUsedAt),
		CreatedAt:  k.CreatedAt.Format(time.RFC3339),
	}
}

func FromAPIKeys(keys []db.ApiKey) []APIKey {
	result := make([]APIKey, len(keys))
	for i, k := range keys {
		result[i] = FromAPIKey(k)
	}
	return result
}

type Campaign struct {
	ID          string  `json:"id"`
	ProjectID   string  `json:"project_id"`
	TemplateID  string  `json:"template_id"`
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	ScheduledAt string  `json:"scheduled_at"`
	SentAt      *string `json:"sent_at"`
	SentCount   int32   `json:"sent_count"`
	FailedCount int32   `json:"failed_count"`
	CreatedAt   string  `json:"created_at"`
}

func FromCampaign(c db.Campaign) Campaign {
	return Campaign{
		ID:          c.ID.String(),
		ProjectID:   c.ProjectID.String(),
		TemplateID:  c.TemplateID.String(),
		Name:        c.Name,
		Status:      c.Status,
		ScheduledAt: c.ScheduledAt.Format(time.RFC3339),
		SentAt:      nullTime(c.SentAt),
		SentCount:   c.SentCount,
		FailedCount: c.FailedCount,
		CreatedAt:   c.CreatedAt.Format(time.RFC3339),
	}
}

func FromCampaigns(campaigns []db.Campaign) []Campaign {
	result := make([]Campaign, len(campaigns))
	for i, c := range campaigns {
		result[i] = FromCampaign(c)
	}
	return result
}
