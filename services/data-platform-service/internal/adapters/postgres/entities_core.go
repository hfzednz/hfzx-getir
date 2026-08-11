package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/nexora/data-platform-service/internal/app/ports"
	"github.com/nexora/data-platform-service/internal/domain"
)

type SchemaRepo struct{ DB *sql.DB }

func (r *SchemaRepo) Save(ctx context.Context, s domain.EventSchema) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO event_schemas (
			id, tenant_id, name, family, version, compatibility, json_schema, active, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (tenant_id, name, version) DO UPDATE SET
			id=EXCLUDED.id, family=EXCLUDED.family, compatibility=EXCLUDED.compatibility,
			json_schema=EXCLUDED.json_schema, active=EXCLUDED.active`,
		s.ID, s.TenantID, s.Name, s.Family, s.Version, s.Compatibility, JSONMap(s.JSONSchema),
		s.Active, s.CreatedAt.UTC())
	return err
}

func (r *SchemaRepo) Get(ctx context.Context, tenantID uuid.UUID, name string, version int) (domain.EventSchema, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, family, version, compatibility, json_schema, active, created_at
		FROM event_schemas WHERE tenant_id=$1 AND name=$2 AND version=$3`, tenantID, name, version)
	return scanSchema(row)
}

func (r *SchemaRepo) GetLatest(ctx context.Context, tenantID uuid.UUID, name string) (domain.EventSchema, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, family, version, compatibility, json_schema, active, created_at
		FROM event_schemas WHERE tenant_id=$1 AND name=$2 ORDER BY version DESC LIMIT 1`, tenantID, name)
	return scanSchema(row)
}

func (r *SchemaRepo) List(ctx context.Context, tenantID uuid.UUID, family string) ([]domain.EventSchema, error) {
	q := `
		SELECT id, tenant_id, name, family, version, compatibility, json_schema, active, created_at
		FROM event_schemas WHERE tenant_id=$1`
	args := []any{tenantID}
	if family != "" {
		q += ` AND family=$2`
		args = append(args, family)
	}
	q += ` ORDER BY name ASC, version DESC`
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.EventSchema{}
	for rows.Next() {
		s, err := scanSchema(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

type scannable interface{ Scan(dest ...any) error }

func scanSchema(row scannable) (domain.EventSchema, error) {
	var s domain.EventSchema
	var schema JSONMap
	err := row.Scan(&s.ID, &s.TenantID, &s.Name, &s.Family, &s.Version, &s.Compatibility, &schema, &s.Active, &s.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.EventSchema{}, domain.ErrNotFound
		}
		return domain.EventSchema{}, err
	}
	s.JSONSchema = map[string]any(schema)
	s.CreatedAt = s.CreatedAt.UTC()
	return s, nil
}

type EventRepo struct{ DB *sql.DB }

func (r *EventRepo) Save(ctx context.Context, e domain.AnalyticsEvent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO analytics_events (
			id, tenant_id, name, family, schema_version, idempotency_key, occurred_at, ingested_at,
			user_id, session_id, city_id, payload, payload_hash, layer, valid, error
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		ON CONFLICT (id) DO UPDATE SET
			name=EXCLUDED.name, family=EXCLUDED.family, schema_version=EXCLUDED.schema_version,
			idempotency_key=EXCLUDED.idempotency_key, occurred_at=EXCLUDED.occurred_at,
			ingested_at=EXCLUDED.ingested_at, user_id=EXCLUDED.user_id, session_id=EXCLUDED.session_id,
			city_id=EXCLUDED.city_id, payload=EXCLUDED.payload, payload_hash=EXCLUDED.payload_hash,
			layer=EXCLUDED.layer, valid=EXCLUDED.valid, error=EXCLUDED.error`,
		e.ID, e.TenantID, e.Name, e.Family, e.SchemaVersion, e.IdempotencyKey,
		e.OccurredAt.UTC(), e.IngestedAt.UTC(), nullUUID(e.UserID), nullUUID(e.SessionID), nullUUID(e.CityID),
		JSONMap(e.Payload), e.PayloadHash, e.Layer, e.Valid, e.Error)
	return err
}

