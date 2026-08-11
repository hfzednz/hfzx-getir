package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/lib/pq"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
)

func init() {
	// Ensure pgx stdlib driver is registered as "pgx".
	_ = stdlib.GetDefaultDriver()
}

// Open opens a Postgres pool via pgx stdlib and verifies connectivity.
func Open(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, errors.New("postgres: empty DATABASE_URL")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: open: %w", err)
	}
	db.SetMaxOpenConns(40)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	return db, nil
}

func mapNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}

func nullUUID(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return *id
}

func scanUUIDPtr(ns uuid.NullUUID) *uuid.UUID {
	if !ns.Valid {
		return nil
	}
	id := ns.UUID
	return &id
}

func ipString(ip *netip.Addr) any {
	if ip == nil {
		return nil
	}
	return ip.String()
}

func parseIP(s sql.NullString) *netip.Addr {
	if !s.Valid || s.String == "" {
		return nil
	}
	a, err := netip.ParseAddr(s.String)
	if err != nil {
		return nil
	}
	return &a
}

func jsonMap(v map[string]any) []byte {
	if v == nil {
		return []byte("{}")
	}
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return b
}

func scanJSONMap(b []byte) map[string]any {
	if len(b) == 0 {
		return map[string]any{}
	}
	out := map[string]any{}
	_ = json.Unmarshal(b, &out)
	return out
}

// ---------------------------------------------------------------------------
// PrincipalRepo
// ---------------------------------------------------------------------------

type PrincipalRepo struct{ DB *sql.DB }

func (r *PrincipalRepo) Create(ctx context.Context, p domain.Principal) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO principals (id, tenant_id, kind, status, display_name, created_at, updated_at, deleted_at)
		VALUES ($1,$2,$3::principal_kind,$4::principal_status,$5,$6,$7,$8)`,
		p.ID, p.TenantID, string(p.Kind), string(p.Status), p.DisplayName, p.CreatedAt, p.UpdatedAt, p.DeletedAt)
	return err
}

func (r *PrincipalRepo) Update(ctx context.Context, p domain.Principal) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE principals SET status=$2::principal_status, display_name=$3, updated_at=$4, deleted_at=$5
		WHERE id=$1`, p.ID, string(p.Status), p.DisplayName, p.UpdatedAt, p.DeletedAt)
	return err
}

func (r *PrincipalRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Principal, error) {
	var p domain.Principal
	var kind, status string
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, kind::text, status::text, COALESCE(display_name,''), created_at, updated_at, deleted_at
		FROM principals WHERE id=$1`, id).
		Scan(&p.ID, &p.TenantID, &kind, &status, &p.DisplayName, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if err != nil {
		return domain.Principal{}, mapNotFound(err)
	}
	p.Kind = domain.PrincipalKind(kind)
	p.Status = domain.PrincipalStatus(status)
	return p, nil
}

func (r *PrincipalRepo) Search(ctx context.Context, tenantID uuid.UUID, query string, limit int) ([]domain.Principal, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT p.id, p.tenant_id, p.kind::text, p.status::text, COALESCE(p.display_name,''), p.created_at, p.updated_at, p.deleted_at
		FROM principals p
		LEFT JOIN identifiers i ON i.principal_id = p.id
		WHERE p.tenant_id=$1 AND (
			p.display_name ILIKE '%'||$2||'%' OR i.value ILIKE '%'||$2||'%'
		)
		GROUP BY p.id
		ORDER BY p.created_at DESC
		LIMIT $3`, tenantID, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Principal{}
	for rows.Next() {
		var p domain.Principal
		var kind, status string
		if err := rows.Scan(&p.ID, &p.TenantID, &kind, &status, &p.DisplayName, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt); err != nil {
			return nil, err
		}
		p.Kind = domain.PrincipalKind(kind)
		p.Status = domain.PrincipalStatus(status)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PrincipalRepo) CreateIdentifier(ctx context.Context, id domain.Identifier) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO identifiers (id, principal_id, tenant_id, type, value, verified_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4::identifier_type,$5,$6,$7,$8)`,
		id.ID, id.PrincipalID, id.TenantID, string(id.Type), id.Value, id.VerifiedAt, id.CreatedAt, id.UpdatedAt)
	return err
}

func (r *PrincipalRepo) FindIdentifier(ctx context.Context, tenantID uuid.UUID, typ domain.IdentifierType, value string) (domain.Identifier, error) {
	var id domain.Identifier
	var t string
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, principal_id, tenant_id, type::text, value, verified_at, created_at, updated_at
		FROM identifiers WHERE tenant_id=$1 AND type=$2::identifier_type AND value=$3`,
		tenantID, string(typ), value).
		Scan(&id.ID, &id.PrincipalID, &id.TenantID, &t, &id.Value, &id.VerifiedAt, &id.CreatedAt, &id.UpdatedAt)
	if err != nil {
		return domain.Identifier{}, mapNotFound(err)
	}
	id.Type = domain.IdentifierType(t)
	return id, nil
}

