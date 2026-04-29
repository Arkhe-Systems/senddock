package handler

import (
	"encoding/json"
	"errors"
	"html"
	"net/http"
	"strconv"
	"time"

	"github.com/arkhe-systems/senddock/internal/cache"
	"github.com/arkhe-systems/senddock/pkg/auth"
	"github.com/arkhe-systems/senddock/internal/response"
	"github.com/arkhe-systems/senddock/internal/service"
)

const (
	maxBatchRecipients     = 500
	sendRateLimit          = 60
	sendRateWindow         = time.Minute
	batchRateLimit         = 10
	batchRateWindow        = time.Minute
	broadcastRateLimit     = 5
	broadcastRateWindow    = time.Hour
)

type EmailHandler struct {
	emailService   *service.EmailService
	projectService *service.ProjectService
	cache          *cache.Redis
	Audit          *service.AuditService
}

func NewEmailHandler(emailService *service.EmailService, projectService *service.ProjectService, redis *cache.Redis) *EmailHandler {
	return &EmailHandler{
		emailService:   emailService,
		projectService: projectService,
		cache:          redis,
	}
}

func (h *EmailHandler) checkRateLimit(w http.ResponseWriter, r *http.Request, prefix, projectID string, limit int64, window time.Duration) bool {
	if h.cache == nil {
		return true
	}
	count, err := h.cache.Increment(r.Context(), "ratelimit:"+prefix+":"+projectID, window)
	if err != nil {
		return true
	}
	if count > limit {
		retryAfter := int(window.Seconds())
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(errorResponse{Error: "rate limit exceeded for this project on this endpoint. Slow down and retry later"})
		return false
	}
	return true
}

type sendRequest struct {
	To           string            `json:"to"`
	SubscriberID string            `json:"subscriber_id"`
	TemplateID   string            `json:"template_id"`
	Subject      string            `json:"subject"`
	HtmlBody     string            `json:"html_body"`
	Data         map[string]string `json:"data"`
}

type broadcastRequest struct {
	TemplateID string            `json:"template_id"`
	Subject    string            `json:"subject"`
	Variables  map[string]string `json:"variables"`
}

type batchRecipient struct {
	To   string            `json:"to"`
	Data map[string]string `json:"data"`
}

type batchSendRequest struct {
	TemplateID string           `json:"template_id"`
	Subject    string           `json:"subject"`
	Recipients []batchRecipient `json:"recipients"`
}

func (h *EmailHandler) verifyAccess(r *http.Request) (string, error) {
	if pid, ok := r.Context().Value(auth.ProjectIDKey).(string); ok {
		return pid, nil
	}

	userID := r.Context().Value(auth.UserIDKey).(string)
	projectID := r.PathValue("id")
	_, err := h.projectService.GetByID(r.Context(), projectID, userID)
	return projectID, err
}

func (h *EmailHandler) Send(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyAccess(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	if !h.checkRateLimit(w, r, "send", projectID, sendRateLimit, sendRateWindow) {
		return
	}

	var req sendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if req.SubscriberID != "" && req.TemplateID != "" {
		result, err := h.emailService.SendToSubscriber(r.Context(), projectID, req.SubscriberID, req.TemplateID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
			return
		}
		json.NewEncoder(w).Encode(result)
		return
	}

	if req.To != "" && req.TemplateID != "" {
		err := h.emailService.SendWithTemplate(r.Context(), projectID, req.TemplateID, req.To, req.Subject, req.Data)
		if errors.Is(err, service.ErrRecipientSuppressed) {
			json.NewEncoder(w).Encode(map[string]any{"message": "suppressed", "suppressed": 1})
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "sent"})
		return
	}

	if req.To != "" && req.HtmlBody != "" {
		if req.Subject == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse{Error: "subject is required for direct send"})
			return
		}
		err := h.emailService.SendDirect(r.Context(), projectID, req.To, req.Subject, req.HtmlBody)
		if errors.Is(err, service.ErrRecipientSuppressed) {
			json.NewEncoder(w).Encode(map[string]any{"message": "suppressed", "suppressed": 1})
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "sent"})
		return
	}

	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(errorResponse{Error: "provide (to + template_id), (subscriber_id + template_id), or (to + subject + html_body)"})
}

func (h *EmailHandler) Broadcast(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyAccess(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	if !h.checkRateLimit(w, r, "broadcast", projectID, broadcastRateLimit, broadcastRateWindow) {
		return
	}

	var req broadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	if req.TemplateID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "template_id is required"})
		return
	}

	var varsJSON []byte
	if len(req.Variables) > 0 {
		varsJSON, _ = json.Marshal(req.Variables)
	}

	result, err := h.emailService.Broadcast(r.Context(), projectID, req.TemplateID, req.Subject, varsJSON)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	if h.Audit != nil {
		userID, _ := r.Context().Value(auth.UserIDKey).(string)
		h.Audit.LogFromRequest(r, projectID, userID, "broadcast.send", "template", req.TemplateID, map[string]any{"recipients": result.Sent, "subject_override": req.Subject})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *EmailHandler) BatchSend(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyAccess(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	if !h.checkRateLimit(w, r, "batch", projectID, batchRateLimit, batchRateWindow) {
		return
	}

	var req batchSendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	if req.TemplateID == "" || len(req.Recipients) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "template_id and recipients are required"})
		return
	}

	if len(req.Recipients) > maxBatchRecipients {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "too many recipients in a single request. Use a broadcast for sending to your subscriber list, or split this batch into chunks of 500 or fewer"})
		return
	}

	sent := 0
	failed := 0
	suppressed := 0
	for _, rcpt := range req.Recipients {
		if rcpt.To == "" {
			failed++
			continue
		}
		err := h.emailService.SendWithTemplate(r.Context(), projectID, req.TemplateID, rcpt.To, req.Subject, rcpt.Data)
		if errors.Is(err, service.ErrRecipientSuppressed) {
			suppressed++
			continue
		}
		if err != nil {
			failed++
		} else {
			sent++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"sent": sent, "failed": failed, "suppressed": suppressed})
}

