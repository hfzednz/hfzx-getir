package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
	"github.com/nexora/identity-service/internal/risk"
	"github.com/nexora/identity-service/internal/security/jwt"
	"github.com/nexora/identity-service/internal/security/password"
	secrefresh "github.com/nexora/identity-service/internal/security/refresh"
	"github.com/nexora/identity-service/internal/security/webauthn"
)

// Default session/token TTLs when security policy is missing fields.
const (
	defaultIdleTTL     = 30 * time.Minute
	defaultAbsoluteTTL = 12 * time.Hour
	defaultRefreshTTL  = 30 * 24 * time.Hour
	defaultAccessTTL   = 15 * time.Minute
	defaultOTPPepper   = "nexora-otp-pepper"
	defaultIssuer      = "nexora-identity"
	defaultAudience    = "nexora"
)

// Deps aggregates application ports for use cases.
type Deps struct {
	Principals ports.PrincipalRepository
	Sessions   ports.SessionRepository
	Devices    ports.DeviceRepository
	Roles      ports.RoleRepository
	Audit      ports.AuditRepository
	OAuth      ports.OAuthRepository
	Risk       ports.RiskRepository
	OTP        ports.OTPSender
	Events     ports.EventPublisher
	Clock      ports.Clock
	IDs        ports.IDGen
	Tokens     ports.TokenIssuer
	Passwords  *password.Hasher
	RiskScore  *risk.Scorer
	WebAuthn   webauthn.Service
	JWTKeys    *jwt.KeyManager
	Issuer     string
	Audience   string
	AccessTTL  time.Duration
	OTPPepper  string
	Social     map[ports.SocialProvider]ports.SocialIdP
}

func (d *Deps) now() time.Time {
	if d.Clock != nil {
		return d.Clock.Now().UTC()
	}
	return time.Now().UTC()
}

func (d *Deps) newID() uuid.UUID {
	if d.IDs != nil {
		return d.IDs.New()
	}
	return uuid.New()
}

func (d *Deps) hasher() *password.Hasher {
	if d.Passwords != nil {
		return d.Passwords
	}
	return password.NewDefaultHasher()
}

func (d *Deps) scorer() *risk.Scorer {
	if d.RiskScore != nil {
		return d.RiskScore
	}
	return risk.NewScorer()
}

func (d *Deps) issuer() string {
	if d.Issuer != "" {
		return d.Issuer
	}
	return defaultIssuer
}

func (d *Deps) audience() string {
	if d.Audience != "" {
		return d.Audience
	}
	return defaultAudience
}

func (d *Deps) accessTTL() time.Duration {
	if d.AccessTTL > 0 {
		return d.AccessTTL
	}
	return defaultAccessTTL
}

func (d *Deps) otpPepper() string {
	if d.OTPPepper != "" {
		return d.OTPPepper
	}
	return defaultOTPPepper
}

func (d *Deps) policy(ctx context.Context, tenantID uuid.UUID) domain.SecurityPolicy {
	if d.Risk == nil {
		return defaultPolicy()
	}
	p, err := d.Risk.GetSecurityPolicy(ctx, tenantID)
	if err != nil {
		return defaultPolicy()
	}
	return p
}

func defaultPolicy() domain.SecurityPolicy {
	return domain.SecurityPolicy{
		ID:                     uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Name:                   "default",
		Enabled:                true,
		PasswordMinLength:      12,
		PasswordRequireUpper:   true,
		PasswordRequireLower:   true,
		PasswordRequireDigit:   true,
		PasswordRequireSymbol:  true,
		PasswordHistoryCount:   5,
		MFARequired:            false,
		MFARequiredAboveRisk:   40,
		SessionIdleSeconds:     int(defaultIdleTTL.Seconds()),
		SessionAbsoluteSeconds: int(defaultAbsoluteTTL.Seconds()),
		RefreshTokenSeconds:    int(defaultRefreshTTL.Seconds()),
		MaxFailedAttempts:      5,
		LockoutSeconds:         900,
		BlockAboveRisk:         90,
	}
}

// AuthResult is returned by successful authentication flows.
type AuthResult struct {
	Principal domain.Principal
	Session   domain.Session
	Tokens    ports.TokenPair
	RiskScore float64
	MFARequired bool
	MFAChallengeID *uuid.UUID
}

func (d *Deps) ensurePrincipalAuthn(p domain.Principal) error {
	switch p.Status {
	case domain.PrincipalStatusActive:
		return nil
	case domain.PrincipalStatusLocked:
		return domain.ErrPrincipalLocked
	case domain.PrincipalStatusSuspended:
		return domain.ErrPrincipalSuspended
	case domain.PrincipalStatusDeleted:
		return domain.ErrPrincipalDeleted
	default:
		return domain.ErrUnauthorized
	}
}

