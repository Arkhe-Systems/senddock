package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/arkhe-systems/senddock/internal/db"
)

const ProjectIDKey contextKey = "projectID"

type APIKeyValidator interface {
	GetAPIKeyByHash(ctx context.Context, keyHash string) (db.ApiKey, error)
	UpdateAPIKeyLastUsed(ctx context.Context, id interface{}) error
}

func APIKey(queries *db.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer sk_") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "missing or invalid api key"})
				return
			}

			rawKey := strings.TrimPrefix(header, "Bearer ")

			hash := sha256.Sum256([]byte(rawKey))
			keyHash := hex.EncodeToString(hash[:])

			apiKey, err := queries.GetAPIKeyByHash(r.Context(), keyHash)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid api key"})
				return
			}

			queries.UpdateAPIKeyLastUsed(r.Context(), apiKey.ID)

			ctx := context.WithValue(r.Context(), ProjectIDKey, apiKey.ProjectID.String())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
