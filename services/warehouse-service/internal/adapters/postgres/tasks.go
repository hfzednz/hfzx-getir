package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/app/ports"
	"github.com/nexora/warehouse-service/internal/domain"
)

// TaskRepo persists pick/pack/dispatch tasks.
type TaskRepo struct{ DB *sql.DB }

var _ ports.TaskRepo = (*TaskRepo)(nil)

func (r *TaskRepo) Create(ctx context.Context, t domain.Task) error {
	if err := ensureWarehouse(ctx, r.DB, t.TenantID, t.WarehouseID); err != nil {
		return err
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO tasks (
			id, tenant_id, warehouse_id, fulfillment_id, station_id, type, status, assignee_id,
			priority, wave_id, batch_id, sla_deadline, claimed_at, started_at, completed_at,
			cancelled_at, escalated_at, escalation_note, history, metadata, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,
			$9,$10,$11,$12,$13,$14,$15,
			$16,$17,$18,$19,$20,$21,$22
		)`,
		t.ID, t.TenantID, t.WarehouseID, nullUUIDValue(t.FulfillmentID), nullUUID(t.StationID),
		string(t.Type), string(t.Status), nullUUID(t.AssigneeID),
		t.Priority, nullUUID(t.WaveID), nullUUID(t.BatchID), nullTime(t.SLADeadline),
		nullTime(t.ClaimedAt), nullTime(t.StartedAt), nullTime(t.CompletedAt),
		nullTime(t.CancelledAt), nullTime(t.EscalatedAt), t.EscalationNote,
		JSONRaw{V: t.History}, JSONMap(metaGetMap(t.Metadata)), t.CreatedAt, t.UpdatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *TaskRepo) Update(ctx context.Context, t domain.Task) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE tasks SET
			warehouse_id=$1, fulfillment_id=$2, station_id=$3, type=$4, status=$5, assignee_id=$6,
			priority=$7, wave_id=$8, batch_id=$9, sla_deadline=$10, claimed_at=$11, started_at=$12,
			completed_at=$13, cancelled_at=$14, escalated_at=$15, escalation_note=$16,
			history=$17, metadata=$18, updated_at=$19
		WHERE id=$20 AND tenant_id=$21`,
		t.WarehouseID, nullUUIDValue(t.FulfillmentID), nullUUID(t.StationID), string(t.Type), string(t.Status), nullUUID(t.AssigneeID),
		t.Priority, nullUUID(t.WaveID), nullUUID(t.BatchID), nullTime(t.SLADeadline), nullTime(t.ClaimedAt), nullTime(t.StartedAt),
		nullTime(t.CompletedAt), nullTime(t.CancelledAt), nullTime(t.EscalatedAt), t.EscalationNote,
		JSONRaw{V: t.History}, JSONMap(metaGetMap(t.Metadata)), t.UpdatedAt,
		t.ID, t.TenantID,
	)
	return rowsAffectedOrNotFound(res, err)
}

func (r *TaskRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Task, error) {
	return r.scanTask(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, warehouse_id, fulfillment_id, station_id, type, status, assignee_id,
			priority, wave_id, batch_id, sla_deadline, claimed_at, started_at, completed_at,
			cancelled_at, escalated_at, escalation_note, history, metadata, created_at, updated_at
		FROM tasks WHERE id=$1 AND tenant_id=$2`, id, tenantID))
}

func (r *TaskRepo) List(ctx context.Context, f ports.TaskFilter) ([]domain.Task, int, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	where := []string{"tenant_id = $1", "warehouse_id = $2"}
	args := []any{f.TenantID, f.WarehouseID}
	n := 3
	if f.Type != nil {
		where = append(where, fmt.Sprintf("type = $%d", n))
		args = append(args, string(*f.Type))
		n++
	}
	if f.Status != nil {
		where = append(where, fmt.Sprintf("status = $%d", n))
		args = append(args, string(*f.Status))
		n++
	}
	if f.AssigneeID != nil {
		where = append(where, fmt.Sprintf("assignee_id = $%d", n))
		args = append(args, *f.AssigneeID)
		n++
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM tasks WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, warehouse_id, fulfillment_id, station_id, type, status, assignee_id,
			priority, wave_id, batch_id, sla_deadline, claimed_at, started_at, completed_at,
			cancelled_at, escalated_at, escalation_note, history, metadata, created_at, updated_at
		FROM tasks WHERE `+whereSQL+`
		ORDER BY priority DESC, created_at ASC
		LIMIT $`+fmt.Sprint(n)+` OFFSET $`+fmt.Sprint(n+1), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]domain.Task, 0)
	for rows.Next() {
		t, err := scanTaskRow(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, t)
	}
	return out, total, rows.Err()
}

func (r *TaskRepo) CountByStatus(ctx context.Context, tenantID, warehouseID uuid.UUID) (map[domain.TaskStatus]int, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT status, COUNT(*) FROM tasks
		WHERE tenant_id=$1 AND warehouse_id=$2
		GROUP BY status`, tenantID, warehouseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[domain.TaskStatus]int{}
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		out[domain.TaskStatus(status)] = count
	}
	return out, rows.Err()
}

func (r *TaskRepo) scanTask(row scannable) (domain.Task, error) {
	t, err := scanTaskRow(row)
	if err != nil {
		return domain.Task{}, mapNotFound(err)
	}
	return t, nil
}

func scanTaskRow(row scannable) (domain.Task, error) {
	var t domain.Task
	var fulfillmentID, stationID, assigneeID, waveID, batchID uuid.NullUUID
	var typ, status string
	var sla, claimed, started, completed, cancelled, escalated sql.NullTime
	var historyRaw []byte
	var meta JSONMap
	err := row.Scan(
		&t.ID, &t.TenantID, &t.WarehouseID, &fulfillmentID, &stationID, &typ, &status, &assigneeID,
		&t.Priority, &waveID, &batchID, &sla, &claimed, &started, &completed,
		&cancelled, &escalated, &t.EscalationNote, &historyRaw, &meta, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return domain.Task{}, err
	}
	t.FulfillmentID = scanUUIDOrNil(fulfillmentID)
	t.StationID = scanNullUUID(stationID)
	t.AssigneeID = scanNullUUID(assigneeID)
	t.WaveID = scanNullUUID(waveID)
	t.BatchID = scanNullUUID(batchID)
	t.Type = domain.TaskType(typ)
	t.Status = domain.TaskStatus(status)
	t.SLADeadline = scanNullTime(sla)
	t.ClaimedAt = scanNullTime(claimed)
	t.StartedAt = scanNullTime(started)
	t.CompletedAt = scanNullTime(completed)
	t.CancelledAt = scanNullTime(cancelled)
	t.EscalatedAt = scanNullTime(escalated)
	t.Metadata = map[string]any(meta)
	t.CreatedAt = t.CreatedAt.UTC()
	t.UpdatedAt = t.UpdatedAt.UTC()
	if len(historyRaw) > 0 && string(historyRaw) != "null" {
		_ = json.Unmarshal(historyRaw, &t.History)
	}
	if t.History == nil {
		t.History = []domain.TaskHistoryEntry{}
	}
	return t, nil
}
