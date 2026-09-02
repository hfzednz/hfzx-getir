// Package memory provides in-memory port implementations for unit tests.
package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
	"github.com/nexora/identity-service/internal/domain/policy"
	"github.com/nexora/identity-service/internal/security/webauthn"
)

// Store is a shared in-memory database for all repositories.
type Store struct {
	mu sync.RWMutex

	Principals   map[uuid.UUID]domain.Principal
	Identifiers  map[uuid.UUID]domain.Identifier
	Credentials  map[uuid.UUID]domain.Credential // by principalID
	PassHistory  []domain.PasswordHistoryEntry
	MFAFactors   map[uuid.UUID]domain.MFAFactor
	BackupCodes  map[uuid.UUID][]domain.BackupCode
	WebAuthn     map[uuid.UUID]domain.WebAuthnCredential
	Consents     map[uuid.UUID]domain.Consent

	Sessions map[uuid.UUID]domain.Session
	Refresh  map[uuid.UUID]domain.RefreshToken
	ByHash   map[string]uuid.UUID

	Devices map[uuid.UUID]domain.Device

	Roles       map[uuid.UUID]domain.Role
	RoleParents map[uuid.UUID][]uuid.UUID
	RolePerms   map[uuid.UUID][]domain.Permission
	Permissions map[uuid.UUID]domain.Permission
	PrincRoles  map[uuid.UUID]domain.PrincipalRole
	TempGrants  map[uuid.UUID]domain.TemporaryGrant

	Audit []domain.AuditEvent

	OTP       map[uuid.UUID]ports.OTPChallenge
	Magic     map[string]ports.MagicLinkChallenge
	PassReset map[string]ports.PasswordResetChallenge
	MFAChal   map[uuid.UUID]ports.MFAChallenge
	Ceremonies map[string]*webauthn.CeremonySession
	Clients   map[string]ports.OAuthClient

	RiskEvents    []domain.RiskEvent
	LoginAttempts []domain.LoginAttempt
	Policies      map[uuid.UUID]domain.SecurityPolicy
}

// NewStore returns an in-memory store with the platform role catalog seeded.
func NewStore() *Store {
	s := &Store{
		Principals: make(map[uuid.UUID]domain.Principal),
		Identifiers: make(map[uuid.UUID]domain.Identifier),
		Credentials: make(map[uuid.UUID]domain.Credential),
		MFAFactors: make(map[uuid.UUID]domain.MFAFactor),
		BackupCodes: make(map[uuid.UUID][]domain.BackupCode),
		WebAuthn: make(map[uuid.UUID]domain.WebAuthnCredential),
		Consents: make(map[uuid.UUID]domain.Consent),
		Sessions: make(map[uuid.UUID]domain.Session),
		Refresh: make(map[uuid.UUID]domain.RefreshToken),
		ByHash: make(map[string]uuid.UUID),
		Devices: make(map[uuid.UUID]domain.Device),
		Roles: make(map[uuid.UUID]domain.Role),
		RoleParents: make(map[uuid.UUID][]uuid.UUID),
		RolePerms: make(map[uuid.UUID][]domain.Permission),
		Permissions: make(map[uuid.UUID]domain.Permission),
		PrincRoles: make(map[uuid.UUID]domain.PrincipalRole),
		TempGrants: make(map[uuid.UUID]domain.TemporaryGrant),
		OTP: make(map[uuid.UUID]ports.OTPChallenge),
		Magic: make(map[string]ports.MagicLinkChallenge),
		PassReset: make(map[string]ports.PasswordResetChallenge),
		MFAChal: make(map[uuid.UUID]ports.MFAChallenge),
		Ceremonies: make(map[string]*webauthn.CeremonySession),
		Clients: make(map[string]ports.OAuthClient),
		Policies: make(map[uuid.UUID]domain.SecurityPolicy),
	}
	seedPlatformRoles(s)
	return s
}

func seedPlatformRoles(s *Store) {
	now := time.Now().UTC()
	for _, name := range []string{
		"customer", "courier", "picker", "packer", "dispatcher", "city_ops",
		"support_agent", "finance_analyst", "admin", "super_admin",
		"service_account", "partner", "supplier", "merchant",
	} {
		id := uuid.NewSHA1(uuid.NameSpaceOID, []byte("nexora.platform.role."+name))
		s.Roles[id] = domain.Role{
			ID: id, Name: name, Kind: domain.RoleKindPlatform,
			Description: "platform " + name, CreatedAt: now, UpdatedAt: now,
		}
	}
}

