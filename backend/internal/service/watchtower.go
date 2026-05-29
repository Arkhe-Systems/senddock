package service

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"
)

type WatchtowerClient struct {
	url   string
	token string
	http  *http.Client

	mu        sync.RWMutex
	healthy   bool
	lastCheck time.Time
	lastErr   string
}

func NewWatchtowerClient(url, token string) *WatchtowerClient {
	if strings.TrimSpace(url) == "" {
		return nil
	}
	return &WatchtowerClient{
		url:   strings.TrimRight(url, "/"),
		token: token,
		http:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (w *WatchtowerClient) Configured() bool {
	return w != nil
}

type WatchtowerStatus struct {
	Configured bool      `json:"configured"`
	Healthy    bool      `json:"healthy"`
	URL        string    `json:"url,omitempty"`
	LastCheck  time.Time `json:"last_check"`
	LastError  string    `json:"last_error,omitempty"`
}

func (w *WatchtowerClient) Status() WatchtowerStatus {
	if w == nil {
		return WatchtowerStatus{}
	}
	w.mu.RLock()
	defer w.mu.RUnlock()
	return WatchtowerStatus{
		Configured: true,
		Healthy:    w.healthy,
		URL:        w.url,
		LastCheck:  w.lastCheck,
		LastError:  w.lastErr,
	}
}

const watchtowerPingCacheTTL = 30 * time.Second

func (w *WatchtowerClient) StatusFresh(ctx context.Context) WatchtowerStatus {
	if w == nil {
		return WatchtowerStatus{}
	}
	w.mu.RLock()
	stale := w.lastCheck.IsZero() || time.Since(w.lastCheck) > watchtowerPingCacheTTL
	w.mu.RUnlock()
	if stale {
		_ = w.Ping(ctx)
	}
	return w.Status()
}

func (w *WatchtowerClient) Ping(ctx context.Context) error {
	if w == nil {
		return errors.New("watchtower not configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, w.url+"/v1/update", nil)
	if err != nil {
		w.recordResult(false, err.Error())
		return err
	}
	if w.token != "" {
		req.Header.Set("Authorization", "Bearer "+w.token)
	}
	resp, err := w.http.Do(req)
	if err != nil {
		w.recordResult(false, err.Error())
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusMethodNotAllowed, http.StatusUnauthorized, http.StatusOK:
		w.recordResult(true, "")
		return nil
	default:
		msg := "unexpected response: " + resp.Status
		w.recordResult(false, msg)
		return errors.New(msg)
	}
}

func (w *WatchtowerClient) Trigger(ctx context.Context) error {
	if w == nil {
		return errors.New("watchtower not configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url+"/v1/update", nil)
	if err != nil {
		return err
	}
	if w.token != "" {
		req.Header.Set("Authorization", "Bearer "+w.token)
	}
	resp, err := w.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return errors.New("watchtower returned " + resp.Status)
	}
	return nil
}

func (w *WatchtowerClient) recordResult(ok bool, errMsg string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.healthy = ok
	w.lastCheck = time.Now()
	w.lastErr = errMsg
}
