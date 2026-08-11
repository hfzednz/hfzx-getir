package domain

import (
	"fmt"
	"net/netip"
	"time"

	"github.com/google/uuid"
)

// Session is an authenticated session with idle and absolute expiry.
type Session struct {
	ID                 uuid.UUID
	PrincipalID        uuid.UUID
	DeviceID           *uuid.UUID
	TenantID           uuid.UUID
	AMR                []string
	ACR                string
	IP                 *netip.Addr
	UserAgent          string
	RiskScore          float64
	IdleExpiresAt      time.Time
	AbsoluteExpiresAt  time.Time
	LastSeenAt         time.Time
	CreatedAt          time.Time
	RevokedAt          *time.Time
	RevokeReason       string
}

func (s Session) IsRevoked() bool {
	return s.RevokedAt != nil
}

// IsExpired reports idle or absolute timeout at the given instant.
func (s Session) IsExpired(now time.Time) bool {
	return now.After(s.IdleExpiresAt) || now.After(s.AbsoluteExpiresAt)
}

// IsUsable is true when the session may mint tokens / authorize requests.
func (s Session) IsUsable(now time.Time) bool {
	return !s.IsRevoked() && !s.IsExpired(now)
}

func (s Session) Validate() error {
	if s.ID == uuid.Nil {
		return fmt.Errorf("%w: session id required", ErrInvalidArgument)
	}
	if s.PrincipalID == uuid.Nil {
		return fmt.Errorf("%w: principal_id required", ErrInvalidArgument)
	}
	if s.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if s.RiskScore < 0 || s.RiskScore > 100 {
		return fmt.Errorf("%w: risk_score must be 0..100", ErrInvalidArgument)
	}
	if s.AbsoluteExpiresAt.Before(s.IdleExpiresAt) {
		return fmt.Errorf("%w: absolute expiry before idle expiry", ErrInvariant)
	}
	return nil
}

// Touch slides the idle window without extending absolute expiry.
func (s *Session) Touch(now time.Time, idleTTL time.Duration) error {
	if !s.IsUsable(now) {
		return ErrSessionExpired
	}
	nextIdle := now.Add(idleTTL)
	if nextIdle.After(s.AbsoluteExpiresAt) {
		nextIdle = s.AbsoluteExpiresAt
	}
	s.IdleExpiresAt = nextIdle
	s.LastSeenAt = now
	return nil
}

// Revoke marks the session revoked.
func (s *Session) Revoke(now time.Time, reason string) {
	if s.RevokedAt != nil {
		return
	}
	s.RevokedAt = &now
	s.RevokeReason = reason
}
