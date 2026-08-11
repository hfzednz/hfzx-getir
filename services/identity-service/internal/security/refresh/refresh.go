// Package refresh provides opaque refresh-token generation, hashing, and family IDs.
package refresh

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

const (
	// DefaultTokenBytes is the entropy size for opaque refresh tokens (32 bytes → 256-bit).
	DefaultTokenBytes = 32
)

var ErrEmptyToken = errors.New("refresh: empty token")

// Token is an opaque refresh token presented to clients.
type Token struct {
	// Raw is the client-facing opaque string (base64url, no padding).
	Raw string
	// Hash is the SHA-256 hex digest for storage (never store Raw).
	Hash string
	// FamilyID groups a rotation chain; reuse of an old token revokes the family.
	FamilyID string
}

// Generate creates a new opaque refresh token and its storage hash.
// If familyID is empty, a new UUID family is created.
func Generate(familyID string) (Token, error) {
	return GenerateN(DefaultTokenBytes, familyID)
}

// GenerateN creates an opaque token with n random bytes.
func GenerateN(n int, familyID string) (Token, error) {
	if n < 16 {
		n = DefaultTokenBytes
	}
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return Token{}, fmt.Errorf("refresh: generate: %w", err)
	}
	raw := base64.RawURLEncoding.EncodeToString(buf)
	hash, err := Hash(raw)
	if err != nil {
		return Token{}, err
	}
	if familyID == "" {
		familyID = NewFamilyID()
	}
	return Token{
		Raw:      raw,
		Hash:     hash,
		FamilyID: familyID,
	}, nil
}

// Hash returns the SHA-256 hex digest of a raw refresh token for storage.
func Hash(raw string) (string, error) {
	if raw == "" {
		return "", ErrEmptyToken
	}
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:]), nil
}

// NewFamilyID returns a new UUID string for a refresh-token family.
func NewFamilyID() string {
	return uuid.NewString()
}

// Matches reports whether raw hashes to storedHash.
func Matches(raw, storedHash string) (bool, error) {
	h, err := Hash(raw)
	if err != nil {
		return false, err
	}
	return h == storedHash, nil
}