func (r *PrincipalRepo) ListIdentifiers(ctx context.Context, principalID uuid.UUID) ([]domain.Identifier, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, principal_id, tenant_id, type::text, value, verified_at, created_at, updated_at
		FROM identifiers WHERE principal_id=$1 ORDER BY created_at`, principalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Identifier{}
	for rows.Next() {
		var id domain.Identifier
		var t string
		if err := rows.Scan(&id.ID, &id.PrincipalID, &id.TenantID, &t, &id.Value, &id.VerifiedAt, &id.CreatedAt, &id.UpdatedAt); err != nil {
			return nil, err
		}
		id.Type = domain.IdentifierType(t)
		out = append(out, id)
	}
	return out, rows.Err()
}

func (r *PrincipalRepo) UpsertCredential(ctx context.Context, c domain.Credential) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO credentials (id, principal_id, password_hash, algorithm, password_changed_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4::credential_algorithm,$5,$6,$7)
		ON CONFLICT (principal_id) DO UPDATE SET
			password_hash=EXCLUDED.password_hash,
			algorithm=EXCLUDED.algorithm,
			password_changed_at=EXCLUDED.password_changed_at,
			updated_at=EXCLUDED.updated_at`,
		c.ID, c.PrincipalID, c.PasswordHash, string(c.Algorithm), c.PasswordChangedAt, c.CreatedAt, c.UpdatedAt)
	return err
}

func (r *PrincipalRepo) GetCredential(ctx context.Context, principalID uuid.UUID) (domain.Credential, error) {
	var c domain.Credential
	var algo string
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, principal_id, password_hash, algorithm::text, password_changed_at, created_at, updated_at
		FROM credentials WHERE principal_id=$1`, principalID).
		Scan(&c.ID, &c.PrincipalID, &c.PasswordHash, &algo, &c.PasswordChangedAt, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return domain.Credential{}, mapNotFound(err)
	}
	c.Algorithm = domain.CredentialAlgorithm(algo)
	return c, nil
}

func (r *PrincipalRepo) AddPasswordHistory(ctx context.Context, e domain.PasswordHistoryEntry) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO password_history (id, principal_id, password_hash, algorithm, created_at)
		VALUES ($1,$2,$3,$4::credential_algorithm,$5)`,
		e.ID, e.PrincipalID, e.PasswordHash, string(e.Algorithm), e.CreatedAt)
	return err
}

func (r *PrincipalRepo) ListPasswordHistory(ctx context.Context, principalID uuid.UUID, limit int) ([]domain.PasswordHistoryEntry, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, principal_id, password_hash, algorithm::text, created_at
		FROM password_history WHERE principal_id=$1 ORDER BY created_at DESC LIMIT $2`, principalID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.PasswordHistoryEntry{}
	for rows.Next() {
		var e domain.PasswordHistoryEntry
		var algo string
		if err := rows.Scan(&e.ID, &e.PrincipalID, &e.PasswordHash, &algo, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.Algorithm = domain.CredentialAlgorithm(algo)
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *PrincipalRepo) CreateMFAFactor(ctx context.Context, f domain.MFAFactor) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO mfa_factors (id, principal_id, type, label, secret_enc, verified, verified_at, disabled_at, created_at, updated_at)
		VALUES ($1,$2,$3::mfa_factor_type,$4,$5,$6,$7,$8,$9,$10)`,
		f.ID, f.PrincipalID, string(f.Type), f.Label, f.SecretEnc, f.Verified, f.VerifiedAt, f.DisabledAt, f.CreatedAt, f.UpdatedAt)
	return err
}

func (r *PrincipalRepo) UpdateMFAFactor(ctx context.Context, f domain.MFAFactor) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE mfa_factors SET label=$2, secret_enc=$3, verified=$4, verified_at=$5, disabled_at=$6, updated_at=$7
		WHERE id=$1`, f.ID, f.Label, f.SecretEnc, f.Verified, f.VerifiedAt, f.DisabledAt, f.UpdatedAt)
	return err
}

func (r *PrincipalRepo) ListMFAFactors(ctx context.Context, principalID uuid.UUID) ([]domain.MFAFactor, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, principal_id, type::text, label, secret_enc, verified, verified_at, disabled_at, created_at, updated_at
		FROM mfa_factors WHERE principal_id=$1 ORDER BY created_at`, principalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.MFAFactor{}
	for rows.Next() {
		var f domain.MFAFactor
		var t string
		if err := rows.Scan(&f.ID, &f.PrincipalID, &t, &f.Label, &f.SecretEnc, &f.Verified, &f.VerifiedAt, &f.DisabledAt, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		f.Type = domain.MFAFactorType(t)
		out = append(out, f)
	}
	return out, rows.Err()
}

