package metrics

import (
	"errors"
	"fmt"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"
)

func TestHandlerExposesSendPipelineMetrics(t *testing.T) {
	EmailAttempt()
	EmailSent()
	EmailFailed("bounce")
	SMTPError(&textproto.Error{Code: 550, Msg: "mailbox unavailable"})
	WebhookDelivery("success")
	BounceIngest("webhook")
	ComplaintIngest()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	Handler().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()

	want := []string{
		"senddock_email_send_attempts_total",
		"senddock_emails_sent_total",
		"senddock_emails_failed_total",
		"senddock_smtp_errors_total",
		"senddock_webhook_deliveries_total",
		"senddock_bounce_ingest_total",
		"senddock_complaint_ingest_total",
		"senddock_bounce_poller_tick_seconds",
	}
	for _, name := range want {
		if !strings.Contains(body, name) {
			t.Errorf("metrics output missing %q", name)
		}
	}
}

func TestSMTPClass(t *testing.T) {
	cases := map[string]struct {
		err  error
		want string
	}{
		"5xx":     {&textproto.Error{Code: 550}, "5xx"},
		"4xx":     {&textproto.Error{Code: 450}, "4xx"},
		"wrapped": {fmt.Errorf("smtp rcpt to failed: %w", &textproto.Error{Code: 550, Msg: "user unknown"}), "5xx"},
		"plain":   {errors.New("dial timeout"), "other"},
	}
	for name, c := range cases {
		if got := smtpClass(c.err); got != c.want {
			t.Errorf("%s: smtpClass = %q, want %q", name, got, c.want)
		}
	}
}
