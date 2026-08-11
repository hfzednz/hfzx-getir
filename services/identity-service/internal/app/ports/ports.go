// Package ports defines application-layer dependency interfaces (hexagonal ports).
package ports

import (
	"context"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/domain"
	"github.com/nexora/identity-service/internal/domain/policy"
	"github.com/nexora/identity-service/internal/security/webauthn"
)

// Clock abstracts time for deterministic tests.
type Clock interface {
	Now() time.Time
}

// IDGen abstracts UUID generation.
type IDGen interface {
	New() uuid.UUID
}

// OTPSender delivers one-time passwords (SMS/email adapters).
type OTPSender interface {
	SendOTP(ctx context.Context, tenantID uuid.UUID, phone, code string) error
}

// EventPublisher publishes domain events (Kafka adapters).
type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload any) error
}

// TokenPair is an issued access + refresh credential set.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	SessionID    uuid.UUID
	RefreshID    uuid.UUID
	FamilyID     uuid.UUID
}

// IssueParams carries claims needed to mint a session token pair.
type IssueParams struct {
	Principal domain.Principal
	Session   domain.Session
	Roles     []string
	AMR       []string
	ACR       string
	DeviceID  *uuid.UUID
}

// TokenIssuer wraps JWT access + opaque refresh issuance and rotation.
type TokenIssuer interface {
	Issue(ctx context.Context, params IssueParams) (TokenPair, error)
	RotateRefresh(ctx context.Context, rawRefresh string) (TokenPair, error)
}

// ---------------------------------------------------------------------------
// Repositories
// ---------------------------------------------------------------------------

// PrincipalRepository persists principals, identifiers, credentials, and factors.
type PrincipalRepository interface {
	Create(ctx context.Context, p domain.Principal) error
	Update(ctx context.Context, p domain.Principal) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.Principal, error)
	Search(ctx context.Context, tenantID uuid.UUID, query string, limit int) ([]domain.Principal, error)

	CreateIdentifier(ctx context.Context, id domain.Identifier) error
	FindIdentifier(ctx context.Context, tenantID uuid.UUID, typ domain.IdentifierType, value string) (domain.Identifier, error)
	ListIdentifiers(ctx context.Context, principalID uuid.UUID) ([]domain.Identifier, error)

	UpsertCredential(ctx context.Context, c domain.Credential) error
	GetCredential(ctx context.Context, principalID uuid.UUID) (domain.Credential, error)
	AddPasswordHistory(ctx context.Context, e domain.PasswordHistoryEntry) error
	ListPasswordHistory(ctx context.Context, principalID uuid.UUID, limit int) ([]domain.PasswordHistoryEntry, error)

	CreateMFAFactor(ctx context.Context, f domain.MFAFactor) error
	UpdateMFAFactor(ctx context.Context, f domain.MFAFactor) error
	ListMFAFactors(ctx context.Context, principalID uuid.UUID) ([]domain.MFAFactor, error)
	GetMFAFactor(ctx context.Context, id uuid.UUID) (domain.MFAFactor, error)

	ReplaceBackupCodes(ctx context.Context, principalID uuid.UUID, codes []domain.BackupCode) error
	ListBackupCodes(ctx context.Context, principalID uuid.UUID) ([]domain.BackupCode, error)
	UpdateBackupCode(ctx context.Context, c domain.BackupCode) error

	CreateWebAuthnCredential(ctx context.Context, c domain.WebAuthnCredential) error
	UpdateWebAuthnCredential(ctx context.Context, c domain.WebAuthnCredential) error
	ListWebAuthnCredentials(ctx context.Context, principalID uuid.UUID) ([]domain.WebAuthnCredential, error)

	CreateConsent(ctx context.Context, c domain.Consent) error
	ListConsents(ctx context.Context, principalID uuid.UUID) ([]domain.Consent, error)
}

// SessionRepository persists sessions and refresh-token families.
type SessionRepository interface {
	Create(ctx context.Context, s domain.Session) error
	Update(ctx context.Context, s domain.Session) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.Session, error)
	ListByPrincipal(ctx context.Context, principalID uuid.UUID) ([]domain.Session, error)
	Revoke(ctx context.Context, id uuid.UUID, at time.Time, reason string) error
	RevokeFamily(ctx context.Context, familyID uuid.UUID, at time.Time, reason string) error

	CreateRefresh(ctx context.Context, t domain.RefreshToken) error
	GetRefreshByHash(ctx context.Context, hash string) (domain.RefreshToken, error)
	UpdateRefresh(ctx context.Context, t domain.RefreshToken) error
	ListRefreshByFamily(ctx context.Context, familyID uuid.UUID) ([]domain.RefreshToken, error)
}

// DeviceRepository persists trusted/known devices.
type DeviceRepository interface {
	Create(ctx context.Context, d domain.Device) error
	Update(ctx context.Context, d domain.Device) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.Device, error)
	FindByFingerprint(ctx context.Context, principalID uuid.UUID, fingerprint string) (domain.Device, error)
	ListByPrincipal(ctx context.Context, principalID uuid.UUID) ([]domain.Device, error)
}