// ---------------------------------------------------------------------------
// PrincipalRepository
// ---------------------------------------------------------------------------

type PrincipalRepo struct{ S *Store }

func (r *PrincipalRepo) Create(_ context.Context, p domain.Principal) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Principals[p.ID]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.Principals[p.ID] = p
	return nil
}

func (r *PrincipalRepo) Update(_ context.Context, p domain.Principal) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Principals[p.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Principals[p.ID] = p
	return nil
}

func (r *PrincipalRepo) GetByID(_ context.Context, id uuid.UUID) (domain.Principal, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	p, ok := r.S.Principals[id]
	if !ok {
		return domain.Principal{}, domain.ErrNotFound
	}
	return p, nil
}

func (r *PrincipalRepo) Search(_ context.Context, tenantID uuid.UUID, query string, limit int) ([]domain.Principal, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	q := strings.ToLower(query)
	out := make([]domain.Principal, 0)
	for _, p := range r.S.Principals {
		if p.TenantID != tenantID {
			continue
		}
		if q == "" || strings.Contains(strings.ToLower(p.DisplayName), q) || strings.Contains(p.ID.String(), q) {
			out = append(out, p)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *PrincipalRepo) CreateIdentifier(_ context.Context, id domain.Identifier) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	norm := domain.NormalizeIdentifier(id.Type, id.Value)
	id.Value = norm
	for _, existing := range r.S.Identifiers {
		if existing.TenantID == id.TenantID && existing.Type == id.Type && existing.Value == norm {
			return domain.ErrAlreadyExists
		}
	}
	r.S.Identifiers[id.ID] = id
	return nil
}

func (r *PrincipalRepo) FindIdentifier(_ context.Context, tenantID uuid.UUID, typ domain.IdentifierType, value string) (domain.Identifier, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	norm := domain.NormalizeIdentifier(typ, value)
	for _, id := range r.S.Identifiers {
		if id.TenantID == tenantID && id.Type == typ && id.Value == norm {
			return id, nil
		}
	}
	return domain.Identifier{}, domain.ErrNotFound
}

func (r *PrincipalRepo) ListIdentifiers(_ context.Context, principalID uuid.UUID) ([]domain.Identifier, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.Identifier, 0)
	for _, id := range r.S.Identifiers {
		if id.PrincipalID == principalID {
			out = append(out, id)
		}
	}
	return out, nil
}

func (r *PrincipalRepo) UpsertCredential(_ context.Context, c domain.Credential) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Credentials[c.PrincipalID] = c
	return nil
}

func (r *PrincipalRepo) GetCredential(_ context.Context, principalID uuid.UUID) (domain.Credential, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	c, ok := r.S.Credentials[principalID]
	if !ok {
		return domain.Credential{}, domain.ErrNotFound
	}
	return c, nil
}

func (r *PrincipalRepo) AddPasswordHistory(_ context.Context, e domain.PasswordHistoryEntry) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.PassHistory = append(r.S.PassHistory, e)
	return nil
}

func (r *PrincipalRepo) ListPasswordHistory(_ context.Context, principalID uuid.UUID, limit int) ([]domain.PasswordHistoryEntry, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.PasswordHistoryEntry, 0)
	for i := len(r.S.PassHistory) - 1; i >= 0; i-- {
		e := r.S.PassHistory[i]
		if e.PrincipalID == principalID {
			out = append(out, e)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *PrincipalRepo) CreateMFAFactor(_ context.Context, f domain.MFAFactor) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.MFAFactors[f.ID] = f
	return nil
}

func (r *PrincipalRepo) UpdateMFAFactor(_ context.Context, f domain.MFAFactor) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.MFAFactors[f.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.MFAFactors[f.ID] = f
	return nil
}

func (r *PrincipalRepo) ListMFAFactors(_ context.Context, principalID uuid.UUID) ([]domain.MFAFactor, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.MFAFactor, 0)
	for _, f := range r.S.MFAFactors {
		if f.PrincipalID == principalID {
			out = append(out, f)
		}
	}
	return out, nil
}

func (r *PrincipalRepo) GetMFAFactor(_ context.Context, id uuid.UUID) (domain.MFAFactor, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	f, ok := r.S.MFAFactors[id]
	if !ok {
		return domain.MFAFactor{}, domain.ErrNotFound
	}
	return f, nil
}

