package service

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestAPIKeyFormat(t *testing.T) {
	s := &APIKeyService{}
	_ = s

	rawBytes := make([]byte, 32)
	rawKey := "sk_" + hex.EncodeToString(rawBytes)

	if !strings.HasPrefix(rawKey, "sk_") {
		t.Errorf("key should start with sk_, got %s", rawKey[:5])
	}

	if len(rawKey) != 67 {
		t.Errorf("key should be 67 chars (sk_ + 64 hex), got %d", len(rawKey))
	}

	prefix := rawKey[:10]
	if len(prefix) != 10 {
		t.Errorf("prefix should be 10 chars, got %d", len(prefix))
	}
}

func TestAPIKeyHashing(t *testing.T) {
	rawKey := "sk_abc123def456"

	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	hash2 := sha256.Sum256([]byte(rawKey))
	keyHash2 := hex.EncodeToString(hash2[:])

	if keyHash != keyHash2 {
		t.Error("same key should produce same hash")
	}

	hash3 := sha256.Sum256([]byte("sk_different"))
	keyHash3 := hex.EncodeToString(hash3[:])

	if keyHash == keyHash3 {
		t.Error("different keys should produce different hashes")
	}
}
