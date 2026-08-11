package app

import (
	"context"
	"fmt"

	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
	"github.com/nexora/identity-service/internal/security/password"
)

// ClientCredentialsInput is the OAuth2 client_credentials grant.
type ClientCredentialsInput struct {
	ClientID     string
	ClientSecret string
	AuthCtx      ports.AuthContext
}

// ClientCredentials issues tokens for a service-account OAuth client.
func (d *Deps) ClientCredentials(ctx context.Context, in ClientCredentialsInput) (AuthResult, error) {
	if in.ClientID == "" || in.ClientSecret == "" {
		return AuthResult{}, fmt.Errorf("%w: client_id and client_secret required", domain.ErrInvalidArgument)
	}
	client, err := d.OAuth.GetClientByClientID(ctx, in.ClientID)
	if err != nil {
		return AuthResult{}, domain.ErrUnauthorized
	}
	if !client.Enabled {
		return AuthResult{}, domain.ErrUnauthorized
	}
	ok := false
	if hasher := d.hasher(); hasher != nil {
		// Secrets may be stored as argon2id hashes (same as passwords) or plaintext in tests.
		if verified, verr := hasher.Verify(in.ClientSecret, client.ClientSecret); verr == nil && verified {
			ok = true
		}
	}
	if !ok && client.ClientSecret == in.ClientSecret {
		ok = true
	}
	if !ok && client.ClientSecret == hashToken(in.ClientSecret) {
		ok = true
	}
	if !ok {
		return AuthResult{}, domain.ErrUnauthorized
	}

	p, err := d.Principals.GetByID(ctx, client.PrincipalID)
	if err != nil {
		return AuthResult{}, err
	}
	if p.Kind != domain.PrincipalKindServiceAccount && p.Kind != domain.PrincipalKindRobot {
		return AuthResult{}, domain.ErrForbidden
	}
	if err := d.ensurePrincipalAuthn(p); err != nil {
		return AuthResult{}, err
	}

	score := d.scoreSignals(in.AuthCtx.Signals)
	res, err := d.createSessionWithTokens(ctx, p, []string{"client_credentials"}, "aal1", nil, in.AuthCtx, score)
	if err != nil {
		return AuthResult{}, err
	}
	d.appendAudit(ctx, domain.AuditEvent{
		TenantID: &p.TenantID, ActorID: &p.ID, ActorKind: string(p.Kind),
		Action: "auth.client_credentials", ResourceType: "oauth_client", ResourceID: client.ID.String(),
		Outcome: domain.AuditOutcomeSuccess, SessionID: &res.Session.ID, IP: in.AuthCtx.IP,
	})
	return res, nil
}

// HashClientSecret hashes a client secret for storage.
func HashClientSecret(secret string, hasher *password.Hasher) (string, error) {
	if hasher == nil {
		hasher = password.NewDefaultHasher()
	}
	return hasher.Hash(secret)
}