func (r *PrincipalRepo) ReplaceBackupCodes(_ context.Context, principalID uuid.UUID, codes []domain.BackupCode) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.BackupCodes[principalID] = codes
	return nil
}

func (r *PrincipalRepo) ListBackupCodes(_ context.Context, principalID uuid.UUID) ([]domain.BackupCode, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	return append([]domain.BackupCode{}, r.S.BackupCodes[principalID]...), nil
}

func (r *PrincipalRepo) UpdateBackupCode(_ context.Context, c domain.BackupCode) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	codes := r.S.BackupCodes[c.PrincipalID]
	for i := range codes {
		if codes[i].ID == c.ID {
			codes[i] = c
			r.S.BackupCodes[c.PrincipalID] = codes
			return nil
		}
	}
	return domain.ErrNotFound
}

func (r *PrincipalRepo) CreateWebAuthnCredential(_ context.Context, c domain.WebAuthnCredential) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.WebAuthn[c.ID] = c
	return nil
}

func (r *PrincipalRepo) UpdateWebAuthnCredential(_ context.Context, c domain.WebAuthnCredential) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.WebAuthn[c.ID] = c
	return nil
}

func (r *PrincipalRepo) ListWebAuthnCredentials(_ context.Context, principalID uuid.UUID) ([]domain.WebAuthnCredential, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.WebAuthnCredential, 0)
	for _, c := range r.S.WebAuthn {
		if c.PrincipalID == principalID {
			out = append(out, c)
		}
	}
	return out, nil
}

func (r *PrincipalRepo) CreateConsent(_ context.Context, c domain.Consent) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Consents[c.ID] = c
	return nil
}

func (r *PrincipalRepo) ListConsents(_ context.Context, principalID uuid.UUID) ([]domain.Consent, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.Consent, 0)
	for _, c := range r.S.Consents {
		if c.PrincipalID == principalID {
			out = append(out, c)
		}
	}
	return out, nil
}

var _ ports.PrincipalRepository = (*PrincipalRepo)(nil)

// ---------------------------------------------------------------------------
// SessionRepository
// ---------------------------------------------------------------------------

type SessionRepo struct{ S *Store }

func (r *SessionRepo) Create(_ context.Context, s domain.Session) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Sessions[s.ID] = s
	return nil
}

func (r *SessionRepo) Update(_ context.Context, s domain.Session) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Sessions[s.ID] = s
	return nil
}

func (r *SessionRepo) GetByID(_ context.Context, id uuid.UUID) (domain.Session, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	s, ok := r.S.Sessions[id]
	if !ok {
		return domain.Session{}, domain.ErrNotFound
	}
	return s, nil
}

func (r *SessionRepo) ListByPrincipal(_ context.Context, principalID uuid.UUID) ([]domain.Session, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.Session, 0)
	for _, s := range r.S.Sessions {
		if s.PrincipalID == principalID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (r *SessionRepo) Revoke(_ context.Context, id uuid.UUID, at time.Time, reason string) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	s, ok := r.S.Sessions[id]
	if !ok {
		return domain.ErrNotFound
	}
	s.Revoke(at, reason)
	r.S.Sessions[id] = s
	return nil
}

func (r *SessionRepo) RevokeFamily(_ context.Context, familyID uuid.UUID, at time.Time, reason string) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	sessionIDs := map[uuid.UUID]struct{}{}
	for id, t := range r.S.Refresh {
		if t.FamilyID == familyID {
			t.Revoke(at, reason)
			r.S.Refresh[id] = t
			sessionIDs[t.SessionID] = struct{}{}
		}
	}
	for sid := range sessionIDs {
		if s, ok := r.S.Sessions[sid]; ok {
			s.Revoke(at, reason)
			r.S.Sessions[sid] = s
		}
	}
	return nil
}

func (r *SessionRepo) CreateRefresh(_ context.Context, t domain.RefreshToken) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Refresh[t.ID] = t
	r.S.ByHash[t.TokenHash] = t.ID
	return nil
}

func (r *SessionRepo) GetRefreshByHash(_ context.Context, hash string) (domain.RefreshToken, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.ByHash[hash]
	if !ok {
		return domain.RefreshToken{}, domain.ErrNotFound
	}
	t, ok := r.S.Refresh[id]
	if !ok {
		return domain.RefreshToken{}, domain.ErrNotFound
	}
	return t, nil
}