func (r *EventRepo) GetByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.AnalyticsEvent, bool, error) {
	if key == "" {
		return domain.AnalyticsEvent{}, false, nil
	}
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, family, schema_version, idempotency_key, occurred_at, ingested_at,
			user_id, session_id, city_id, payload, payload_hash, layer, valid, error
		FROM analytics_events WHERE tenant_id=$1 AND idempotency_key=$2`, tenantID, key)
	e, err := scanEvent(row)
	if err != nil {
		if isNoRows(err) {
			return domain.AnalyticsEvent{}, false, nil
		}
		return domain.AnalyticsEvent{}, false, err
	}
	return e, true, nil
}

func (r *EventRepo) List(ctx context.Context, tenantID uuid.UUID, name string, limit int) ([]domain.AnalyticsEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	q := `
		SELECT id, tenant_id, name, family, schema_version, idempotency_key, occurred_at, ingested_at,
			user_id, session_id, city_id, payload, payload_hash, layer, valid, error
		FROM analytics_events WHERE tenant_id=$1`
	args := []any{tenantID}
	n := 2
	if name != "" {
		q += ` AND name=$2`
		args = append(args, name)
		n = 3
	}
	q += fmt.Sprintf(` ORDER BY ingested_at DESC LIMIT $%d`, n)
	args = append(args, limit)
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.AnalyticsEvent{}
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *EventRepo) CountByHash(ctx context.Context, tenantID uuid.UUID, hash string, since time.Time) (int, error) {
	var n int
	err := r.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM analytics_events
		WHERE tenant_id=$1 AND payload_hash=$2 AND ingested_at >= $3`, tenantID, hash, since.UTC()).Scan(&n)
	return n, err
}

func scanEvent(row scannable) (domain.AnalyticsEvent, error) {
	var e domain.AnalyticsEvent
	var payload JSONMap
	var user, session, city uuid.NullUUID
	err := row.Scan(
		&e.ID, &e.TenantID, &e.Name, &e.Family, &e.SchemaVersion, &e.IdempotencyKey,
		&e.OccurredAt, &e.IngestedAt, &user, &session, &city, &payload, &e.PayloadHash,
		&e.Layer, &e.Valid, &e.Error)
	if err != nil {
		return domain.AnalyticsEvent{}, err
	}
	e.UserID = scanNullUUID(user)
	e.SessionID = scanNullUUID(session)
	e.CityID = scanNullUUID(city)
	e.Payload = map[string]any(payload)
	e.OccurredAt = e.OccurredAt.UTC()
	e.IngestedAt = e.IngestedAt.UTC()
	return e, nil
}

type StreamRepo struct{ DB *sql.DB }

func (r *StreamRepo) SaveJob(ctx context.Context, j domain.StreamJob) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO stream_jobs (
			id, tenant_id, name, event_name, window_sec, metric_field, agg, enabled, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			name=EXCLUDED.name, event_name=EXCLUDED.event_name, window_sec=EXCLUDED.window_sec,
			metric_field=EXCLUDED.metric_field, agg=EXCLUDED.agg, enabled=EXCLUDED.enabled,
			updated_at=EXCLUDED.updated_at`,
		j.ID, j.TenantID, j.Name, j.EventName, j.WindowSec, j.MetricField, j.Agg, j.Enabled, j.UpdatedAt.UTC())
	return err
}

func (r *StreamRepo) ListJobs(ctx context.Context, tenantID uuid.UUID) ([]domain.StreamJob, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, name, event_name, window_sec, metric_field, agg, enabled, updated_at
		FROM stream_jobs WHERE tenant_id=$1 ORDER BY name ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.StreamJob{}
	for rows.Next() {
		var j domain.StreamJob
		if err := rows.Scan(&j.ID, &j.TenantID, &j.Name, &j.EventName, &j.WindowSec, &j.MetricField, &j.Agg, &j.Enabled, &j.UpdatedAt); err != nil {
			return nil, err
		}
		j.UpdatedAt = j.UpdatedAt.UTC()
		out = append(out, j)
	}
	return out, rows.Err()
}

func (r *StreamRepo) UpsertAggregate(ctx context.Context, a domain.AggregateWindow) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO aggregate_windows (
			tenant_id, job_id, window_start, window_end, value, count, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (tenant_id, job_id, window_start) DO UPDATE SET
			window_end=EXCLUDED.window_end, value=EXCLUDED.value, count=EXCLUDED.count,
			updated_at=EXCLUDED.updated_at`,
		a.TenantID, a.JobID, a.WindowStart.UTC(), a.WindowEnd.UTC(), a.Value, a.Count, a.UpdatedAt.UTC())
	return err
}

