package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/app/ports"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// TagRepo persists tag definitions and profile assignments.
type TagRepo struct{ DB *sql.DB }

var _ ports.TagRepository = (*TagRepo)(nil)

func (r *TagRepo) UpsertTag(ctx context.Context, t domain.Tag) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO tags (id, tenant_id, kind, name, description, color, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET
			kind=EXCLUDED.kind, name=EXCLUDED.name, description=EXCLUDED.description,
			color=EXCLUDED.color, updated_at=EXCLUDED.updated_at`,
		t.ID, t.TenantID, string(t.Kind), t.Name, t.Description, t.Color, t.CreatedAt, t.UpdatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *TagRepo) GetTag(ctx context.Context, id uuid.UUID) (domain.Tag, error) {
	var t domain.Tag
	var kind string
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, kind, name, description, color, created_at, updated_at
		FROM tags WHERE id=$1`, id).Scan(
		&t.ID, &t.TenantID, &kind, &t.Name, &t.Description, &t.Color, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return domain.Tag{}, mapNotFound(err)
	}
	t.Kind = domain.TagKind(kind)
	t.CreatedAt = t.CreatedAt.UTC()
	t.UpdatedAt = t.UpdatedAt.UTC()
	return t, nil
}

func (r *TagRepo) Add(ctx context.Context, pt domain.ProfileTag) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO profile_tags (profile_id, tag_id, assigned_by, assigned_at, note)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (profile_id, tag_id) DO UPDATE SET
			assigned_by=EXCLUDED.assigned_by, assigned_at=EXCLUDED.assigned_at, note=EXCLUDED.note`,
		pt.ProfileID, pt.TagID, nullUUID(pt.AssignedBy), pt.AssignedAt, pt.Note,
	)
	return mapUniqueViolation(err)
}

func (r *TagRepo) Remove(ctx context.Context, profileID, tagID uuid.UUID) error {
	res, err := r.DB.ExecContext(ctx, `
		DELETE FROM profile_tags WHERE profile_id=$1 AND tag_id=$2`, profileID, tagID)
	return rowsAffectedOrNotFound(res, err)
}

func (r *TagRepo) List(ctx context.Context, profileID uuid.UUID) ([]domain.ProfileTag, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT profile_id, tag_id, assigned_by, assigned_at, note
		FROM profile_tags WHERE profile_id=$1 ORDER BY assigned_at DESC`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.ProfileTag, 0)
	for rows.Next() {
		var pt domain.ProfileTag
		var assignedBy uuid.NullUUID
		if err := rows.Scan(&pt.ProfileID, &pt.TagID, &assignedBy, &pt.AssignedAt, &pt.Note); err != nil {
			return nil, err
		}
		pt.AssignedBy = scanNullUUID(assignedBy)
		pt.AssignedAt = pt.AssignedAt.UTC()
		out = append(out, pt)
	}
	return out, rows.Err()
}