func (d *Deps) scoreSignals(names []string) float64 {
	sigs := make([]risk.Signal, 0, len(names))
	for _, n := range names {
		sigs = append(sigs, risk.Signal(n))
	}
	return float64(d.scorer().Score(sigs))
}

func (d *Deps) appendAudit(ctx context.Context, e domain.AuditEvent) {
	if e.ID == uuid.Nil {
		e.ID = d.newID()
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = d.now()
	}
	if d.Audit != nil {
		_ = d.Audit.Append(ctx, e)
	}
	if d.Events != nil {
		_ = d.Events.Publish(ctx, ports.TopicAuditEvents, e.ID.String(), e)
	}
}

func (d *Deps) publishSession(ctx context.Context, action string, s domain.Session) {
	if d.Events == nil {
		return
	}
	_ = d.Events.Publish(ctx, ports.TopicSessionEvents, s.ID.String(), map[string]any{
		"action":       action,
		"session_id":   s.ID,
		"principal_id": s.PrincipalID,
		"tenant_id":    s.TenantID,
	})
}

func (d *Deps) publishPrincipal(ctx context.Context, action string, p domain.Principal) {
	if d.Events == nil {
		return
	}
	_ = d.Events.Publish(ctx, ports.TopicPrincipalLifecycle, p.ID.String(), map[string]any{
		"action":       action,
		"principal_id": p.ID,
		"tenant_id":    p.TenantID,
		"kind":         p.Kind,
		"status":       p.Status,
	})
}

func (d *Deps) recordLoginAttempt(ctx context.Context, a domain.LoginAttempt) {
	if a.ID == uuid.Nil {
		a.ID = d.newID()
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = d.now()
	}
	if d.Risk != nil {
		_ = d.Risk.AppendLoginAttempt(ctx, a)
	}
}

func (d *Deps) recordRisk(ctx context.Context, e domain.RiskEvent) {
	if e.ID == uuid.Nil {
		e.ID = d.newID()
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = d.now()
	}
	if d.Risk != nil {
		_ = d.Risk.AppendRiskEvent(ctx, e)
	}
	if d.Events != nil {
		_ = d.Events.Publish(ctx, ports.TopicSecurityRisk, e.ID.String(), e)
	}
}

func (d *Deps) createSessionWithTokens(
	ctx context.Context,
	p domain.Principal,
	amr []string,
	acr string,
	deviceID *uuid.UUID,
	authCtx ports.AuthContext,
	riskScore float64,
) (AuthResult, error) {
	pol := d.policy(ctx, p.TenantID)
	now := d.now()
	idleTTL := pol.SessionIdleTTL()
	if idleTTL <= 0 {
		idleTTL = defaultIdleTTL
	}
	absTTL := pol.SessionAbsoluteTTL()
	if absTTL <= 0 {
		absTTL = defaultAbsoluteTTL
	}

	sess := domain.Session{
		ID:                d.newID(),
		PrincipalID:       p.ID,
		DeviceID:          deviceID,
		TenantID:          p.TenantID,
		AMR:               amr,
		ACR:               acr,
		IP:                authCtx.IP,
		UserAgent:         authCtx.UserAgent,
		RiskScore:         domain.ClampRiskScore(riskScore),
		IdleExpiresAt:     now.Add(idleTTL),
		AbsoluteExpiresAt: now.Add(absTTL),
		LastSeenAt:        now,
		CreatedAt:         now,
	}
	if err := sess.Validate(); err != nil {
		return AuthResult{}, err
	}
	if err := d.Sessions.Create(ctx, sess); err != nil {
		return AuthResult{}, err
	}

	roles, err := d.roleNames(ctx, p.ID)
	if err != nil {
		roles = nil
	}

	pair, err := d.Tokens.Issue(ctx, ports.IssueParams{
		Principal: p,
		Session:   sess,
		Roles:     roles,
		AMR:       amr,
		ACR:       acr,
		DeviceID:  deviceID,
	})
	if err != nil {
		return AuthResult{}, err
	}

	d.publishSession(ctx, "created", sess)
	d.appendAudit(ctx, domain.AuditEvent{
		TenantID:     &p.TenantID,
		ActorID:      &p.ID,
		ActorKind:    string(p.Kind),
		Action:       "session.create",
		ResourceType: "session",
		ResourceID:   sess.ID.String(),
		Outcome:      domain.AuditOutcomeSuccess,
		IP:           authCtx.IP,
		UserAgent:    authCtx.UserAgent,
		SessionID:    &sess.ID,
		RequestID:    authCtx.RequestID,
	})

	return AuthResult{
		Principal: p,
		Session:   sess,
		Tokens:    pair,
		RiskScore: riskScore,
	}, nil
}

