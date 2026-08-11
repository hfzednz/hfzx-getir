package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/crm-service/internal/app/ports"
	"github.com/nexora/crm-service/internal/domain"
)

// CaseRepo persists investigation cases.
type CaseRepo struct{ DB *sql.DB }

const caseSelect = `
	SELECT id, tenant_id, customer_id, ticket_id, case_type, status, title, details, assignee_id, created_at, updated_at
	FROM cases`

func (r *CaseRepo) Save(ctx context.Context, c domain.Case) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO cases (
			id, tenant_id, customer_id, ticket_id, case_type, status, title, details, assignee_id, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET
			customer_id=EXCLUDED.customer_id,
			ticket_id=EXCLUDED.ticket_id,
			case_type=EXCLUDED.case_type,
			status=EXCLUDED.status,
			title=EXCLUDED.title,
			details=EXCLUDED.details,
			assignee_id=EXCLUDED.assignee_id,
			updated_at=EXCLUDED.updated_at`,
		c.ID, c.TenantID, c.CustomerID, nullUUID(c.TicketID), c.Type, c.Status,
		c.Title, c.Details, nullUUID(c.AssigneeID), c.CreatedAt.UTC(), c.UpdatedAt.UTC())
	return err
}

func (r *CaseRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Case, error) {
	row := r.DB.QueryRowContext(ctx, caseSelect+` WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	return scanCase(row)
}

func (r *CaseRepo) ListByCustomer(ctx context.Context, tenantID, customerID uuid.UUID) ([]domain.Case, error) {
	rows, err := r.DB.QueryContext(ctx, caseSelect+`
		WHERE tenant_id=$1 AND customer_id=$2 ORDER BY created_at DESC`, tenantID, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Case{}
	for rows.Next() {
		c, err := scanCase(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func scanCase(row scannable) (domain.Case, error) {
	var c domain.Case
	var ticket, assignee uuid.NullUUID
	err := row.Scan(
		&c.ID, &c.TenantID, &c.CustomerID, &ticket, &c.Type, &c.Status,
		&c.Title, &c.Details, &assignee, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return domain.Case{}, mapNotFound(err)
	}
	c.TicketID = scanNullUUID(ticket)
	c.AssigneeID = scanNullUUID(assignee)
	c.CreatedAt = c.CreatedAt.UTC()
	c.UpdatedAt = c.UpdatedAt.UTC()
	return c, nil
}

var _ ports.CaseRepo = (*CaseRepo)(nil)
