package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/notification-service/internal/app/ports"
	"github.com/nexora/notification-service/internal/domain"
)

// InboxRepo persists in-app inbox items.
type InboxRepo struct{ DB *sql.DB }

func (r *InboxRepo) Create(ctx context.Context, item domain.InboxItem) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO inbox_items (
			id, tenant_id, principal_id, message_id, title, body, "read", archived, created_at, read_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		item.ID, item.TenantID, item.PrincipalID, item.MessageID, item.Title, item.Body,
		item.Read, item.Archived, item.CreatedAt.UTC(), nullTime(item.ReadAt))
	return err
}

func (r *InboxRepo) List(ctx context.Context, tenantID, principalID uuid.UUID, includeArchived bool) ([]domain.InboxItem, error) {
	q := `
		SELECT id, tenant_id, principal_id, message_id, title, body, "read", archived, created_at, read_at
		FROM inbox_items WHERE tenant_id=$1 AND principal_id=$2`
	if !includeArchived {
		q += ` AND archived=false`
	}
	q += ` ORDER BY created_at DESC`
	rows, err := r.DB.QueryContext(ctx, q, tenantID, principalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.InboxItem{}
	for rows.Next() {
		it, err := scanInbox(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *InboxRepo) MarkRead(ctx context.Context, tenantID, id uuid.UUID, now time.Time) (domain.InboxItem, error) {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE inbox_items SET "read"=true, read_at=$3 WHERE id=$1 AND tenant_id=$2`, id, tenantID, now.UTC())
	if err != nil {
		return domain.InboxItem{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.InboxItem{}, domain.ErrNotFound
	}
	return r.get(ctx, tenantID, id)
}

func (r *InboxRepo) Archive(ctx context.Context, tenantID, id uuid.UUID) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE inbox_items SET archived=true WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *InboxRepo) get(ctx context.Context, tenantID, id uuid.UUID) (domain.InboxItem, error) {
	it, err := scanInbox(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, principal_id, message_id, title, body, "read", archived, created_at, read_at
		FROM inbox_items WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if isNoRows(err) {
		return domain.InboxItem{}, domain.ErrNotFound
	}
	return it, err
}

func scanInbox(row scanner) (domain.InboxItem, error) {
	var it domain.InboxItem
	var readAt sql.NullTime
	err := row.Scan(
		&it.ID, &it.TenantID, &it.PrincipalID, &it.MessageID, &it.Title, &it.Body,
		&it.Read, &it.Archived, &it.CreatedAt, &readAt)
	if err != nil {
		return domain.InboxItem{}, err
	}
	it.ReadAt = scanNullTime(readAt)
	it.CreatedAt = it.CreatedAt.UTC()
	return it, nil
}

var _ ports.InboxRepo = (*InboxRepo)(nil)
