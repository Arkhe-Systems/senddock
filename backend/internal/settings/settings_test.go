package settings

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/arkhe-systems/senddock/internal/db"
)

type reversibleCipher struct{}

func (reversibleCipher) Encrypt(plaintext string) (string, error) {
	return "sealed:" + plaintext, nil
}

func (reversibleCipher) Decrypt(ciphertext string) (string, error) {
	if !strings.HasPrefix(ciphertext, "sealed:") {
		return "", errors.New("not sealed by this cipher")
	}
	return strings.TrimPrefix(ciphertext, "sealed:"), nil
}

func TestNewProviderStartsWithDefaultIdleTimeout(t *testing.T) {
	p := NewProvider(nil, nil)

	if got := p.SessionIdleTimeoutMinutes(); got != DefaultSessionIdleTimeoutMinutes {
		t.Errorf("expected %d, got %d", DefaultSessionIdleTimeoutMinutes, got)
	}
	if got := p.PublicURL(); got != "" {
		t.Errorf("expected empty public url, got %q", got)
	}
}

func TestUpdateRejectsIdleTimeoutOutOfRange(t *testing.T) {
	p := NewProvider(nil, nil)

	cases := []int{0, MinSessionIdleTimeoutMinutes - 1, MaxSessionIdleTimeoutMinutes + 1}
	for _, minutes := range cases {
		_, err := p.Update(context.Background(), Settings{PublicURL: "https://example.com", SessionIdleTimeoutMinutes: minutes})
		if !errors.Is(err, ErrInvalidIdleTimeout) {
			t.Errorf("minutes=%d: expected ErrInvalidIdleTimeout, got %v", minutes, err)
		}
	}
}

func TestStoreDecryptsLicenseKey(t *testing.T) {
	p := NewProvider(nil, reversibleCipher{})

	p.store(db.InstanceSetting{LicenseKeyEncrypted: "sealed:lk_live_abc123"})

	if got := p.LicenseKey(); got != "lk_live_abc123" {
		t.Errorf("expected decrypted key, got %q", got)
	}
}

func TestStoreTreatsUndecryptableKeyAsUnset(t *testing.T) {
	p := NewProvider(nil, reversibleCipher{})

	p.store(db.InstanceSetting{LicenseKeyEncrypted: "garbage-from-a-rotated-secret"})

	if got := p.LicenseKey(); got != "" {
		t.Errorf("a key that cannot be decrypted must not leak through, got %q", got)
	}
}

func TestSetLicenseKeyWithoutCipherIsRefused(t *testing.T) {
	p := NewProvider(nil, nil)

	err := p.SetLicenseKey(context.Background(), "lk_live_abc123")
	if !errors.Is(err, ErrLicenseStorageUnavailable) {
		t.Errorf("expected ErrLicenseStorageUnavailable, got %v", err)
	}
}

func TestLicenseChangeListenersFire(t *testing.T) {
	p := NewProvider(nil, reversibleCipher{})

	var seen []string
	p.OnLicenseKeyChange(func(key string) { seen = append(seen, key) })
	p.OnLicenseKeyChange(func(key string) { seen = append(seen, "second:"+key) })

	p.notifyLicenseChange("lk_live_abc123")

	if len(seen) != 2 || seen[0] != "lk_live_abc123" || seen[1] != "second:lk_live_abc123" {
		t.Errorf("every listener must be notified in order, got %v", seen)
	}
}
