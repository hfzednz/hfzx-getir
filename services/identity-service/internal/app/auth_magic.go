package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
	secrefresh "github.com/nexora/identity-service/internal/security/refresh"
)

// StartMagicLinkInput begins email magic-link auth.
type StartMagicLinkInput struct {
	TenantID uuid.UUID
	Email    string
	AuthCtx  ports.AuthContext
}

// StartMagicLink issues a one-time magic link token for an existing (or new) email principal.
// Returns the raw token for the delivery adapter (email).
func (d *Deps) StartMagicLink(ctx context.Context, in StartMagicLinkInput) (token string, err error) {
	email := domain.NormalizeIdentifier(domain.IdentifierTypeEmail, in.Email)
	if in.TenantID == uuid.Nil || email == "" {
		return "", fmt.Errorf("%w: tenant_id and email required", domain.ErrInvalidArgument)
	}
	now := d.now()
	ident, err := d.Principals.FindIdentifier(ctx, in.TenantID, domain.IdentifierTypeEmail, email)
	var principalID uuid.UUID
	if err == domain.ErrNotFound {
		p := domain.Principal{
			ID: d.newID(), TenantID: in.TenantID, Kind: domain.PrincipalKindUser,
			Status: domain.PrincipalStatusActive, CreatedAt: now, UpdatedAt: now,
		}
		if err := d.Principals.Create(ctx, p); err != nil {
			return "", err
		}
		ident = domain.Identifier{
			ID: d.newID(), PrincipalID: p.ID, TenantID: in.TenantID,
			Type: domain.IdentifierTypeEmail, Value: email, CreatedAt: now, UpdatedAt: now,
		}
		if err := d.Principals.CreateIdentifier(ctx, ident); err != nil {
			return "", err
		}
		principalID = p.ID
		d.publishPrincipal(ctx, "created", p)
	} else if err != nil {
		return "", err
	} else {
		principalID = ident.PrincipalID
	}

	raw, err := secrefresh.Generate("")
	if err != nil {
		return "", err
	}
	ch := ports.MagicLinkChallenge{
		ID: d.newID(), TenantID: in.TenantID, PrincipalID: principalID,
		TokenHash: raw.Hash, ExpiresAt: now.Add(15 * time.Minute), CreatedAt: now,
	}
	if err := d.OAuth.SaveMagicLink(ctx, ch); err != nil {
		return "", err
	}
	d.appendAudit(ctx, domain.AuditEvent{
		TenantID: &in.TenantID, ActorID: &principalID, ActorKind: "user",
		Action: "auth.magic.start", ResourceType: "magic_link", ResourceID: ch.ID.String(),
		Outcome: domain.AuditOutcomeSuccess, IP: in.AuthCtx.IP, RequestID: in.AuthCtx.RequestID,
	})
	return raw.Raw, nil
}

// ConsumeMagicLinkInput redeems a magic-link token.
type ConsumeMagicLinkInput struct {
	Token   string
	AuthCtx ports.AuthContext
}

// ConsumeMagicLink verifies and consumes a magic link, issuing a session.
func (d *Deps) ConsumeMagicLink(ctx context.Context, in ConsumeMagicLinkInput) (AuthResult, error) {
	if in.Token == "" {
		return AuthResult{}, fmt.Errorf("%w: token required", domain.ErrInvalidArgument)
	}
	hash, err := secrefresh.Hash(in.Token)
	if err != nil {
		return AuthResult{}, domain.ErrUnauthorized
	}
	ch, err := d.OAuth.GetMagicLinkByHash(ctx, hash)
	if err != nil {
		return AuthResult{}, domain.ErrUnauthorized
	}
	now := d.now()
	if ch.ConsumedAt != nil || now.After(ch.ExpiresAt) {
		return AuthResult{}, domain.ErrUnauthorized
	}
	p, err := d.Principals.GetByID(ctx, ch.PrincipalID)
	if err != nil {
		return AuthResult{}, err
	}
	if err := d.ensurePrincipalAuthn(p); err != nil {
		return AuthResult{}, err
	}
	_ = d.OAuth.ConsumeMagicLink(ctx, ch.ID, now)

	score := d.scoreSignals(in.AuthCtx.Signals)
	pol := d.policy(ctx, p.TenantID)
	if score >= pol.BlockAboveRisk {
		return AuthResult{}, domain.ErrRiskBlocked
	}
	res, err := d.createSessionWithTokens(ctx, p, []string{"magic"}, "aal1", nil, in.AuthCtx, score)
	if err != nil {
		return AuthResult{}, err
	}
	d.recordLoginAttempt(ctx, domain.LoginAttempt{
		TenantID: &p.TenantID, PrincipalID: &p.ID,
		Result: domain.LoginAttemptSuccess, IP: in.AuthCtx.IP, UserAgent: in.AuthCtx.UserAgent,
	})
	return res, nil
}
