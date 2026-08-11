package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CredentialAlgorithm is the password hashing algorithm.
type CredentialAlgorithm string

const (
	CredentialAlgorithmArgon2id CredentialAlgorithm = "argon2id"
)

func (a CredentialAlgorithm) Valid() bool {
	return a == CredentialAlgorithmArgon2id
}

// MFAFactorType enumerates enrolled second-factor kinds.
type MFAFactorType string

const (
	MFAFactorTOTP     MFAFactorType = "totp"
	MFAFactorSMS      MFAFactorType = "sms"
	MFAFactorEmail    MFAFactorType = "email"
	MFAFactorWebAuthn MFAFactorType = "webauthn"
	MFAFactorPush     MFAFactorType = "push"
	MFAFactorHardware MFAFactorType = "hardware"
)

func (t MFAFactorType) Valid() bool {
	switch t {
	case MFAFactorTOTP, MFAFactorSMS, MFAFactorEmail, MFAFactorWebAuthn, MFAFactorPush, MFAFactorHardware:
		return true
	default:
		return false
	}
}

// Credential is the primary password material for a principal.
type Credential struct {
	ID                uuid.UUID
	PrincipalID       uuid.UUID
	PasswordHash      string
	Algorithm         CredentialAlgorithm
	PasswordChangedAt time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (c Credential) Validate() error {
	if c.ID == uuid.Nil {
		return fmt.Errorf("%w: credential id required", ErrInvalidArgument)
	}
	if c.PrincipalID == uuid.Nil {
		return fmt.Errorf("%w: principal_id required", ErrInvalidArgument)
	}
	if c.PasswordHash == "" {
		return fmt.Errorf("%w: password_hash required", ErrInvalidArgument)
	}
	if !c.Algorithm.Valid() {
		return fmt.Errorf("%w: unsupported algorithm %q", ErrInvalidArgument, c.Algorithm)
	}
	return nil
}

// PasswordHistoryEntry records a past password hash for reuse checks.
type PasswordHistoryEntry struct {
	ID           uuid.UUID
	PrincipalID  uuid.UUID
	PasswordHash string
	Algorithm    CredentialAlgorithm
	CreatedAt    time.Time
}

// MFAFactor is an enrolled second factor.
type MFAFactor struct {
	ID          uuid.UUID
	PrincipalID uuid.UUID
	Type        MFAFactorType
	Label       string
	SecretEnc   []byte
	Verified    bool
	VerifiedAt  *time.Time
	DisabledAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (f MFAFactor) IsActive() bool {
	return f.Verified && f.DisabledAt == nil
}

func (f MFAFactor) Validate() error {
	if f.ID == uuid.Nil {
		return fmt.Errorf("%w: mfa factor id required", ErrInvalidArgument)
	}
	if f.PrincipalID == uuid.Nil {
		return fmt.Errorf("%w: principal_id required", ErrInvalidArgument)
	}
	if !f.Type.Valid() {
		return fmt.Errorf("%w: invalid mfa factor type %q", ErrInvalidArgument, f.Type)
	}
	if f.Verified && f.VerifiedAt == nil {
		return fmt.Errorf("%w: verified factor requires verified_at", ErrInvariant)
	}
	return nil
}

// BackupCode is a one-time MFA recovery code (stored hashed).
type BackupCode struct {
	ID          uuid.UUID
	PrincipalID uuid.UUID
	CodeHash    string
	UsedAt      *time.Time
	CreatedAt   time.Time
}

func (b BackupCode) IsConsumed() bool {
	return b.UsedAt != nil
}

// WebAuthnCredential is a registered passkey / security key.
type WebAuthnCredential struct {
	ID             uuid.UUID
	PrincipalID    uuid.UUID
	CredentialID   []byte
	PublicKey      []byte
	AAGUID         []byte
	SignCount      uint64
	Transports     []string
	Nickname       string
	BackupEligible bool
	BackupState    bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LastUsedAt     *time.Time
	RevokedAt      *time.Time
}

func (w WebAuthnCredential) IsActive() bool {
	return w.RevokedAt == nil
}

func (w WebAuthnCredential) Validate() error {
	if w.ID == uuid.Nil {
		return fmt.Errorf("%w: webauthn credential id required", ErrInvalidArgument)
	}
	if w.PrincipalID == uuid.Nil {
		return fmt.Errorf("%w: principal_id required", ErrInvalidArgument)
	}
	if len(w.CredentialID) == 0 {
		return fmt.Errorf("%w: credential_id required", ErrInvalidArgument)
	}
	if len(w.PublicKey) == 0 {
		return fmt.Errorf("%w: public_key required", ErrInvalidArgument)
	}
	return nil
}

// AdvanceSignCount updates the signature counter; rejects regressions (clone signal).
func (w *WebAuthnCredential) AdvanceSignCount(next uint64) error {
	if next < w.SignCount {
		return fmt.Errorf("%w: sign_count regression (%d < %d)", ErrSecurityViolation, next, w.SignCount)
	}
	w.SignCount = next
	return nil
}