func (h *EmailHandler) TestSMTP(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyAccess(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	err = h.emailService.TestSMTP(r.Context(), projectID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "SMTP connection successful. Test email sent."})
}

func (h *EmailHandler) Logs(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyAccess(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	limit := int32(50)
	offset := int32(0)

	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 && v <= 100 {
		limit = int32(v)
	}
	if v, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && v >= 0 {
		offset = int32(v)
	}

	status := r.URL.Query().Get("status")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	logs, total, err := h.emailService.GetLogs(r.Context(), projectID, limit, offset, status, from, to)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs":  response.FromEmailLogs(logs),
		"total": total,
	})
}

func (h *EmailHandler) Stats(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyAccess(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	stats, err := h.emailService.GetStats(r.Context(), projectID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *EmailHandler) UnsubscribePage(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	subscriberID := r.PathValue("subscriberId")
	token := r.URL.Query().Get("t")

	ctx, err := h.emailService.ResolveUnsubscribe(r.Context(), projectID, subscriberID, token)
	if err != nil {
		writeUnsubscribeHTML(w, http.StatusNotFound, unsubscribeInvalidPage())
		return
	}

	writeUnsubscribeHTML(w, http.StatusOK, unsubscribeConfirmPage(ctx.ProjectName, ctx.Email, projectID, subscriberID, token))
}

func (h *EmailHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	subscriberID := r.PathValue("subscriberId")
	token := r.URL.Query().Get("t")

	if err := h.emailService.Unsubscribe(r.Context(), projectID, subscriberID, token); err != nil {
		writeUnsubscribeHTML(w, http.StatusNotFound, unsubscribeInvalidPage())
		return
	}

	writeUnsubscribeHTML(w, http.StatusOK, unsubscribeDonePage())
}

func writeUnsubscribeHTML(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	w.Write([]byte(body))
}

func unsubscribePageShell(content string) string {
	return `<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>Unsubscribe</title><style>
*,*::before,*::after{box-sizing:border-box}
html,body{margin:0;padding:0}
body{min-height:100vh;display:flex;align-items:center;justify-content:center;background:#09090b;color:#fafafa;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,Oxygen,Ubuntu,sans-serif;line-height:1.5;padding:24px}
.card{width:100%;max-width:420px;background:#18181b;border:1px solid #27272a;border-radius:12px;padding:32px;text-align:center}
.card h1{margin:0 0 8px;font-size:18px;font-weight:600;letter-spacing:-0.01em}
.card p{margin:0 0 8px;font-size:14px;color:#a1a1aa}
.email{display:inline-block;margin:8px 0 24px;padding:6px 12px;background:#27272a;border-radius:6px;font-family:ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace;font-size:13px;color:#fafafa}
.project{color:#fafafa;font-weight:500}
button,.button{display:inline-block;width:100%;padding:10px 16px;border:0;border-radius:8px;font-size:14px;font-weight:500;cursor:pointer;font-family:inherit}
.danger{background:#fafafa;color:#09090b}
.danger:hover{background:#e4e4e7}
.muted{margin-top:8px;background:transparent;color:#a1a1aa}
.muted:hover{color:#fafafa}
.hint{margin-top:24px;font-size:12px;color:#52525b}
.icon{margin:0 auto 16px;width:40px;height:40px;border-radius:999px;background:#27272a;display:flex;align-items:center;justify-content:center;color:#a1a1aa;font-size:20px;line-height:1}
.success .icon{background:rgba(34,197,94,0.1);color:#4ade80}
.error .icon{background:rgba(239,68,68,0.1);color:#f87171}
form{margin:0}
</style></head><body>` + content + `</body></html>`
}

func unsubscribeConfirmPage(projectName, email, projectID, subscriberID, token string) string {
	safeProject := html.EscapeString(projectName)
	safeEmail := html.EscapeString(email)
	action := "/unsubscribe/" + projectID + "/" + subscriberID + "?t=" + token
	return unsubscribePageShell(`
<div class="card">
	<div class="icon">&#9993;</div>
	<h1>Unsubscribe from <span class="project">` + safeProject + `</span>?</h1>
	<p>You are about to stop receiving emails sent to</p>
	<span class="email">` + safeEmail + `</span>
	<form method="POST" action="` + action + `">
		<button type="submit" class="danger">Confirm unsubscribe</button>
	</form>
	<p class="hint">If you reached this page by accident, just close this tab.</p>
</div>`)
}

func unsubscribeDonePage() string {
	return unsubscribePageShell(`
<div class="card success">
	<div class="icon">&#10003;</div>
	<h1>You have been unsubscribed</h1>
	<p>You will no longer receive emails from this list. You can close this tab.</p>
</div>`)
}

func unsubscribeInvalidPage() string {
	return unsubscribePageShell(`
<div class="card error">
	<div class="icon">&#9888;</div>
	<h1>Link expired or invalid</h1>
	<p>This unsubscribe link is no longer valid. If you keep receiving emails you don't want, reply to one of them and the sender can remove you manually.</p>
</div>`)
}
