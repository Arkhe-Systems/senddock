package response

import (
	"database/sql"
	"encoding/json"
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
	ID             string         `json:"id"`
	ProjectID      string         `json:"project_id"`
	Email          string         `json:"email"`
	Name           string         `json:"name"`
	Status         string         `json:"status"`
	Fields         map[string]any `json:"fields"`
	Tags           []string       `json:"tags"`
	SubscribedAt   string         `json:"subscribed_at"`
	UnsubscribedAt *string        `json:"unsubscribed_at"`
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
}

type FieldDefinition struct {
	ID        string   `json:"id"`
	ProjectID string   `json:"project_id"`
	Key       string   `json:"key"`
	Label     string   `json:"label"`
	FieldType string   `json:"field_type"`
	Options   []string `json:"options"`
	Required  bool     `json:"required"`
	CreatedAt string   `json:"created_at"`
}

func FromFieldDefinition(f db.SubscriberFieldDefinition) FieldDefinition {
	options := []string{}
	if f.Options.Valid && len(f.Options.RawMessage) > 0 {
		_ = json.Unmarshal(f.Options.RawMessage, &options)
	}
	return FieldDefinition{
		ID:        f.ID.String(),
		ProjectID: f.ProjectID.String(),
		Key:       f.Key,
		Label:     f.Label,
		FieldType: f.FieldType,
		Options:   options,
		Required:  f.Required,
		CreatedAt: f.CreatedAt.Format(time.RFC3339),
	}
}

