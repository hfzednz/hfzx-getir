package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/crm-service/internal/app/ports"
	"github.com/nexora/crm-service/internal/domain"
)

// FeedbackRepo persists feedback and CSAT responses.
type FeedbackRepo struct{ DB *sql.DB }

func (r *FeedbackRepo) SaveFeedback(ctx context.Context, f domain.Feedback) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO feedback (id, tenant_id, customer_id, ticket_id, conversation_id, kind, score, comment, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		f.ID, f.TenantID, f.CustomerID, nullUUID(f.TicketID), nullUUID(f.ConversationID),
		f.Kind, f.Score, f.Comment, f.CreatedAt.UTC())
	return err
}

func (r *FeedbackRepo) SaveCSAT(ctx context.Context, c domain.CSATResponse) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO csat_responses (id, tenant_id, customer_id, ticket_id, conversation_id, score, comment, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		c.ID, c.TenantID, c.CustomerID, nullUUID(c.TicketID), nullUUID(c.ConversationID),
		c.Score, c.Comment, c.CreatedAt.UTC())
	return err
}

func (r *FeedbackRepo) ListCSAT(ctx context.Context, tenantID uuid.UUID) ([]domain.CSATResponse, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, customer_id, ticket_id, conversation_id, score, comment, created_at
		FROM csat_responses WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.CSATResponse{}
	for rows.Next() {
		var c domain.CSATResponse
		var ticket, conv uuid.NullUUID
		if err := rows.Scan(&c.ID, &c.TenantID, &c.CustomerID, &ticket, &conv, &c.Score, &c.Comment, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.TicketID = scanNullUUID(ticket)
		c.ConversationID = scanNullUUID(conv)
		c.CreatedAt = c.CreatedAt.UTC()
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *FeedbackRepo) ListFeedback(ctx context.Context, tenantID uuid.UUID) ([]domain.Feedback, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, customer_id, ticket_id, conversation_id, kind, score, comment, created_at
		FROM feedback WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Feedback{}
	for rows.Next() {
		var f domain.Feedback
		var ticket, conv uuid.NullUUID
		if err := rows.Scan(&f.ID, &f.TenantID, &f.CustomerID, &ticket, &conv, &f.Kind, &f.Score, &f.Comment, &f.CreatedAt); err != nil {
			return nil, err
		}
		f.TicketID = scanNullUUID(ticket)
		f.ConversationID = scanNullUUID(conv)
		f.CreatedAt = f.CreatedAt.UTC()
		out = append(out, f)
	}
	return out, rows.Err()
}

var _ ports.FeedbackRepo = (*FeedbackRepo)(nil)
