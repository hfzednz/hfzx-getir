package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// FraudDecision is the allow/challenge/block outcome.
type FraudDecision string

const (
	FraudAllow     FraudDecision = "allow"
	FraudChallenge FraudDecision = "challenge"
	FraudBlock     FraudDecision = "block"
)

// FraudScore is a persisted risk assessment for an intent.
type FraudScore struct {
	ID         uuid.UUID
	IntentID   uuid.UUID
	TenantID   uuid.UUID
	Score      int // 0–100
	Decision   FraudDecision
	Reasons    []string
	CreatedAt  time.Time
}

// Validate checks fraud score invariants.
func (f FraudScore) Validate() error {
	if f.ID == uuid.Nil {
		return fmt.Errorf("%w: fraud score id required", ErrInvalidArgument)
	}
	if f.Score < 0 || f.Score > 100 {
		return fmt.Errorf("%w: score must be 0–100", ErrInvalidArgument)
	}
	switch f.Decision {
	case FraudAllow, FraudChallenge, FraudBlock:
	default:
		return fmt.Errorf("%w: invalid decision %q", ErrInvalidArgument, f.Decision)
	}
	return nil
}
