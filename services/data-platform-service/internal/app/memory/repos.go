package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/data-platform-service/internal/domain"
)

type Store struct {
	mu           sync.RWMutex
	Schemas      map[string]domain.EventSchema
	Events       []domain.AnalyticsEvent
	Idem         map[string]uuid.UUID
	Jobs         []domain.StreamJob
	Aggs         []domain.AggregateWindow
	Datasets     []domain.LakeDataset
	Facts        []domain.FactSnapshot
	KPIs         map[string]domain.KPIValue
	Realtime     map[string]domain.RealtimeMetric
	Experiments  map[string]domain.Experiment
	Assignments  map[string]domain.ExperimentAssignment
	Reports      []domain.ReportDef
	ReportRuns   []domain.ReportRun
	Obs          []domain.ObservabilitySignal
	AlertRules   []domain.AlertRule
	AlertEvents  []domain.AlertEvent
	Assets       []domain.CatalogAsset
	Lineage      []domain.LineageEdge
	Quality      []domain.QualityCheck
	Outbox       []domain.OutboxMessage
	OLAP         []olapRow
}

type olapRow struct {
	TenantID uuid.UUID
	Metric   string
	Value    float64
	TS       time.Time
}

func NewStore() *Store {
	return &Store{
		Schemas: make(map[string]domain.EventSchema),
		Idem: make(map[string]uuid.UUID),
		KPIs: make(map[string]domain.KPIValue),
		Realtime: make(map[string]domain.RealtimeMetric),
		Experiments: make(map[string]domain.Experiment),
		Assignments: make(map[string]domain.ExperimentAssignment),
	}
}

func schemaKey(tenantID uuid.UUID, name string, version int) string {
	return tenantID.String() + "|" + name + "|" + itoa(version)
}

func kpiKey(tenantID uuid.UUID, key string) string {
	return tenantID.String() + "|" + key
}

func rtKey(tenantID uuid.UUID, key string) string {
	return tenantID.String() + "|" + key
}

func expKey(tenantID uuid.UUID, key string) string {
	return tenantID.String() + "|" + key
}

func assignKey(tenantID, expID, subjectID uuid.UUID) string {
	return tenantID.String() + "|" + expID.String() + "|" + subjectID.String()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	d := []byte{}
	for n > 0 {
		d = append([]byte{byte('0' + n%10)}, d...)
		n /= 10
	}
	return string(d)
}

type Repos struct {
	Schemas     *SchemaRepo
	Events      *EventRepo
	Streams     *StreamRepo
	Lake        *LakeRepo
	Warehouse   *WarehouseRepo
	Realtime    *RealtimeRepo
	Experiments *ExperimentRepo
	Reports     *ReportRepo
	Obs         *ObsRepo
	Alerts      *AlertRepo
	Catalog     *CatalogRepo
	Quality     *QualityRepo
	Outbox      *OutboxRepo
	OLAP        *OLAPWriter
}

func NewRepos(s *Store) *Repos {
	return &Repos{
		Schemas: &SchemaRepo{s: s}, Events: &EventRepo{s: s}, Streams: &StreamRepo{s: s},
		Lake: &LakeRepo{s: s}, Warehouse: &WarehouseRepo{s: s}, Realtime: &RealtimeRepo{s: s},
		Experiments: &ExperimentRepo{s: s}, Reports: &ReportRepo{s: s}, Obs: &ObsRepo{s: s},
		Alerts: &AlertRepo{s: s}, Catalog: &CatalogRepo{s: s}, Quality: &QualityRepo{s: s},
		Outbox: &OutboxRepo{s: s}, OLAP: &OLAPWriter{s: s},
	}
}

type SchemaRepo struct{ s *Store }

func (r *SchemaRepo) Save(_ context.Context, s domain.EventSchema) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Schemas[schemaKey(s.TenantID, s.Name, s.Version)] = s
	return nil
}

func (r *SchemaRepo) Get(_ context.Context, tenantID uuid.UUID, name string, version int) (domain.EventSchema, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	s, ok := r.s.Schemas[schemaKey(tenantID, name, version)]
	if !ok {
		return domain.EventSchema{}, domain.ErrNotFound
	}
	return s, nil
}

func (r *SchemaRepo) GetLatest(_ context.Context, tenantID uuid.UUID, name string) (domain.EventSchema, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var best domain.EventSchema
	found := false
	prefix := tenantID.String() + "|" + name + "|"
	for k, s := range r.s.Schemas {
		if strings.HasPrefix(k, prefix) {
			if !found || s.Version > best.Version {
				best, found = s, true
			}
		}
	}
	if !found {
		return domain.EventSchema{}, domain.ErrNotFound
	}
	return best, nil
}

