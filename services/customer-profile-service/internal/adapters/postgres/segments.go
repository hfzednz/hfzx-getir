package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/app/ports"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// SegmentRepo persists segments and memberships.
type SegmentRepo struct{ DB *sql.DB }

var _ ports.SegmentRepository = (*SegmentRepo)(nil)

func (r *SegmentRepo) GetSegment(ctx context.Context, id uuid.UUID) (domain.Segment, error) {
	return r.scanSegment(r.DB.QueryRowContext(ctx, segmentSelect+` WHERE id=$1`, id))
}

func (r *SegmentRepo) ListSegments(ctx context.Context, tenantID uuid.UUID) ([]domain.Segment, error) {
	rows, err := r.DB.QueryContext(ctx, segmentSelect+` WHERE tenant_id=$1 ORDER BY name`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Segment, 0)
	for rows.Next() {
		s, err := scanSegmentRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *SegmentRepo) UpsertSegment(ctx context.Context, s domain.Segment) error {
	kind := s.Kind
	if kind == "" {
		kind = domain.SegmentKindDynamic
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO segments (id, tenant_id, name, kind, description, rules, active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			name=EXCLUDED.name, kind=EXCLUDED.kind, description=EXCLUDED.description,
			rules=EXCLUDED.rules, active=EXCLUDED.active, updated_at=EXCLUDED.updated_at`,
		s.ID, s.TenantID, s.Name, string(kind), s.Description, JSONMap(metaGetMap(s.Rules)), s.Active, s.CreatedAt, s.UpdatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *SegmentRepo) Assign(ctx context.Context, m domain.SegmentMembership) error {
	source := m.Source
	if source == "" {
		source = "rules"
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO segment_members (segment_id, profile_id, joined_at, expires_at, source)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (segment_id, profile_id) DO UPDATE SET
			joined_at=EXCLUDED.joined_at, expires_at=EXCLUDED.expires_at, source=EXCLUDED.source`,
		m.SegmentID, m.ProfileID, m.JoinedAt, nullTime(m.ExpiresAt), source,
	)
	return err
}

func (r *SegmentRepo) RemoveMembership(ctx context.Context, segmentID, profileID uuid.UUID) error {
	res, err := r.DB.ExecContext(ctx, `
		DELETE FROM segment_members WHERE segment_id=$1 AND profile_id=$2`, segmentID, profileID)
	return rowsAffectedOrNotFound(res, err)
}

func (r *SegmentRepo) ListMembers(ctx context.Context, segmentID uuid.UUID) ([]domain.SegmentMembership, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT segment_id, profile_id, joined_at, expires_at, source
		FROM segment_members WHERE segment_id=$1 ORDER BY joined_at`, segmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMemberships(rows)
}

func (r *SegmentRepo) ListByProfile(ctx context.Context, profileID uuid.UUID) ([]domain.SegmentMembership, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT segment_id, profile_id, joined_at, expires_at, source
		FROM segment_members WHERE profile_id=$1 ORDER BY joined_at`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMemberships(rows)
}

const segmentSelect = `
	SELECT id, tenant_id, name, kind, description, rules, active, created_at, updated_at
	FROM segments`

func (r *SegmentRepo) scanSegment(row scannable) (domain.Segment, error) {
	s, err := scanSegmentRow(row)
	if err != nil {
		return domain.Segment{}, mapNotFound(err)
	}
	return s, nil
}

func scanSegmentRow(row scannable) (domain.Segment, error) {
	var s domain.Segment
	var kind string
	var rules JSONMap
	err := row.Scan(&s.ID, &s.TenantID, &s.Name, &kind, &s.Description, &rules, &s.Active, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return domain.Segment{}, err
	}
	s.Kind = domain.SegmentKind(kind)
	s.Rules = map[string]any(rules)
	s.CreatedAt = s.CreatedAt.UTC()
	s.UpdatedAt = s.UpdatedAt.UTC()
	return s, nil
}

func scanMemberships(rows *sql.Rows) ([]domain.SegmentMembership, error) {
	out := make([]domain.SegmentMembership, 0)
	for rows.Next() {
		var m domain.SegmentMembership
		var expires sql.NullTime
		if err := rows.Scan(&m.SegmentID, &m.ProfileID, &m.JoinedAt, &expires, &m.Source); err != nil {
			return nil, err
		}
		m.ExpiresAt = scanNullTime(expires)
		m.JoinedAt = m.JoinedAt.UTC()
		out = append(out, m)
	}
	return out, rows.Err()
}
