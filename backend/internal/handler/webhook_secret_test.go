package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/arkhe-systems/senddock/internal/service"
	"github.com/arkhe-systems/senddock/pkg/auth"
)

// The signing secret is handed out once, by the response that creates the webhook,
// and no read path may return it again. These tests drive the real handler over HTTP
// against a migrated database, because the leak this guards against lived in the
// response mapping shared by every read, not in a function that can be called alone.
//
// They skip unless DATABASE_URL points at a migrated instance, so the default test run
// stays free of external dependencies.

type webhookSecretFixture struct {
	mux       http.Handler
	conn      *sql.DB
	projectID string
	userID    string
	base      string
}

func newWebhookSecretFixture(t *testing.T, role string) webhookSecretFixture {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set; skipping webhook secret integration test")
	}

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		t.Fatalf("ping database: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	ctx := context.Background()

	var userID, workspaceID, projectID uuid.UUID
	if err := conn.QueryRowContext(ctx,
		`INSERT INTO users (email, name) VALUES ($1, $2) RETURNING id`,
		"webhook-secret-"+uuid.NewString()+"@example.test", "Webhook Secret Test",
	).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if err := conn.QueryRowContext(ctx,
		`INSERT INTO workspaces (name, created_by) VALUES ($1, $2) RETURNING id`,
		"webhook-secret-test", userID,
	).Scan(&workspaceID); err != nil {
		t.Fatalf("insert workspace: %v", err)
	}
	t.Cleanup(func() {
		if _, err := conn.ExecContext(ctx, `DELETE FROM workspaces WHERE id = $1`, workspaceID); err != nil {
			t.Logf("cleanup workspace: %v", err)
		}
	})
	if _, err := conn.ExecContext(ctx,
		`INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, $3)`,
		workspaceID, userID, role,
	); err != nil {
		t.Fatalf("insert membership: %v", err)
	}
	if err := conn.QueryRowContext(ctx,
		`INSERT INTO projects (workspace_id, user_id, name, from_email) VALUES ($1, $2, $3, $4) RETURNING id`,
		workspaceID, userID, "webhook-secret-test", "sender@example.test",
	).Scan(&projectID); err != nil {
		t.Fatalf("insert project: %v", err)
	}

	queries := db.New(conn)
	h := NewWebhookHandler(service.NewWebhookService(queries), service.NewProjectService(queries, "test-encryption-secret"))

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/projects/{id}/webhooks", h.Create)
	mux.HandleFunc("GET /api/v1/projects/{id}/webhooks", h.List)
	mux.HandleFunc("GET /api/v1/projects/{id}/webhooks/{webhookId}", h.Get)
	mux.HandleFunc("PATCH /api/v1/projects/{id}/webhooks/{webhookId}", h.Patch)

	return webhookSecretFixture{
		mux:       mux,
		conn:      conn,
		projectID: projectID.String(),
		userID:    userID.String(),
		base:      "/api/v1/projects/" + projectID.String() + "/webhooks",
	}
}

// insertWebhook writes a webhook with a known secret, for tests that only exercise a
// read path and so do not need a role allowed to create one.
func (f webhookSecretFixture) insertWebhook(t *testing.T, secret string) string {
	t.Helper()

	var id uuid.UUID
	if err := f.conn.QueryRowContext(context.Background(),
		`INSERT INTO webhooks (project_id, url, secret, events) VALUES ($1, $2, $3, $4) RETURNING id`,
		f.projectID, "https://example.com/webhooks/senddock", secret, []string{"email.sent"},
	).Scan(&id); err != nil {
		t.Fatalf("insert webhook: %v", err)
	}
	return id.String()
}

func (f webhookSecretFixture) request(t *testing.T, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, f.userID))

	rec := httptest.NewRecorder()
	f.mux.ServeHTTP(rec, req)
	return rec
}

type webhookBody struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
	Active bool   `json:"active"`
}

func decodeWebhook(t *testing.T, body string) webhookBody {
	t.Helper()

	var hook webhookBody
	if err := json.Unmarshal([]byte(body), &hook); err != nil {
		t.Fatalf("decode webhook: %v (body %s)", err, body)
	}
	return hook
}

func TestWebhookSecretIsReturnedOnceAndNeverReadBack(t *testing.T) {
	f := newWebhookSecretFixture(t, "owner")

	rec := f.request(t, http.MethodPost, f.base,
		`{"url":"https://example.com/webhooks/senddock","events":["email.sent"]}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201 (body %s)", rec.Code, rec.Body.String())
	}

	created := decodeWebhook(t, rec.Body.String())
	if len(created.Secret) != 64 {
		t.Fatalf("created secret = %q, want a 64-character secret", created.Secret)
	}
	if created.ID == "" {
		t.Fatal("created webhook has no id")
	}

	// Every read below must be silent about a secret that is known to exist, so the
	// assertions cannot pass merely because nothing was generated.
	reads := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"list", http.MethodGet, f.base, ""},
		{"get", http.MethodGet, f.base + "/" + created.ID, ""},
		{"patch", http.MethodPatch, f.base + "/" + created.ID, `{"active":false}`},
	}

	for _, read := range reads {
		t.Run(read.name, func(t *testing.T) {
			rec := f.request(t, read.method, read.path, read.body)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200 (body %s)", rec.Code, rec.Body.String())
			}

			body := rec.Body.String()
			if strings.Contains(body, created.Secret) {
				t.Errorf("response leaked the signing secret: %s", body)
			}

			if read.name == "list" {
				var payload struct {
					Webhooks []webhookBody `json:"webhooks"`
				}
				if err := json.Unmarshal([]byte(body), &payload); err != nil {
					t.Fatalf("decode list: %v", err)
				}
				if len(payload.Webhooks) != 1 {
					t.Fatalf("returned %d webhooks, want 1", len(payload.Webhooks))
				}
				if payload.Webhooks[0].Secret != "" {
					t.Errorf("listed secret = %q, want empty", payload.Webhooks[0].Secret)
				}
				return
			}

			if hook := decodeWebhook(t, body); hook.Secret != "" {
				t.Errorf("secret = %q, want empty", hook.Secret)
			}
		})
	}
}

// A read-only member can list a project's webhooks, and that role has no business
// holding the signing key: with the secret readable it could sign deliveries of its
// own. Reading a webhook needs no capability, so this is the widest audience the
// field was exposed to.
func TestWebhookSecretIsHiddenFromAViewer(t *testing.T) {
	f := newWebhookSecretFixture(t, "viewer")

	const secret = "9f0a3b1c2d3e4f5061728394a5b6c7d8e9f0a1b2c3d4e5f60718293a4b5c6d7e"
	id := f.insertWebhook(t, secret)

	for _, path := range []string{f.base, f.base + "/" + id} {
		rec := f.request(t, http.MethodGet, path, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body %s)", rec.Code, rec.Body.String())
		}
		if strings.Contains(rec.Body.String(), secret) {
			t.Errorf("viewer-visible response leaked the signing secret: %s", rec.Body.String())
		}
	}
}
