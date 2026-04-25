package service

import (
	"strings"
	"testing"
)

func TestIsPublicURLReachable(t *testing.T) {
	cases := []struct {
		url  string
		want bool
	}{
		{"", false},
		{"http://localhost", false},
		{"http://localhost:8080", false},
		{"https://localhost", false},
		{"http://127.0.0.1", false},
		{"http://127.0.0.1:8080", false},
		{"http://[::1]", false},
		{"http://0.0.0.0:8080", false},
		{"https://email.mycompany.com", true},
		{"http://senddock.example.com", true},
		{"https://app.senddock.dev", true},
		{"not-a-url", false},
	}

	for _, tc := range cases {
		t.Run(tc.url, func(t *testing.T) {
			got := IsPublicURLReachable(tc.url)
			if got != tc.want {
				t.Errorf("IsPublicURLReachable(%q) = %v, want %v", tc.url, got, tc.want)
			}
		})
	}
}

func TestEnsureUnsubscribeFooter(t *testing.T) {
	t.Run("no placeholder appends footer", func(t *testing.T) {
		body := "<p>Hello</p>"
		got := ensureUnsubscribeFooter(body)
		if !strings.Contains(got, "{{unsubscribe_url}}") {
			t.Error("expected footer with unsubscribe placeholder appended")
		}
		if !strings.HasPrefix(got, body) {
			t.Error("original body should still be at the start")
		}
	})

	t.Run("placeholder present is left alone", func(t *testing.T) {
		body := `<p>Hi <a href="{{unsubscribe_url}}">unsub</a></p>`
		got := ensureUnsubscribeFooter(body)
		if got != body {
			t.Errorf("body with placeholder should not be modified, got %q", got)
		}
	})

	t.Run("inserts before closing body tag", func(t *testing.T) {
		body := "<html><body><p>Hi</p></body></html>"
		got := ensureUnsubscribeFooter(body)
		if !strings.Contains(got, "{{unsubscribe_url}}") {
			t.Errorf("expected unsubscribe placeholder in result, got %q", got)
		}
		if !strings.HasSuffix(got, "</body></html>") {
			t.Errorf("closing tags should remain at the end, got %q", got)
		}
		footerStart := strings.Index(got, "<hr")
		bodyClose := strings.LastIndex(got, "</body>")
		if footerStart == -1 || footerStart > bodyClose {
			t.Errorf("footer must be inserted before </body>, footerStart=%d bodyClose=%d", footerStart, bodyClose)
		}
	})
}

func TestSignUnsubVerifyRoundTrip(t *testing.T) {
	s := &EmailService{publicURL: "https://example.com", encSecret: "test-secret"}
	pid := "11111111-1111-1111-1111-111111111111"
	sid := "22222222-2222-2222-2222-222222222222"

	token := s.signUnsub(pid, sid)
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	if !s.verifyUnsubToken(pid, sid, token) {
		t.Error("freshly signed token should verify")
	}

	if s.verifyUnsubToken(pid, sid, token+"x") {
		t.Error("tampered token must not verify")
	}

	if s.verifyUnsubToken(pid, "different-sid", token) {
		t.Error("token bound to a different subscriber must not verify")
	}

	otherSecret := &EmailService{publicURL: "https://example.com", encSecret: "other-secret"}
	if otherSecret.verifyUnsubToken(pid, sid, token) {
		t.Error("token signed with one secret must not verify under another")
	}
}

func TestUnsubURLIncludesSignedToken(t *testing.T) {
	s := &EmailService{publicURL: "https://email.example.com", encSecret: "k"}
	pid := "abcd"
	sid := "efgh"

	url := s.unsubURL(pid, sid)
	if !strings.HasPrefix(url, "https://email.example.com/unsubscribe/abcd/efgh?t=") {
		t.Errorf("unexpected URL shape: %s", url)
	}
}