func FromFieldDefinitions(defs []db.SubscriberFieldDefinition) []FieldDefinition {
	result := make([]FieldDefinition, len(defs))
	for i, f := range defs {
		result[i] = FromFieldDefinition(f)
	}
	return result
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

func subscriberFields(raw json.RawMessage) map[string]any {
	fields := map[string]any{}
	if len(raw) == 0 {
		return fields
	}
	_ = json.Unmarshal(raw, &fields)
	return fields
}

func FromSubscriber(s db.Subscriber) Subscriber {
	tags := s.Tags
	if tags == nil {
		tags = []string{}
	}
	return Subscriber{
		ID:             s.ID.String(),
		ProjectID:      s.ProjectID.String(),
		Email:          s.Email,
		Name:           s.Name,
		Status:         s.Status,
		Fields:         subscriberFields(s.Metadata),
		Tags:           tags,
		SubscribedAt:   s.SubscribedAt.Format(time.RFC3339),
		UnsubscribedAt: nullTime(s.UnsubscribedAt),
		CreatedAt:      s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      s.UpdatedAt.Format(time.RFC3339),
	}
}

type Webhook struct {
	ID        string   `json:"id"`
	URL       string   `json:"url"`
	Secret    string   `json:"secret"`
	Events    []string `json:"events"`
	Active    bool     `json:"active"`
	CreatedAt string   `json:"created_at"`
}

func FromWebhook(w db.Webhook) Webhook {
	events := w.Events
	if events == nil {
		events = []string{}
	}
	return Webhook{
		ID:        w.ID.String(),
		URL:       w.Url,
		Secret:    w.Secret,
		Events:    events,
		Active:    w.Active,
		CreatedAt: w.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func FromWebhooks(hooks []db.Webhook) []Webhook {
	result := make([]Webhook, len(hooks))
	for i, w := range hooks {
		result[i] = FromWebhook(w)
	}
	return result
}

type WebhookDelivery struct {
	ID             string `json:"id"`
	EventType      string `json:"event_type"`
	Status         string `json:"status"`
	Attempts       int32  `json:"attempts"`
	LastStatusCode int32  `json:"last_status_code,omitempty"`
	LastError      string `json:"last_error,omitempty"`
	NextAttemptAt  string `json:"next_attempt_at,omitempty"`
	DeliveredAt    string `json:"delivered_at,omitempty"`
	CreatedAt      string `json:"created_at"`
}

func FromWebhookDelivery(d db.WebhookDelivery) WebhookDelivery {
	out := WebhookDelivery{
		ID:        d.ID.String(),
		EventType: d.EventType,
		Status:    d.Status,
		Attempts:  d.Attempts,
		CreatedAt: d.CreatedAt.UTC().Format(time.RFC3339),
	}
	if d.LastStatusCode.Valid {
		out.LastStatusCode = d.LastStatusCode.Int32
	}
	if d.LastError.Valid {
		out.LastError = d.LastError.String
	}
	if !d.NextAttemptAt.IsZero() && d.Status == "pending" {
		out.NextAttemptAt = d.NextAttemptAt.UTC().Format(time.RFC3339)
	}
	if d.DeliveredAt.Valid {
		out.DeliveredAt = d.DeliveredAt.Time.UTC().Format(time.RFC3339)
	}
	return out
}

func FromWebhookDeliveries(deliveries []db.WebhookDelivery) []WebhookDelivery {
	result := make([]WebhookDelivery, len(deliveries))
	for i, d := range deliveries {
		result[i] = FromWebhookDelivery(d)
	}
	return result
}

type Segment struct {
	ID        string          `json:"id"`
	ProjectID string          `json:"project_id"`
	Name      string          `json:"name"`
	Predicate json.RawMessage `json:"predicate"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

func FromSegment(s db.Segment) Segment {
	return Segment{
		ID:        s.ID.String(),
		ProjectID: s.ProjectID.String(),
		Name:      s.Name,
		Predicate: s.Predicate,
		CreatedAt: s.CreatedAt.Format(time.RFC3339),
		UpdatedAt: s.UpdatedAt.Format(time.RFC3339),
	}
}

func FromSegments(segments []db.Segment) []Segment {
	result := make([]Segment, len(segments))
	for i, s := range segments {
		result[i] = FromSegment(s)
	}
	return result
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
	OpenedAt     *string `json:"opened_at"`
	ClickedAt    *string `json:"clicked_at"`
}

type EmailClick struct {
	ID        string  `json:"id"`
	URL       string  `json:"url"`
	ClickedAt string  `json:"clicked_at"`
	UserAgent *string `json:"user_agent"`
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
		OpenedAt:     nullTimeStr(l.OpenedAt),
		ClickedAt:    nullTimeStr(l.ClickedAt),
	}
}

func nullTimeStr(t sql.NullTime) *string {
	if !t.Valid {
		return nil
	}
	s := t.Time.Format(time.RFC3339)
	return &s
}

func FromEmailClicks(rows []db.ListEmailClicksByLogRow) []EmailClick {
	out := make([]EmailClick, len(rows))
	for i, c := range rows {
		var ua *string
		if c.UserAgent.Valid {
			s := c.UserAgent.String
			ua = &s
		}
		out[i] = EmailClick{
			ID:        c.ID.String(),
			URL:       c.Url,
			ClickedAt: c.ClickedAt.Format(time.RFC3339),
			UserAgent: ua,
		}
	}
	return out
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
	ID          string          `json:"id"`
	ProjectID   string          `json:"project_id"`
	TemplateID  string          `json:"template_id"`
	Name        string          `json:"name"`
	Subject     string          `json:"subject"`
	Status      string          `json:"status"`
	ScheduledAt string          `json:"scheduled_at"`
	SentAt      *string         `json:"sent_at"`
	SentCount   int32           `json:"sent_count"`
	FailedCount int32           `json:"failed_count"`
	CreatedAt   string          `json:"created_at"`
	Variables   json.RawMessage `json:"variables"`
}

func FromCampaign(c db.Campaign) Campaign {
	return Campaign{
		ID:          c.ID.String(),
		ProjectID:   c.ProjectID.String(),
		TemplateID:  c.TemplateID.String(),
		Name:        c.Name,
		Subject:     c.Subject,
		Status:      c.Status,
		ScheduledAt: c.ScheduledAt.Format(time.RFC3339),
		SentAt:      nullTime(c.SentAt),
		SentCount:   c.SentCount,
		FailedCount: c.FailedCount,
		CreatedAt:   c.CreatedAt.Format(time.RFC3339),
		Variables:   c.Variables,
	}
}

func FromCampaigns(campaigns []db.Campaign) []Campaign {
	result := make([]Campaign, len(campaigns))
	for i, c := range campaigns {
		result[i] = FromCampaign(c)
	}
	return result
}

type Broadcast struct {
	ID              string          `json:"id"`
	ProjectID       string          `json:"project_id"`
	TemplateID      string          `json:"template_id"`
	Subject         string          `json:"subject"`
	Variables       json.RawMessage `json:"variables"`
	Status          string          `json:"status"`
	TotalRecipients int32           `json:"total_recipients"`
	SentCount       int32           `json:"sent_count"`
	FailedCount     int32           `json:"failed_count"`
	SuppressedCount int32           `json:"suppressed_count"`
	StartedAt       string          `json:"started_at"`
	FinishedAt      *string         `json:"finished_at"`
}

func FromBroadcast(b db.Broadcast) Broadcast {
	return Broadcast{
		ID:              b.ID.String(),
		ProjectID:       b.ProjectID.String(),
		TemplateID:      b.TemplateID.String(),
		Subject:         b.Subject,
		Variables:       b.Variables,
		Status:          b.Status,
		TotalRecipients: b.TotalRecipients,
		SentCount:       b.SentCount,
		FailedCount:     b.FailedCount,
		SuppressedCount: b.SuppressedCount,
		StartedAt:       b.StartedAt.UTC().Format(time.RFC3339),
		FinishedAt:      nullTime(b.FinishedAt),
	}
}

func FromBroadcasts(broadcasts []db.Broadcast) []Broadcast {
	result := make([]Broadcast, len(broadcasts))
	for i, b := range broadcasts {
		result[i] = FromBroadcast(b)
	}
	return result
}
