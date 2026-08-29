package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
	"github.com/nexora/identity-service/internal/risk"
	secotp "github.com/nexora/identity-service/internal/security/otp"
)

// StartOTPInput begins a phone OTP authentication challenge.
type StartOTPInput struct {
	TenantID uuid.UUID
	Phone    string
	AuthCtx  ports.AuthContext
}

// StartOTP generates and sends a phone OTP. Returns the challenge id.
func (d *Deps) StartOTP(ctx context.Context, in StartOTPInput) (uuid.UUID, error) {
	phone := domain.NormalizeIdentifier(domain.IdentifierTypePhone, in.Phone)
	if phone == "" || in.TenantID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("%w: tenant_id and phone required", domain.ErrInvalidArgument)
	}

	plain, ch, err := secotp.NewChallenge(secotp.DefaultLength, d.otpPepper(), secotp.DefaultTTL)
	if err != nil {
		return uuid.Nil, err
	}
	now := d.now()
	id := d.newID()
	stored := ports.OTPChallenge{
		ID:        id,
		TenantID:  in.TenantID,
		Phone:     phone,
		CodeHash:  ch.Hash,
		ExpiresAt: now.Add(secotp.DefaultTTL),
		CreatedAt: now,
	}
	if err := d.OAuth.SaveOTPChallenge(ctx, stored); err != nil {
		return uuid.Nil, err
	}
	if d.OTP != nil {
		if err := d.OTP.SendOTP(ctx, in.TenantID, phone, plain); err != nil {
			return uuid.Nil, err
		}
	}
	d.appendAudit(ctx, domain.AuditEvent{
		TenantID:     &in.TenantID,
		ActorKind:    "anonymous",
		Action:       "auth.otp.start",
		ResourceType: "otp_challenge",
		ResourceID:   id.String(),
		Outcome:      domain.AuditOutcomeSuccess,
		IP:           in.AuthCtx.IP,
		UserAgent:    in.AuthCtx.UserAgent,
		RequestID:    in.AuthCtx.RequestID,
		Details:      map[string]any{"phone": phone},
	})
	return id, nil
}

// VerifyOTPInput completes phone OTP auth.
type VerifyOTPInput struct {
	ChallengeID uuid.UUID
	Code        string
	AuthCtx     ports.AuthContext
}

// VerifyOTP validates the OTP, creates/finds the principal by phone, and issues a session.
func (d *Deps) VerifyOTP(ctx context.Context, in VerifyOTPInput) (AuthResult, error) {
	ch, err := d.OAuth.GetOTPChallenge(ctx, in.ChallengeID)
	if err != nil {
		return AuthResult{}, domain.ErrUnauthorized
	}
	now := d.now()
	if now.After(ch.ExpiresAt) {
		_ = d.OAuth.DeleteOTPChallenge(ctx, ch.ID)
		return AuthResult{}, domain.ErrUnauthorized
	}
	ok, err := secotp.VerifyOTP(in.Code, d.otpPepper(), ch.CodeHash)
	if err != nil || !ok {
		ch.Attempts++
		_ = d.OAuth.UpdateOTPChallenge(ctx, ch)
		d.recordLoginAttempt(ctx, domain.LoginAttempt{
			TenantID:          &ch.TenantID,
			Identifier:        ch.Phone,
			Result:            domain.LoginAttemptInvalidCredentials,
			IP:                in.AuthCtx.IP,
			UserAgent:         in.AuthCtx.UserAgent,
			DeviceFingerprint: in.AuthCtx.DeviceFingerprint,
			FailureReason:     "invalid_otp",
		})
		return AuthResult{}, domain.ErrUnauthorized
	}
	_ = d.OAuth.DeleteOTPChallenge(ctx, ch.ID)

	ident, err := d.Principals.FindIdentifier(ctx, ch.TenantID, domain.IdentifierTypePhone, ch.Phone)
	var p domain.Principal
	if err != nil {
		if err != domain.ErrNotFound {
			return AuthResult{}, err
		}
		p = domain.Principal{
			ID:        d.newID(),
			TenantID:  ch.TenantID,
			Kind:      domain.PrincipalKindUser,
			Status:    domain.PrincipalStatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := d.Principals.Create(ctx, p); err != nil {
			return AuthResult{}, err
		}
		verified := now
		idrow := domain.Identifier{
			ID:          d.newID(),
			PrincipalID: p.ID,
			TenantID:    ch.TenantID,
			Type:        domain.IdentifierTypePhone,
			Value:       ch.Phone,
			VerifiedAt:  &verified,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := d.Principals.CreateIdentifier(ctx, idrow); err != nil {
			return AuthResult{}, err
		}
		d.publishPrincipal(ctx, "created", p)
	} else {
		p, err = d.Principals.GetByID(ctx, ident.PrincipalID)
		if err != nil {
			return AuthResult{}, err
		}
	}
	if err := d.ensurePrincipalAuthn(p); err != nil {
		return AuthResult{}, err
	}
	if err := d.ensureOTPRoles(ctx, p.ID, ch.Phone); err != nil {
		return AuthResult{}, err
	}

	signals := append([]string{}, in.AuthCtx.Signals...)
	score := d.scoreSignals(signals)
	pol := d.policy(ctx, p.TenantID)
	if score >= pol.BlockAboveRisk {
		d.recordRisk(ctx, domain.RiskEvent{
			PrincipalID: &p.ID,
			TenantID:    &p.TenantID,
			EventType:   "login_blocked",
			Severity:    domain.RiskSeverityHigh,
			ScoreDelta:  score,
			IP:          in.AuthCtx.IP,
		})
		return AuthResult{}, domain.ErrRiskBlocked
	}

	var deviceID *uuid.UUID
	if in.AuthCtx.DeviceFingerprint != "" {
		if dev, derr := d.Devices.FindByFingerprint(ctx, p.ID, in.AuthCtx.DeviceFingerprint); derr == nil && !dev.IsRevoked() {
			deviceID = &dev.ID
			_ = dev.See(now)
			_ = d.Devices.Update(ctx, dev)
		} else {
			signals = append(signals, string(risk.SignalNewDevice))
			score = d.scoreSignals(signals)
		}
	}

	res, err := d.createSessionWithTokens(ctx, p, []string{"otp"}, "aal1", deviceID, in.AuthCtx, score)
	if err != nil {
		return AuthResult{}, err
	}
	d.recordLoginAttempt(ctx, domain.LoginAttempt{
		TenantID:          &p.TenantID,
		PrincipalID:       &p.ID,
		Identifier:        ch.Phone,
		Result:            domain.LoginAttemptSuccess,
		IP:                in.AuthCtx.IP,
		UserAgent:         in.AuthCtx.UserAgent,
		DeviceFingerprint: in.AuthCtx.DeviceFingerprint,
	})
	return res, nil
}
