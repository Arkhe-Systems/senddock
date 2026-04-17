package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/google/uuid"
)

type APIKeyService struct {
	queries *db.Queries
}

func NewAPIKeyService(queries *db.Queries) *APIKeyService {
	return &APIKeyService{queries: queries}
}

type APIKeyCreateResult struct {
	Key    string     `json:"key"`
	APIKey db.ApiKey  `json:"api_key"`
}

func (s *APIKeyService) Create(ctx context.Context, projectID, name string) (APIKeyCreateResult, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return APIKeyCreateResult{}, errors.New("invalid project id")
	}

	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return APIKeyCreateResult{}, err
	}
	rawKey := "sk_" + hex.EncodeToString(rawBytes)
	prefix := rawKey[:10]

	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	apiKey, err := s.queries.CreateAPIKey(ctx, db.CreateAPIKeyParams{
		ProjectID: pid,
		Name:      name,
		KeyHash:   keyHash,
		KeyPrefix: prefix,
	})
	if err != nil {
		return APIKeyCreateResult{}, err
	}

	return APIKeyCreateResult{
		Key:    rawKey,
		APIKey: apiKey,
	}, nil
}

func (s *APIKeyService) ListByProject(ctx context.Context, projectID string) ([]db.ApiKey, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, errors.New("invalid project id")
	}

	return s.queries.ListAPIKeysByProject(ctx, pid)
}

func (s *APIKeyService) Delete(ctx context.Context, keyID, projectID string) error {
	kid, err := uuid.Parse(keyID)
	if err != nil {
		return errors.New("invalid key id")
	}

	pid, err := uuid.Parse(projectID)
	if err != nil {
		return errors.New("invalid project id")
	}

	return s.queries.DeleteAPIKey(ctx, db.DeleteAPIKeyParams{
		ID:        kid,
		ProjectID: pid,
	})
}

// ValidateKey checks an API key and returns the project ID if valid
func (s *APIKeyService) ValidateKey(ctx context.Context, rawKey string) (db.ApiKey, error) {
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	apiKey, err := s.queries.GetAPIKeyByHash(ctx, keyHash)
	if err != nil {
		return db.ApiKey{}, errors.New("invalid api key")
	}

	s.queries.UpdateAPIKeyLastUsed(ctx, apiKey.ID)

	return apiKey, nil
}