func (r *StreamRepo) ListAggregates(ctx context.Context, tenantID, jobID uuid.UUID, limit int) ([]domain.AggregateWindow, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT tenant_id, job_id, window_start, window_end, value, count, updated_at
		FROM aggregate_windows WHERE tenant_id=$1 AND job_id=$2
		ORDER BY window_start DESC LIMIT $3`, tenantID, jobID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.AggregateWindow{}
	for rows.Next() {
		var a domain.AggregateWindow
		if err := rows.Scan(&a.TenantID, &a.JobID, &a.WindowStart, &a.WindowEnd, &a.Value, &a.Count, &a.UpdatedAt); err != nil {
			return nil, err
		}
		a.WindowStart = a.WindowStart.UTC()
		a.WindowEnd = a.WindowEnd.UTC()
		a.UpdatedAt = a.UpdatedAt.UTC()
		out = append(out, a)
	}
	return out, rows.Err()
}

type LakeRepo struct{ DB *sql.DB }

func (r *LakeRepo) SaveDataset(ctx context.Context, d domain.LakeDataset) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO lake_datasets (
			id, tenant_id, name, layer, format, location, partition_by, retention_days, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			name=EXCLUDED.name, layer=EXCLUDED.layer, format=EXCLUDED.format, location=EXCLUDED.location,
			partition_by=EXCLUDED.partition_by, retention_days=EXCLUDED.retention_days,
			updated_at=EXCLUDED.updated_at`,
		d.ID, d.TenantID, d.Name, d.Layer, d.Format, d.Location, textArray(d.PartitionBy),
		d.RetentionDays, d.UpdatedAt.UTC())
	return err
}

func (r *LakeRepo) ListDatasets(ctx context.Context, tenantID uuid.UUID, layer string) ([]domain.LakeDataset, error) {
	q := `
		SELECT id, tenant_id, name, layer, format, location, partition_by, retention_days, updated_at
		FROM lake_datasets WHERE tenant_id=$1`
	args := []any{tenantID}
	if layer != "" {
		q += ` AND layer=$2`
		args = append(args, layer)
	}
	q += ` ORDER BY name ASC`
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.LakeDataset{}
	for rows.Next() {
		var d domain.LakeDataset
		var parts []string
		if err := rows.Scan(&d.ID, &d.TenantID, &d.Name, &d.Layer, &d.Format, &d.Location, pq.Array(&parts), &d.RetentionDays, &d.UpdatedAt); err != nil {
			return nil, err
		}
		d.PartitionBy = parts
		d.UpdatedAt = d.UpdatedAt.UTC()
		out = append(out, d)
	}
	return out, rows.Err()
}

type WarehouseRepo struct{ DB *sql.DB }

func (r *WarehouseRepo) SaveFact(ctx context.Context, f domain.FactSnapshot) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO fact_snapshots (
			id, tenant_id, fact_table, grain_key, measures, dims, as_of, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		f.ID, f.TenantID, f.FactTable, f.GrainKey, FloatMap(f.Measures), StringMap(f.Dims),
		f.AsOf.UTC(), f.CreatedAt.UTC())
	return err
}

func (r *WarehouseRepo) ListFacts(ctx context.Context, tenantID uuid.UUID, factTable string, limit int) ([]domain.FactSnapshot, error) {
	if limit <= 0 {
		limit = 100
	}
	q := `
		SELECT id, tenant_id, fact_table, grain_key, measures, dims, as_of, created_at
		FROM fact_snapshots WHERE tenant_id=$1`
	args := []any{tenantID}
	n := 2
	if factTable != "" {
		q += ` AND fact_table=$2`
		args = append(args, factTable)
		n = 3
	}
	q += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d`, n)
	args = append(args, limit)
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.FactSnapshot{}
	for rows.Next() {
		var f domain.FactSnapshot
		var measures FloatMap
		var dims StringMap
		if err := rows.Scan(&f.ID, &f.TenantID, &f.FactTable, &f.GrainKey, &measures, &dims, &f.AsOf, &f.CreatedAt); err != nil {
			return nil, err
		}
		f.Measures = map[string]float64(measures)
		f.Dims = map[string]string(dims)
		f.AsOf = f.AsOf.UTC()
		f.CreatedAt = f.CreatedAt.UTC()
		out = append(out, f)
	}
	return out, rows.Err()
}

func (r *WarehouseRepo) SaveKPI(ctx context.Context, k domain.KPIValue) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO kpi_values (tenant_id, key, value, unit, dims, as_of)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (tenant_id, key) DO UPDATE SET
			value=EXCLUDED.value, unit=EXCLUDED.unit, dims=EXCLUDED.dims, as_of=EXCLUDED.as_of`,
		k.TenantID, k.Key, k.Value, k.Unit, StringMap(k.Dims), k.AsOf.UTC())
	return err
}

