package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
)

// Refresh rotates a refresh token (with reuse detection → family revoke).
func (d *Deps) Refresh(ctx context.Context, rawRefresh string) (ports.TokenPair, error) {
	if rawRefresh == "" {
		return ports.TokenPair{}, fmt.Errorf("%w: refresh token required", domain.ErrInvalidArgument)
	}
	return d.Tokens.RotateRefresh(ctx, rawRefresh)
}

// LogoutInput revokes a session (and optionally its refresh family via session revoke).
type LogoutInput struct {
	SessionID uuid.UUID
	ActorID   uuid.UUID
	Reason    string
	AuthCtx   ports.AuthContext
}

// Logout revokes the given session.
func (d *Deps) Logout(ctx context.Context, in LogoutInput) error {
	if in.SessionID == uuid.Nil {
		return fmt.Errorf("%w: session_id required", domain.ErrInvalidArgument)
	}
	sess, err := d.Sessions.GetByID(ctx, in.SessionID)
	if err != nil {
		return err
	}
	if in.ActorID != uuid.Nil && sess.PrincipalID != in.ActorID {
		return domain.ErrForbidden
	}
	now := d.now()
	reason := in.Reason
	if reason == "" {
		reason = "logout"
	}
	if err := d.Sessions.Revoke(ctx, sess.ID, now, reason); err != nil {
		return err
	}
	// Best-effort: revoke refresh tokens bound to this session by re-loading and updating.
	// Full family revoke happens on reuse; here we mark the session revoked so RotateRefresh fails.
	d.publishSession(ctx, "revoked", sess)
	d.appendAudit(ctx, domain.AuditEvent{
		TenantID: &sess.TenantID, ActorID: &sess.PrincipalID, ActorKind: "user",
		Action: "session.logout", ResourceType: "session", ResourceID: sess.ID.String(),
		Outcome: domain.AuditOutcomeSuccess, SessionID: &sess.ID, IP: in.AuthCtx.IP,
	})
	return nil
}

// ListSessions returns active sessions for a principal.
func (d *Deps) ListSessions(ctx context.Context, principalID uuid.UUID) ([]domain.Session, error) {
	if principalID == uuid.Nil {
		return nil, fmt.Errorf("%w: principal_id required", domain.ErrInvalidArgument)
	}
	all, err := d.Sessions.ListByPrincipal(ctx, principalID)
	if err != nil {
		return nil, err
	}
	now := d.now()
	out := make([]domain.Session, 0, len(all))
	for _, s := range all {
		if s.IsUsable(now) {
			out = append(out, s)
		}
	}
	return out, nil
}

// RevokeSessionInput is admin/self session revocation.
type RevokeSessionInput struct {
	SessionID   uuid.UUID
	ActorID     uuid.UUID
	AllowAdmin  bool
	Reason      string
	AuthCtx     ports.AuthContext
}

// RevokeSession revokes a specific session (owner or admin).
func (d *Deps) RevokeSession(ctx context.Context, in RevokeSessionInput) error {
	sess, err := d.Sessions.GetByID(ctx, in.SessionID)
	if err != nil {
		return err
	}
	if !in.AllowAdmin && sess.PrincipalID != in.ActorID {
		return domain.ErrForbidden
	}
	now := d.now()
	reason := in.Reason
	if reason == "" {
		reason = "revoked"
	}
	if err := d.Sessions.Revoke(ctx, sess.ID, now, reason); err != nil {
		return err
	}
	d.publishSession(ctx, "revoked", sess)
	d.appendAudit(ctx, domain.AuditEvent{
		TenantID: &sess.TenantID, ActorID: &in.ActorID, ActorKind: "user",
		Action: "session.revoke", ResourceType: "session", ResourceID: sess.ID.String(),
		Outcome: domain.AuditOutcomeSuccess, SessionID: &sess.ID, IP: in.AuthCtx.IP,
	})
	return nil
}
