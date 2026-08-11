package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
)

type AuditRepo struct{ DB *sql.DB }

func (r *AuditRepo) Append(ctx context.Context, e domain.AuditEvent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO audit_events (
			id, tenant_id, actor_id, actor_kind, action, resource_type, resource_id, outcome,
			ip, user_agent, session_id, request_id, details, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::inet,$10,$11,$12,$13::jsonb,$14)`,
		e.ID, nullUUID(e.TenantID), nullUUID(e.ActorID), e.ActorKind, e.Action, e.ResourceType, nullStr(e.ResourceID),
		string(e.Outcome), ipString(e.IP), e.UserAgent, nullUUID(e.SessionID), nullStr(e.RequestID), jsonMap(e.Details), e.CreatedAt)
	return err
}

func (r *AuditRepo) ListByPrincipal(ctx context.Context, principalID uuid.UUID, limit int) ([]domain.AuditEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, actor_id, actor_kind, action, resource_type, COALESCE(resource_id,''), outcome,
			host(ip)::text, user_agent, session_id, COALESCE(request_id,''), details, created_at
		FROM audit_events WHERE actor_id=$1 ORDER BY created_at DESC LIMIT $2`, principalID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.AuditEvent{}
	for rows.Next() {
		var e domain.AuditEvent
		var tenant, actor, session uuid.NullUUID
		var outcome string
		var ip sql.NullString
		var details []byte
		if err := rows.Scan(&e.ID, &tenant, &actor, &e.ActorKind, &e.Action, &e.ResourceType, &e.ResourceID, &outcome,
			&ip, &e.UserAgent, &session, &e.RequestID, &details, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.TenantID = scanUUIDPtr(tenant)
		e.ActorID = scanUUIDPtr(actor)
		e.SessionID = scanUUIDPtr(session)
		e.Outcome = domain.AuditOutcome(outcome)
		e.IP = parseIP(ip)
		e.Details = scanJSONMap(details)
		out = append(out, e)
	}
	return out, rows.Err()
}

var _ ports.AuditRepository = (*AuditRepo)(nil)
