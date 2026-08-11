package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
	"github.com/nexora/identity-service/internal/security/totp"
)

// EnrollTOTPResult is the provisioning payload for authenticator apps.
type EnrollTOTPResult struct {
	FactorID  uuid.UUID
	Secret    string
	OTPAuthURI string
	BackupCodes []string
}

// EnrollTOTP creates an unverified TOTP factor and one-time backup codes.
func (d *Deps) EnrollTOTP(ctx context.Context, principalID uuid.UUID, accountName string) (EnrollTOTPResult, error) {
	p, err := d.Principals.GetByID(ctx, principalID)
	if err != nil {
		return EnrollTOTPResult{}, err
	}
	secret, err := totp.GenerateSecret()
	if err != nil {
		return EnrollTOTPResult{}, err
	}
	if accountName == "" {
		accountName = p.ID.String()
	}
	uri, err := totp.OTPAuthURI(secret, accountName, totp.Options{Issuer: "NEXORA"})
	if err != nil {
		return EnrollTOTPResult{}, err
	}
	now := d.now()
	factor := domain.MFAFactor{
		ID: d.newID(), PrincipalID: p.ID, Type: domain.MFAFactorTOTP,
		Label: "Authenticator", SecretEnc: []byte(secret), Verified: false,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := d.Principals.CreateMFAFactor(ctx, factor); err != nil {
		return EnrollTOTPResult{}, err
	}

	plainCodes := make([]string, 0, 10)
	stored := make([]domain.BackupCode, 0, 10)
	for i := 0; i < 10; i++ {
		code, err := randomBackupCode()
		if err != nil {
			return EnrollTOTPResult{}, err
		}
		plainCodes = append(plainCodes, code)
		stored = append(stored, domain.BackupCode{
			ID: d.newID(), PrincipalID: p.ID, CodeHash: hashToken(code), CreatedAt: now,
		})
	}
	if err := d.Principals.ReplaceBackupCodes(ctx, p.ID, stored); err != nil {
		return EnrollTOTPResult{}, err
	}
	d.appendAudit(ctx, domain.AuditEvent{
		TenantID: &p.TenantID, ActorID: &p.ID, ActorKind: string(p.Kind),
		Action: "mfa.totp.enroll", ResourceType: "mfa_factor", ResourceID: factor.ID.String(),
		Outcome: domain.AuditOutcomeSuccess,
	})
	return EnrollTOTPResult{
		FactorID: factor.ID, Secret: secret, OTPAuthURI: uri, BackupCodes: plainCodes,
	}, nil
}

func randomBackupCode() (string, error) {
	var b [5]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

// VerifyMFAChallengeInput verifies a pending MFA challenge with a TOTP code.
type VerifyMFAChallengeInput struct {
	ChallengeID uuid.UUID
	Code        string
	AuthCtx     ports.AuthContext
}

// VerifyMFAChallenge validates TOTP against an enrolled factor and issues a session.
func (d *Deps) VerifyMFAChallenge(ctx context.Context, in VerifyMFAChallengeInput) (AuthResult, error) {
	ch, err := d.OAuth.GetMFAChallenge(ctx, in.ChallengeID)
	if err != nil {
		return AuthResult{}, domain.ErrUnauthorized
	}
	now := d.now()
	if now.After(ch.ExpiresAt) {
		_ = d.OAuth.DeleteMFAChallenge(ctx, ch.ID)
		return AuthResult{}, domain.ErrUnauthorized
	}
	factors, err := d.Principals.ListMFAFactors(ctx, ch.PrincipalID)
	if err != nil {
		return AuthResult{}, err
	}
	ok := false
	for _, f := range factors {
		if f.Type != domain.MFAFactorTOTP || f.DisabledAt != nil {
			continue
		}
		secret := string(f.SecretEnc)
		valid, verr := totp.ValidateCode(secret, in.Code, totp.Options{Now: d.now})
		if verr == nil && valid {
			ok = true
			if !f.Verified {
				f.Verified = true
				f.VerifiedAt = &now
				f.UpdatedAt = now
				_ = d.Principals.UpdateMFAFactor(ctx, f)
			}
			break
		}
	}
	if !ok {
		d.recordLoginAttempt(ctx, domain.LoginAttempt{
			PrincipalID: &ch.PrincipalID, Result: domain.LoginAttemptMFAFailed,
			IP: in.AuthCtx.IP, FailureReason: "bad_totp",
		})
		return AuthResult{}, domain.ErrMFAFailed
	}
	_ = d.OAuth.DeleteMFAChallenge(ctx, ch.ID)
	p, err := d.Principals.GetByID(ctx, ch.PrincipalID)
	if err != nil {
		return AuthResult{}, err
	}
	score := d.scoreSignals(in.AuthCtx.Signals)
	return d.createSessionWithTokens(ctx, p, []string{"pwd", "otp"}, "aal2", nil, in.AuthCtx, score)
}

// VerifyBackupCodeInput consumes a one-time backup code for MFA recovery.
type VerifyBackupCodeInput struct {
	PrincipalID uuid.UUID
	Code        string
	AuthCtx     ports.AuthContext
}

// VerifyBackupCode validates and consumes a backup code, then issues a session.
func (d *Deps) VerifyBackupCode(ctx context.Context, in VerifyBackupCodeInput) (AuthResult, error) {
	if in.PrincipalID == uuid.Nil || in.Code == "" {
		return AuthResult{}, fmt.Errorf("%w: principal_id and code required", domain.ErrInvalidArgument)
	}
	codes, err := d.Principals.ListBackupCodes(ctx, in.PrincipalID)
	if err != nil {
		return AuthResult{}, err
	}
	want := hashToken(in.Code)
	now := d.now()
	found := false
	for _, c := range codes {
		if c.IsConsumed() {
			continue
		}
		if c.CodeHash == want {
			c.UsedAt = &now
			if err := d.Principals.UpdateBackupCode(ctx, c); err != nil {
				return AuthResult{}, err
			}
			found = true
			break
		}
	}
	if !found {
		return AuthResult{}, domain.ErrMFAFailed
	}
	p, err := d.Principals.GetByID(ctx, in.PrincipalID)
	if err != nil {
		return AuthResult{}, err
	}
	if err := d.ensurePrincipalAuthn(p); err != nil {
		return AuthResult{}, err
	}
	score := d.scoreSignals(in.AuthCtx.Signals)
	d.appendAudit(ctx, domain.AuditEvent{
		TenantID: &p.TenantID, ActorID: &p.ID, ActorKind: string(p.Kind),
		Action: "mfa.backup.verify", ResourceType: "principal", ResourceID: p.ID.String(),
		Outcome: domain.AuditOutcomeSuccess, IP: in.AuthCtx.IP,
	})
	return d.createSessionWithTokens(ctx, p, []string{"backup"}, "aal2", nil, in.AuthCtx, score)
}

// ConfirmTOTPEnrollment verifies the first TOTP code and marks the factor verified.
func (d *Deps) ConfirmTOTPEnrollment(ctx context.Context, factorID uuid.UUID, code string) error {
	f, err := d.Principals.GetMFAFactor(ctx, factorID)
	if err != nil {
		return err
	}
	ok, err := totp.ValidateCode(string(f.SecretEnc), code, totp.Options{Now: d.now})
	if err != nil || !ok {
		return domain.ErrMFAFailed
	}
	now := d.now()
	f.Verified = true
	f.VerifiedAt = &now
	f.UpdatedAt = now
	return d.Principals.UpdateMFAFactor(ctx, f)
}
