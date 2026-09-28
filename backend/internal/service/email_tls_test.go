package service

import (
	"errors"
	"strings"
	"testing"
)

// Certificates are verified unless a project explicitly opted out. This is the property
// that stops an on-path attacker from reading SMTP credentials and mail.
func TestSMTPTLSConfigVerifiesCertificatesByDefault(t *testing.T) {
	config := smtpTLSConfig("smtp.example.test", false)

	if config.InsecureSkipVerify {
		t.Fatal("certificate verification is disabled without an opt-out")
	}
	if config.ServerName != "smtp.example.test" {
		t.Fatalf("ServerName = %q, want the relay host so SNI and verification work", config.ServerName)
	}
}

// The opt-out exists for relays with a self-signed or expired certificate, and it has to
// stay scoped to the project that asked for it.
func TestSMTPTLSConfigHonoursThePerProjectOptOut(t *testing.T) {
	if !smtpTLSConfig("smtp.example.test", true).InsecureSkipVerify {
		t.Fatal("the per-project opt-out did not disable verification")
	}
}

func TestIsCertificateError(t *testing.T) {
	cases := []struct {
		err  error
		want bool
	}{
		{errors.New("x509: certificate signed by unknown authority"), true},
		{errors.New("tls: failed to verify certificate: x509: certificate has expired"), true},
		{errors.New("dial tcp 10.0.0.1:587: connection refused"), false},
		{errors.New("smtp auth failed: 535 authentication failed"), false},
	}
	for _, tc := range cases {
		if got := isCertificateError(tc.err); got != tc.want {
			t.Errorf("isCertificateError(%q) = %v, want %v", tc.err, got, tc.want)
		}
	}
}

// A verification failure has to name the fix. Without the hint, turning verification on
// would turn every self-signed relay into a support thread.
func TestCertificateErrorPointsAtTheOptOut(t *testing.T) {
	cause := errors.New("x509: certificate signed by unknown authority")

	blocked := certificateError("smtp.example.test", false, cause).Error()
	if !strings.Contains(blocked, "smtp.example.test") {
		t.Errorf("message does not name the relay: %q", blocked)
	}
	if !strings.Contains(blocked, "Allow insecure TLS") {
		t.Errorf("message does not point at the project setting: %q", blocked)
	}
	if !strings.Contains(blocked, cause.Error()) {
		t.Errorf("message dropped the underlying cause: %q", blocked)
	}

	// With the opt-out already on there is nothing to suggest, and the underlying
	// handshake failure is still reported.
	allowed := certificateError("smtp.example.test", true, cause).Error()
	if strings.Contains(allowed, "Allow insecure TLS") {
		t.Errorf("message suggests a setting that is already enabled: %q", allowed)
	}
	if !strings.Contains(allowed, cause.Error()) {
		t.Errorf("message dropped the underlying cause: %q", allowed)
	}
}
