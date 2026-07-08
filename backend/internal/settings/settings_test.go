package settings

import (
	"context"
	"errors"
	"testing"
)

func TestNewProviderStartsWithDefaultIdleTimeout(t *testing.T) {
	p := NewProvider(nil)

	if got := p.SessionIdleTimeoutMinutes(); got != DefaultSessionIdleTimeoutMinutes {
		t.Errorf("expected %d, got %d", DefaultSessionIdleTimeoutMinutes, got)
	}
	if got := p.PublicURL(); got != "" {
		t.Errorf("expected empty public url, got %q", got)
	}
}

func TestUpdateRejectsIdleTimeoutOutOfRange(t *testing.T) {
	p := NewProvider(nil)

	cases := []int{0, MinSessionIdleTimeoutMinutes - 1, MaxSessionIdleTimeoutMinutes + 1}
	for _, minutes := range cases {
		_, err := p.Update(context.Background(), Settings{PublicURL: "https://example.com", SessionIdleTimeoutMinutes: minutes})
		if !errors.Is(err, ErrInvalidIdleTimeout) {
			t.Errorf("minutes=%d: expected ErrInvalidIdleTimeout, got %v", minutes, err)
		}
	}
}
