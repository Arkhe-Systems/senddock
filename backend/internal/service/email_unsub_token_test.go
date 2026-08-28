package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func legacyToken(secret, projectID, subscriberID string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(projectID + ":" + subscriberID))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func TestUnsubTokenRoundtrips(t *testing.T) {
	s := &EmailService{encSecret: "test-secret-0123456789abcdef"}
	project := "11111111-1111-1111-1111-111111111111"
	subscriber := "22222222-2222-2222-2222-222222222222"
	newsletter := "33333333-3333-3333-3333-333333333333"

	twoPart := s.signUnsub(project, subscriber)
	threePart := s.signUnsubNewsletter(project, subscriber, newsletter)

	cases := []struct {
		name         string
		newsletterID string
		token        string
		want         bool
	}{
		{"two-part roundtrip", "", twoPart, true},
		{"three-part roundtrip", newsletter, threePart, true},
		{"stripped newsletter fails closed", "", threePart, false},
		{"injected newsletter fails closed", newsletter, twoPart, false},
		{"tampered newsletter fails", "44444444-4444-4444-4444-444444444444", threePart, false},
		{"wrong subscriber fails", "", s.signUnsub(project, "55555555-5555-5555-5555-555555555555"), false},
		{"wrong project fails", "", s.signUnsub("66666666-6666-6666-6666-666666666666", subscriber), false},
		{"empty token fails", "", "", false},
		{"empty token with newsletter fails", newsletter, "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := s.verifyAnyUnsubToken(project, subscriber, tc.newsletterID, tc.token)
			if got != tc.want {
				t.Fatalf("verifyAnyUnsubToken(newsletter=%q) = %v, want %v", tc.newsletterID, got, tc.want)
			}
		})
	}
}

func TestUnsubTokenBackwardCompatible(t *testing.T) {
	secret := "compat-secret-0123456789abcdef"
	s := &EmailService{encSecret: secret}
	project := "11111111-1111-1111-1111-111111111111"
	subscriber := "22222222-2222-2222-2222-222222222222"

	preChange := legacyToken(secret, project, subscriber)
	if !s.verifyUnsubToken(project, subscriber, preChange) {
		t.Fatal("a token produced by the pre-change algorithm must still verify")
	}
	if !s.verifyAnyUnsubToken(project, subscriber, "", preChange) {
		t.Fatal("the dispatcher must accept legacy tokens when no newsletter is present")
	}
}

func TestUnsubNewsletterURLCarriesSignedContext(t *testing.T) {
	s := &EmailService{encSecret: "url-secret-0123456789abcdef", settings: staticURL("https://mail.example.com")}
	project := "11111111-1111-1111-1111-111111111111"
	subscriber := "22222222-2222-2222-2222-222222222222"
	newsletter := "33333333-3333-3333-3333-333333333333"

	url := s.unsubNewsletterURL(project, subscriber, newsletter)
	want := "https://mail.example.com/unsubscribe/" + project + "/" + subscriber + "?n=" + newsletter + "&t=" + s.signUnsubNewsletter(project, subscriber, newsletter)
	if url != want {
		t.Fatalf("url = %q, want %q", url, want)
	}
}

type staticURL string

func (u staticURL) PublicURL() string { return string(u) }
