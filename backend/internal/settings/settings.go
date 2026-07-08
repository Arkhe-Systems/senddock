package settings

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"

	"github.com/arkhe-systems/senddock/internal/db"
)

const (
	DefaultSessionIdleTimeoutMinutes = 120
	MinSessionIdleTimeoutMinutes     = 5
	MaxSessionIdleTimeoutMinutes     = 1440
)

var ErrInvalidIdleTimeout = errors.New("session idle timeout must be between 5 and 1440 minutes")

type Settings struct {
	PublicURL                 string
	SessionIdleTimeoutMinutes int
}

type Provider struct {
	queries *db.Queries

	mu     sync.RWMutex
	cached Settings
}

func NewProvider(queries *db.Queries) *Provider {
	return &Provider{
		queries: queries,
		cached: Settings{
			SessionIdleTimeoutMinutes: DefaultSessionIdleTimeoutMinutes,
		},
	}
}

func (p *Provider) Load(ctx context.Context, envPublicURL string) error {
	row, err := p.queries.GetInstanceSettings(ctx)
	if err != nil {
		return err
	}

	envPublicURL = strings.TrimRight(strings.TrimSpace(envPublicURL), "/")

	if row.PublicUrl == "" && envPublicURL != "" {
		log.Println("DEPRECATION: PUBLIC_URL is now configured from the dashboard under Instance Settings. The value in your environment has been imported and support for it will be removed in v0.9.")
		row, err = p.queries.SetInstancePublicURL(ctx, envPublicURL)
		if err != nil {
			return err
		}
	}

	p.store(row)
	return nil
}

func (p *Provider) store(row db.InstanceSetting) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cached = Settings{
		PublicURL:                 row.PublicUrl,
		SessionIdleTimeoutMinutes: int(row.SessionIdleTimeoutMinutes),
	}
}

func (p *Provider) Current() Settings {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.cached
}

func (p *Provider) PublicURL() string {
	return p.Current().PublicURL
}

func (p *Provider) SessionIdleTimeoutMinutes() int {
	return p.Current().SessionIdleTimeoutMinutes
}

func (p *Provider) Update(ctx context.Context, next Settings) (Settings, error) {
	if next.SessionIdleTimeoutMinutes < MinSessionIdleTimeoutMinutes || next.SessionIdleTimeoutMinutes > MaxSessionIdleTimeoutMinutes {
		return Settings{}, ErrInvalidIdleTimeout
	}

	row, err := p.queries.UpdateInstanceSettings(ctx, db.UpdateInstanceSettingsParams{
		PublicUrl:                 strings.TrimRight(strings.TrimSpace(next.PublicURL), "/"),
		SessionIdleTimeoutMinutes: int32(next.SessionIdleTimeoutMinutes),
	})
	if err != nil {
		return Settings{}, err
	}

	p.store(row)
	return p.Current(), nil
}
