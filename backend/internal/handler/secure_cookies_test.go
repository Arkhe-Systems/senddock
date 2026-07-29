package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arkhe-systems/senddock/internal/service"
)

func TestAuthCookiesFollowTheSecureResolver(t *testing.T) {
	original := secureCookieResolver
	t.Cleanup(func() { secureCookieResolver = original })

	cases := []struct {
		name   string
		secure bool
	}{
		{name: "https deployment marks cookies secure", secure: true},
		{name: "plain http deployment does not", secure: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			SetSecureCookieResolver(func() bool { return tc.secure })

			rec := httptest.NewRecorder()
			setAuthCookies(rec, service.AuthTokens{AccessToken: "a", RefreshToken: "r"})

			cookies := (&http.Response{Header: rec.Header()}).Cookies()
			if len(cookies) != 2 {
				t.Fatalf("expected both auth cookies, got %d", len(cookies))
			}
			for _, c := range cookies {
				if c.Secure != tc.secure {
					t.Errorf("cookie %s: expected Secure=%v, got %v", c.Name, tc.secure, c.Secure)
				}
			}
		})
	}
}

func TestSecureCookieResolverIgnoresNil(t *testing.T) {
	original := secureCookieResolver
	t.Cleanup(func() { secureCookieResolver = original })

	SetSecureCookieResolver(func() bool { return true })
	SetSecureCookieResolver(nil)

	if !secureCookieResolver() {
		t.Error("passing nil must not clear a configured resolver")
	}
}
