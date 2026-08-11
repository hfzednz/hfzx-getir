package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
	"github.com/nexora/identity-service/internal/risk"
	"github.com/nexora/identity-service/internal/security/password"
	secrefresh "github.com/nexora/identity-service/internal/security/refresh"
)

// RegisterInput creates a new user with email/password.
type RegisterInput struct {
	TenantID    uuid.UUID
	Email       string
	Password    string
	DisplayName string
	AuthCtx     ports.AuthContext
}

// Register creates a principal, email identifier, and password credential.
func (d *Deps) Register(ctx context.Context, in RegisterInput) (AuthResult, error) {
	email := domain.NormalizeIdentifier(domain.IdentifierTypeEmail, in.Email)
	if in.TenantID == uuid.Nil || email == "" || in.Password == "" {
		return AuthResult{}, fmt.Errorf("%w: tenant_id, email, password required", domain.ErrInvalidArgument)
	}
	if _, err := d.Principals.FindIdentifier(ctx, in.TenantID, domain.IdentifierTypeEmail, email); err == nil {
		return AuthResult{}, domain.ErrAlreadyExists
	} else if err != domain.ErrNotFound {
		return AuthResult{}, err
	}

	pol := d.policy(ctx, in.TenantID)
	if err := password.ValidateComplexity(in.Password, complexityFromPolicy(pol)); err != nil {
		return AuthResult{}, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
	}
	hash, err := d.hasher().Hash(in.Password)
	if err != nil {
		return AuthResult{}, err
	}

	now := d.now()
	p := domain.Principal{
		ID:          d.newID(),
		TenantID:    in.TenantID,
		Kind:        domain.PrincipalKindUser,
		Status:      domain.PrincipalStatusActive,
		DisplayName: in.DisplayName,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := d.Principals.Create(ctx, p); err != nil {
		return AuthResult{}, err
	}
	ident := domain.Identifier{
		ID:          d.newID(),
		PrincipalID: p.ID,
		TenantID:    in.TenantID,
		Type:        domain.IdentifierTypeEmail,
		Value:       email,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := d.Principals.CreateIdentifier(ctx, ident); err != nil {
		return AuthResult{}, err
	}
	cred := domain.Credential{
		ID:                d.newID(),
		PrincipalID:       p.ID,
		PasswordHash:      hash,
		Algorithm:         domain.CredentialAlgorithmArgon2id,
		PasswordChangedAt: now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := d.Principals.UpsertCredential(ctx, cred); err != nil {
		return AuthResult{}, err
	}
	d.publishPrincipal(ctx, "created", p)
	d.appendAudit(ctx, domain.AuditEvent{
		TenantID:     &in.TenantID,
		ActorID:      &p.ID,
		ActorKind:    string(p.Kind),
		Action:       "auth.register",
		ResourceType: "principal",
		ResourceID:   p.ID.String(),
		Outcome:      domain.AuditOutcomeSuccess,
		IP:           in.AuthCtx.IP,
		UserAgent:    in.AuthCtx.UserAgent,
		RequestID:    in.AuthCtx.RequestID,
	})

	score := d.scoreSignals(in.AuthCtx.Signals)
	return d.createSessionWithTokens(ctx, p, []string{"pwd"}, "aal1", nil, in.AuthCtx, score)
}

// LoginInput is email/password login.
type LoginInput struct {
	TenantID uuid.UUID
	Email    string
	Password string
	AuthCtx  ports.AuthContext
}

// Login authenticates with email/password, applies risk scoring and lockout.
func (d *Deps) Login(ctx context.Context, in LoginInput) (AuthResult, error) {
	email := domain.NormalizeIdentifier(domain.IdentifierTypeEmail, in.Email)
	if in.TenantID == uuid.Nil || email == "" || in.Password == "" {
		return AuthResult{}, fmt.Errorf("%w: tenant_id, email, password required", domain.ErrInvalidArgument)
	}
	pol := d.policy(ctx, in.TenantID)

	ident, err := d.Principals.FindIdentifier(ctx, in.TenantID, domain.IdentifierTypeEmail, email)
	if err != nil {
		d.recordLoginAttempt(ctx, domain.LoginAttempt{
			TenantID:          &in.TenantID,
			Identifier:        email,
			Result:            domain.LoginAttemptInvalidCredentials,
			IP:                in.AuthCtx.IP,
			UserAgent:         in.AuthCtx.UserAgent,
			DeviceFingerprint: in.AuthCtx.DeviceFingerprint,
			FailureReason:     "unknown_identifier",
		})
		return AuthResult{}, domain.ErrUnauthorized
	}
	p, err := d.Principals.GetByID(ctx, ident.PrincipalID)
	if err != nil {
		return AuthResult{}, err
	}
	if err := d.ensurePrincipalAuthn(p); err != nil {
		result := domain.LoginAttemptLocked
		if err == domain.ErrPrincipalSuspended {
			result = domain.LoginAttemptSuspended
		}
		d.recordLoginAttempt(ctx, domain.LoginAttempt{
			TenantID:    &in.TenantID,
			PrincipalID: &p.ID,
			Identifier:  email,
			Result:      result,
			IP:          in.AuthCtx.IP,
			UserAgent:   in.AuthCtx.UserAgent,
		})
		return AuthResult{}, err
	}

	lockoutWindow := time.Duration(pol.LockoutSeconds) * time.Second
	if lockoutWindow <= 0 {
		lockoutWindow = 15 * time.Minute
	}
	// Lockout check via recent failures.
	if pol.MaxFailedAttempts > 0 {
		since := d.now().Add(-lockoutWindow)
		n, _ := d.Risk.CountRecentFailures(ctx, in.TenantID, email, since)
		if n >= pol.MaxFailedAttempts {
			d.recordLoginAttempt(ctx, domain.LoginAttempt{
				TenantID:      &in.TenantID,
				PrincipalID:   &p.ID,
				Identifier:    email,
				Result:        domain.LoginAttemptLocked,
				IP:            in.AuthCtx.IP,
				FailureReason: "lockout",
			})
			// Hint for callers: principal is temporarily locked due to failures.
			return AuthResult{}, fmt.Errorf("%w: too many failed attempts", domain.ErrPrincipalLocked)
		}
	}

	cred, err := d.Principals.GetCredential(ctx, p.ID)
	if err != nil {
		return AuthResult{}, domain.ErrUnauthorized
	}
	ok, err := d.hasher().Verify(in.Password, cred.PasswordHash)
	if err != nil || !ok {
		d.recordLoginAttempt(ctx, domain.LoginAttempt{
			TenantID:          &in.TenantID,
			PrincipalID:       &p.ID,
			Identifier:        email,
			Result:            domain.LoginAttemptInvalidCredentials,
			IP:                in.AuthCtx.IP,
			UserAgent:         in.AuthCtx.UserAgent,
			DeviceFingerprint: in.AuthCtx.DeviceFingerprint,
			FailureReason:     "bad_password",
		})
		// Re-check if this failure tips into lockout (hint for UX).
		// Count includes the attempt just recorded.
		if pol.MaxFailedAttempts > 0 {
			since := d.now().Add(-lockoutWindow)
			n, _ := d.Risk.CountRecentFailures(ctx, in.TenantID, email, since)
			if n >= pol.MaxFailedAttempts {
				return AuthResult{}, fmt.Errorf("%w: account locked after failed attempts", domain.ErrPrincipalLocked)
			}
		}
		return AuthResult{}, domain.ErrUnauthorized
	}

	signals := append([]string{}, in.AuthCtx.Signals...)
	var deviceID *uuid.UUID
	if in.AuthCtx.DeviceFingerprint != "" {
		if dev, derr := d.Devices.FindByFingerprint(ctx, p.ID, in.AuthCtx.DeviceFingerprint); derr == nil && !dev.IsRevoked() {
			deviceID = &dev.ID
		} else {
			signals = append(signals, string(risk.SignalNewDevice))
		}
	}
	score := d.scoreSignals(signals)
	if score >= pol.BlockAboveRisk {
		d.recordLoginAttempt(ctx, domain.LoginAttempt{
			TenantID: &in.TenantID, PrincipalID: &p.ID, Identifier: email,
			Result: domain.LoginAttemptBlockedRisk, IP: in.AuthCtx.IP,
		})
		d.recordRisk(ctx, domain.RiskEvent{
			PrincipalID: &p.ID, TenantID: &in.TenantID,
			EventType: "login_blocked", Severity: domain.RiskSeverityHigh,
			ScoreDelta: score, IP: in.AuthCtx.IP,
		})
		return AuthResult{}, domain.ErrRiskBlocked
	}

	if risk.AdaptiveMFARequired(int(score), risk.MFAPolicy{
		Threshold:     int(pol.MFARequiredAboveRisk),
		AlwaysRequire: pol.MFARequired,
	}) {
		factors, _ := d.Principals.ListMFAFactors(ctx, p.ID)
		hasActive := false
		for _, f := range factors {
			if f.IsActive() {
				hasActive = true
				break
			}
		}
		if hasActive {
			chID := d.newID()
			_ = d.OAuth.SaveMFAChallenge(ctx, ports.MFAChallenge{
				ID:          chID,
				PrincipalID: p.ID,
				FactorType:  domain.MFAFactorTOTP,
				ExpiresAt:   d.now().Add(5 * time.Minute),
				CreatedAt:   d.now(),
			})
			d.recordLoginAttempt(ctx, domain.LoginAttempt{
				TenantID: &in.TenantID, PrincipalID: &p.ID, Identifier: email,
				Result: domain.LoginAttemptMFARequired, IP: in.AuthCtx.IP,
			})
			return AuthResult{
				Principal:      p,
				RiskScore:      score,
				MFARequired:    true,
				MFAChallengeID: &chID,
			}, domain.ErrMFARequired
		}
	}

	res, err := d.createSessionWithTokens(ctx, p, []string{"pwd"}, "aal1", deviceID, in.AuthCtx, score)
	if err != nil {
		return AuthResult{}, err
	}
	d.recordLoginAttempt(ctx, domain.LoginAttempt{
		TenantID: &in.TenantID, PrincipalID: &p.ID, Identifier: email,
		Result: domain.LoginAttemptSuccess, IP: in.AuthCtx.IP,
		UserAgent: in.AuthCtx.UserAgent, DeviceFingerprint: in.AuthCtx.DeviceFingerprint,
	})
	return res, nil
}

// ForgotPasswordInput starts a password-reset flow.
type ForgotPasswordInput struct {
	TenantID uuid.UUID
	Email    string
	AuthCtx  ports.AuthContext
}

// ForgotPassword issues a one-time reset token. Always succeeds for unknown emails (anti-enumeration).
func (d *Deps) ForgotPassword(ctx context.Context, in ForgotPasswordInput) (token string, err error) {
	email := domain.NormalizeIdentifier(domain.IdentifierTypeEmail, in.Email)
	if in.TenantID == uuid.Nil || email == "" {
		return "", fmt.Errorf("%w: tenant_id and email required", domain.ErrInvalidArgument)
	}
	raw, err := secrefresh.Generate("")
	if err != nil {
		return "", err
	}
	ident, err := d.Principals.FindIdentifier(ctx, in.TenantID, domain.IdentifierTypeEmail, email)
	if err == domain.ErrNotFound {
		return raw.Raw, nil // do not leak
	}
	if err != nil {
		return "", err
	}
	now := d.now()
	ch := ports.PasswordResetChallenge{
		ID:          d.newID(),
		TenantID:    in.TenantID,
		PrincipalID: ident.PrincipalID,
		TokenHash:   raw.Hash,
		ExpiresAt:   now.Add(defaultIdleTTL),
		CreatedAt:   now,
	}
	if err := d.OAuth.SavePasswordReset(ctx, ch); err != nil {
		return "", err
	}
	d.appendAudit(ctx, domain.AuditEvent{
		TenantID: &in.TenantID, ActorID: &ident.PrincipalID, ActorKind: "user",
		Action: "auth.password.forgot", ResourceType: "principal", ResourceID: ident.PrincipalID.String(),
		Outcome: domain.AuditOutcomeSuccess, IP: in.AuthCtx.IP, RequestID: in.AuthCtx.RequestID,
	})
	return raw.Raw, nil
}

// ResetPasswordInput consumes a reset token and sets a new password.
type ResetPasswordInput struct {
	Token       string
	NewPassword string
	AuthCtx     ports.AuthContext
}

// ResetPassword sets a new password using a valid reset token.
func (d *Deps) ResetPassword(ctx context.Context, in ResetPasswordInput) error {
	if in.Token == "" || in.NewPassword == "" {
		return fmt.Errorf("%w: token and new_password required", domain.ErrInvalidArgument)
	}
	hash, err := secrefresh.Hash(in.Token)
	if err != nil {
		return domain.ErrUnauthorized
	}
	ch, err := d.OAuth.GetPasswordResetByHash(ctx, hash)
	if err != nil {
		return domain.ErrUnauthorized
	}
	now := d.now()
	if ch.ConsumedAt != nil || now.After(ch.ExpiresAt) {
		return domain.ErrUnauthorized
	}
	p, err := d.Principals.GetByID(ctx, ch.PrincipalID)
	if err != nil {
		return err
	}
	pol := d.policy(ctx, p.TenantID)
	if err := password.ValidateComplexity(in.NewPassword, complexityFromPolicy(pol)); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
	}
	hist, _ := d.Principals.ListPasswordHistory(ctx, p.ID, pol.PasswordHistoryCount)
	prev := make([]string, 0, len(hist)+1)
	if cred, cerr := d.Principals.GetCredential(ctx, p.ID); cerr == nil {
		prev = append(prev, cred.PasswordHash)
	}
	for _, h := range hist {
		prev = append(prev, h.PasswordHash)
	}
	if err := password.CheckHistory(in.NewPassword, prev, d.hasher()); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
	}
	newHash, err := d.hasher().Hash(in.NewPassword)
	if err != nil {
		return err
	}
	if old, cerr := d.Principals.GetCredential(ctx, p.ID); cerr == nil {
		_ = d.Principals.AddPasswordHistory(ctx, domain.PasswordHistoryEntry{
			ID: d.newID(), PrincipalID: p.ID, PasswordHash: old.PasswordHash,
			Algorithm: old.Algorithm, CreatedAt: now,
		})
	}
	cred := domain.Credential{
		ID: d.newID(), PrincipalID: p.ID, PasswordHash: newHash,
		Algorithm: domain.CredentialAlgorithmArgon2id, PasswordChangedAt: now,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := d.Principals.UpsertCredential(ctx, cred); err != nil {
		return err
	}
	_ = d.OAuth.ConsumePasswordReset(ctx, ch.ID, now)
	d.appendAudit(ctx, domain.AuditEvent{
		TenantID: &p.TenantID, ActorID: &p.ID, ActorKind: string(p.Kind),
		Action: "auth.password.reset", ResourceType: "principal", ResourceID: p.ID.String(),
		Outcome: domain.AuditOutcomeSuccess, IP: in.AuthCtx.IP,
	})
	return nil
}

// ChangePasswordInput changes password for an authenticated principal.
type ChangePasswordInput struct {
	PrincipalID uuid.UUID
	OldPassword string
	NewPassword string
	AuthCtx     ports.AuthContext
}

// ChangePassword verifies the old password and sets a new one.
func (d *Deps) ChangePassword(ctx context.Context, in ChangePasswordInput) error {
	if in.PrincipalID == uuid.Nil || in.OldPassword == "" || in.NewPassword == "" {
		return fmt.Errorf("%w: principal_id, old_password, new_password required", domain.ErrInvalidArgument)
	}
	p, err := d.Principals.GetByID(ctx, in.PrincipalID)
	if err != nil {
		return err
	}
	cred, err := d.Principals.GetCredential(ctx, p.ID)
	if err != nil {
		return err
	}
	ok, err := d.hasher().Verify(in.OldPassword, cred.PasswordHash)
	if err != nil || !ok {
		return domain.ErrUnauthorized
	}
	pol := d.policy(ctx, p.TenantID)
	if err := password.ValidateComplexity(in.NewPassword, complexityFromPolicy(pol)); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
	}
	hist, _ := d.Principals.ListPasswordHistory(ctx, p.ID, pol.PasswordHistoryCount)
	prev := []string{cred.PasswordHash}
	for _, h := range hist {
		prev = append(prev, h.PasswordHash)
	}
	if err := password.CheckHistory(in.NewPassword, prev, d.hasher()); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
	}
	newHash, err := d.hasher().Hash(in.NewPassword)
	if err != nil {
		return err
	}
	now := d.now()
	_ = d.Principals.AddPasswordHistory(ctx, domain.PasswordHistoryEntry{
		ID: d.newID(), PrincipalID: p.ID, PasswordHash: cred.PasswordHash,
		Algorithm: cred.Algorithm, CreatedAt: now,
	})
	cred.PasswordHash = newHash
	cred.PasswordChangedAt = now
	cred.UpdatedAt = now
	if err := d.Principals.UpsertCredential(ctx, cred); err != nil {
		return err
	}
	d.appendAudit(ctx, domain.AuditEvent{
		TenantID: &p.TenantID, ActorID: &p.ID, ActorKind: string(p.Kind),
		Action: "auth.password.change", ResourceType: "principal", ResourceID: p.ID.String(),
		Outcome: domain.AuditOutcomeSuccess, IP: in.AuthCtx.IP,
	})
	return nil
}

// hashToken is a small helper for magic-link style tokens when refresh package is overkill.
func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
