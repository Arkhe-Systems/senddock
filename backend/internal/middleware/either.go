package middleware

import (
	"net/http"
	"strings"
)

func EitherAuth(cookieAuth, apiKeyAuth func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if strings.HasPrefix(header, "Bearer sk_") {
				apiKeyAuth(next).ServeHTTP(w, r)
				return
			}

			cookieAuth(next).ServeHTTP(w, r)
		})
	}
}
