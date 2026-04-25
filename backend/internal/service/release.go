package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/arkhe-systems/senddock/internal/cache"
	"github.com/arkhe-systems/senddock/internal/version"
)

const (
	githubReleaseURL = "https://api.github.com/repos/arkhe-systems/senddock/releases/latest"
	releaseCacheKey  = "release:latest"
	releaseCacheTTL  = time.Hour
)

type ReleaseInfo struct {
	Current    string `json:"current"`
	Latest     string `json:"latest"`
	Outdated   bool   `json:"outdated"`
	ReleaseURL string `json:"release_url"`
	Notes      string `json:"notes"`
	CheckedAt  string `json:"checked_at"`
	Available  bool   `json:"available"`
}

type ReleaseService struct {
	cache *cache.Redis
	http  *http.Client
}

func NewReleaseService(redis *cache.Redis) *ReleaseService {
	return &ReleaseService{
		cache: redis,
		http:  &http.Client{Timeout: 5 * time.Second},
	}
}

type cachedRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Body    string `json:"body"`
}

func (s *ReleaseService) GetRelease(ctx context.Context) ReleaseInfo {
	info := ReleaseInfo{
		Current:   version.Version,
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
		Available: true,
	}

	release, err := s.fetchLatest(ctx)
	if err != nil || release.TagName == "" {
		info.Available = false
		return info
	}

	info.Latest = strings.TrimPrefix(release.TagName, "v")
	info.ReleaseURL = release.HTMLURL
	info.Notes = release.Body
	info.Outdated = isNewer(info.Latest, info.Current)
	return info
}

func (s *ReleaseService) fetchLatest(ctx context.Context) (cachedRelease, error) {
	var cached cachedRelease
	if s.cache != nil && s.cache.Get(ctx, releaseCacheKey, &cached) {
		return cached, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubReleaseURL, nil)
	if err != nil {
		return cachedRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "senddock-update-check")

	resp, err := s.http.Do(req)
	if err != nil {
		return cachedRelease{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return cachedRelease{}, errors.New("github returned " + resp.Status)
	}

	var release cachedRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return cachedRelease{}, err
	}

	if s.cache != nil {
		s.cache.Set(ctx, releaseCacheKey, release, releaseCacheTTL)
	}
	return release, nil
}

func isNewer(remote, local string) bool {
	rp := splitVersion(remote)
	lp := splitVersion(local)
	for i := range 3 {
		if rp[i] > lp[i] {
			return true
		}
		if rp[i] < lp[i] {
			return false
		}
	}
	return false
}

func splitVersion(v string) [3]int {
	parts := strings.SplitN(v, ".", 3)
	out := [3]int{}
	for i := 0; i < len(parts) && i < 3; i++ {
		n := 0
		for _, ch := range parts[i] {
			if ch < '0' || ch > '9' {
				break
			}
			n = n*10 + int(ch-'0')
		}
		out[i] = n
	}
	return out
}
