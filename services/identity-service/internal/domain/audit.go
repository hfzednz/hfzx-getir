package domain

import (
	"fmt"
	"net/netip"
	"time"

	"github.com/google/uuid"
)

// AuditOutcome is the result of an audited action.
type AuditOutcome string

const (
	AuditOutcomeSuccess AuditOutcome = "success"
	AuditOutcomeFailure AuditOutcome = "failure"
	AuditOutcomeDenied  AuditOutcome = "denied"
)

func (o AuditOutcome) Valid() bool {
	switch o {
	case AuditOutcomeSuccess, AuditOutcomeFailure, AuditOutcomeDenied:
		return true
	default:
		return false
	}
}

// AuditEvent is an append-only security/compliance event.
type AuditEvent struct {
	ID           uuid.UUID
	TenantID     *uuid.UUID
	ActorID      *uuid.UUID
	ActorKind    string
	Action       string
	ResourceType string
	ResourceID   string
	Outcome      AuditOutcome
	IP           *netip.Addr
	UserAgent    string
	SessionID    *uuid.UUID
	RequestID    string
	Details      map[string]any
	CreatedAt    time.Time
}

func (e AuditEvent) Validate() error {
	if e.ID == uuid.Nil {
		return fmt.Errorf("%w: audit event id required", ErrInvalidArgument)
	}
	if e.Action == "" {
		return fmt.Errorf("%w: action required", ErrInvalidArgument)
	}
	if e.ResourceType == "" {
		return fmt.Errorf("%w: resource_type required", ErrInvalidArgument)
	}
	if !e.Outcome.Valid() {
		return fmt.Errorf("%w: invalid audit outcome %q", ErrInvalidArgument, e.Outcome)
	}
	if e.ActorKind == "" {
		return fmt.Errorf("%w: actor_kind required", ErrInvalidArgument)
	}
	return nil
}

// Consent records a versioned privacy/terms consent decision.
type Consent struct {
	ID          uuid.UUID
	PrincipalID uuid.UUID
	TenantID    uuid.UUID
	Purpose     string
	Version     string
	Granted     bool
	GrantedAt   *time.Time
	RevokedAt   *time.Time
	IP          *netip.Addr
	UserAgent   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (c Consent) IsActive() bool {
	return c.Granted && c.RevokedAt == nil
}

func (c Consent) Validate() error {
	if c.ID == uuid.Nil {
		return fmt.Errorf("%w: consent id required", ErrInvalidArgument)
	}
	if c.PrincipalID == uuid.Nil {
		return fmt.Errorf("%w: principal_id required", ErrInvalidArgument)
	}
	if c.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if c.Purpose == "" || c.Version == "" {
		return fmt.Errorf("%w: purpose and version required", ErrInvalidArgument)
	}
	if c.Granted && c.GrantedAt == nil {
		return fmt.Errorf("%w: granted consent requires granted_at", ErrInvariant)
	}
	return nil
}

// SecurityPolicy is a tenant/platform authn and session policy bundle.
type SecurityPolicy struct {
	ID                       uuid.UUID
	TenantID                 *uuid.UUID
	Name                     string
	Description              string
	Enabled                  bool
	PasswordMinLength        int
	PasswordRequireUpper     bool
	PasswordRequireLower     bool
	PasswordRequireDigit     bool
	PasswordRequireSymbol    bool
	PasswordHistoryCount     int
	MFARequired              bool
	MFARequiredAboveRisk     float64
	SessionIdleSeconds       int
	SessionAbsoluteSeconds   int
	RefreshTokenSeconds      int
	MaxConcurrentSessions    *int
	MaxFailedAttempts        int
	LockoutSeconds           int
	BlockAboveRisk           float64
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

func (p SecurityPolicy) SessionIdleTTL() time.Duration {
	return time.Duration(p.SessionIdleSeconds) * time.Second
}

func (p SecurityPolicy) SessionAbsoluteTTL() time.Duration {
	return time.Duration(p.SessionAbsoluteSeconds) * time.Second
}

func (p SecurityPolicy) RefreshTokenTTL() time.Duration {
	return time.Duration(p.RefreshTokenSeconds) * time.Second
}

func (p SecurityPolicy) Validate() error {
	if p.ID == uuid.Nil {
		return fmt.Errorf("%w: security policy id required", ErrInvalidArgument)
	}
	if p.Name == "" {
		return fmt.Errorf("%w: name required", ErrInvalidArgument)
	}
	if p.PasswordMinLength < 8 {
		return fmt.Errorf("%w: password_min_length must be >= 8", ErrInvalidArgument)
	}
	if p.SessionIdleSeconds <= 0 || p.SessionAbsoluteSeconds < p.SessionIdleSeconds {
		return fmt.Errorf("%w: invalid session TTL configuration", ErrInvalidArgument)
	}
	if p.RefreshTokenSeconds <= 0 {
		return fmt.Errorf("%w: refresh_token_seconds must be > 0", ErrInvalidArgument)
	}
	if p.MFARequiredAboveRisk < 0 || p.MFARequiredAboveRisk > 100 {
		return fmt.Errorf("%w: mfa_required_above_risk must be 0..100", ErrInvalidArgument)
	}
	if p.BlockAboveRisk < 0 || p.BlockAboveRisk > 100 {
		return fmt.Errorf("%w: block_above_risk must be 0..100", ErrInvalidArgument)
	}
	return nil
}