func (r *SessionRepo) UpdateRefresh(_ context.Context, t domain.RefreshToken) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Refresh[t.ID] = t
	r.S.ByHash[t.TokenHash] = t.ID
	return nil
}

func (r *SessionRepo) ListRefreshByFamily(_ context.Context, familyID uuid.UUID) ([]domain.RefreshToken, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.RefreshToken, 0)
	for _, t := range r.S.Refresh {
		if t.FamilyID == familyID {
			out = append(out, t)
		}
	}
	return out, nil
}

var _ ports.SessionRepository = (*SessionRepo)(nil)

// ---------------------------------------------------------------------------
// DeviceRepository
// ---------------------------------------------------------------------------

type DeviceRepo struct{ S *Store }

func (r *DeviceRepo) Create(_ context.Context, d domain.Device) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Devices[d.ID] = d
	return nil
}

func (r *DeviceRepo) Update(_ context.Context, d domain.Device) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Devices[d.ID] = d
	return nil
}

func (r *DeviceRepo) GetByID(_ context.Context, id uuid.UUID) (domain.Device, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	d, ok := r.S.Devices[id]
	if !ok {
		return domain.Device{}, domain.ErrNotFound
	}
	return d, nil
}

func (r *DeviceRepo) FindByFingerprint(_ context.Context, principalID uuid.UUID, fingerprint string) (domain.Device, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	for _, d := range r.S.Devices {
		if d.PrincipalID == principalID && d.Fingerprint == fingerprint {
			return d, nil
		}
	}
	return domain.Device{}, domain.ErrNotFound
}

func (r *DeviceRepo) ListByPrincipal(_ context.Context, principalID uuid.UUID) ([]domain.Device, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.Device, 0)
	for _, d := range r.S.Devices {
		if d.PrincipalID == principalID {
			out = append(out, d)
		}
	}
	return out, nil
}

var _ ports.DeviceRepository = (*DeviceRepo)(nil)

// ---------------------------------------------------------------------------
// RoleRepository
// ---------------------------------------------------------------------------

type RoleRepo struct{ S *Store }

func (r *RoleRepo) GetRole(_ context.Context, id uuid.UUID) (domain.Role, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	role, ok := r.S.Roles[id]
	if !ok {
		return domain.Role{}, domain.ErrNotFound
	}
	return role, nil
}

func (r *RoleRepo) GetRoleByName(_ context.Context, tenantID *uuid.UUID, name string) (domain.Role, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	for _, role := range r.S.Roles {
		if role.Name != name {
			continue
		}
		if tenantID == nil && role.TenantID == nil {
			return role, nil
		}
		if tenantID != nil && role.TenantID != nil && *tenantID == *role.TenantID {
			return role, nil
		}
	}
	return domain.Role{}, domain.ErrNotFound
}

func (r *RoleRepo) ListRolePermissions(_ context.Context, roleID uuid.UUID) ([]domain.Permission, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	return append([]domain.Permission{}, r.S.RolePerms[roleID]...), nil
}

func (r *RoleRepo) RoleGraph(_ context.Context, tenantID uuid.UUID) (policy.RoleGraph, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	g := make(policy.RoleGraph)
	for id, role := range r.S.Roles {
		if role.TenantID != nil && *role.TenantID != tenantID && role.Kind != domain.RoleKindPlatform {
			continue
		}
		g[id] = policy.RoleNode{
			Role:        role,
			ParentIDs:   append([]uuid.UUID{}, r.S.RoleParents[id]...),
			Permissions: append([]domain.Permission{}, r.S.RolePerms[id]...),
		}
	}
	return g, nil
}

func (r *RoleRepo) AssignRole(_ context.Context, pr domain.PrincipalRole) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.PrincRoles[pr.ID] = pr
	return nil
}

func (r *RoleRepo) ListPrincipalRoles(_ context.Context, principalID uuid.UUID) ([]domain.PrincipalRole, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.PrincipalRole, 0)
	for _, pr := range r.S.PrincRoles {
		if pr.PrincipalID == principalID {
			out = append(out, pr)
		}
	}
	return out, nil
}

func (r *RoleRepo) CreateTemporaryGrant(_ context.Context, g domain.TemporaryGrant) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.TempGrants[g.ID] = g
	return nil
}

