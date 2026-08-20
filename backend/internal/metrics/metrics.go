package metrics

import (
	"context"
	"errors"
	"math"
	"net/http"
	"net/textproto"
	"sync"
	"time"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	emailAttempts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "senddock_email_send_attempts_total",
		Help: "Emails handed to the SMTP send path.",
	})
	emailsSent = promauto.NewCounter(prometheus.CounterOpts{
		Name: "senddock_emails_sent_total",
		Help: "Emails accepted by the upstream SMTP server.",
	})
	emailsFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "senddock_emails_failed_total",
		Help: "Emails that failed to send, by reason.",
	}, []string{"reason"})
	smtpErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "senddock_smtp_errors_total",
		Help: "SMTP errors by response class.",
	}, []string{"class"})
	webhookDeliveries = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "senddock_webhook_deliveries_total",
		Help: "Outbound webhook delivery outcomes.",
	}, []string{"result"})
	bounceIngest = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "senddock_bounce_ingest_total",
		Help: "Bounces ingested, by source.",
	}, []string{"source"})
	complaintIngest = promauto.NewCounter(prometheus.CounterOpts{
		Name: "senddock_complaint_ingest_total",
		Help: "Spam complaints ingested.",
	})
	pollerTickDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "senddock_bounce_poller_tick_seconds",
		Help:    "Duration of a bounce IMAP poller tick.",
		Buckets: prometheus.DefBuckets,
	})
)

func EmailAttempt() { emailAttempts.Inc() }

func EmailSent() { emailsSent.Inc() }

func EmailFailed(reason string) { emailsFailed.WithLabelValues(reason).Inc() }

func SMTPError(err error) { smtpErrors.WithLabelValues(smtpClass(err)).Inc() }

func WebhookDelivery(result string) { webhookDeliveries.WithLabelValues(result).Inc() }

func BounceIngest(source string) { bounceIngest.WithLabelValues(source).Inc() }

func ComplaintIngest() { complaintIngest.Inc() }

func ObservePollerTick(d time.Duration) { pollerTickDuration.Observe(d.Seconds()) }

func smtpClass(err error) string {
	var protoErr *textproto.Error
	if errors.As(err, &protoErr) {
		switch {
		case protoErr.Code >= 500 && protoErr.Code < 600:
			return "5xx"
		case protoErr.Code >= 400 && protoErr.Code < 500:
			return "4xx"
		}
	}
	return "other"
}

var queueDepthOnce sync.Once

func RegisterQueueDepth(queries *db.Queries) {
	queueDepthOnce.Do(func() {
		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "senddock_broadcast_queue_depth",
			Help: "Broadcast jobs waiting or in flight (pending, retry, sending).",
		}, func() float64 {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			n, err := queries.CountActiveBroadcastJobs(ctx)
			if err != nil {
				return math.NaN()
			}
			return float64(n)
		})
	})
}

func Handler() http.Handler { return promhttp.Handler() }
