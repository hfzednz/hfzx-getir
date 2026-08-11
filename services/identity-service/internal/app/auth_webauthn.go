package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
	"github.com/nexora/identity-service/internal/security/webauthn"
)

// BeginWebAuthnRegistration starts a passkey registration ceremony.
func (d *Deps) BeginWebAuthnRegistration(ctx context.Context, principalID uuid.UUID) (*webauthn.PublicKeyCredentialCreationOptions, error) {
	if d.WebAuthn == nil {
		return nil, webauthn.ErrNotImplemented
	}
	p, err := d.Principals.GetByID(ctx, principalID)
	if err != nil {
		return nil, err
	}
	existing, _ := d.Principals.ListWebAuthnCredentials(ctx, p.ID)
	exclude := make([]webauthn.Credential, 0, len(existing))
	for _, c := range existing {
		if !c.IsActive() {
			continue
		}
		exclude = append(exclude, webauthn.Credential{
			ID: c.CredentialID, PublicKey: c.PublicKey, SignCount: uint32(c.SignCount),
			Transport: c.Transports, AAGUID: c.AAGUID, BackupEligible: c.BackupEligible, BackupState: c.BackupState,
		})
	}
	name := p.DisplayName
	if name == "" {
		name = p.ID.String()
	}
	opts, session, err := d.WebAuthn.BeginRegistration(ctx, webauthn.User{
		ID: p.ID[:], Name: name, DisplayName: name,
	}, exclude)
	if err != nil {
		return nil, err
	}
	if session != nil {
		_ = d.OAuth.SaveWebAuthnCeremony(ctx, session)
		if opts != nil {
			opts.SessionID = session.ID
		}
	}
	return opts, nil
}

// FinishWebAuthnRegistration completes passkey registration.
func (d *Deps) FinishWebAuthnRegistration(ctx context.Context, principalID uuid.UUID, resp *webauthn.AuthenticatorAttestationResponse) error {
	if d.WebAuthn == nil {
		return webauthn.ErrNotImplemented
	}
	if resp == nil || resp.SessionID == "" {
		return fmt.Errorf("%w: session required", domain.ErrInvalidArgument)
	}
	session, err := d.OAuth.GetWebAuthnCeremony(ctx, resp.SessionID)
	if err != nil {
		return webauthn.ErrInvalidSession
	}
	cred, err := d.WebAuthn.FinishRegistration(ctx, session, resp)
	if err != nil {
		return err
	}
	_ = d.OAuth.DeleteWebAuthnCeremony(ctx, resp.SessionID)
	now := d.now()
	wc := domain.WebAuthnCredential{
		ID: d.newID(), PrincipalID: principalID, CredentialID: cred.ID, PublicKey: cred.PublicKey,
		AAGUID: cred.AAGUID, SignCount: uint64(cred.SignCount), Transports: cred.Transport,
		BackupEligible: cred.BackupEligible, BackupState: cred.BackupState,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := d.Principals.CreateWebAuthnCredential(ctx, wc); err != nil {
		return err
	}
	d.appendAudit(ctx, domain.AuditEvent{
		ActorID: &principalID, ActorKind: "user",
		Action: "webauthn.register", ResourceType: "webauthn_credential", ResourceID: wc.ID.String(),
		Outcome: domain.AuditOutcomeSuccess,
	})
	return nil
}

// BeginWebAuthnLogin starts a passkey authentication ceremony.
func (d *Deps) BeginWebAuthnLogin(ctx context.Context, tenantID uuid.UUID, identifier string) (*webauthn.PublicKeyCredentialRequestOptions, error) {
	if d.WebAuthn == nil {
		return nil, webauthn.ErrNotImplemented
	}
	email := domain.NormalizeIdentifier(domain.IdentifierTypeEmail, identifier)
	ident, err := d.Principals.FindIdentifier(ctx, tenantID, domain.IdentifierTypeEmail, email)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}
	p, err := d.Principals.GetByID(ctx, ident.PrincipalID)
	if err != nil {
		return nil, err
	}
	stored, _ := d.Principals.ListWebAuthnCredentials(ctx, p.ID)
	allowed := make([]webauthn.Credential, 0, len(stored))
	for _, c := range stored {
		if !c.IsActive() {
			continue
		}
		allowed = append(allowed, webauthn.Credential{
			ID: c.CredentialID, PublicKey: c.PublicKey, SignCount: uint32(c.SignCount),
			Transport: c.Transports, AAGUID: c.AAGUID,
		})
	}
	name := p.DisplayName
	if name == "" {
		name = p.ID.String()
	}
	opts, session, err := d.WebAuthn.BeginLogin(ctx, webauthn.User{
		ID: p.ID[:], Name: name, DisplayName: name,
	}, allowed)
	if err != nil {
		return nil, err
	}
	if session != nil {
		_ = d.OAuth.SaveWebAuthnCeremony(ctx, session)
		if opts != nil {
			opts.SessionID = session.ID
		}
	}
	return opts, nil
}

// FinishWebAuthnLogin completes passkey login and issues a session.
func (d *Deps) FinishWebAuthnLogin(ctx context.Context, tenantID uuid.UUID, resp *webauthn.AuthenticatorAssertionResponse, authCtx ports.AuthContext) (AuthResult, error) {
	if d.WebAuthn == nil {
		return AuthResult{}, webauthn.ErrNotImplemented
	}
	if resp == nil || resp.SessionID == "" {
		return AuthResult{}, fmt.Errorf("%w: session required", domain.ErrInvalidArgument)
	}
	session, err := d.OAuth.GetWebAuthnCeremony(ctx, resp.SessionID)
	if err != nil {
		return AuthResult{}, webauthn.ErrInvalidSession
	}
	cred, err := d.WebAuthn.FinishLogin(ctx, session, resp)
	if err != nil {
		return AuthResult{}, err
	}
	_ = d.OAuth.DeleteWebAuthnCeremony(ctx, resp.SessionID)

	// Resolve principal from user handle or credential id match.
	var principalID uuid.UUID
	if len(session.UserID) == 16 {
		copy(principalID[:], session.UserID)
	}
	if principalID == uuid.Nil {
		return AuthResult{}, domain.ErrUnauthorized
	}
	p, err := d.Principals.GetByID(ctx, principalID)
	if err != nil {
		return AuthResult{}, err
	}
	if p.TenantID != tenantID {
		return AuthResult{}, domain.ErrForbidden
	}
	if err := d.ensurePrincipalAuthn(p); err != nil {
		return AuthResult{}, err
	}

	// Advance sign count on matching credential.
	stored, _ := d.Principals.ListWebAuthnCredentials(ctx, p.ID)
	now := d.now()
	for _, c := range stored {
		if string(c.CredentialID) == string(cred.ID) {
			if err := c.AdvanceSignCount(uint64(cred.SignCount)); err != nil {
				return AuthResult{}, err
			}
			c.UpdatedAt = now
			c.LastUsedAt = &now
			_ = d.Principals.UpdateWebAuthnCredential(ctx, c)
			break
		}
	}

	score := d.scoreSignals(authCtx.Signals)
	return d.createSessionWithTokens(ctx, p, []string{"webauthn"}, "aal2", nil, authCtx, score)
}
