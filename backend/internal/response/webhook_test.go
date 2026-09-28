package response

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/google/uuid"
)

const testWebhookSecret = "9f0a3b1c2d3e4f5061728394a5b6c7d8e9f0a1b2c3d4e5f60718293a4b5c6d7e"

func testWebhook() db.Webhook {
	return db.Webhook{
		ID:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		ProjectID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		Url:       "https://example.com/webhooks/senddock",
		Secret:    testWebhookSecret,
		Events:    []string{"email.sent"},
		Active:    true,
		CreatedAt: time.Date(2026, 4, 29, 5, 12, 0, 0, time.UTC),
	}
}

func TestFromWebhookHidesSecret(t *testing.T) {
	hook := FromWebhook(testWebhook())

	if hook.Secret != "" {
		t.Errorf("secret = %q, want empty", hook.Secret)
	}
	if hook.ID != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("id = %q, want the webhook id", hook.ID)
	}
	if hook.URL != "https://example.com/webhooks/senddock" {
		t.Errorf("url = %q, want the webhook url", hook.URL)
	}
	if !hook.Active {
		t.Error("active = false, want true")
	}
	if len(hook.Events) != 1 || hook.Events[0] != "email.sent" {
		t.Errorf("events = %v, want [email.sent]", hook.Events)
	}
}

func TestFromWebhookListHidesEverySecret(t *testing.T) {
	hooks := []db.Webhook{testWebhook(), testWebhook()}

	for i, hook := range FromWebhooks(hooks) {
		if hook.Secret != "" {
			t.Errorf("webhooks[%d].secret = %q, want empty", i, hook.Secret)
		}
	}
}

// The creation response is the one place the secret is allowed through, otherwise
// no client could ever verify a signature.
func TestFromCreatedWebhookKeepsSecret(t *testing.T) {
	if got := FromCreatedWebhook(testWebhook()).Secret; got != testWebhookSecret {
		t.Errorf("secret = %q, want the generated secret", got)
	}
}

// Guards the documented wire format: reads carry the field and it is empty, so a
// client holding a stored webhook has nothing it could sign with.
func TestFromWebhookJSONNeverCarriesSecret(t *testing.T) {
	body, err := json.Marshal(struct {
		Webhooks []Webhook `json:"webhooks"`
	}{Webhooks: FromWebhooks([]db.Webhook{testWebhook()})})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if strings.Contains(string(body), testWebhookSecret) {
		t.Errorf("list response leaked the secret: %s", body)
	}
	if !strings.Contains(string(body), `"secret":""`) {
		t.Errorf("list response should carry an empty secret field: %s", body)
	}
}

func TestFromCreatedWebhookJSONCarriesSecret(t *testing.T) {
	body, err := json.Marshal(FromCreatedWebhook(testWebhook()))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if !strings.Contains(string(body), testWebhookSecret) {
		t.Errorf("creation response dropped the secret: %s", body)
	}
}