func (r *SchemaRepo) List(_ context.Context, tenantID uuid.UUID, family string) ([]domain.EventSchema, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.EventSchema, 0)
	for _, s := range r.s.Schemas {
		if s.TenantID == tenantID && (family == "" || s.Family == family) {
			out = append(out, s)
		}
	}
	return out, nil
}

type EventRepo struct{ s *Store }

func (r *EventRepo) Save(_ context.Context, e domain.AnalyticsEvent) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Events {
		if r.s.Events[i].ID == e.ID {
			r.s.Events[i] = e
			return nil
		}
	}
	r.s.Events = append(r.s.Events, e)
	if e.IdempotencyKey != "" {
		r.s.Idem[e.TenantID.String()+"|"+e.IdempotencyKey] = e.ID
	}
	return nil
}

func (r *EventRepo) GetByIdempotency(_ context.Context, tenantID uuid.UUID, key string) (domain.AnalyticsEvent, bool, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	id, ok := r.s.Idem[tenantID.String()+"|"+key]
	if !ok {
		return domain.AnalyticsEvent{}, false, nil
	}
	for _, e := range r.s.Events {
		if e.ID == id {
			return e, true, nil
		}
	}
	return domain.AnalyticsEvent{}, false, nil
}

func (r *EventRepo) List(_ context.Context, tenantID uuid.UUID, name string, limit int) ([]domain.AnalyticsEvent, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.AnalyticsEvent, 0)
	for i := len(r.s.Events) - 1; i >= 0; i-- {
		e := r.s.Events[i]
		if e.TenantID == tenantID && (name == "" || e.Name == name) {
			out = append(out, e)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *EventRepo) CountByHash(_ context.Context, tenantID uuid.UUID, hash string, since time.Time) (int, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	n := 0
	for _, e := range r.s.Events {
		if e.TenantID == tenantID && e.PayloadHash == hash && !e.IngestedAt.Before(since) {
			n++
		}
	}
	return n, nil
}

type StreamRepo struct{ s *Store }

func (r *StreamRepo) SaveJob(_ context.Context, j domain.StreamJob) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Jobs {
		if r.s.Jobs[i].ID == j.ID {
			r.s.Jobs[i] = j
			return nil
		}
	}
	r.s.Jobs = append(r.s.Jobs, j)
	return nil
}

func (r *StreamRepo) ListJobs(_ context.Context, tenantID uuid.UUID) ([]domain.StreamJob, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.StreamJob, 0)
	for _, j := range r.s.Jobs {
		if j.TenantID == tenantID {
			out = append(out, j)
		}
	}
	return out, nil
}

func (r *StreamRepo) UpsertAggregate(_ context.Context, a domain.AggregateWindow) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Aggs {
		if r.s.Aggs[i].JobID == a.JobID && r.s.Aggs[i].WindowStart.Equal(a.WindowStart) {
			r.s.Aggs[i] = a
			return nil
		}
	}
	r.s.Aggs = append(r.s.Aggs, a)
	return nil
}

func (r *StreamRepo) ListAggregates(_ context.Context, tenantID, jobID uuid.UUID, limit int) ([]domain.AggregateWindow, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.AggregateWindow, 0)
	for _, a := range r.s.Aggs {
		if a.TenantID == tenantID && a.JobID == jobID {
			out = append(out, a)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

type LakeRepo struct{ s *Store }

func (r *LakeRepo) SaveDataset(_ context.Context, d domain.LakeDataset) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Datasets {
		if r.s.Datasets[i].TenantID == d.TenantID && r.s.Datasets[i].Name == d.Name && r.s.Datasets[i].Layer == d.Layer {
			r.s.Datasets[i] = d
			return nil
		}
	}
	r.s.Datasets = append(r.s.Datasets, d)
	return nil
}

func (r *LakeRepo) ListDatasets(_ context.Context, tenantID uuid.UUID, layer string) ([]domain.LakeDataset, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.LakeDataset, 0)
	for _, d := range r.s.Datasets {
		if d.TenantID == tenantID && (layer == "" || d.Layer == layer) {
			out = append(out, d)
		}
	}
	return out, nil
}

type WarehouseRepo struct{ s *Store }

func (r *WarehouseRepo) SaveFact(_ context.Context, f domain.FactSnapshot) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Facts = append(r.s.Facts, f)
	return nil
}