func (r *RoleRepo) ListTemporaryGrants(_ context.Context, principalID uuid.UUID) ([]domain.TemporaryGrant, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.TemporaryGrant, 0)
	for _, g := range r.S.TempGrants {
		if g.PrincipalID == principalID {
			out = append(out, g)
		}
	}
	return out, nil
}

func (r *RoleRepo) GetPermission(_ context.Context, id uuid.UUID) (domain.Permission, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	p, ok := r.S.Permissions[id]
	if !ok {
		return domain.Permission{}, domain.ErrNotFound
	}
	return p, nil
}

func (r *RoleRepo) FindPermission(_ context.Context, resource, action string) (domain.Permission, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	for _, p := range r.S.Permissions {
		if p.Resource == resource && p.Action == action {
			return p, nil
		}
	}
	return domain.Permission{}, domain.ErrNotFound
}

// Seed helpers for tests.
func (r *RoleRepo) SeedRole(role domain.Role, parents []uuid.UUID, perms []domain.Permission) {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Roles[role.ID] = role
	r.S.RoleParents[role.ID] = parents
	r.S.RolePerms[role.ID] = perms
	for _, p := range perms {
		r.S.Permissions[p.ID] = p
	}
}

func (r *RoleRepo) SeedPermission(p domain.Permission) {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Permissions[p.ID] = p
}

var _ ports.RoleRepository = (*RoleRepo)(nil)

// ---------------------------------------------------------------------------
// AuditRepository
// ---------------------------------------------------------------------------

type AuditRepo struct{ S *Store }

func (r *AuditRepo) Append(_ context.Context, e domain.AuditEvent) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Audit = append(r.S.Audit, e)
	return nil
}

func (r *AuditRepo) ListByPrincipal(_ context.Context, principalID uuid.UUID, limit int) ([]domain.AuditEvent, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.AuditEvent, 0)
	for _, e := range r.S.Audit {
		if e.ActorID != nil && *e.ActorID == principalID {
			out = append(out, e)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

var _ ports.AuditRepository = (*AuditRepo)(nil)

// ---------------------------------------------------------------------------
// OAuthRepository
// ---------------------------------------------------------------------------

type OAuthRepo struct{ S *Store }

func (r *OAuthRepo) GetClientByClientID(_ context.Context, clientID string) (ports.OAuthClient, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	c, ok := r.S.Clients[clientID]
	if !ok {
		return ports.OAuthClient{}, domain.ErrNotFound
	}
	return c, nil
}

func (r *OAuthRepo) SaveOTPChallenge(_ context.Context, c ports.OTPChallenge) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.OTP[c.ID] = c
	return nil
}

func (r *OAuthRepo) GetOTPChallenge(_ context.Context, id uuid.UUID) (ports.OTPChallenge, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	c, ok := r.S.OTP[id]
	if !ok {
		return ports.OTPChallenge{}, domain.ErrNotFound
	}
	return c, nil
}

func (r *OAuthRepo) DeleteOTPChallenge(_ context.Context, id uuid.UUID) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	delete(r.S.OTP, id)
	return nil
}

func (r *OAuthRepo) UpdateOTPChallenge(_ context.Context, c ports.OTPChallenge) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.OTP[c.ID] = c
	return nil
}

func (r *OAuthRepo) SaveMagicLink(_ context.Context, c ports.MagicLinkChallenge) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Magic[c.TokenHash] = c
	return nil
}

func (r *OAuthRepo) GetMagicLinkByHash(_ context.Context, hash string) (ports.MagicLinkChallenge, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	c, ok := r.S.Magic[hash]
	if !ok {
		return ports.MagicLinkChallenge{}, domain.ErrNotFound
	}
	return c, nil
}

func (r *OAuthRepo) ConsumeMagicLink(_ context.Context, id uuid.UUID, at time.Time) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	for k, c := range r.S.Magic {
		if c.ID == id {
			c.ConsumedAt = &at
			r.S.Magic[k] = c
			return nil
		}
	}
	return domain.ErrNotFound
}

func (r *OAuthRepo) SavePasswordReset(_ context.Context, c ports.PasswordResetChallenge) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.PassReset[c.TokenHash] = c
	return nil
}

func (r *OAuthRepo) GetPasswordResetByHash(_ context.Context, hash string) (ports.PasswordResetChallenge, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	c, ok := r.S.PassReset[hash]
	if !ok {
		return ports.PasswordResetChallenge{}, domain.ErrNotFound
	}
	return c, nil
}

