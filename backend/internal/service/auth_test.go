package service

import (
	"testing"
)

func TestHashToken(t *testing.T) {
	token := "test-token-123"
	hash1 := hashToken(token)
	hash2 := hashToken(token)

	if hash1 != hash2 {
		t.Errorf("same token should produce same hash, got %s and %s", hash1, hash2)
	}

	if len(hash1) != 64 {
		t.Errorf("hash should be 64 chars (sha256 hex), got %d", len(hash1))
	}

	differentHash := hashToken("different-token")
	if hash1 == differentHash {
		t.Error("different tokens should produce different hashes")
	}
}

func TestGenerateRandomToken(t *testing.T) {
	token1, err := generateRandomToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	token2, err := generateRandomToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token1 == token2 {
		t.Error("two random tokens should not be equal")
	}

	if len(token1) != 64 {
		t.Errorf("token should be 64 chars (32 bytes hex), got %d", len(token1))
	}
}

func TestGenerateAccessToken(t *testing.T) {
	s := &AuthService{
		jwtSecret: []byte("test-secret"),
	}

	uid := mustParseUUID("550e8400-e29b-41d4-a716-446655440000")

	token, err := s.generateAccessToken(uid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token == "" {
		t.Error("token should not be empty")
	}

	if len(token) < 50 {
		t.Errorf("token seems too short: %d chars", len(token))
	}
}