func (r *WarehouseRepo) ListFacts(_ context.Context, tenantID uuid.UUID, factTable string, limit int) ([]domain.FactSnapshot, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.FactSnapshot, 0)
	for _, f := range r.s.Facts {
		if f.TenantID == tenantID && (factTable == "" || f.FactTable == factTable) {
			out = append(out, f)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *WarehouseRepo) SaveKPI(_ context.Context, k domain.KPIValue) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.KPIs[kpiKey(k.TenantID, k.Key)] = k
	return nil
}

func (r *WarehouseRepo) GetKPI(_ context.Context, tenantID uuid.UUID, key string) (domain.KPIValue, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	k, ok := r.s.KPIs[kpiKey(tenantID, key)]
	if !ok {
		return domain.KPIValue{}, domain.ErrNotFound
	}
	return k, nil
}

func (r *WarehouseRepo) ListKPIs(_ context.Context, tenantID uuid.UUID) ([]domain.KPIValue, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.KPIValue, 0)
	for _, k := range r.s.KPIs {
		if k.TenantID == tenantID {
			out = append(out, k)
		}
	}
	return out, nil
}

type RealtimeRepo struct{ s *Store }

func (r *RealtimeRepo) Incr(_ context.Context, tenantID uuid.UUID, key string, delta float64, now time.Time) (domain.RealtimeMetric, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	k := rtKey(tenantID, key)
	m := r.s.Realtime[k]
	m.TenantID = tenantID
	m.Key = key
	m.Value += delta
	m.UpdatedAt = now
	r.s.Realtime[k] = m
	return m, nil
}

func (r *RealtimeRepo) Get(_ context.Context, tenantID uuid.UUID, key string) (domain.RealtimeMetric, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	m, ok := r.s.Realtime[rtKey(tenantID, key)]
	if !ok {
		return domain.RealtimeMetric{}, domain.ErrNotFound
	}
	return m, nil
}

func (r *RealtimeRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.RealtimeMetric, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.RealtimeMetric, 0)
	for _, m := range r.s.Realtime {
		if m.TenantID == tenantID {
			out = append(out, m)
		}
	}
	return out, nil
}

type ExperimentRepo struct{ s *Store }

func (r *ExperimentRepo) Save(_ context.Context, e domain.Experiment) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Experiments[expKey(e.TenantID, e.Key)] = e
	return nil
}

func (r *ExperimentRepo) Get(_ context.Context, tenantID uuid.UUID, key string) (domain.Experiment, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	e, ok := r.s.Experiments[expKey(tenantID, key)]
	if !ok {
		return domain.Experiment{}, domain.ErrNotFound
	}
	return e, nil
}

func (r *ExperimentRepo) SaveAssignment(_ context.Context, a domain.ExperimentAssignment) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Assignments[assignKey(a.TenantID, a.ExperimentID, a.SubjectID)] = a
	return nil
}

func (r *ExperimentRepo) GetAssignment(_ context.Context, tenantID, experimentID, subjectID uuid.UUID) (domain.ExperimentAssignment, bool, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	a, ok := r.s.Assignments[assignKey(tenantID, experimentID, subjectID)]
	return a, ok, nil
}

type ReportRepo struct{ s *Store }

func (r *ReportRepo) SaveDef(_ context.Context, def domain.ReportDef) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Reports {
		if r.s.Reports[i].ID == def.ID {
			r.s.Reports[i] = def
			return nil
		}
	}
	r.s.Reports = append(r.s.Reports, def)
	return nil
}

func (r *ReportRepo) ListDefs(_ context.Context, tenantID uuid.UUID) ([]domain.ReportDef, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.ReportDef, 0)
	for _, def := range r.s.Reports {
		if def.TenantID == tenantID {
			out = append(out, def)
		}
	}
	return out, nil
}

func (r *ReportRepo) SaveRun(_ context.Context, run domain.ReportRun) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.ReportRuns = append(r.s.ReportRuns, run)
	return nil
}

