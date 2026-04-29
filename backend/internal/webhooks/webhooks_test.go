package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestGenerateSecret_Uniqueness(t *testing.T) {
	a, err := GenerateSecret()
	if err != nil {
		t.Fatal(err)
	}
	b, err := GenerateSecret()
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("two generated secrets should not be equal")
	}
	if len(a) != 64 {
		t.Fatalf("expected 64-char hex secret, got %d", len(a))
	}
}

func TestSend_HmacSignatureRoundtrip(t *testing.T) {
	secret := "test-webhook-secret"
	body := []byte(`{"id":"evt_xxx","type":"email.opened"}`)

	receivedSig := atomic.Value{}
	receivedTimestamp := atomic.Value{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sigHeader := r.Header.Get("X-SendDock-Signature")
		parts := strings.Split(sigHeader, ",")
		if len(parts) != 2 {
			t.Fatalf("malformed signature header: %s", sigHeader)
		}
		ts := strings.TrimPrefix(parts[0], "t=")
		sig := strings.TrimPrefix(parts[1], "v1=")
		receivedTimestamp.Store(ts)
		receivedSig.Store(sig)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := NewService(nil)
	hook := stubHook(server.URL, secret)

	statusCode, err := svc.send(t.Context(), hook, body)
	if err != nil {
		t.Fatalf("send returned error: %v", err)
	}
	if statusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", statusCode)
	}

	ts := receivedTimestamp.Load().(string)
	sig := receivedSig.Load().(string)

	mac := hmac.New(sha256.New, []byte(secret))
	fmt.Fprintf(mac, "%s.", ts)
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sig), []byte(expected)) {
		t.Fatalf("signature mismatch:\n  got:      %s\n  expected: %s", sig, expected)
	}
}
