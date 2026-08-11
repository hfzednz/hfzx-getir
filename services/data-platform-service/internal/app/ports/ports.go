package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/data-platform-service/internal/domain"
)

type Clock interface{ Now() time.Time }
type IDGen interface{ New() uuid.UUID }

type SchemaRepo interface {
	Save(ctx context.Context, s domain.EventSchema) error
	Get(ctx context.Context, tenantID uuid.UUID, name string, version int) (domain.EventSchema, error)
	GetLatest(ctx context.Context, tenantID uuid.UUID, name string) (domain.EventSchema, error)
	List(ctx context.Context, tenantID uuid.UUID, family string) ([]domain.EventSchema, error)
}

type EventRepo interface {
	Save(ctx context.Context, e domain.AnalyticsEvent) error
	GetByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.AnalyticsEvent, bool, error)
	List(ctx context.Context, tenantID uuid.UUID, name string, limit int) ([]domain.AnalyticsEvent, error)
	CountByHash(ctx context.Context, tenantID uuid.UUID, hash string, since time.Time) (int, error)
}

type StreamRepo interface {
	SaveJob(ctx context.Context, j domain.StreamJob) error
	ListJobs(ctx context.Context, tenantID uuid.UUID) ([]domain.StreamJob, error)
	UpsertAggregate(ctx context.Context, a domain.AggregateWindow) error
	ListAggregates(ctx context.Context, tenantID, jobID uuid.UUID, limit int) ([]domain.AggregateWindow, error)
}

type LakeRepo interface {
	SaveDataset(ctx context.Context, d domain.LakeDataset) error
	ListDatasets(ctx context.Context, tenantID uuid.UUID, layer string) ([]domain.LakeDataset, error)
}

type WarehouseRepo interface {
	SaveFact(ctx context.Context, f domain.FactSnapshot) error
	ListFacts(ctx context.Context, tenantID uuid.UUID, factTable string, limit int) ([]domain.FactSnapshot, error)
	SaveKPI(ctx context.Context, k domain.KPIValue) error
	GetKPI(ctx context.Context, tenantID uuid.UUID, key string) (domain.KPIValue, error)
	ListKPIs(ctx context.Context, tenantID uuid.UUID) ([]domain.KPIValue, error)
}

type RealtimeRepo interface {
	Incr(ctx context.Context, tenantID uuid.UUID, key string, delta float64, now time.Time) (domain.RealtimeMetric, error)
	Get(ctx context.Context, tenantID uuid.UUID, key string) (domain.RealtimeMetric, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.RealtimeMetric, error)
}

type ExperimentRepo interface {
	Save(ctx context.Context, e domain.Experiment) error
	Get(ctx context.Context, tenantID uuid.UUID, key string) (domain.Experiment, error)
	SaveAssignment(ctx context.Context, a domain.ExperimentAssignment) error
	GetAssignment(ctx context.Context, tenantID, experimentID, subjectID uuid.UUID) (domain.ExperimentAssignment, bool, error)
}

type ReportRepo interface {
	SaveDef(ctx context.Context, r domain.ReportDef) error
	ListDefs(ctx context.Context, tenantID uuid.UUID) ([]domain.ReportDef, error)
	SaveRun(ctx context.Context, r domain.ReportRun) error
	ListRuns(ctx context.Context, tenantID, reportID uuid.UUID, limit int) ([]domain.ReportRun, error)
}

type ObsRepo interface {
	SaveSignal(ctx context.Context, s domain.ObservabilitySignal) error
	ListSignals(ctx context.Context, tenantID uuid.UUID, kind, service string, limit int) ([]domain.ObservabilitySignal, error)
}

type AlertRepo interface {
	SaveRule(ctx context.Context, r domain.AlertRule) error
	ListRules(ctx context.Context, tenantID uuid.UUID) ([]domain.AlertRule, error)
	SaveEvent(ctx context.Context, e domain.AlertEvent) error
	ListEvents(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.AlertEvent, error)
}

type CatalogRepo interface {
	SaveAsset(ctx context.Context, a domain.CatalogAsset) error
	ListAssets(ctx context.Context, tenantID uuid.UUID, typ string) ([]domain.CatalogAsset, error)
	SaveLineage(ctx context.Context, e domain.LineageEdge) error
	ListLineage(ctx context.Context, tenantID uuid.UUID, name string) ([]domain.LineageEdge, error)
}

type QualityRepo interface {
	Save(ctx context.Context, q domain.QualityCheck) error
	List(ctx context.Context, tenantID uuid.UUID, asset string, limit int) ([]domain.QualityCheck, error)
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
	Update(ctx context.Context, m domain.OutboxMessage) error
}

type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, payload map[string]any) error
}

// OLAPWriter writes aggregates to ClickHouse (stub in memory mode).
type OLAPWriter interface {
	InsertAggregate(ctx context.Context, tenantID uuid.UUID, metric string, value float64, ts time.Time) error
	QuerySum(ctx context.Context, tenantID uuid.UUID, metric string, from, to time.Time) (float64, error)
}
