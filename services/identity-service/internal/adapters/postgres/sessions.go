package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
)

type SessionRepo struct{ DB *sql.DB }

func (r *SessionRepo) Create(ctx context.Context, s domain.Session) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sessions (
			id, principal_id, device_id, tenant_id, amr, acr, ip, user_agent, risk_score,
			idle_expires_at, absolute_expires_at, last_seen_at, created_at, revoked_at, revoke_reason
		) VALUES ($1,$2,$3,$4,$5,$6,$7::inet,$8,$9,$10,$11,$12,$13,$14,$15)`,
		s.ID, s.PrincipalID, nullUUID(s.DeviceID), s.TenantID, pq.Array(s.AMR), s.ACR, ipString(s.IP), s.UserAgent, s.RiskScore,
		s.IdleExpiresAt, s.AbsoluteExpiresAt, s.LastSeenAt, s.CreatedAt, s.RevokedAt, nullStr(s.RevokeReason))
	return err
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (r *SessionRepo) Update(ctx context.Context, s domain.Session) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE sessions SET device_id=$2, amr=$3, acr=$4, ip=$5::inet, user_agent=$6, risk_score=$7,
			idle_expires_at=$8, absolute_expires_at=$9, last_seen_at=$10, revoked_at=$11, revoke_reason=$12
		WHERE id=$1`,
		s.ID, nullUUID(s.DeviceID), pq.Array(s.AMR), s.ACR, ipString(s.IP), s.UserAgent, s.RiskScore,
		s.IdleExpiresAt, s.AbsoluteExpiresAt, s.LastSeenAt, s.RevokedAt, nullStr(s.RevokeReason))
	return err
}

func (r *SessionRepo) scan(row interface {
	Scan(dest ...any) error
}) (domain.Session, error) {
	var s domain.Session
	var device uuid.NullUUID
	var amr []string
	var ip sql.NullString
	var revoke sql.NullString
	err := row.Scan(&s.ID, &s.PrincipalID, &device, &s.TenantID, pq.Array(&amr), &s.ACR, &ip, &s.UserAgent, &s.RiskScore,
		&s.IdleExpiresAt, &s.AbsoluteExpiresAt, &s.LastSeenAt, &s.CreatedAt, &s.RevokedAt, &revoke)
	if err != nil {
		return domain.Session{}, mapNotFound(err)
	}
	s.DeviceID = scanUUIDPtr(device)
	s.AMR = amr
	s.IP = parseIP(ip)
	if revoke.Valid {
		s.RevokeReason = revoke.String
	}
	return s, nil
}

func (r *SessionRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Session, error) {
	return r.scan(r.DB.QueryRowContext(ctx, `
		SELECT id, principal_id, device_id, tenant_id, amr, acr, host(ip)::text, user_agent, risk_score,
			idle_expires_at, absolute_expires_at, last_seen_at, created_at, revoked_at, revoke_reason
		FROM sessions WHERE id=$1`, id))
}

func (r *SessionRepo) ListByPrincipal(ctx context.Context, principalID uuid.UUID) ([]domain.Session, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, principal_id, device_id, tenant_id, amr, acr, host(ip)::text, user_agent, risk_score,
			idle_expires_at, absolute_expires_at, last_seen_at, created_at, revoked_at, revoke_reason
		FROM sessions WHERE principal_id=$1 ORDER BY created_at DESC`, principalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Session{}
	for rows.Next() {
		s, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *SessionRepo) Revoke(ctx context.Context, id uuid.UUID, at time.Time, reason string) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE sessions SET revoked_at=$2, revoke_reason=$3 WHERE id=$1 AND revoked_at IS NULL`, id, at, reason)
	return err
}

func (r *SessionRepo) RevokeFamily(ctx context.Context, familyID uuid.UUID, at time.Time, reason string) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE sessions SET revoked_at=$2, revoke_reason=$3
		WHERE id IN (SELECT DISTINCT session_id FROM refresh_tokens WHERE family_id=$1) AND revoked_at IS NULL`,
		familyID, at, reason)
	if err != nil {
		return err
	}
	_, err = r.DB.ExecContext(ctx, `
		UPDATE refresh_tokens SET revoked_at=$2, revoke_reason=$3 WHERE family_id=$1 AND revoked_at IS NULL`,
		familyID, at, reason)
	return err
}

func (r *SessionRepo) CreateRefresh(ctx context.Context, t domain.RefreshToken) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO refresh_tokens (id, session_id, principal_id, token_hash, family_id, rotated_from, expires_at, created_at, revoked_at, revoke_reason)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		t.ID, t.SessionID, t.PrincipalID, t.TokenHash, t.FamilyID, nullUUID(t.RotatedFrom), t.ExpiresAt, t.CreatedAt, t.RevokedAt, nullStr(t.RevokeReason))
	return err
}

func (r *SessionRepo) GetRefreshByHash(ctx context.Context, hash string) (domain.RefreshToken, error) {
	var t domain.RefreshToken
	var rotated uuid.NullUUID
	var revoke sql.NullString
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, session_id, principal_id, token_hash, family_id, rotated_from, expires_at, created_at, revoked_at, revoke_reason
		FROM refresh_tokens WHERE token_hash=$1`, hash).
		Scan(&t.ID, &t.SessionID, &t.PrincipalID, &t.TokenHash, &t.FamilyID, &rotated, &t.ExpiresAt, &t.CreatedAt, &t.RevokedAt, &revoke)
	if err != nil {
		return domain.RefreshToken{}, mapNotFound(err)
	}
	t.RotatedFrom = scanUUIDPtr(rotated)
	if revoke.Valid {
		t.RevokeReason = revoke.String
	}
	return t, nil
}

func (r *SessionRepo) UpdateRefresh(ctx context.Context, t domain.RefreshToken) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE refresh_tokens SET revoked_at=$2, revoke_reason=$3 WHERE id=$1`, t.ID, t.RevokedAt, nullStr(t.RevokeReason))
	return err
}

func (r *SessionRepo) ListRefreshByFamily(ctx context.Context, familyID uuid.UUID) ([]domain.RefreshToken, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, session_id, principal_id, token_hash, family_id, rotated_from, expires_at, created_at, revoked_at, revoke_reason
		FROM refresh_tokens WHERE family_id=$1 ORDER BY created_at`, familyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.RefreshToken{}
	for rows.Next() {
		var t domain.RefreshToken
		var rotated uuid.NullUUID
		var revoke sql.NullString
		if err := rows.Scan(&t.ID, &t.SessionID, &t.PrincipalID, &t.TokenHash, &t.FamilyID, &rotated, &t.ExpiresAt, &t.CreatedAt, &t.RevokedAt, &revoke); err != nil {
			return nil, err
		}
		t.RotatedFrom = scanUUIDPtr(rotated)
		if revoke.Valid {
			t.RevokeReason = revoke.String
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

var _ ports.SessionRepository = (*SessionRepo)(nil)
