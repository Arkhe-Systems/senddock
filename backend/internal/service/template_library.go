package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/arkhe-systems/senddock/internal/cache"
)

const (
	templateLibraryManifestKey  = "template_library:manifest"
	templateLibraryManifestTTL  = time.Hour
	templateLibraryFetchTimeout = 10 * time.Second
	templateLibraryMaxBytes     = 1 << 20
)

var (
	ErrTemplateLibraryUnavailable = errors.New("template library is unavailable")
	ErrTemplateLibraryNotFound    = errors.New("template not found in library")
)

type TemplateLibraryEntry struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Category     string   `json:"category"`
	Description  string   `json:"description"`
	ThumbnailURL string   `json:"thumbnail_url"`
	HTMLURL      string   `json:"html_url"`
	Variables    []string `json:"variables"`
}

type templateLibraryManifest struct {
	Version   int                    `json:"version"`
	Templates []TemplateLibraryEntry `json:"templates"`
}

type TemplateLibraryService struct {
	manifestURL string
	cache       *cache.Redis
	httpClient  *http.Client
}

func NewTemplateLibraryService(manifestURL string, cacheClient *cache.Redis) *TemplateLibraryService {
	return &TemplateLibraryService{
		manifestURL: manifestURL,
		cache:       cacheClient,
		httpClient:  &http.Client{Timeout: templateLibraryFetchTimeout},
	}
}

func (s *TemplateLibraryService) List(ctx context.Context) ([]TemplateLibraryEntry, error) {
	var cached templateLibraryManifest
	if s.cache.Get(ctx, templateLibraryManifestKey, &cached) {
		return cached.Templates, nil
	}

	manifest, err := s.fetchManifest(ctx)
	if err != nil {
		return nil, err
	}

	s.cache.Set(ctx, templateLibraryManifestKey, manifest, templateLibraryManifestTTL)
	return manifest.Templates, nil
}

func (s *TemplateLibraryService) Find(ctx context.Context, id string) (TemplateLibraryEntry, error) {
	entries, err := s.List(ctx)
	if err != nil {
		return TemplateLibraryEntry{}, err
	}
	for _, entry := range entries {
		if entry.ID == id {
			return entry, nil
		}
	}
	return TemplateLibraryEntry{}, ErrTemplateLibraryNotFound
}

func (s *TemplateLibraryService) FetchHTML(ctx context.Context, entry TemplateLibraryEntry) (string, error) {
	if entry.HTMLURL == "" {
		return "", ErrTemplateLibraryUnavailable
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, entry.HTMLURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrTemplateLibraryUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: status %d", ErrTemplateLibraryUnavailable, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, templateLibraryMaxBytes))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrTemplateLibraryUnavailable, err)
	}

	return string(body), nil
}

func (s *TemplateLibraryService) fetchManifest(ctx context.Context) (templateLibraryManifest, error) {
	if s.manifestURL == "" {
		return templateLibraryManifest{}, ErrTemplateLibraryUnavailable
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.manifestURL, nil)
	if err != nil {
		return templateLibraryManifest{}, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return templateLibraryManifest{}, fmt.Errorf("%w: %v", ErrTemplateLibraryUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return templateLibraryManifest{}, fmt.Errorf("%w: status %d", ErrTemplateLibraryUnavailable, resp.StatusCode)
	}

	var manifest templateLibraryManifest
	if err := json.NewDecoder(io.LimitReader(resp.Body, templateLibraryMaxBytes)).Decode(&manifest); err != nil {
		return templateLibraryManifest{}, fmt.Errorf("%w: %v", ErrTemplateLibraryUnavailable, err)
	}

	return manifest, nil
}
