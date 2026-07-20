package settings

import (
	"context"
	"errors"
	"log"
	"net/url"
	"strings"
	"sync"

	"github.com/arkhe-systems/senddock/internal/db"
)

const (
	DefaultSessionIdleTimeoutMinutes = 120
	MinSessionIdleTimeoutMinutes     = 5
	MaxSessionIdleTimeoutMinutes     = 1440
)

var (
	ErrInvalidIdleTimeout        = errors.New("session idle timeout must be between 5 and 1440 minutes")
	ErrInvalidPublicURL          = errors.New("the public URL must start with http:// or https:// and include a host")
	ErrLicenseStorageUnavailable = errors.New("license storage is not configured")
)

func normalizePublicURL(raw string) (string, error) {
	trimmed := strings.TrimRight(strings.TrimSpace(raw), "/")
	if trimmed == "" {
		return "", nil
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", ErrInvalidPublicURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", ErrInvalidPublicURL
	}
	if parsed.Hostname() == "" {
		return "", ErrInvalidPublicURL
	}
	return trimmed, nil
}

type Settings struct {
	PublicURL                 string
	SessionIdleTimeoutMinutes int
}

type Cipher interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

type Provider struct {
	queries *db.Queries
	cipher  Cipher

	mu         sync.RWMutex
	cached     Settings
	licenseKey string

	listenersMu      sync.Mutex
	licenseListeners []func(key string)
}

func NewProvider(queries *db.Queries, cipher Cipher) *Provider {
	return &Provider{
		queries: queries,
		cipher:  cipher,
		cached: Settings{
			SessionIdleTimeoutMinutes: DefaultSessionIdleTimeoutMinutes,
		},
	}
}

func (p *Provider) Load(ctx context.Context, envPublicURL, envLicenseKey string) error {
	if err := p.queries.EnsureInstanceSettingsRow(ctx); err != nil {
		return err
	}

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

	envLicenseKey = strings.TrimSpace(envLicenseKey)

	if row.LicenseKeyEncrypted == "" && envLicenseKey != "" && p.cipher != nil {
		log.Println("DEPRECATION: SENDDOCK_LICENSE_KEY is now configured from the dashboard under Instance Settings. The value in your environment has been imported and support for it will be removed in v0.9.")
		sealed, err := p.cipher.Encrypt(envLicenseKey)
		if err != nil {
			return err
		}
		row, err = p.queries.SetInstanceLicenseKey(ctx, sealed)
		if err != nil {
			return err
		}
	}

	p.store(row)
	return nil
}

func (p *Provider) store(row db.InstanceSetting) {
	key := ""
	if row.LicenseKeyEncrypted != "" && p.cipher != nil {
		decrypted, err := p.cipher.Decrypt(row.LicenseKeyEncrypted)
		if err != nil {
			log.Printf("settings: stored license key could not be decrypted, treating it as unset: %v", err)
		} else {
			key = decrypted
		}
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.cached = Settings{
		PublicURL:                 row.PublicUrl,
		SessionIdleTimeoutMinutes: int(row.SessionIdleTimeoutMinutes),
	}
	p.licenseKey = key
}

func (p *Provider) LicenseKey() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.licenseKey
}

func (p *Provider) OnLicenseKeyChange(fn func(key string)) {
	p.listenersMu.Lock()
	defer p.listenersMu.Unlock()
	p.licenseListeners = append(p.licenseListeners, fn)
}

func (p *Provider) notifyLicenseChange(key string) {
	p.listenersMu.Lock()
	listeners := make([]func(string), len(p.licenseListeners))
	copy(listeners, p.licenseListeners)
	p.listenersMu.Unlock()

	for _, fn := range listeners {
		fn(key)
	}
}

func (p *Provider) SetLicenseKey(ctx context.Context, key string) error {
	if p.cipher == nil {
		return ErrLicenseStorageUnavailable
	}

	key = strings.TrimSpace(key)

	encrypted := ""
	if key != "" {
		sealed, err := p.cipher.Encrypt(key)
		if err != nil {
			return err
		}
		encrypted = sealed
	}

	row, err := p.queries.SetInstanceLicenseKey(ctx, encrypted)
	if err != nil {
		return err
	}

	p.store(row)
	p.notifyLicenseChange(key)
	return nil
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

	publicURL, err := normalizePublicURL(next.PublicURL)
	if err != nil {
		return Settings{}, err
	}

	row, err := p.queries.UpdateInstanceSettings(ctx, db.UpdateInstanceSettingsParams{
		PublicUrl:                 publicURL,
		SessionIdleTimeoutMinutes: int32(next.SessionIdleTimeoutMinutes),
	})
	if err != nil {
		return Settings{}, err
	}

	p.store(row)
	return p.Current(), nil
}
