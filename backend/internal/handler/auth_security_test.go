package handler

import (
	"context"
	"testing"
	"time"

	"github.com/arkhe-systems/senddock/internal/service"
)

func TestAPIKeyCapabilityAllowlist(t *testing.T) {
	// Machine capabilities an API key legitimately exercises.
	allowed := []service.Capability{
		service.CapSendTransactional,
		service.CapBroadcast,
		service.CapSubscribersWrite,
	}
	for _, c := range allowed {
		if !apiKeyCapabilities[c] {
			t.Errorf("expected API keys to be allowed capability %q", c)
		}
	}
	// Capabilities an API key must never bypass into.
	denied := []service.Capability{
		service.CapMembersManage,
		service.CapWorkspaceDelete,
		service.CapProjectSettings,
		service.CapAPIKeysManage,
		service.CapWebhooksWrite,
		service.CapTemplatesWrite,
		service.CapCampaignsWrite,
		service.CapSuppressionsWrite,
	}
	for _, c := range denied {
		if apiKeyCapabilities[c] {
			t.Errorf("API keys must NOT be allowed capability %q", c)
		}
	}
}

func TestLockoutKeyDeterministicAndCaseInsensitive(t *testing.T) {
	a := lockoutKey("login_fail:", "User@Example.com")
	b := lockoutKey("login_fail:", "  user@example.com ")
	if a != b {
		t.Fatalf("lockout key should be case/space-insensitive: %q vs %q", a, b)
	}
	if lockoutKey("login_fail:", "a@b.com") == lockoutKey("2fa_fail:", "a@b.com") {
		t.Fatal("different prefixes must produce different keys")
	}
}

type fakeLimiter struct {
	counts map[string]int64
}

func (f *fakeLimiter) Count(_ context.Context, key string) int64 { return f.counts[key] }
func (f *fakeLimiter) Increment(_ context.Context, key string, _ time.Duration) (int64, error) {
	f.counts[key]++
	return f.counts[key], nil
}
func (f *fakeLimiter) Delete(_ context.Context, keys ...string) {
	for _, k := range keys {
		delete(f.counts, k)
	}
}

func TestLockoutThreshold(t *testing.T) {
	h := &AuthHandler{limiter: &fakeLimiter{counts: map[string]int64{}}}
	ctx := context.Background()
	key := lockoutKey("login_fail:", "brute@example.com")

	for i := 0; i < maxLoginAttempts; i++ {
		if h.locked(ctx, key, maxLoginAttempts) {
			t.Fatalf("locked too early at attempt %d", i)
		}
		h.recordFailure(ctx, key)
	}
	if !h.locked(ctx, key, maxLoginAttempts) {
		t.Fatal("should be locked after reaching the attempt limit")
	}
	h.clearFailures(ctx, key)
	if h.locked(ctx, key, maxLoginAttempts) {
		t.Fatal("clearing failures should unlock the account")
	}
}

func TestLockoutNoopWithoutLimiter(t *testing.T) {
	h := &AuthHandler{} // no limiter configured
	ctx := context.Background()
	if h.locked(ctx, "k", 1) {
		t.Fatal("must never lock when no limiter is configured")
	}
	h.recordFailure(ctx, "k") // must not panic
	h.clearFailures(ctx, "k")
}
