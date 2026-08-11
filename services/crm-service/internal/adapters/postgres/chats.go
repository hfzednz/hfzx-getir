package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/crm-service/internal/app/ports"
	"github.com/nexora/crm-service/internal/domain"
)

// ChatRepo persists conversations and messages.
type ChatRepo struct{ DB *sql.DB }

const conversationSelect = `
	SELECT id, tenant_id, customer_id, agent_id, ticket_id, status, channel,
		transferred_from, started_at, ended_at, created_at, updated_at
	FROM conversations`

func (r *ChatRepo) SaveConversation(ctx context.Context, c domain.Conversation) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO conversations (
			id, tenant_id, customer_id, agent_id, ticket_id, status, channel,
			transferred_from, started_at, ended_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (id) DO UPDATE SET
			customer_id=EXCLUDED.customer_id,
			agent_id=EXCLUDED.agent_id,
			ticket_id=EXCLUDED.ticket_id,
			status=EXCLUDED.status,
			channel=EXCLUDED.channel,
			transferred_from=EXCLUDED.transferred_from,
			started_at=EXCLUDED.started_at,
			ended_at=EXCLUDED.ended_at,
			updated_at=EXCLUDED.updated_at`,
		c.ID, c.TenantID, c.CustomerID, nullUUID(c.AgentID), nullUUID(c.TicketID),
		c.Status, c.Channel, nullUUID(c.TransferredFrom), c.StartedAt.UTC(),
		nullTime(c.EndedAt), c.CreatedAt.UTC(), c.UpdatedAt.UTC())
	return err
}

func (r *ChatRepo) GetConversation(ctx context.Context, tenantID, id uuid.UUID) (domain.Conversation, error) {
	row := r.DB.QueryRowContext(ctx, conversationSelect+` WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var c domain.Conversation
	var agent, ticket, transferred uuid.NullUUID
	var ended sql.NullTime
	err := row.Scan(
		&c.ID, &c.TenantID, &c.CustomerID, &agent, &ticket, &c.Status, &c.Channel,
		&transferred, &c.StartedAt, &ended, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return domain.Conversation{}, mapNotFound(err)
	}
	c.AgentID = scanNullUUID(agent)
	c.TicketID = scanNullUUID(ticket)
	c.TransferredFrom = scanNullUUID(transferred)
	c.EndedAt = scanNullTime(ended)
	c.StartedAt = c.StartedAt.UTC()
	c.CreatedAt = c.CreatedAt.UTC()
	c.UpdatedAt = c.UpdatedAt.UTC()
	return c, nil
}

func (r *ChatRepo) AddMessage(ctx context.Context, m domain.Message) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO messages (id, tenant_id, conversation_id, sender_role, sender_id, body, sentiment, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		m.ID, m.TenantID, m.ConversationID, m.SenderRole, nullUUID(m.SenderID),
		m.Body, m.Sentiment, m.CreatedAt.UTC())
	return err
}

func (r *ChatRepo) ListMessages(ctx context.Context, tenantID, conversationID uuid.UUID) ([]domain.Message, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, conversation_id, sender_role, sender_id, body, sentiment, created_at
		FROM messages WHERE tenant_id=$1 AND conversation_id=$2 ORDER BY created_at ASC`,
		tenantID, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Message{}
	for rows.Next() {
		var m domain.Message
		var sender uuid.NullUUID
		if err := rows.Scan(&m.ID, &m.TenantID, &m.ConversationID, &m.SenderRole, &sender, &m.Body, &m.Sentiment, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.SenderID = scanNullUUID(sender)
		m.CreatedAt = m.CreatedAt.UTC()
		out = append(out, m)
	}
	return out, rows.Err()
}

var _ ports.ChatRepo = (*ChatRepo)(nil)
