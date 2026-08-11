package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/checkout-service/internal/domain"
)

// AbandonInput marks a session abandoned (with recovery token).
type AbandonInput struct {
	TenantID  uuid.UUID
	SessionID uuid.UUID
}

// Abandon marks a non-terminal session as abandoned.
func (d *Deps) Abandon(ctx context.Context, in AbandonInput) (domain.Session, error) {
	s, err := d.Sessions.GetByID(ctx, in.TenantID, in.SessionID)
	if err != nil {
		return domain.Session{}, err
	}
	if s.Status.IsTerminal() {
		return domain.Session{}, fmt.Errorf("%w: already terminal %s", domain.ErrInvalidTransition, s.Status)
	}
	if s.RecoveryToken == "" {
		s.RecoveryToken = d.newID().String()
	}
	if err := d.transition(&s, domain.StatusAbandoned); err != nil {
		return domain.Session{}, err
	}
	if err := d.Sessions.Update(ctx, s); err != nil {
		return domain.Session{}, err
	}
	_ = d.appendEvent(ctx, s.ID, s.TenantID, domain.EventCheckoutAbandoned, map[string]any{
		"recoveryToken": s.RecoveryToken,
	})
	return s, nil
}

// RecoverAbandonedInput restores an abandoned session via recovery token.
type RecoverAbandonedInput struct {
	RecoveryToken string
	PrincipalID   uuid.UUID
}

// RecoverAbandoned brings an abandoned session back to started for re-validation.
func (d *Deps) RecoverAbandoned(ctx context.Context, in RecoverAbandonedInput) (domain.Session, error) {
	token := strings.TrimSpace(in.RecoveryToken)
	if token == "" {
		return domain.Session{}, fmt.Errorf("%w: recovery_token required", domain.ErrInvalidArgument)
	}
	s, err := d.Sessions.GetByRecoveryToken(ctx, token)
	if err != nil {
		return domain.Session{}, err
	}
	if s.Status != domain.StatusAbandoned {
		return domain.Session{}, fmt.Errorf("%w: session not abandoned", domain.ErrConflict)
	}
	if in.PrincipalID != uuid.Nil && in.PrincipalID != s.PrincipalID {
		return domain.Session{}, fmt.Errorf("%w: principal mismatch", domain.ErrForbidden)
	}
	// Abandoned is terminal in the transition table; allow recovery by direct reset
	// (explicit recovery path, not a normal transition).
	now := d.now()
	s.Status = domain.StatusStarted
	s.Validation = domain.ValidationResults{}
	s.AbandonedAt = nil
	s.UpdatedAt = now
	s.Version++
	s.RecoveryToken = d.newID().String() // rotate
	if err := s.Validate(); err != nil {
		return domain.Session{}, err
	}
	if err := d.Sessions.Update(ctx, s); err != nil {
		return domain.Session{}, err
	}
	_ = d.appendEvent(ctx, s.ID, s.TenantID, domain.EventCheckoutRecovered, map[string]any{
		"status": string(s.Status),
	})
	return s, nil
}