func (r *OAuthRepo) ConsumePasswordReset(_ context.Context, id uuid.UUID, at time.Time) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	for k, c := range r.S.PassReset {
		if c.ID == id {
			c.ConsumedAt = &at
			r.S.PassReset[k] = c
			return nil
		}
	}
	return domain.ErrNotFound
}

func (r *OAuthRepo) SaveMFAChallenge(_ context.Context, c ports.MFAChallenge) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.MFAChal[c.ID] = c
	return nil
}

func (r *OAuthRepo) GetMFAChallenge(_ context.Context, id uuid.UUID) (ports.MFAChallenge, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	c, ok := r.S.MFAChal[id]
	if !ok {
		return ports.MFAChallenge{}, domain.ErrNotFound
	}
	return c, nil
}

func (r *OAuthRepo) DeleteMFAChallenge(_ context.Context, id uuid.UUID) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	delete(r.S.MFAChal, id)
	return nil
}

func (r *OAuthRepo) SaveWebAuthnCeremony(_ context.Context, session *webauthn.CeremonySession) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Ceremonies[session.ID] = session
	return nil
}

func (r *OAuthRepo) GetWebAuthnCeremony(_ context.Context, id string) (*webauthn.CeremonySession, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	s, ok := r.S.Ceremonies[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return s, nil
}

func (r *OAuthRepo) DeleteWebAuthnCeremony(_ context.Context, id string) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	delete(r.S.Ceremonies, id)
	return nil
}

var _ ports.OAuthRepository = (*OAuthRepo)(nil)

// ---------------------------------------------------------------------------
// RiskRepository
// ---------------------------------------------------------------------------

type RiskRepo struct{ S *Store }

func (r *RiskRepo) AppendRiskEvent(_ context.Context, e domain.RiskEvent) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.RiskEvents = append(r.S.RiskEvents, e)
	return nil
}

func (r *RiskRepo) AppendLoginAttempt(_ context.Context, a domain.LoginAttempt) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.LoginAttempts = append(r.S.LoginAttempts, a)
	return nil
}

func (r *RiskRepo) CountRecentFailures(_ context.Context, tenantID uuid.UUID, identifier string, since time.Time) (int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	n := 0
	for _, a := range r.S.LoginAttempts {
		if a.Identifier != identifier {
			continue
		}
		if a.TenantID != nil && *a.TenantID != tenantID {
			continue
		}
		if a.CreatedAt.Before(since) {
			continue
		}
		if a.Result == domain.LoginAttemptInvalidCredentials {
			n++
		}
	}
	return n, nil
}

func (r *RiskRepo) GetSecurityPolicy(_ context.Context, tenantID uuid.UUID) (domain.SecurityPolicy, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	if p, ok := r.S.Policies[tenantID]; ok {
		return p, nil
	}
	return domain.SecurityPolicy{}, domain.ErrNotFound
}

func (r *RiskRepo) SeedPolicy(tenantID uuid.UUID, p domain.SecurityPolicy) {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Policies[tenantID] = p
}

var _ ports.RiskRepository = (*RiskRepo)(nil)

// ---------------------------------------------------------------------------
// Small fakes
// ---------------------------------------------------------------------------

type Clock struct{ T time.Time }

func (c *Clock) Now() time.Time {
	if c.T.IsZero() {
		return time.Now().UTC()
	}
	return c.T.UTC()
}

type IDGen struct{}

func (IDGen) New() uuid.UUID { return uuid.New() }

type OTPSender struct {
	LastPhone string
	LastCode  string
}

func (o *OTPSender) SendOTP(_ context.Context, _ uuid.UUID, phone, code string) error {
	o.LastPhone = phone
	o.LastCode = code
	return nil
}

type EventPublisher struct {
	Events []struct {
		Topic   string
		Key     string
		Payload any
	}
}

func (e *EventPublisher) Publish(_ context.Context, topic, key string, payload any) error {
	e.Events = append(e.Events, struct {
		Topic   string
		Key     string
		Payload any
	}{topic, key, payload})
	return nil
}

var _ ports.Clock = (*Clock)(nil)
var _ ports.IDGen = (IDGen{})
var _ ports.OTPSender = (*OTPSender)(nil)
var _ ports.EventPublisher = (*EventPublisher)(nil)
