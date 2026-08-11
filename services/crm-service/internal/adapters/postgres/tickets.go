package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/crm-service/internal/app/ports"
	"github.com/nexora/crm-service/internal/domain"
)

// TicketRepo persists tickets, events, and notes.
type TicketRepo struct{ DB *sql.DB }

const ticketSelect = `
	SELECT id, tenant_id, customer_id, assignee_id, team_id, status, priority, category,
		subject, description, COALESCE(idempotency_key, ''), merged_into_id,
		first_response_due, resolve_due, first_responded_at, resolved_at, closed_at,
		sla_breached, tags, created_at, updated_at
	FROM tickets`

func (r *TicketRepo) Save(ctx context.Context, t domain.Ticket) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO tickets (
			id, tenant_id, customer_id, assignee_id, team_id, status, priority, category,
			subject, description, idempotency_key, merged_into_id,
			first_response_due, resolve_due, first_responded_at, resolved_at, closed_at,
			sla_breached, tags, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21
		)
		ON CONFLICT (id) DO UPDATE SET
			customer_id=EXCLUDED.customer_id,
			assignee_id=EXCLUDED.assignee_id,
			team_id=EXCLUDED.team_id,
			status=EXCLUDED.status,
			priority=EXCLUDED.priority,
			category=EXCLUDED.category,
			subject=EXCLUDED.subject,
			description=EXCLUDED.description,
			idempotency_key=EXCLUDED.idempotency_key,
			merged_into_id=EXCLUDED.merged_into_id,
			first_response_due=EXCLUDED.first_response_due,
			resolve_due=EXCLUDED.resolve_due,
			first_responded_at=EXCLUDED.first_responded_at,
			resolved_at=EXCLUDED.resolved_at,
			closed_at=EXCLUDED.closed_at,
			sla_breached=EXCLUDED.sla_breached,
			tags=EXCLUDED.tags,
			updated_at=EXCLUDED.updated_at`,
		t.ID, t.TenantID, t.CustomerID, nullUUID(t.AssigneeID), nullUUID(t.TeamID),
		t.Status, t.Priority, t.Category, t.Subject, t.Description,
		nullString(t.IdempotencyKey), nullUUID(t.MergedIntoID),
		nullTime(t.FirstResponseDue), nullTime(t.ResolveDue), nullTime(t.FirstRespondedAt),
		nullTime(t.ResolvedAt), nullTime(t.ClosedAt), t.SLABreached, TextArray(t.Tags),
		t.CreatedAt.UTC(), t.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *TicketRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Ticket, error) {
	return r.scanTicket(r.DB.QueryRowContext(ctx, ticketSelect+`
		WHERE id=$1 AND tenant_id=$2`, id, tenantID))
}

func (r *TicketRepo) GetByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.Ticket, error) {
	return r.scanTicket(r.DB.QueryRowContext(ctx, ticketSelect+`
		WHERE tenant_id=$1 AND idempotency_key=$2`, tenantID, key))
}

func (r *TicketRepo) ListByCustomer(ctx context.Context, tenantID, customerID uuid.UUID) ([]domain.Ticket, error) {
	rows, err := r.DB.QueryContext(ctx, ticketSelect+`
		WHERE tenant_id=$1 AND customer_id=$2 ORDER BY created_at DESC`, tenantID, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTickets(rows)
}

func (r *TicketRepo) AddEvent(ctx context.Context, e domain.TicketEvent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO ticket_events (id, tenant_id, ticket_id, actor_id, event_type, payload, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		e.ID, e.TenantID, e.TicketID, nullUUID(e.ActorID), e.Type, JSONMap(e.Payload), e.CreatedAt.UTC())
	return err
}

func (r *TicketRepo) ListEvents(ctx context.Context, tenantID, ticketID uuid.UUID) ([]domain.TicketEvent, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, ticket_id, actor_id, event_type, payload, created_at
		FROM ticket_events WHERE tenant_id=$1 AND ticket_id=$2 ORDER BY created_at ASC`,
		tenantID, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.TicketEvent{}
	for rows.Next() {
		var e domain.TicketEvent
		var actor uuid.NullUUID
		var payload JSONMap
		if err := rows.Scan(&e.ID, &e.TenantID, &e.TicketID, &actor, &e.Type, &payload, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.ActorID = scanNullUUID(actor)
		e.Payload = map[string]any(payload)
		e.CreatedAt = e.CreatedAt.UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *TicketRepo) AddNote(ctx context.Context, n domain.TicketNote) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO ticket_notes (id, tenant_id, ticket_id, author_id, body, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		n.ID, n.TenantID, n.TicketID, n.AuthorID, n.Body, n.CreatedAt.UTC())
	return err
}

