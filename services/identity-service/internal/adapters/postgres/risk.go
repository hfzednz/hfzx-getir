package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
)

type RiskRepo struct{ DB *sql.DB }

func (r *RiskRepo) AppendRiskEvent(ctx context.Context, e domain.RiskEvent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO risk_events (
			id, principal_id, session_id, device_id, tenant_id, event_type, severity,
			score_delta, score_after, ip, details, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7::risk_event_severity,$8,$9,$10::inet,$11::jsonb,$12)`,
		e.ID, nullUUID(e.PrincipalID), nullUUID(e.SessionID), nullUUID(e.DeviceID), nullUUID(e.TenantID),
		e.EventType, string(e.Severity), e.ScoreDelta, e.ScoreAfter, ipString(e.IP), jsonMap(e.Details), e.CreatedAt)
	return err
}

func (r *RiskRepo) AppendLoginAttempt(ctx context.Context, a domain.LoginAttempt) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO login_attempts (
			id, tenant_id, principal_id, identifier, result, ip, user_agent, device_fingerprint, failure_reason, created_at
		) VALUES ($1,$2,$3,$4,$5::login_attempt_result,$6::inet,$7,$8,$9,$10)`,
		a.ID, nullUUID(a.TenantID), nullUUID(a.PrincipalID), a.Identifier, string(a.Result),
		ipString(a.IP), a.UserAgent, nullStr(a.DeviceFingerprint), nullStr(a.FailureReason), a.CreatedAt)
	return err
}

func (r *RiskRepo) CountRecentFailures(ctx context.Context, tenantID uuid.UUID, identifier string, since time.Time) (int, error) {
	var n int
	err := r.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM login_attempts
		WHERE tenant_id=$1 AND identifier=$2 AND created_at >= $3
			AND result IN ('invalid_credentials','mfa_failed','blocked_risk','locked')`,
		tenantID, identifier, since).Scan(&n)
	return n, err
}

func (r *RiskRepo) GetSecurityPolicy(ctx context.Context, tenantID uuid.UUID) (domain.SecurityPolicy, error) {
	var p domain.SecurityPolicy
	var tenant uuid.NullUUID
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, description, enabled,
			password_min_length, password_require_upper, password_require_lower, password_require_digit,
			password_require_symbol, password_history_count, mfa_required, mfa_required_above_risk,
			session_idle_seconds, session_absolute_seconds, refresh_token_seconds, max_concurrent_sessions,
			max_failed_attempts, lockout_seconds, block_above_risk, created_at, updated_at
		FROM security_policies
		WHERE enabled=TRUE AND (tenant_id=$1 OR tenant_id IS NULL)
		ORDER BY CASE WHEN tenant_id=$1 THEN 0 ELSE 1 END
		LIMIT 1`, tenantID).
		Scan(&p.ID, &tenant, &p.Name, &p.Description, &p.Enabled,
			&p.PasswordMinLength, &p.PasswordRequireUpper, &p.PasswordRequireLower, &p.PasswordRequireDigit,
			&p.PasswordRequireSymbol, &p.PasswordHistoryCount, &p.MFARequired, &p.MFARequiredAboveRisk,
			&p.SessionIdleSeconds, &p.SessionAbsoluteSeconds, &p.RefreshTokenSeconds, &p.MaxConcurrentSessions,
			&p.MaxFailedAttempts, &p.LockoutSeconds, &p.BlockAboveRisk, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return domain.SecurityPolicy{}, mapNotFound(err)
	}
	p.TenantID = scanUUIDPtr(tenant)
	return p, nil
}

var _ ports.RiskRepository = (*RiskRepo)(nil)
