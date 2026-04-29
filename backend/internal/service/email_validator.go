package service

import (
	"bufio"
	"context"
	_ "embed"
	"net"
	"net/mail"
	"os"
	"strings"
	"sync"
	"time"
)

//go:embed disposable_domains.txt
var embeddedDisposableList string

type EmailValidator struct {
	mxResolver  *net.Resolver
	mxTimeout   time.Duration
	disposable  map[string]struct{}
	mu          sync.RWMutex
}

func NewEmailValidator() *EmailValidator {
	v := &EmailValidator{
		mxResolver: net.DefaultResolver,
		mxTimeout:  3 * time.Second,
		disposable: make(map[string]struct{}),
	}
	v.loadDisposable(strings.NewReader(embeddedDisposableList))
	if path := os.Getenv("DISPOSABLE_DOMAINS_FILE"); path != "" {
		if f, err := os.Open(path); err == nil {
			v.loadDisposable(f)
			f.Close()
		}
	}
	return v
}

func (v *EmailValidator) loadDisposable(r interface {
	Read(p []byte) (n int, err error)
}) {
	scanner := bufio.NewScanner(r)
	v.mu.Lock()
	defer v.mu.Unlock()
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		v.disposable[strings.ToLower(line)] = struct{}{}
	}
}

func (v *EmailValidator) Syntax(email string) (string, bool) {
	addr, err := mail.ParseAddress(strings.TrimSpace(email))
	if err != nil {
		return "", false
	}
	at := strings.LastIndex(addr.Address, "@")
	if at <= 0 || at == len(addr.Address)-1 {
		return "", false
	}
	return strings.ToLower(addr.Address), true
}

func (v *EmailValidator) IsDisposable(email string) bool {
	at := strings.LastIndex(email, "@")
	if at < 0 {
		return false
	}
	domain := strings.ToLower(email[at+1:])
	v.mu.RLock()
	_, ok := v.disposable[domain]
	v.mu.RUnlock()
	return ok
}

func (v *EmailValidator) HasMX(ctx context.Context, email string, cache map[string]bool) bool {
	at := strings.LastIndex(email, "@")
	if at < 0 {
		return false
	}
	domain := strings.ToLower(email[at+1:])
	if cached, ok := cache[domain]; ok {
		return cached
	}
	lookupCtx, cancel := context.WithTimeout(ctx, v.mxTimeout)
	defer cancel()
	records, err := v.mxResolver.LookupMX(lookupCtx, domain)
	hasMX := err == nil && len(records) > 0
	if !hasMX {
		_, fallbackErr := v.mxResolver.LookupHost(lookupCtx, domain)
		hasMX = fallbackErr == nil
	}
	cache[domain] = hasMX
	return hasMX
}
