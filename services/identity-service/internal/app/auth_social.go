package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
)

// StubSocialProviders returns no-op IdP stubs for Google/Apple/Facebook/Microsoft/GitHub.
func StubSocialProviders() map[ports.SocialProvider]ports.SocialIdP {
	return map[ports.SocialProvider]ports.SocialIdP{
		ports.SocialGoogle:    &stubIdP{provider: ports.SocialGoogle},
		ports.SocialApple:     &stubIdP{provider: ports.SocialApple},
		ports.SocialFacebook:  &stubIdP{provider: ports.SocialFacebook},
		ports.SocialMicrosoft: &stubIdP{provider: ports.SocialMicrosoft},
		ports.SocialGitHub:    &stubIdP{provider: ports.SocialGitHub},
	}
}

type stubIdP struct {
	provider ports.SocialProvider
}

func (s *stubIdP) Provider() ports.SocialProvider { return s.provider }

func (s *stubIdP) Exchange(_ context.Context, _, _ string) (ports.SocialProfile, error) {
	return ports.SocialProfile{}, fmt.Errorf("%w: social provider %s not configured", domain.ErrUnauthorized, s.provider)
}

// SocialCallbackInput is an OIDC/OAuth authorization-code callback.
type SocialCallbackInput struct {
	TenantID    uuid.UUID
	Provider    ports.SocialProvider
	Code        string
	RedirectURI string
	AuthCtx     ports.AuthContext
}

// HandleSocialCallback exchanges an IdP code, links/creates a principal, and issues a session.
func (d *Deps) HandleSocialCallback(ctx context.Context, in SocialCallbackInput) (AuthResult, error) {
	if in.TenantID == uuid.Nil || in.Code == "" {
		return AuthResult{}, fmt.Errorf("%w: tenant_id and code required", domain.ErrInvalidArgument)
	}
	var idp ports.SocialIdP
	if d.Social != nil {
		idp = d.Social[in.Provider]
	}
	if idp == nil {
		return AuthResult{}, fmt.Errorf("%w: unknown provider %s", domain.ErrInvalidArgument, in.Provider)
	}
	profile, err := idp.Exchange(ctx, in.Code, in.RedirectURI)
	if err != nil {
		return AuthResult{}, err
	}
	extValue := string(profile.Provider) + ":" + profile.Subject
	now := d.now()

	ident, err := d.Principals.FindIdentifier(ctx, in.TenantID, domain.IdentifierTypeExternal, extValue)
	var p domain.Principal
	if err == domain.ErrNotFound {
		p = domain.Principal{
			ID: d.newID(), TenantID: in.TenantID, Kind: domain.PrincipalKindUser,
			Status: domain.PrincipalStatusActive, DisplayName: profile.DisplayName,
			CreatedAt: now, UpdatedAt: now,
		}
		if err := d.Principals.Create(ctx, p); err != nil {
			return AuthResult{}, err
		}
		ext := domain.Identifier{
			ID: d.newID(), PrincipalID: p.ID, TenantID: in.TenantID,
			Type: domain.IdentifierTypeExternal, Value: extValue,
			CreatedAt: now, UpdatedAt: now,
		}
		if profile.EmailVerified {
			t := now
			ext.VerifiedAt = &t
		}
		if err := d.Principals.CreateIdentifier(ctx, ext); err != nil {
			return AuthResult{}, err
		}
		if profile.Email != "" {
			email := domain.NormalizeIdentifier(domain.IdentifierTypeEmail, profile.Email)
			em := domain.Identifier{
				ID: d.newID(), PrincipalID: p.ID, TenantID: in.TenantID,
				Type: domain.IdentifierTypeEmail, Value: email, CreatedAt: now, UpdatedAt: now,
			}
			if profile.EmailVerified {
				em.VerifiedAt = &now
			}
			_ = d.Principals.CreateIdentifier(ctx, em)
		}
		d.publishPrincipal(ctx, "created", p)
	} else if err != nil {
		return AuthResult{}, err
	} else {
		p, err = d.Principals.GetByID(ctx, ident.PrincipalID)
		if err != nil {
			return AuthResult{}, err
		}
	}
	if err := d.ensurePrincipalAuthn(p); err != nil {
		return AuthResult{}, err
	}

	score := d.scoreSignals(in.AuthCtx.Signals)
	pol := d.policy(ctx, p.TenantID)
	if score >= pol.BlockAboveRisk {
		return AuthResult{}, domain.ErrRiskBlocked
	}
	amr := []string{"federated", string(in.Provider)}
	res, err := d.createSessionWithTokens(ctx, p, amr, "aal1", nil, in.AuthCtx, score)
	if err != nil {
		return AuthResult{}, err
	}
	d.recordLoginAttempt(ctx, domain.LoginAttempt{
		TenantID: &p.TenantID, PrincipalID: &p.ID, Identifier: extValue,
		Result: domain.LoginAttemptSuccess, IP: in.AuthCtx.IP, UserAgent: in.AuthCtx.UserAgent,
	})
	return res, nil
}

// Ensure stub satisfies port.
var _ ports.SocialIdP = (*stubIdP)(nil)
