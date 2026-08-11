package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/app/ports"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// CRMRepo persists CRM notes and timeline.
type CRMRepo struct{ DB *sql.DB }

var _ ports.CRMRepository = (*CRMRepo)(nil)

func (r *CRMRepo) AddNote(ctx context.Context, n domain.CRMNote) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO crm_notes (
			id, profile_id, tenant_id, author_id, body, pinned, created_at, updated_at, deleted_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		n.ID, n.ProfileID, n.TenantID, n.AuthorID, n.Body, n.Pinned, n.CreatedAt, n.UpdatedAt, nullTime(n.DeletedAt),
	)
	return mapUniqueViolation(err)
}

func (r *CRMRepo) ListNotes(ctx context.Context, profileID uuid.UUID, limit int) ([]domain.CRMNote, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, profile_id, tenant_id, author_id, body, pinned, created_at, updated_at, deleted_at
		FROM crm_notes
		WHERE profile_id=$1 AND deleted_at IS NULL
		ORDER BY pinned DESC, created_at DESC
		LIMIT $2`, profileID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.CRMNote, 0)
	for rows.Next() {
		var n domain.CRMNote
		var deleted sql.NullTime
		if err := rows.Scan(&n.ID, &n.ProfileID, &n.TenantID, &n.AuthorID, &n.Body, &n.Pinned, &n.CreatedAt, &n.UpdatedAt, &deleted); err != nil {
			return nil, err
		}
		n.DeletedAt = scanNullTime(deleted)
		n.CreatedAt = n.CreatedAt.UTC()
		n.UpdatedAt = n.UpdatedAt.UTC()
		out = append(out, n)
	}
	return out, rows.Err()
}

func (r *CRMRepo) AppendTimeline(ctx context.Context, e domain.TimelineEvent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO timeline_events (
			id, profile_id, tenant_id, type, payload, actor_id, occurred_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		e.ID, e.ProfileID, e.TenantID, e.Type, JSONMap(metaGetMap(e.Payload)), nullUUID(e.ActorID), e.OccurredAt, e.CreatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *CRMRepo) ListTimeline(ctx context.Context, profileID uuid.UUID, limit int) ([]domain.TimelineEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, profile_id, tenant_id, type, payload, actor_id, occurred_at, created_at
		FROM timeline_events
		WHERE profile_id=$1
		ORDER BY occurred_at DESC
		LIMIT $2`, profileID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.TimelineEvent, 0)
	for rows.Next() {
		var e domain.TimelineEvent
		var payload JSONMap
		var actorID uuid.NullUUID
		if err := rows.Scan(&e.ID, &e.ProfileID, &e.TenantID, &e.Type, &payload, &actorID, &e.OccurredAt, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.Payload = map[string]any(payload)
		e.ActorID = scanNullUUID(actorID)
		e.OccurredAt = e.OccurredAt.UTC()
		e.CreatedAt = e.CreatedAt.UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}
