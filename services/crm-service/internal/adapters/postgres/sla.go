package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/crm-service/internal/app/ports"
	"github.com/nexora/crm-service/internal/domain"
)

// SLARepo persists SLA policies and escalations.
type SLARepo struct{ DB *sql.DB }

func (r *SLARepo) SavePolicy(ctx context.Context, p domain.SLAPolicy) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sla_policies (
			id, tenant_id, name, priority, first_response_minutes, resolve_minutes, active, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			name=EXCLUDED.name,
			priority=EXCLUDED.priority,
			first_response_minutes=EXCLUDED.first_response_minutes,
			resolve_minutes=EXCLUDED.resolve_minutes,
			active=EXCLUDED.active,
			updated_at=EXCLUDED.updated_at`,
		p.ID, p.TenantID, p.Name, p.Priority, p.FirstResponseMinutes, p.ResolveMinutes,
		p.Active, p.CreatedAt.UTC(), p.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *SLARepo) GetPolicyByPriority(ctx context.Context, tenantID uuid.UUID, priority string) (domain.SLAPolicy, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, priority, first_response_minutes, resolve_minutes, active, created_at, updated_at
		FROM sla_policies WHERE tenant_id=$1 AND priority=$2`, tenantID, priority)
	var p domain.SLAPolicy
	err := row.Scan(
		&p.ID, &p.TenantID, &p.Name, &p.Priority, &p.FirstResponseMinutes, &p.ResolveMinutes,
		&p.Active, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return domain.SLAPolicy{}, mapNotFound(err)
	}
	p.CreatedAt = p.CreatedAt.UTC()
	p.UpdatedAt = p.UpdatedAt.UTC()
	return p, nil
}

func (r *SLARepo) ListPolicies(ctx context.Context, tenantID uuid.UUID) ([]domain.SLAPolicy, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, name, priority, first_response_minutes, resolve_minutes, active, created_at, updated_at
		FROM sla_policies WHERE tenant_id=$1 ORDER BY priority ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.SLAPolicy{}
	for rows.Next() {
		var p domain.SLAPolicy
		if err := rows.Scan(
			&p.ID, &p.TenantID, &p.Name, &p.Priority, &p.FirstResponseMinutes, &p.ResolveMinutes,
			&p.Active, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.CreatedAt = p.CreatedAt.UTC()
		p.UpdatedAt = p.UpdatedAt.UTC()
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *SLARepo) SaveEscalation(ctx context.Context, e domain.Escalation) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO escalations (
			id, tenant_id, ticket_id, from_priority, to_priority, reason, triggered_by_sla, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		e.ID, e.TenantID, e.TicketID, e.FromPriority, e.ToPriority, e.Reason, e.TriggeredBySLA, e.CreatedAt.UTC())
	return err
}

func (r *SLARepo) ListEscalations(ctx context.Context, tenantID, ticketID uuid.UUID) ([]domain.Escalation, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, ticket_id, from_priority, to_priority, reason, triggered_by_sla, created_at
		FROM escalations WHERE tenant_id=$1 AND ticket_id=$2 ORDER BY created_at ASC`,
		tenantID, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Escalation{}
	for rows.Next() {
		var e domain.Escalation
		if err := rows.Scan(
			&e.ID, &e.TenantID, &e.TicketID, &e.FromPriority, &e.ToPriority,
			&e.Reason, &e.TriggeredBySLA, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.CreatedAt = e.CreatedAt.UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}

var _ ports.SLARepo = (*SLARepo)(nil)