func (r *ReportRepo) ListRuns(_ context.Context, tenantID, reportID uuid.UUID, limit int) ([]domain.ReportRun, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.ReportRun, 0)
	for _, run := range r.s.ReportRuns {
		if run.TenantID == tenantID && run.ReportID == reportID {
			out = append(out, run)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

type ObsRepo struct{ s *Store }

func (r *ObsRepo) SaveSignal(_ context.Context, s domain.ObservabilitySignal) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Obs = append(r.s.Obs, s)
	return nil
}

func (r *ObsRepo) ListSignals(_ context.Context, tenantID uuid.UUID, kind, service string, limit int) ([]domain.ObservabilitySignal, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.ObservabilitySignal, 0)
	for i := len(r.s.Obs) - 1; i >= 0; i-- {
		s := r.s.Obs[i]
		if s.TenantID == tenantID && (kind == "" || s.Kind == kind) && (service == "" || s.Service == service) {
			out = append(out, s)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

type AlertRepo struct{ s *Store }

func (r *AlertRepo) SaveRule(_ context.Context, rule domain.AlertRule) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.AlertRules {
		if r.s.AlertRules[i].ID == rule.ID {
			r.s.AlertRules[i] = rule
			return nil
		}
	}
	r.s.AlertRules = append(r.s.AlertRules, rule)
	return nil
}

func (r *AlertRepo) ListRules(_ context.Context, tenantID uuid.UUID) ([]domain.AlertRule, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.AlertRule, 0)
	for _, rule := range r.s.AlertRules {
		if rule.TenantID == tenantID {
			out = append(out, rule)
		}
	}
	return out, nil
}

func (r *AlertRepo) SaveEvent(_ context.Context, e domain.AlertEvent) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.AlertEvents = append(r.s.AlertEvents, e)
	return nil
}

func (r *AlertRepo) ListEvents(_ context.Context, tenantID uuid.UUID, limit int) ([]domain.AlertEvent, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.AlertEvent, 0)
	for i := len(r.s.AlertEvents) - 1; i >= 0; i-- {
		e := r.s.AlertEvents[i]
		if e.TenantID == tenantID {
			out = append(out, e)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

type CatalogRepo struct{ s *Store }

func (r *CatalogRepo) SaveAsset(_ context.Context, a domain.CatalogAsset) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Assets {
		if r.s.Assets[i].TenantID == a.TenantID && r.s.Assets[i].Name == a.Name {
			r.s.Assets[i] = a
			return nil
		}
	}
	r.s.Assets = append(r.s.Assets, a)
	return nil
}

func (r *CatalogRepo) ListAssets(_ context.Context, tenantID uuid.UUID, typ string) ([]domain.CatalogAsset, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.CatalogAsset, 0)
	for _, a := range r.s.Assets {
		if a.TenantID == tenantID && (typ == "" || a.Type == typ) {
			out = append(out, a)
		}
	}
	return out, nil
}

func (r *CatalogRepo) SaveLineage(_ context.Context, e domain.LineageEdge) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Lineage = append(r.s.Lineage, e)
	return nil
}

func (r *CatalogRepo) ListLineage(_ context.Context, tenantID uuid.UUID, name string) ([]domain.LineageEdge, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.LineageEdge, 0)
	for _, e := range r.s.Lineage {
		if e.TenantID == tenantID && (name == "" || e.FromName == name || e.ToName == name) {
			out = append(out, e)
		}
	}
	return out, nil
}

type QualityRepo struct{ s *Store }

func (r *QualityRepo) Save(_ context.Context, q domain.QualityCheck) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Quality = append(r.s.Quality, q)
	return nil
}

func (r *QualityRepo) List(_ context.Context, tenantID uuid.UUID, asset string, limit int) ([]domain.QualityCheck, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.QualityCheck, 0)
	for i := len(r.s.Quality) - 1; i >= 0; i-- {
		q := r.s.Quality[i]
		if q.TenantID == tenantID && (asset == "" || q.AssetName == asset) {
			out = append(out, q)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

type OutboxRepo struct{ s *Store }

func (r *OutboxRepo) Enqueue(_ context.Context, m domain.OutboxMessage) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Outbox = append(r.s.Outbox, m)
	return nil
}

func (r *OutboxRepo) ListPending(_ context.Context, limit int) ([]domain.OutboxMessage, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.OutboxMessage, 0)
	for _, m := range r.s.Outbox {
		if m.Status == domain.OutboxStatusPending {
			out = append(out, m)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *OutboxRepo) Update(_ context.Context, m domain.OutboxMessage) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Outbox {
		if r.s.Outbox[i].ID == m.ID {
			r.s.Outbox[i] = m
			return nil
		}
	}
	return domain.ErrNotFound
}

type OLAPWriter struct{ s *Store }

func (r *OLAPWriter) InsertAggregate(_ context.Context, tenantID uuid.UUID, metric string, value float64, ts time.Time) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.OLAP = append(r.s.OLAP, olapRow{TenantID: tenantID, Metric: metric, Value: value, TS: ts})
	return nil
}

func (r *OLAPWriter) QuerySum(_ context.Context, tenantID uuid.UUID, metric string, from, to time.Time) (float64, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var sum float64
	for _, row := range r.s.OLAP {
		if row.TenantID == tenantID && row.Metric == metric && !row.TS.Before(from) && row.TS.Before(to) {
			sum += row.Value
		}
	}
	return sum, nil
}