func (d *Deps) roleNames(ctx context.Context, principalID uuid.UUID) ([]string, error) {
	if d.Roles == nil {
		return nil, nil
	}
	bindings, err := d.Roles.ListPrincipalRoles(ctx, principalID)
	if err != nil {
		return nil, err
	}
	now := d.now()
	names := make([]string, 0, len(bindings))
	for _, b := range bindings {
		if !b.IsActive(now) {
			continue
		}
		role, err := d.Roles.GetRole(ctx, b.RoleID)
		if err != nil {
			continue
		}
		names = append(names, role.Name)
	}
	return names, nil
}

func complexityFromPolicy(p domain.SecurityPolicy) password.ComplexityPolicy {
	return password.ComplexityPolicy{
		MinLength:     p.PasswordMinLength,
		RequireUpper:  p.PasswordRequireUpper,
		RequireLower:  p.PasswordRequireLower,
		RequireDigit:  p.PasswordRequireDigit,
		RequireSymbol: p.PasswordRequireSymbol,
		MaxLength:     128,
	}
}

// SystemClock is a real-time Clock.
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

// UUIDGen generates random UUIDs.
type UUIDGen struct{}

func (UUIDGen) New() uuid.UUID { return uuid.New() }

// ---------------------------------------------------------------------------
// Default TokenIssuer implementation
// ---------------------------------------------------------------------------

// DefaultTokenIssuer issues JWT access tokens and opaque refresh tokens.
type DefaultTokenIssuer struct {
	Deps *Deps
}

// Issue creates a new refresh family and access JWT for a session.
func (t *DefaultTokenIssuer) Issue(ctx context.Context, params ports.IssueParams) (ports.TokenPair, error) {
	d := t.Deps
	now := d.now()
	pol := d.policy(ctx, params.Principal.TenantID)
	refreshTTL := pol.RefreshTokenTTL()
	if refreshTTL <= 0 {
		refreshTTL = defaultRefreshTTL
	}

	raw, err := secrefresh.Generate("")
	if err != nil {
		return ports.TokenPair{}, err
	}
	familyID, err := uuid.Parse(raw.FamilyID)
	if err != nil {
		return ports.TokenPair{}, fmt.Errorf("invalid family id: %w", err)
	}
	rt := domain.RefreshToken{
		ID:          d.newID(),
		SessionID:   params.Session.ID,
		PrincipalID: params.Principal.ID,
		TokenHash:   raw.Hash,
		FamilyID:    familyID,
		ExpiresAt:   now.Add(refreshTTL),
		CreatedAt:   now,
	}
	if err := rt.Validate(); err != nil {
		return ports.TokenPair{}, err
	}
	if err := d.Sessions.CreateRefresh(ctx, rt); err != nil {
		return ports.TokenPair{}, err
	}

	access, exp, err := t.signAccess(params, now)
	if err != nil {
		return ports.TokenPair{}, err
	}
	return ports.TokenPair{
		AccessToken:  access,
		RefreshToken: raw.Raw,
		ExpiresAt:    exp,
		SessionID:    params.Session.ID,
		RefreshID:    rt.ID,
		FamilyID:     familyID,
	}, nil
}