// RoleRepository persists RBAC definitions and bindings.
type RoleRepository interface {
	GetRole(ctx context.Context, id uuid.UUID) (domain.Role, error)
	GetRoleByName(ctx context.Context, tenantID *uuid.UUID, name string) (domain.Role, error)
	ListRolePermissions(ctx context.Context, roleID uuid.UUID) ([]domain.Permission, error)
	RoleGraph(ctx context.Context, tenantID uuid.UUID) (policy.RoleGraph, error)

	AssignRole(ctx context.Context, pr domain.PrincipalRole) error
	ListPrincipalRoles(ctx context.Context, principalID uuid.UUID) ([]domain.PrincipalRole, error)
	CreateTemporaryGrant(ctx context.Context, g domain.TemporaryGrant) error
	ListTemporaryGrants(ctx context.Context, principalID uuid.UUID) ([]domain.TemporaryGrant, error)
	GetPermission(ctx context.Context, id uuid.UUID) (domain.Permission, error)
	FindPermission(ctx context.Context, resource, action string) (domain.Permission, error)
}

// AuditRepository appends local audit events.
type AuditRepository interface {
	Append(ctx context.Context, e domain.AuditEvent) error
	ListByPrincipal(ctx context.Context, principalID uuid.UUID, limit int) ([]domain.AuditEvent, error)
}

// OTPChallenge is a hashed OTP pending verification.
type OTPChallenge struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Phone     string
	CodeHash  string
	ExpiresAt time.Time
	Attempts  int
	CreatedAt time.Time
}

// MagicLinkChallenge is a one-time magic-link token.
type MagicLinkChallenge struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	TokenHash   string
	ExpiresAt   time.Time
	ConsumedAt  *time.Time
	CreatedAt   time.Time
}

// PasswordResetChallenge is a one-time password-reset token.
type PasswordResetChallenge struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	TokenHash   string
	ExpiresAt   time.Time
	ConsumedAt  *time.Time
	CreatedAt   time.Time
}

// MFAChallenge is a pending step-up MFA verification.
type MFAChallenge struct {
	ID          uuid.UUID
	PrincipalID uuid.UUID
	SessionHint uuid.UUID
	FactorType  domain.MFAFactorType
	ExpiresAt   time.Time
	CreatedAt   time.Time
}

// OAuthClient is a registered OAuth2/OIDC client (incl. service accounts).
type OAuthClient struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	ClientID     string
	ClientSecret string // hashed
	PrincipalID  uuid.UUID
	GrantTypes   []string
	Scopes       []string
	Enabled      bool
	CreatedAt    time.Time
}

// OAuthRepository persists OAuth clients and ephemeral auth challenges.
type OAuthRepository interface {
	GetClientByClientID(ctx context.Context, clientID string) (OAuthClient, error)

	SaveOTPChallenge(ctx context.Context, c OTPChallenge) error
	GetOTPChallenge(ctx context.Context, id uuid.UUID) (OTPChallenge, error)
	DeleteOTPChallenge(ctx context.Context, id uuid.UUID) error
	UpdateOTPChallenge(ctx context.Context, c OTPChallenge) error

	SaveMagicLink(ctx context.Context, c MagicLinkChallenge) error
	GetMagicLinkByHash(ctx context.Context, hash string) (MagicLinkChallenge, error)
	ConsumeMagicLink(ctx context.Context, id uuid.UUID, at time.Time) error

	SavePasswordReset(ctx context.Context, c PasswordResetChallenge) error
	GetPasswordResetByHash(ctx context.Context, hash string) (PasswordResetChallenge, error)
	ConsumePasswordReset(ctx context.Context, id uuid.UUID, at time.Time) error

	SaveMFAChallenge(ctx context.Context, c MFAChallenge) error
	GetMFAChallenge(ctx context.Context, id uuid.UUID) (MFAChallenge, error)
	DeleteMFAChallenge(ctx context.Context, id uuid.UUID) error

	SaveWebAuthnCeremony(ctx context.Context, session *webauthn.CeremonySession) error
	GetWebAuthnCeremony(ctx context.Context, id string) (*webauthn.CeremonySession, error)
	DeleteWebAuthnCeremony(ctx context.Context, id string) error
}

// RiskRepository persists risk events, login attempts, and security policies.
type RiskRepository interface {
	AppendRiskEvent(ctx context.Context, e domain.RiskEvent) error
	AppendLoginAttempt(ctx context.Context, a domain.LoginAttempt) error
	CountRecentFailures(ctx context.Context, tenantID uuid.UUID, identifier string, since time.Time) (int, error)
	GetSecurityPolicy(ctx context.Context, tenantID uuid.UUID) (domain.SecurityPolicy, error)
}

// AuthContext carries request metadata for risk/audit.
type AuthContext struct {
	IP                *netip.Addr
	UserAgent         string
	DeviceFingerprint string
	RequestID         string
	CountryCode       string
	Signals           []string // risk signal names
}

// Kafka topic constants used by EventPublisher.
const (
	TopicPrincipalLifecycle = "identity.principal.lifecycle"
	TopicSessionEvents      = "identity.session.events"
	TopicSecurityRisk       = "identity.security.risk"
	TopicAuditEvents        = "identity.audit.events"
)

// SocialProvider identifies an external IdP.
type SocialProvider string

const (
	SocialGoogle    SocialProvider = "google"
	SocialApple     SocialProvider = "apple"
	SocialFacebook  SocialProvider = "facebook"
	SocialMicrosoft SocialProvider = "microsoft"
	SocialGitHub    SocialProvider = "github"
)

// SocialProfile is the normalized profile returned by an IdP token exchange.
type SocialProfile struct {
	Provider      SocialProvider
	Subject       string
	Email         string
	EmailVerified bool
	DisplayName   string
}

// SocialIdP exchanges authorization codes / id_tokens with an external IdP.
type SocialIdP interface {
	Provider() SocialProvider
	Exchange(ctx context.Context, code string, redirectURI string) (SocialProfile, error)
}
