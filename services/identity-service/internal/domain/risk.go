package domain

import (
	"fmt"
	"net/netip"
	"time"

	"github.com/google/uuid"
)

// RiskSeverity classifies a risk signal.
type RiskSeverity string

const (
	RiskSeverityInfo     RiskSeverity = "info"
	RiskSeverityLow      RiskSeverity = "low"
	RiskSeverityMedium   RiskSeverity = "medium"
	RiskSeverityHigh     RiskSeverity = "high"
	RiskSeverityCritical RiskSeverity = "critical"
)

func (s RiskSeverity) Valid() bool {
	switch s {
	case RiskSeverityInfo, RiskSeverityLow, RiskSeverityMedium, RiskSeverityHigh, RiskSeverityCritical:
		return true
	default:
		return false
	}
}

// Common risk event type constants.
const (
	RiskEventImpossibleTravel   = "impossible_travel"
	RiskEventNewDevice          = "new_device"
	RiskEventVPNDetected        = "vpn_detected"
	RiskEventCredentialStuffing = "credential_stuffing"
	RiskEventTokenReuse         = "refresh_token_reuse"
	RiskEventAnomalousBehavior  = "anomalous_behavior"
)

// RiskEvent is a single risk signal against a principal/session/device.
type RiskEvent struct {
	ID          uuid.UUID
	PrincipalID *uuid.UUID
	SessionID   *uuid.UUID
	DeviceID    *uuid.UUID
	TenantID    *uuid.UUID
	EventType   string
	Severity    RiskSeverity
	ScoreDelta  float64
	ScoreAfter  *float64
	IP          *netip.Addr
	Details     map[string]any
	CreatedAt   time.Time
}

func (e RiskEvent) Validate() error {
	if e.ID == uuid.Nil {
		return fmt.Errorf("%w: risk event id required", ErrInvalidArgument)
	}
	if e.EventType == "" {
		return fmt.Errorf("%w: event_type required", ErrInvalidArgument)
	}
	if !e.Severity.Valid() {
		return fmt.Errorf("%w: invalid severity %q", ErrInvalidArgument, e.Severity)
	}
	if e.ScoreAfter != nil && (*e.ScoreAfter < 0 || *e.ScoreAfter > 100) {
		return fmt.Errorf("%w: score_after must be 0..100", ErrInvalidArgument)
	}
	return nil
}

// LoginAttemptResult is the outcome of an authentication attempt.
type LoginAttemptResult string

const (
	LoginAttemptSuccess            LoginAttemptResult = "success"
	LoginAttemptInvalidCredentials LoginAttemptResult = "invalid_credentials"
	LoginAttemptLocked             LoginAttemptResult = "locked"
	LoginAttemptSuspended          LoginAttemptResult = "suspended"
	LoginAttemptMFARequired        LoginAttemptResult = "mfa_required"
	LoginAttemptMFAFailed          LoginAttemptResult = "mfa_failed"
	LoginAttemptBlockedRisk        LoginAttemptResult = "blocked_risk"
	LoginAttemptRateLimited        LoginAttemptResult = "rate_limited"
)

func (r LoginAttemptResult) Valid() bool {
	switch r {
	case LoginAttemptSuccess, LoginAttemptInvalidCredentials, LoginAttemptLocked,
		LoginAttemptSuspended, LoginAttemptMFARequired, LoginAttemptMFAFailed,
		LoginAttemptBlockedRisk, LoginAttemptRateLimited:
		return true
	default:
		return false
	}
}

// LoginAttempt records an auth attempt for lockout and risk.
type LoginAttempt struct {
	ID                uuid.UUID
	TenantID          *uuid.UUID
	PrincipalID       *uuid.UUID
	Identifier        string
	Result            LoginAttemptResult
	IP                *netip.Addr
	UserAgent         string
	DeviceFingerprint string
	FailureReason     string
	CreatedAt         time.Time
}

func (a LoginAttempt) Validate() error {
	if a.ID == uuid.Nil {
		return fmt.Errorf("%w: login attempt id required", ErrInvalidArgument)
	}
	if !a.Result.Valid() {
		return fmt.Errorf("%w: invalid login attempt result %q", ErrInvalidArgument, a.Result)
	}
	return nil
}

// ClampRiskScore keeps a score within the canonical 0..100 range.
func ClampRiskScore(score float64) float64 {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}