// RotateRefresh rotates a usable refresh token; reuse of an already-rotated
// token revokes the entire family and returns ErrTokenReuse.
func (t *DefaultTokenIssuer) RotateRefresh(ctx context.Context, rawRefresh string) (ports.TokenPair, error) {
	d := t.Deps
	now := d.now()
	hash, err := secrefresh.Hash(rawRefresh)
	if err != nil {
		return ports.TokenPair{}, domain.ErrUnauthorized
	}
	existing, err := d.Sessions.GetRefreshByHash(ctx, hash)
	if err != nil {
		return ports.TokenPair{}, domain.ErrUnauthorized
	}

	// Reuse detection: token already revoked (typically after prior rotation).
	if existing.IsRevoked() {
		_ = d.Sessions.RevokeFamily(ctx, existing.FamilyID, now, "refresh_reuse")
		pid := existing.PrincipalID
		fid := existing.FamilyID
		d.recordRisk(ctx, domain.RiskEvent{
			PrincipalID: &pid,
			SessionID:   &existing.SessionID,
			EventType:   domain.RiskEventTokenReuse,
			Severity:    domain.RiskSeverityCritical,
			ScoreDelta:  50,
			Details:     map[string]any{"family_id": fid.String()},
		})
		d.appendAudit(ctx, domain.AuditEvent{
			ActorID:      &pid,
			ActorKind:    "user",
			Action:       "token.refresh.reuse",
			ResourceType: "refresh_family",
			ResourceID:   fid.String(),
			Outcome:      domain.AuditOutcomeDenied,
			SessionID:    &existing.SessionID,
			Details:      map[string]any{"family_id": fid.String()},
		})
		return ports.TokenPair{}, domain.ErrTokenReuse
	}
	if existing.IsExpired(now) {
		return ports.TokenPair{}, domain.ErrTokenExpired
	}

	sess, err := d.Sessions.GetByID(ctx, existing.SessionID)
	if err != nil {
		return ports.TokenPair{}, err
	}
	if !sess.IsUsable(now) {
		existing.Revoke(now, "session_unusable")
		_ = d.Sessions.UpdateRefresh(ctx, existing)
		if sess.IsRevoked() {
			return ports.TokenPair{}, domain.ErrSessionRevoked
		}
		return ports.TokenPair{}, domain.ErrSessionExpired
	}

	p, err := d.Principals.GetByID(ctx, existing.PrincipalID)
	if err != nil {
		return ports.TokenPair{}, err
	}
	if err := d.ensurePrincipalAuthn(p); err != nil {
		return ports.TokenPair{}, err
	}

	pol := d.policy(ctx, p.TenantID)
	idleTTL := pol.SessionIdleTTL()
	if idleTTL <= 0 {
		idleTTL = defaultIdleTTL
	}
	_ = sess.Touch(now, idleTTL)
	_ = d.Sessions.Update(ctx, sess)

	// Revoke current refresh (rotation).
	existing.Revoke(now, "rotated")
	if err := d.Sessions.UpdateRefresh(ctx, existing); err != nil {
		return ports.TokenPair{}, err
	}

	refreshTTL := pol.RefreshTokenTTL()
	if refreshTTL <= 0 {
		refreshTTL = defaultRefreshTTL
	}
	nextRaw, err := secrefresh.Generate(existing.FamilyID.String())
	if err != nil {
		return ports.TokenPair{}, err
	}
	rotatedFrom := existing.ID
	next := domain.RefreshToken{
		ID:          d.newID(),
		SessionID:   sess.ID,
		PrincipalID: p.ID,
		TokenHash:   nextRaw.Hash,
		FamilyID:    existing.FamilyID,
		RotatedFrom: &rotatedFrom,
		ExpiresAt:   now.Add(refreshTTL),
		CreatedAt:   now,
	}
	if err := d.Sessions.CreateRefresh(ctx, next); err != nil {
		return ports.TokenPair{}, err
	}

	roles, _ := d.roleNames(ctx, p.ID)
	access, exp, err := t.signAccess(ports.IssueParams{
		Principal: p,
		Session:   sess,
		Roles:     roles,
		AMR:       sess.AMR,
		ACR:       sess.ACR,
		DeviceID:  sess.DeviceID,
	}, now)
	if err != nil {
		return ports.TokenPair{}, err
	}

	d.appendAudit(ctx, domain.AuditEvent{
		TenantID:     &p.TenantID,
		ActorID:      &p.ID,
		ActorKind:    string(p.Kind),
		Action:       "token.refresh",
		ResourceType: "session",
		ResourceID:   sess.ID.String(),
		Outcome:      domain.AuditOutcomeSuccess,
		SessionID:    &sess.ID,
	})

	return ports.TokenPair{
		AccessToken:  access,
		RefreshToken: nextRaw.Raw,
		ExpiresAt:    exp,
		SessionID:    sess.ID,
		RefreshID:    next.ID,
		FamilyID:     next.FamilyID,
	}, nil
}

func (t *DefaultTokenIssuer) signAccess(params ports.IssueParams, now time.Time) (string, time.Time, error) {
	d := t.Deps
	exp := now.Add(d.accessTTL())
	deviceID := ""
	if params.DeviceID != nil {
		deviceID = params.DeviceID.String()
	}
	if d.JWTKeys == nil {
		// Tests may inject a nil key manager only when Tokens is fully mocked.
		return "", time.Time{}, jwt.ErrNoKey
	}
	claims := jwt.AccessClaims{
		Subject:  params.Principal.ID.String(),
		Session:  params.Session.ID.String(),
		Tenant:   params.Principal.TenantID.String(),
		Roles:    params.Roles,
		AMR:      params.AMR,
		ACR:      params.ACR,
		DeviceID: deviceID,
		Issuer:   d.issuer(),
		Audience: d.audience(),
		Expires:  exp,
		IssuedAt: now,
		JTI:      d.newID().String(),
	}
	tok, err := d.JWTKeys.IssueAccessToken(claims)
	if err != nil {
		return "", time.Time{}, err
	}
	return tok, exp, nil
}

// Ensure DefaultTokenIssuer implements ports.TokenIssuer.
var _ ports.TokenIssuer = (*DefaultTokenIssuer)(nil)
