// Package password provides Argon2id hashing, complexity validation,
// password history checks, and a breach-check port for NEXORA identity-service.
package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 3
	argonMemory  = 64 * 1024 // 64 MiB
	argonThreads = 4
	argonKeyLen  = 32
	saltLen      = 16
)

var (
	ErrEmptyPassword       = errors.New("password: empty password")
	ErrInvalidHash         = errors.New("password: invalid hash encoding")
	ErrComplexityFailed    = errors.New("password: complexity requirements not met")
	ErrPasswordReused      = errors.New("password: matches a previous password")
	ErrPasswordBreached    = errors.New("password: found in known breach corpus")
)

// Params holds Argon2id parameters used when hashing.
type Params struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
	SaltLen uint32
}

// DefaultParams returns production-sensible Argon2id parameters.
func DefaultParams() Params {
	return Params{
		Time:    argonTime,
		Memory:  argonMemory,
		Threads: argonThreads,
		KeyLen:  argonKeyLen,
		SaltLen: saltLen,
	}
}

// Hasher hashes and verifies passwords with Argon2id.
type Hasher struct {
	params Params
}

// NewHasher returns a Hasher with the given parameters.
func NewHasher(p Params) *Hasher {
	if p.Time == 0 {
		p.Time = argonTime
	}
	if p.Memory == 0 {
		p.Memory = argonMemory
	}
	if p.Threads == 0 {
		p.Threads = argonThreads
	}
	if p.KeyLen == 0 {
		p.KeyLen = argonKeyLen
	}
	if p.SaltLen == 0 {
		p.SaltLen = saltLen
	}
	return &Hasher{params: p}
}

// NewDefaultHasher returns a Hasher with DefaultParams.
func NewDefaultHasher() *Hasher {
	return NewHasher(DefaultParams())
}

// Hash returns a PHC-style encoded Argon2id hash:
// $argon2id$v=19$m=65536,t=3,p=4$<salt_b64>$<hash_b64>
func (h *Hasher) Hash(password string) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}
	salt := make([]byte, h.params.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("password: generate salt: %w", err)
	}
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.params.Time,
		h.params.Memory,
		h.params.Threads,
		h.params.KeyLen,
	)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.params.Memory,
		h.params.Time,
		h.params.Threads,
		b64Salt,
		b64Hash,
	)
	return encoded, nil
}

// Verify checks password against an encoded Argon2id hash using a constant-time compare.
func (h *Hasher) Verify(password, encodedHash string) (bool, error) {
	if password == "" || encodedHash == "" {
		return false, ErrEmptyPassword
	}
	p, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}
	other := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Threads, p.KeyLen)
	if subtle.ConstantTimeCompare(hash, other) == 1 {
		return true, nil
	}
	return false, nil
}

func decodeHash(encoded string) (Params, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	// "", "argon2id", "v=19", "m=...,t=...,p=...", salt, hash
	if len(parts) != 6 || parts[1] != "argon2id" {
		return Params{}, nil, nil, ErrInvalidHash
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return Params{}, nil, nil, ErrInvalidHash
	}
	if version != argon2.Version {
		return Params{}, nil, nil, ErrInvalidHash
	}
	var p Params
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Time, &p.Threads); err != nil {
		return Params{}, nil, nil, ErrInvalidHash
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return Params{}, nil, nil, ErrInvalidHash
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return Params{}, nil, nil, ErrInvalidHash
	}
	p.KeyLen = uint32(len(hash))
	p.SaltLen = uint32(len(salt))
	return p, salt, hash, nil
}

// ComplexityPolicy configures password complexity rules.
type ComplexityPolicy struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSymbol  bool
	MaxLength      int // 0 = no max
}

// DefaultComplexityPolicy returns a sensible production policy.
func DefaultComplexityPolicy() ComplexityPolicy {
	return ComplexityPolicy{
		MinLength:     12,
		RequireUpper:  true,
		RequireLower:  true,
		RequireDigit:  true,
		RequireSymbol: true,
		MaxLength:     128,
	}
}

// ValidateComplexity checks password against the policy.
func ValidateComplexity(password string, policy ComplexityPolicy) error {
	if policy.MinLength <= 0 {
		policy.MinLength = 8
	}
	runes := []rune(password)
	n := len(runes)
	if n < policy.MinLength {
		return fmt.Errorf("%w: minimum length %d", ErrComplexityFailed, policy.MinLength)
	}
	if policy.MaxLength > 0 && n > policy.MaxLength {
		return fmt.Errorf("%w: maximum length %d", ErrComplexityFailed, policy.MaxLength)
	}

	var hasUpper, hasLower, hasDigit, hasSymbol bool
	for _, r := range runes {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbol = true
		}
	}

	var missing []string
	if policy.RequireUpper && !hasUpper {
		missing = append(missing, "uppercase")
	}
	if policy.RequireLower && !hasLower {
		missing = append(missing, "lowercase")
	}
	if policy.RequireDigit && !hasDigit {
		missing = append(missing, "digit")
	}
	if policy.RequireSymbol && !hasSymbol {
		missing = append(missing, "symbol")
	}
	if len(missing) > 0 {
		return fmt.Errorf("%w: missing %s", ErrComplexityFailed, strings.Join(missing, ", "))
	}
	return nil
}

// CheckHistory returns ErrPasswordReused if password matches any previous encoded hash.
func CheckHistory(password string, previousHashes []string, hasher *Hasher) error {
	if hasher == nil {
		hasher = NewDefaultHasher()
	}
	for _, prev := range previousHashes {
		ok, err := hasher.Verify(password, prev)
		if err != nil {
			// skip malformed historical hashes
			continue
		}
		if ok {
			return ErrPasswordReused
		}
	}
	return nil
}

// BreachChecker is the HaveIBeenPwned-style port for checking breached passwords.
// Production implementations call the Pwned Passwords k-anonymity API (or a local corpus).
type BreachChecker interface {
	// IsBreached reports whether the password appears in known breach data.
	IsBreached(password string) (bool, error)
}

// MockBreachChecker is a test/stub BreachChecker backed by an in-memory set.
type MockBreachChecker struct {
	// Breached is a set of plaintext passwords considered breached.
	Breached map[string]struct{}
	// Err, when set, is returned from every IsBreached call.
	Err error
}

// IsBreached implements BreachChecker.
func (m *MockBreachChecker) IsBreached(password string) (bool, error) {
	if m == nil {
		return false, nil
	}
	if m.Err != nil {
		return false, m.Err
	}
	if m.Breached == nil {
		return false, nil
	}
	_, ok := m.Breached[password]
	return ok, nil
}

// NewMockBreachChecker returns a MockBreachChecker seeded with the given passwords.
func NewMockBreachChecker(breached ...string) *MockBreachChecker {
	m := &MockBreachChecker{Breached: make(map[string]struct{}, len(breached))}
	for _, p := range breached {
		m.Breached[p] = struct{}{}
	}
	return m
}

// EnsureNotBreached returns ErrPasswordBreached when the checker reports a hit.
func EnsureNotBreached(password string, checker BreachChecker) error {
	if checker == nil {
		return nil
	}
	hit, err := checker.IsBreached(password)
	if err != nil {
		return fmt.Errorf("password: breach check: %w", err)
	}
	if hit {
		return ErrPasswordBreached
	}
	return nil
}
