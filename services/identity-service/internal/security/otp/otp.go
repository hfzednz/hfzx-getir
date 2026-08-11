// Package otp provides numeric OTP generation, hashing, verification, and TTL helpers.
package otp

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"
)

var (
	ErrInvalidLength = errors.New("otp: invalid length")
	ErrEmptyOTP      = errors.New("otp: empty otp")
	ErrExpired       = errors.New("otp: expired")
)

const (
	// DefaultLength is a typical 6-digit OTP length.
	DefaultLength = 6
	// DefaultTTL is a typical OTP lifetime.
	DefaultTTL = 5 * time.Minute
	// MaxLength caps GenerateNumeric to avoid absurd sizes.
	MaxLength = 12
)

// GenerateNumeric returns an n-digit numeric OTP (leading zeros preserved as string).
func GenerateNumeric(n int) (string, error) {
	if n <= 0 || n > MaxLength {
		return "", fmt.Errorf("%w: must be 1..%d", ErrInvalidLength, MaxLength)
	}
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
	v, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("otp: generate: %w", err)
	}
	format := fmt.Sprintf("%%0%dd", n)
	return fmt.Sprintf(format, v), nil
}

// HashOTP returns a SHA-256 hex digest of the OTP (optionally peppered).
// Pepper should be a server-side secret; empty pepper is allowed for tests.
func HashOTP(otp, pepper string) (string, error) {
	if otp == "" {
		return "", ErrEmptyOTP
	}
	h := sha256.New()
	_, _ = h.Write([]byte(pepper))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(otp))
	return hex.EncodeToString(h.Sum(nil)), nil
}

// VerifyOTP compares a candidate OTP against a stored hash using a timing-safe compare.
func VerifyOTP(otp, pepper, storedHash string) (bool, error) {
	if otp == "" || storedHash == "" {
		return false, ErrEmptyOTP
	}
	candidate, err := HashOTP(otp, pepper)
	if err != nil {
		return false, err
	}
	a := []byte(candidate)
	b := []byte(storedHash)
	if len(a) != len(b) {
		// Normalize lengths for ConstantTimeCompare contract; still reject.
		return false, nil
	}
	return subtle.ConstantTimeCompare(a, b) == 1, nil
}

// ExpiresAt returns now+ttl.
func ExpiresAt(ttl time.Duration) time.Time {
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	return time.Now().UTC().Add(ttl)
}

// IsExpired reports whether expiresAt is before now (UTC).
func IsExpired(expiresAt time.Time) bool {
	return time.Now().UTC().After(expiresAt.UTC())
}

// Remaining returns the time left until expiry; zero if already expired.
func Remaining(expiresAt time.Time) time.Duration {
	d := expiresAt.UTC().Sub(time.Now().UTC())
	if d < 0 {
		return 0
	}
	return d
}

// EnsureValid returns ErrExpired when the OTP window has elapsed.
func EnsureValid(expiresAt time.Time) error {
	if IsExpired(expiresAt) {
		return ErrExpired
	}
	return nil
}

// Challenge holds a hashed OTP and its expiry for persistence.
type Challenge struct {
	Hash      string
	ExpiresAt time.Time
}

// NewChallenge generates a numeric OTP, hashes it, and returns the plaintext
// (for delivery) plus the Challenge to store.
func NewChallenge(length int, pepper string, ttl time.Duration) (plaintext string, ch Challenge, err error) {
	if length <= 0 {
		length = DefaultLength
	}
	plaintext, err = GenerateNumeric(length)
	if err != nil {
		return "", Challenge{}, err
	}
	hash, err := HashOTP(plaintext, pepper)
	if err != nil {
		return "", Challenge{}, err
	}
	ch = Challenge{
		Hash:      hash,
		ExpiresAt: ExpiresAt(ttl),
	}
	return plaintext, ch, nil
}

// VerifyChallenge verifies an OTP against a stored challenge and expiry.
func VerifyChallenge(otp, pepper string, ch Challenge) (bool, error) {
	if err := EnsureValid(ch.ExpiresAt); err != nil {
		return false, err
	}
	return VerifyOTP(otp, pepper, ch.Hash)
}
