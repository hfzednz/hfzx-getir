package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
	"github.com/nexora/identity-service/internal/security/webauthn"
)

type OAuthRepo struct{ DB *sql.DB }

func (r *OAuthRepo) GetClientByClientID(ctx context.Context, clientID string) (ports.OAuthClient, error) {
	var c ports.OAuthClient
	var secret sql.NullString
	var tenant, principal uuid.NullUUID
	var grants, scopes []string
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, client_id, client_secret_hash, principal_id, grant_types, scopes, active, created_at
		FROM oauth_clients WHERE client_id=$1`, clientID).
		Scan(&c.ID, &tenant, &c.ClientID, &secret, &principal, pq.Array(&grants), pq.Array(&scopes), &c.Enabled, &c.CreatedAt)
	if err != nil {
		return ports.OAuthClient{}, mapNotFound(err)
	}
	if tenant.Valid {
		c.TenantID = tenant.UUID
	}
	if principal.Valid {
		c.PrincipalID = principal.UUID
	}
	if secret.Valid {
		c.ClientSecret = secret.String
	}
	c.GrantTypes = grants
	c.Scopes = scopes
	return c, nil
}

func (r *OAuthRepo) SaveOTPChallenge(ctx context.Context, c ports.OTPChallenge) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO otp_challenges (id, tenant_id, phone, code_hash, expires_at, attempts, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET code_hash=EXCLUDED.code_hash, expires_at=EXCLUDED.expires_at, attempts=EXCLUDED.attempts`,
		c.ID, c.TenantID, c.Phone, c.CodeHash, c.ExpiresAt, c.Attempts, c.CreatedAt)
	return err
}

func (r *OAuthRepo) GetOTPChallenge(ctx context.Context, id uuid.UUID) (ports.OTPChallenge, error) {
	var c ports.OTPChallenge
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, phone, code_hash, expires_at, attempts, created_at FROM otp_challenges WHERE id=$1`, id).
		Scan(&c.ID, &c.TenantID, &c.Phone, &c.CodeHash, &c.ExpiresAt, &c.Attempts, &c.CreatedAt)
	if err != nil {
		return ports.OTPChallenge{}, mapNotFound(err)
	}
	return c, nil
}

func (r *OAuthRepo) DeleteOTPChallenge(ctx context.Context, id uuid.UUID) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM otp_challenges WHERE id=$1`, id)
	return err
}

func (r *OAuthRepo) UpdateOTPChallenge(ctx context.Context, c ports.OTPChallenge) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE otp_challenges SET attempts=$2, expires_at=$3 WHERE id=$1`, c.ID, c.Attempts, c.ExpiresAt)
	return err
}

func (r *OAuthRepo) SaveMagicLink(ctx context.Context, c ports.MagicLinkChallenge) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO magic_link_challenges (id, tenant_id, principal_id, token_hash, expires_at, consumed_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		c.ID, c.TenantID, c.PrincipalID, c.TokenHash, c.ExpiresAt, c.ConsumedAt, c.CreatedAt)
	return err
}

func (r *OAuthRepo) GetMagicLinkByHash(ctx context.Context, hash string) (ports.MagicLinkChallenge, error) {
	var c ports.MagicLinkChallenge
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, principal_id, token_hash, expires_at, consumed_at, created_at
		FROM magic_link_challenges WHERE token_hash=$1`, hash).
		Scan(&c.ID, &c.TenantID, &c.PrincipalID, &c.TokenHash, &c.ExpiresAt, &c.ConsumedAt, &c.CreatedAt)
	if err != nil {
		return ports.MagicLinkChallenge{}, mapNotFound(err)
	}
	return c, nil
}

