package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/app/ports"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// PrivacyRepo persists privacy requests.
type PrivacyRepo struct{ DB *sql.DB }

var _ ports.PrivacyRepository = (*PrivacyRepo)(nil)

func (r *PrivacyRepo) Create(ctx context.Context, req domain.PrivacyRequest) error {
	status := req.Status
	if status == "" {
		status = domain.PrivacyStatusPending
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO privacy_requests (
			id, profile_id, tenant_id, kind, status, payload_ref, requested_by, reason, error_message,
			completed_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		req.ID, req.ProfileID, req.TenantID, string(req.Kind), string(status), req.PayloadRef,
		nullUUID(req.RequestedBy), req.Reason, req.ErrorMessage, nullTime(req.CompletedAt), req.CreatedAt, req.UpdatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *PrivacyRepo) Update(ctx context.Context, req domain.PrivacyRequest) error {
	status := req.Status
	if status == "" {
		status = domain.PrivacyStatusPending
	}
	res, err := r.DB.ExecContext(ctx, `
		UPDATE privacy_requests SET
			kind=$1, status=$2, payload_ref=$3, requested_by=$4, reason=$5, error_message=$6,
			completed_at=$7, updated_at=$8
		WHERE id=$9`,
		string(req.Kind), string(status), req.PayloadRef, nullUUID(req.RequestedBy), req.Reason, req.ErrorMessage,
		nullTime(req.CompletedAt), req.UpdatedAt, req.ID,
	)
	return rowsAffectedOrNotFound(res, err)
}

func (r *PrivacyRepo) Get(ctx context.Context, id uuid.UUID) (domain.PrivacyRequest, error) {
	return r.scanPrivacy(r.DB.QueryRowContext(ctx, privacySelect+` WHERE id=$1`, id))
}

func (r *PrivacyRepo) ListPending(ctx context.Context, limit int) ([]domain.PrivacyRequest, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.DB.QueryContext(ctx, privacySelect+`
		WHERE status IN ('pending','processing')
		ORDER BY created_at ASC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.PrivacyRequest, 0)
	for rows.Next() {
		req, err := scanPrivacyRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, req)
	}
	return out, rows.Err()
}

const privacySelect = `
	SELECT id, profile_id, tenant_id, kind, status, payload_ref, requested_by, reason, error_message,
		completed_at, created_at, updated_at
	FROM privacy_requests`

func (r *PrivacyRepo) scanPrivacy(row scannable) (domain.PrivacyRequest, error) {
	req, err := scanPrivacyRow(row)
	if err != nil {
		return domain.PrivacyRequest{}, mapNotFound(err)
	}
	return req, nil
}

func scanPrivacyRow(row scannable) (domain.PrivacyRequest, error) {
	var req domain.PrivacyRequest
	var kind, status string
	var requestedBy uuid.NullUUID
	var completed sql.NullTime
	err := row.Scan(
		&req.ID, &req.ProfileID, &req.TenantID, &kind, &status, &req.PayloadRef, &requestedBy,
		&req.Reason, &req.ErrorMessage, &completed, &req.CreatedAt, &req.UpdatedAt,
	)
	if err != nil {
		return domain.PrivacyRequest{}, err
	}
	req.Kind = domain.PrivacyRequestKind(kind)
	req.Status = domain.PrivacyRequestStatus(status)
	req.RequestedBy = scanNullUUID(requestedBy)
	req.CompletedAt = scanNullTime(completed)
	req.CreatedAt = req.CreatedAt.UTC()
	req.UpdatedAt = req.UpdatedAt.UTC()
	return req, nil
}

// ActivityRepo persists profile-side activity.
type ActivityRepo struct{ DB *sql.DB }

var _ ports.ActivityRepository = (*ActivityRepo)(nil)

func (r *ActivityRepo) Record(ctx context.Context, e domain.ActivityEntry) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO profile_activity (
			id, profile_id, tenant_id, actor_id, action, resource_type, resource_id, payload,
			ip, user_agent, occurred_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::inet,$10,$11,$12)`,
		e.ID, e.ProfileID, e.TenantID, nullUUID(e.ActorID), e.Action, e.ResourceType, nullUUID(e.ResourceID),
		JSONMap(metaGetMap(e.Payload)), nullIP(e.IP), e.UserAgent, e.OccurredAt, e.CreatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *ActivityRepo) List(ctx context.Context, profileID uuid.UUID, limit int) ([]domain.ActivityEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, profile_id, tenant_id, actor_id, action, resource_type, resource_id, payload,
			host(ip)::text, user_agent, occurred_at, created_at
		FROM profile_activity
		WHERE profile_id=$1
		ORDER BY occurred_at DESC
		LIMIT $2`, profileID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.ActivityEntry, 0)
	for rows.Next() {
		var e domain.ActivityEntry
		var actorID, resourceID uuid.NullUUID
		var payload JSONMap
		var ip sql.NullString
		if err := rows.Scan(
			&e.ID, &e.ProfileID, &e.TenantID, &actorID, &e.Action, &e.ResourceType, &resourceID, &payload,
			&ip, &e.UserAgent, &e.OccurredAt, &e.CreatedAt,
		); err != nil {
			return nil, err
		}
		e.ActorID = scanNullUUID(actorID)
		e.ResourceID = scanNullUUID(resourceID)
		e.Payload = map[string]any(payload)
		e.IP = scanIP(ip)
		e.OccurredAt = e.OccurredAt.UTC()
		e.CreatedAt = e.CreatedAt.UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}