func (r *TicketRepo) ListNotes(ctx context.Context, tenantID, ticketID uuid.UUID) ([]domain.TicketNote, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, ticket_id, author_id, body, created_at
		FROM ticket_notes WHERE tenant_id=$1 AND ticket_id=$2 ORDER BY created_at ASC`,
		tenantID, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.TicketNote{}
	for rows.Next() {
		var n domain.TicketNote
		if err := rows.Scan(&n.ID, &n.TenantID, &n.TicketID, &n.AuthorID, &n.Body, &n.CreatedAt); err != nil {
			return nil, err
		}
		n.CreatedAt = n.CreatedAt.UTC()
		out = append(out, n)
	}
	return out, rows.Err()
}

func (r *TicketRepo) ListOpen(ctx context.Context, tenantID uuid.UUID) ([]domain.Ticket, error) {
	rows, err := r.DB.QueryContext(ctx, ticketSelect+`
		WHERE tenant_id=$1 AND status <> $2 ORDER BY created_at DESC`,
		tenantID, domain.TicketStatusClosed)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTickets(rows)
}

type scannable interface {
	Scan(dest ...any) error
}

func (r *TicketRepo) scanTicket(row scannable) (domain.Ticket, error) {
	var t domain.Ticket
	var assignee, team, merged uuid.NullUUID
	var frDue, resolveDue, firstResp, resolved, closed sql.NullTime
	var tags TextArray
	err := row.Scan(
		&t.ID, &t.TenantID, &t.CustomerID, &assignee, &team, &t.Status, &t.Priority, &t.Category,
		&t.Subject, &t.Description, &t.IdempotencyKey, &merged,
		&frDue, &resolveDue, &firstResp, &resolved, &closed,
		&t.SLABreached, &tags, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return domain.Ticket{}, mapNotFound(err)
	}
	t.AssigneeID = scanNullUUID(assignee)
	t.TeamID = scanNullUUID(team)
	t.MergedIntoID = scanNullUUID(merged)
	t.FirstResponseDue = scanNullTime(frDue)
	t.ResolveDue = scanNullTime(resolveDue)
	t.FirstRespondedAt = scanNullTime(firstResp)
	t.ResolvedAt = scanNullTime(resolved)
	t.ClosedAt = scanNullTime(closed)
	t.Tags = []string(tags)
	t.CreatedAt = t.CreatedAt.UTC()
	t.UpdatedAt = t.UpdatedAt.UTC()
	return t, nil
}

func scanTickets(rows *sql.Rows) ([]domain.Ticket, error) {
	out := []domain.Ticket{}
	for rows.Next() {
		var t domain.Ticket
		var assignee, team, merged uuid.NullUUID
		var frDue, resolveDue, firstResp, resolved, closed sql.NullTime
		var tags TextArray
		if err := rows.Scan(
			&t.ID, &t.TenantID, &t.CustomerID, &assignee, &team, &t.Status, &t.Priority, &t.Category,
			&t.Subject, &t.Description, &t.IdempotencyKey, &merged,
			&frDue, &resolveDue, &firstResp, &resolved, &closed,
			&t.SLABreached, &tags, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		t.AssigneeID = scanNullUUID(assignee)
		t.TeamID = scanNullUUID(team)
		t.MergedIntoID = scanNullUUID(merged)
		t.FirstResponseDue = scanNullTime(frDue)
		t.ResolveDue = scanNullTime(resolveDue)
		t.FirstRespondedAt = scanNullTime(firstResp)
		t.ResolvedAt = scanNullTime(resolved)
		t.ClosedAt = scanNullTime(closed)
		t.Tags = []string(tags)
		t.CreatedAt = t.CreatedAt.UTC()
		t.UpdatedAt = t.UpdatedAt.UTC()
		out = append(out, t)
	}
	return out, rows.Err()
}

var _ ports.TicketRepo = (*TicketRepo)(nil)
