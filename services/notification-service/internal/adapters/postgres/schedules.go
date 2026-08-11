package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/notification-service/internal/app/ports"
	"github.com/nexora/notification-service/internal/domain"
)

// ScheduleRepo persists delayed send schedules.
type ScheduleRepo struct{ DB *sql.DB }

func (r *ScheduleRepo) Create(ctx context.Context, s domain.Schedule) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO schedules (
			id, tenant_id, principal_id, channel, priority, template_key, locale, recipient,
			subject, body, vars_json, idempotency_key, send_at, status, message_id, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		s.ID, s.TenantID, s.PrincipalID, string(s.Channel), string(s.Priority), s.TemplateKey, s.Locale,
		s.Recipient, s.Subject, s.Body, JSONStringMap(s.Vars), s.IdempotencyKey, s.SendAt.UTC(),
		string(s.Status), nullUUID(s.MessageID), s.CreatedAt.UTC(), s.UpdatedAt.UTC())
	return err
}

func (r *ScheduleRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Schedule, error) {
	s, err := scanSchedule(r.DB.QueryRowContext(ctx, scheduleSelect+` WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if isNoRows(err) {
		return domain.Schedule{}, domain.ErrNotFound
	}
	return s, err
}

func (r *ScheduleRepo) Update(ctx context.Context, s domain.Schedule) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE schedules SET
			principal_id=$3, channel=$4, priority=$5, template_key=$6, locale=$7, recipient=$8,
			subject=$9, body=$10, vars_json=$11, idempotency_key=$12, send_at=$13, status=$14,
			message_id=$15, updated_at=$16
		WHERE id=$1 AND tenant_id=$2`,
		s.ID, s.TenantID, s.PrincipalID, string(s.Channel), string(s.Priority), s.TemplateKey, s.Locale,
		s.Recipient, s.Subject, s.Body, JSONStringMap(s.Vars), s.IdempotencyKey, s.SendAt.UTC(),
		string(s.Status), nullUUID(s.MessageID), s.UpdatedAt.UTC())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ScheduleRepo) ListDue(ctx context.Context, now time.Time, limit int) ([]domain.Schedule, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.DB.QueryContext(ctx, scheduleSelect+`
		WHERE status=$1 AND send_at <= $2 ORDER BY send_at ASC LIMIT $3`,
		string(domain.SchedulePending), now.UTC(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Schedule{}
	for rows.Next() {
		s, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *ScheduleRepo) Cancel(ctx context.Context, tenantID, id uuid.UUID, now time.Time) (domain.Schedule, error) {
	s, err := r.Get(ctx, tenantID, id)
	if err != nil {
		return domain.Schedule{}, err
	}
	if s.Status != domain.SchedulePending {
		return domain.Schedule{}, domain.ErrConflict
	}
	s.Status = domain.ScheduleCancelled
	s.UpdatedAt = now.UTC()
	if err := r.Update(ctx, s); err != nil {
		return domain.Schedule{}, err
	}
	return s, nil
}

const scheduleSelect = `
	SELECT id, tenant_id, principal_id, channel, priority, template_key, locale, recipient,
		subject, body, vars_json, idempotency_key, send_at, status, message_id, created_at, updated_at
	FROM schedules`

func scanSchedule(row scanner) (domain.Schedule, error) {
	var s domain.Schedule
	var channel, priority, status string
	var vars JSONStringMap
	var messageID uuid.NullUUID
	err := row.Scan(
		&s.ID, &s.TenantID, &s.PrincipalID, &channel, &priority, &s.TemplateKey, &s.Locale, &s.Recipient,
		&s.Subject, &s.Body, &vars, &s.IdempotencyKey, &s.SendAt, &status, &messageID,
		&s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return domain.Schedule{}, err
	}
	s.Channel = domain.Channel(channel)
	s.Priority = domain.Priority(priority)
	s.Status = domain.ScheduleStatus(status)
	s.Vars = map[string]string(vars)
	s.MessageID = scanNullUUID(messageID)
	s.SendAt = s.SendAt.UTC()
	s.CreatedAt = s.CreatedAt.UTC()
	s.UpdatedAt = s.UpdatedAt.UTC()
	return s, nil
}

var _ ports.ScheduleRepo = (*ScheduleRepo)(nil)