func (r *WarehouseRepo) GetKPI(ctx context.Context, tenantID uuid.UUID, key string) (domain.KPIValue, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT tenant_id, key, value, unit, dims, as_of FROM kpi_values WHERE tenant_id=$1 AND key=$2`,
		tenantID, key)
	var k domain.KPIValue
	var dims StringMap
	err := row.Scan(&k.TenantID, &k.Key, &k.Value, &k.Unit, &dims, &k.AsOf)
	if err != nil {
		if isNoRows(err) {
			return domain.KPIValue{}, domain.ErrNotFound
		}
		return domain.KPIValue{}, err
	}
	k.Dims = map[string]string(dims)
	k.AsOf = k.AsOf.UTC()
	return k, nil
}

func (r *WarehouseRepo) ListKPIs(ctx context.Context, tenantID uuid.UUID) ([]domain.KPIValue, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT tenant_id, key, value, unit, dims, as_of FROM kpi_values WHERE tenant_id=$1 ORDER BY key ASC`,
		tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.KPIValue{}
	for rows.Next() {
		var k domain.KPIValue
		var dims StringMap
		if err := rows.Scan(&k.TenantID, &k.Key, &k.Value, &k.Unit, &dims, &k.AsOf); err != nil {
			return nil, err
		}
		k.Dims = map[string]string(dims)
		k.AsOf = k.AsOf.UTC()
		out = append(out, k)
	}
	return out, rows.Err()
}

type RealtimeRepo struct{ DB *sql.DB }

func (r *RealtimeRepo) Incr(ctx context.Context, tenantID uuid.UUID, key string, delta float64, now time.Time) (domain.RealtimeMetric, error) {
	row := r.DB.QueryRowContext(ctx, `
		INSERT INTO realtime_metrics (tenant_id, key, value, updated_at)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (tenant_id, key) DO UPDATE SET
			value = realtime_metrics.value + EXCLUDED.value,
			updated_at = EXCLUDED.updated_at
		RETURNING tenant_id, key, value, updated_at`,
		tenantID, key, delta, now.UTC())
	var m domain.RealtimeMetric
	if err := row.Scan(&m.TenantID, &m.Key, &m.Value, &m.UpdatedAt); err != nil {
		return domain.RealtimeMetric{}, err
	}
	m.UpdatedAt = m.UpdatedAt.UTC()
	return m, nil
}

func (r *RealtimeRepo) Get(ctx context.Context, tenantID uuid.UUID, key string) (domain.RealtimeMetric, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT tenant_id, key, value, updated_at FROM realtime_metrics WHERE tenant_id=$1 AND key=$2`,
		tenantID, key)
	var m domain.RealtimeMetric
	err := row.Scan(&m.TenantID, &m.Key, &m.Value, &m.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.RealtimeMetric{}, domain.ErrNotFound
		}
		return domain.RealtimeMetric{}, err
	}
	m.UpdatedAt = m.UpdatedAt.UTC()
	return m, nil
}

func (r *RealtimeRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.RealtimeMetric, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT tenant_id, key, value, updated_at FROM realtime_metrics WHERE tenant_id=$1 ORDER BY key ASC`,
		tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.RealtimeMetric{}
	for rows.Next() {
		var m domain.RealtimeMetric
		if err := rows.Scan(&m.TenantID, &m.Key, &m.Value, &m.UpdatedAt); err != nil {
			return nil, err
		}
		m.UpdatedAt = m.UpdatedAt.UTC()
		out = append(out, m)
	}
	return out, rows.Err()
}

var (
	_ ports.SchemaRepo    = (*SchemaRepo)(nil)
	_ ports.EventRepo     = (*EventRepo)(nil)
	_ ports.StreamRepo    = (*StreamRepo)(nil)
	_ ports.LakeRepo      = (*LakeRepo)(nil)
	_ ports.WarehouseRepo = (*WarehouseRepo)(nil)
	_ ports.RealtimeRepo  = (*RealtimeRepo)(nil)
)
