package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RefreshToken is an opaque refresh credential with rotation family tracking.
type RefreshToken struct {
	ID           uuid.UUID
	SessionID    uuid.UUID
	PrincipalID  uuid.UUID
	TokenHash    string
	FamilyID     uuid.UUID
	RotatedFrom  *uuid.UUID
	ExpiresAt    time.Time
	CreatedAt    time.Time
	RevokedAt    *time.Time
	RevokeReason string
}

func (t RefreshToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

func (t RefreshToken) IsExpired(now time.Time) bool {
	return now.After(t.ExpiresAt)
}

func (t RefreshToken) IsUsable(now time.Time) bool {
	return !t.IsRevoked() && !t.IsExpired(now)
}

func (t RefreshToken) Validate() error {
	if t.ID == uuid.Nil {
		return fmt.Errorf("%w: refresh token id required", ErrInvalidArgument)
	}
	if t.SessionID == uuid.Nil {
		return fmt.Errorf("%w: session_id required", ErrInvalidArgument)
	}
	if t.PrincipalID == uuid.Nil {
		return fmt.Errorf("%w: principal_id required", ErrInvalidArgument)
	}
	if t.TokenHash == "" {
		return fmt.Errorf("%w: token_hash required", ErrInvalidArgument)
	}
	if t.FamilyID == uuid.Nil {
		return fmt.Errorf("%w: family_id required", ErrInvalidArgument)
	}
	return nil
}

// Revoke marks the token revoked.
func (t *RefreshToken) Revoke(now time.Time, reason string) {
	if t.RevokedAt != nil {
		return
	}
	t.RevokedAt = &now
	t.RevokeReason = reason
}

// AccessTokenClaims are the logical claims carried by a short-lived access JWT.
// This is a domain view — encoding/signing lives in security adapters.
type AccessTokenClaims struct {
	Subject  uuid.UUID // sub
	SessionID uuid.UUID // sid
	TenantID uuid.UUID // tid
	Roles    []string
	AMR      []string
	ACR      string
	DeviceID *uuid.UUID
	ExpiresAt time.Time
	IssuedAt  time.Time
}

func (c AccessTokenClaims) Validate() error {
	if c.Subject == uuid.Nil {
		return fmt.Errorf("%w: sub required", ErrInvalidArgument)
	}
	if c.SessionID == uuid.Nil {
		return fmt.Errorf("%w: sid required", ErrInvalidArgument)
	}
	if c.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tid required", ErrInvalidArgument)
	}
	if !c.ExpiresAt.After(c.IssuedAt) {
		return fmt.Errorf("%w: exp must be after iat", ErrInvariant)
	}
	return nil
}

func (c AccessTokenClaims) IsExpired(now time.Time) bool {
	return now.After(c.ExpiresAt)
}