func (r *OAuthRepo) ConsumeMagicLink(ctx context.Context, id uuid.UUID, at time.Time) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE magic_link_challenges SET consumed_at=$2 WHERE id=$1 AND consumed_at IS NULL`, id, at)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *OAuthRepo) SavePasswordReset(ctx context.Context, c ports.PasswordResetChallenge) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO password_reset_challenges (id, tenant_id, principal_id, token_hash, expires_at, consumed_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		c.ID, c.TenantID, c.PrincipalID, c.TokenHash, c.ExpiresAt, c.ConsumedAt, c.CreatedAt)
	return err
}

func (r *OAuthRepo) GetPasswordResetByHash(ctx context.Context, hash string) (ports.PasswordResetChallenge, error) {
	var c ports.PasswordResetChallenge
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, principal_id, token_hash, expires_at, consumed_at, created_at
		FROM password_reset_challenges WHERE token_hash=$1`, hash).
		Scan(&c.ID, &c.TenantID, &c.PrincipalID, &c.TokenHash, &c.ExpiresAt, &c.ConsumedAt, &c.CreatedAt)
	if err != nil {
		return ports.PasswordResetChallenge{}, mapNotFound(err)
	}
	return c, nil
}

func (r *OAuthRepo) ConsumePasswordReset(ctx context.Context, id uuid.UUID, at time.Time) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE password_reset_challenges SET consumed_at=$2 WHERE id=$1 AND consumed_at IS NULL`, id, at)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *OAuthRepo) SaveMFAChallenge(ctx context.Context, c ports.MFAChallenge) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO mfa_challenges (id, principal_id, session_hint, factor_type, expires_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		c.ID, c.PrincipalID, c.SessionHint, string(c.FactorType), c.ExpiresAt, c.CreatedAt)
	return err
}

func (r *OAuthRepo) GetMFAChallenge(ctx context.Context, id uuid.UUID) (ports.MFAChallenge, error) {
	var c ports.MFAChallenge
	var ft string
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, principal_id, session_hint, factor_type, expires_at, created_at FROM mfa_challenges WHERE id=$1`, id).
		Scan(&c.ID, &c.PrincipalID, &c.SessionHint, &ft, &c.ExpiresAt, &c.CreatedAt)
	if err != nil {
		return ports.MFAChallenge{}, mapNotFound(err)
	}
	c.FactorType = domain.MFAFactorType(ft)
	return c, nil
}

func (r *OAuthRepo) DeleteMFAChallenge(ctx context.Context, id uuid.UUID) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM mfa_challenges WHERE id=$1`, id)
	return err
}

func (r *OAuthRepo) SaveWebAuthnCeremony(ctx context.Context, session *webauthn.CeremonySession) error {
	if session == nil {
		return domain.ErrInvalidArgument
	}
	payload, _ := json.Marshal(session)
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO webauthn_ceremonies (id, user_id, challenge, type, expires_at, payload, created_at)
		VALUES ($1,$2,$3,$4,$5,$6::jsonb,NOW())
		ON CONFLICT (id) DO UPDATE SET user_id=EXCLUDED.user_id, challenge=EXCLUDED.challenge,
			type=EXCLUDED.type, expires_at=EXCLUDED.expires_at, payload=EXCLUDED.payload`,
		session.ID, session.UserID, session.Challenge, session.Type, session.ExpiresAt, payload)
	return err
}

func (r *OAuthRepo) GetWebAuthnCeremony(ctx context.Context, id string) (*webauthn.CeremonySession, error) {
	var payload []byte
	var s webauthn.CeremonySession
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, user_id, challenge, type, expires_at, payload FROM webauthn_ceremonies WHERE id=$1`, id).
		Scan(&s.ID, &s.UserID, &s.Challenge, &s.Type, &s.ExpiresAt, &payload)
	if err != nil {
		return nil, mapNotFound(err)
	}
	_ = json.Unmarshal(payload, &s)
	return &s, nil
}

func (r *OAuthRepo) DeleteWebAuthnCeremony(ctx context.Context, id string) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM webauthn_ceremonies WHERE id=$1`, id)
	return err
}

var _ ports.OAuthRepository = (*OAuthRepo)(nil)
