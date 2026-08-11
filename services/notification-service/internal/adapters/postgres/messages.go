package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/notification-service/internal/app/ports"
	"github.com/nexora/notification-service/internal/domain"
)

// MessageRepo persists notification messages.
type MessageRepo struct{ DB *sql.DB }

func (r *MessageRepo) Create(ctx context.Context, m domain.Message) error {
	if m.IdempotencyKey != "" {
		if _, err := r.GetByIdempotency(ctx, m.TenantID, m.IdempotencyKey); err == nil {
			return domain.ErrIdempotencyConflict
		} else if err != domain.ErrNotFound {
			return err
		}
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO messages (
			id, tenant_id, principal_id, order_id, channel, priority, template_key, template_id,
			locale, subject, body, recipient, status, idempotency_key, vars_json, suppress_reason,
			attempts, max_attempts, last_error, provider, provider_ref, created_at, updated_at, sent_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24
		)`,
		m.ID, m.TenantID, m.PrincipalID, nullUUID(m.OrderID), string(m.Channel), string(m.Priority),
		m.TemplateKey, nullUUID(m.TemplateID), m.Locale, m.Subject, m.Body, m.Recipient, string(m.Status),
		m.IdempotencyKey, JSONStringMap(m.Vars), m.SuppressReason, m.Attempts, m.MaxAttempts,
		m.LastError, m.Provider, m.ProviderRef, m.CreatedAt.UTC(), m.UpdatedAt.UTC(), nullTime(m.SentAt))
	return mapUniqueViolation(err)
}

func (r *MessageRepo) Update(ctx context.Context, m domain.Message) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE messages SET
			principal_id=$3, order_id=$4, channel=$5, priority=$6, template_key=$7, template_id=$8,
			locale=$9, subject=$10, body=$11, recipient=$12, status=$13, idempotency_key=$14, vars_json=$15,
			suppress_reason=$16, attempts=$17, max_attempts=$18, last_error=$19, provider=$20, provider_ref=$21,
			updated_at=$22, sent_at=$23
		WHERE id=$1 AND tenant_id=$2`,
		m.ID, m.TenantID, m.PrincipalID, nullUUID(m.OrderID), string(m.Channel), string(m.Priority),
		m.TemplateKey, nullUUID(m.TemplateID), m.Locale, m.Subject, m.Body, m.Recipient, string(m.Status),
		m.IdempotencyKey, JSONStringMap(m.Vars), m.SuppressReason, m.Attempts, m.MaxAttempts,
		m.LastError, m.Provider, m.ProviderRef, m.UpdatedAt.UTC(), nullTime(m.SentAt))
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *MessageRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Message, error) {
	m, err := r.scan(r.DB.QueryRowContext(ctx, messageSelect+` WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if isNoRows(err) {
		return domain.Message{}, domain.ErrNotFound
	}
	return m, err
}

func (r *MessageRepo) GetByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.Message, error) {
	m, err := r.scan(r.DB.QueryRowContext(ctx, messageSelect+`
		WHERE tenant_id=$1 AND idempotency_key=$2 AND idempotency_key <> ''`, tenantID, key))
	if isNoRows(err) {
		return domain.Message{}, domain.ErrNotFound
	}
	return m, err
}

func (r *MessageRepo) ListFailed(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.Message, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.DB.QueryContext(ctx, messageSelect+`
		WHERE tenant_id=$1 AND status=$2 ORDER BY updated_at DESC LIMIT $3`,
		tenantID, string(domain.MessageFailed), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMessages(rows)
}

func (r *MessageRepo) CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[domain.MessageStatus]int, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT status, COUNT(*) FROM messages WHERE tenant_id=$1 GROUP BY status`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[domain.MessageStatus]int{}
	for rows.Next() {
		var status string
		var n int
		if err := rows.Scan(&status, &n); err != nil {
			return nil, err
		}
		out[domain.MessageStatus(status)] = n
	}
	return out, rows.Err()
}

const messageSelect = `
	SELECT id, tenant_id, principal_id, order_id, channel, priority, template_key, template_id,
		locale, subject, body, recipient, status, idempotency_key, vars_json, suppress_reason,
		attempts, max_attempts, last_error, provider, provider_ref, created_at, updated_at, sent_at
	FROM messages`

func (r *MessageRepo) scan(row *sql.Row) (domain.Message, error) {
	return scanMessage(row)
}

func scanMessage(row scanner) (domain.Message, error) {
	var m domain.Message
	var channel, priority, status string
	var orderID, templateID uuid.NullUUID
	var vars JSONStringMap
	var sentAt sql.NullTime
	err := row.Scan(
		&m.ID, &m.TenantID, &m.PrincipalID, &orderID, &channel, &priority, &m.TemplateKey, &templateID,
		&m.Locale, &m.Subject, &m.Body, &m.Recipient, &status, &m.IdempotencyKey, &vars, &m.SuppressReason,
		&m.Attempts, &m.MaxAttempts, &m.LastError, &m.Provider, &m.ProviderRef,
		&m.CreatedAt, &m.UpdatedAt, &sentAt)
	if err != nil {
		return domain.Message{}, err
	}
	m.Channel = domain.Channel(channel)
	m.Priority = domain.Priority(priority)
	m.Status = domain.MessageStatus(status)
	m.OrderID = scanNullUUID(orderID)
	m.TemplateID = scanNullUUID(templateID)
	m.Vars = map[string]string(vars)
	m.SentAt = scanNullTime(sentAt)
	m.CreatedAt = m.CreatedAt.UTC()
	m.UpdatedAt = m.UpdatedAt.UTC()
	return m, nil
}

func scanMessages(rows *sql.Rows) ([]domain.Message, error) {
	out := []domain.Message{}
	for rows.Next() {
		m, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

var _ ports.MessageRepo = (*MessageRepo)(nil)
