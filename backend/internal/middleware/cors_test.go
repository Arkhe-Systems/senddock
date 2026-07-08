package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func staticOrigin(url string) func() string {
	return func() string { return url }
}

func TestCORS_ReadsOriginPerRequest(t *testing.T) {
	origin := "https://old.example.com"
	handler := CORS(func() string { return origin })(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, httptest.NewRequest("GET", "/test", nil))
	if got := first.Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Fatalf("expected %s, got %s", origin, got)
	}

	origin = "https://new.example.com"
	second := httptest.NewRecorder()
	handler.ServeHTTP(second, httptest.NewRequest("GET", "/test", nil))
	if got := second.Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Errorf("origin change must apply without rebuilding the middleware, got %s", got)
	}
}

func TestCORS_SetsHeaders(t *testing.T) {
	frontendURL := "http://localhost:5173"
	handler := CORS(staticOrigin(frontendURL))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != frontendURL {
		t.Errorf("expected origin %s, got %s", frontendURL, rec.Header().Get("Access-Control-Allow-Origin"))
	}

	if rec.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("expected credentials true")
	}
}

func TestCORS_PreflightReturns200(t *testing.T) {
	handler := CORS(staticOrigin("http://localhost:5173"))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("preflight should return 200, got %d", rec.Code)
	}
}

func TestCORS_DynamicOrigin(t *testing.T) {
	handler := CORS(staticOrigin("https://senddock.dev"))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "https://senddock.dev" {
		t.Errorf("expected origin https://senddock.dev, got %s", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}