func (r *PrincipalRepo) GetMFAFactor(ctx context.Context, id uuid.UUID) (domain.MFAFactor, error) {
	var f domain.MFAFactor
	var t string
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, principal_id, type::text, label, secret_enc, verified, verified_at, disabled_at, created_at, updated_at
		FROM mfa_factors WHERE id=$1`, id).
		Scan(&f.ID, &f.PrincipalID, &t, &f.Label, &f.SecretEnc, &f.Verified, &f.VerifiedAt, &f.DisabledAt, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return domain.MFAFactor{}, mapNotFound(err)
	}
	f.Type = domain.MFAFactorType(t)
	return f, nil
}

func (r *PrincipalRepo) ReplaceBackupCodes(ctx context.Context, principalID uuid.UUID, codes []domain.BackupCode) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM backup_codes WHERE principal_id=$1`, principalID); err != nil {
		return err
	}
	for _, c := range codes {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO backup_codes (id, principal_id, code_hash, used_at, created_at)
			VALUES ($1,$2,$3,$4,$5)`, c.ID, c.PrincipalID, c.CodeHash, c.UsedAt, c.CreatedAt); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *PrincipalRepo) ListBackupCodes(ctx context.Context, principalID uuid.UUID) ([]domain.BackupCode, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, principal_id, code_hash, used_at, created_at FROM backup_codes WHERE principal_id=$1`, principalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.BackupCode{}
	for rows.Next() {
		var c domain.BackupCode
		if err := rows.Scan(&c.ID, &c.PrincipalID, &c.CodeHash, &c.UsedAt, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *PrincipalRepo) UpdateBackupCode(ctx context.Context, c domain.BackupCode) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE backup_codes SET used_at=$2 WHERE id=$1`, c.ID, c.UsedAt)
	return err
}

func (r *PrincipalRepo) CreateWebAuthnCredential(ctx context.Context, c domain.WebAuthnCredential) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO webauthn_credentials (
			id, principal_id, credential_id, public_key, aaguid, sign_count, transports,
			nickname, backup_eligible, backup_state, created_at, updated_at, last_used_at, revoked_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		c.ID, c.PrincipalID, c.CredentialID, c.PublicKey, c.AAGUID, int64(c.SignCount), pq.Array(c.Transports),
		c.Nickname, c.BackupEligible, c.BackupState, c.CreatedAt, c.UpdatedAt, c.LastUsedAt, c.RevokedAt)
	return err
}

func (r *PrincipalRepo) UpdateWebAuthnCredential(ctx context.Context, c domain.WebAuthnCredential) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE webauthn_credentials SET public_key=$2, aaguid=$3, sign_count=$4, transports=$5,
			nickname=$6, backup_eligible=$7, backup_state=$8, updated_at=$9, last_used_at=$10, revoked_at=$11
		WHERE id=$1`,
		c.ID, c.PublicKey, c.AAGUID, int64(c.SignCount), pq.Array(c.Transports),
		c.Nickname, c.BackupEligible, c.BackupState, c.UpdatedAt, c.LastUsedAt, c.RevokedAt)
	return err
}

func (r *PrincipalRepo) ListWebAuthnCredentials(ctx context.Context, principalID uuid.UUID) ([]domain.WebAuthnCredential, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, principal_id, credential_id, public_key, aaguid, sign_count, transports,
			nickname, backup_eligible, backup_state, created_at, updated_at, last_used_at, revoked_at
		FROM webauthn_credentials WHERE principal_id=$1`, principalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.WebAuthnCredential{}
	for rows.Next() {
		var c domain.WebAuthnCredential
		var sign int64
		var transports []string
		if err := rows.Scan(&c.ID, &c.PrincipalID, &c.CredentialID, &c.PublicKey, &c.AAGUID, &sign, pq.Array(&transports),
			&c.Nickname, &c.BackupEligible, &c.BackupState, &c.CreatedAt, &c.UpdatedAt, &c.LastUsedAt, &c.RevokedAt); err != nil {
			return nil, err
		}
		c.SignCount = uint64(sign)
		c.Transports = transports
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *PrincipalRepo) CreateConsent(ctx context.Context, c domain.Consent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO consents (id, principal_id, tenant_id, purpose, version, granted, granted_at, revoked_at, ip, user_agent, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::inet,$10,$11,$12)`,
		c.ID, c.PrincipalID, c.TenantID, c.Purpose, c.Version, c.Granted, c.GrantedAt, c.RevokedAt, ipString(c.IP), c.UserAgent, c.CreatedAt, c.UpdatedAt)
	return err
}

func (r *PrincipalRepo) ListConsents(ctx context.Context, principalID uuid.UUID) ([]domain.Consent, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, principal_id, tenant_id, purpose, version, granted, granted_at, revoked_at, host(ip)::text, user_agent, created_at, updated_at
		FROM consents WHERE principal_id=$1`, principalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Consent{}
	for rows.Next() {
		var c domain.Consent
		var ip sql.NullString
		if err := rows.Scan(&c.ID, &c.PrincipalID, &c.TenantID, &c.Purpose, &c.Version, &c.Granted, &c.GrantedAt, &c.RevokedAt, &ip, &c.UserAgent, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		c.IP = parseIP(ip)
		out = append(out, c)
	}
	return out, rows.Err()
}

var _ ports.PrincipalRepository = (*PrincipalRepo)(nil)
