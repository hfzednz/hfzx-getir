package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/liveops-service/internal/domain"
)

type Store struct {
	mu           sync.RWMutex
	Flags        []domain.FeatureFlag
	Configs      []domain.ConfigDocument
	Experiments  []domain.Experiment
	Assignments  []domain.Assignment
	Events       []domain.LiveOpsEvent
	Changes      []domain.ChangeRequest
	Rollbacks    []domain.RollbackRecord
	Outbox       []domain.OutboxMessage
	Cache        map[string]cacheEntry
}

type cacheEntry struct {
	Value     string
	ExpiresAt time.Time
}

func NewStore() *Store {
	return &Store{Cache: make(map[string]cacheEntry)}
}

type Repos struct {
	Flags       *FlagRepo
	Configs     *ConfigRepo
	Experiments *ExperimentRepo
	Events      *EventRepo
	Changes     *ChangeRepo
	Rollbacks   *RollbackRepo
	Outbox      *OutboxRepo
	Cache       *Cache
	Metrics     *MockMetrics
	AI          *MockAI
}

func NewRepos(s *Store) *Repos {
	return &Repos{
		Flags: &FlagRepo{s: s}, Configs: &ConfigRepo{s: s}, Experiments: &ExperimentRepo{s: s},
		Events: &EventRepo{s: s}, Changes: &ChangeRepo{s: s}, Rollbacks: &RollbackRepo{s: s},
		Outbox: &OutboxRepo{s: s}, Cache: &Cache{s: s}, Metrics: &MockMetrics{}, AI: &MockAI{},
	}
}

type FlagRepo struct{ s *Store }

func (r *FlagRepo) Save(_ context.Context, f domain.FeatureFlag) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Flags {
		if r.s.Flags[i].ID == f.ID {
			r.s.Flags[i] = f
			return nil
		}
	}
	r.s.Flags = append(r.s.Flags, f)
	return nil
}

func (r *FlagRepo) GetByKey(_ context.Context, tenantID uuid.UUID, key string) (domain.FeatureFlag, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	key = domain.NormalizeKey(key)
	for _, f := range r.s.Flags {
		if f.TenantID == tenantID && f.Key == key {
			return f, nil
		}
	}
	return domain.FeatureFlag{}, domain.ErrNotFound
}

func (r *FlagRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.FeatureFlag, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.FeatureFlag{}
	for _, f := range r.s.Flags {
		if f.TenantID == tenantID {
			out = append(out, f)
		}
	}
	return out, nil
}

type ConfigRepo struct{ s *Store }

func (r *ConfigRepo) Save(_ context.Context, c domain.ConfigDocument) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Configs {
		if r.s.Configs[i].ID == c.ID {
			r.s.Configs[i] = c
			return nil
		}
	}
	r.s.Configs = append(r.s.Configs, c)
	return nil
}

func (r *ConfigRepo) GetByKey(_ context.Context, tenantID uuid.UUID, key string) (domain.ConfigDocument, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	key = domain.NormalizeKey(key)
	for _, c := range r.s.Configs {
		if c.TenantID == tenantID && c.Key == key {
			return c, nil
		}
	}
	return domain.ConfigDocument{}, domain.ErrNotFound
}

func (r *ConfigRepo) List(_ context.Context, tenantID uuid.UUID, namespace string) ([]domain.ConfigDocument, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.ConfigDocument{}
	for _, c := range r.s.Configs {
		if c.TenantID == tenantID && (namespace == "" || c.Namespace == namespace) {
			out = append(out, c)
		}
	}
	return out, nil
}

type ExperimentRepo struct{ s *Store }

func (r *ExperimentRepo) Save(_ context.Context, e domain.Experiment) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Experiments {
		if r.s.Experiments[i].ID == e.ID {
			r.s.Experiments[i] = e
			return nil
		}
	}
	r.s.Experiments = append(r.s.Experiments, e)
	return nil
}

func (r *ExperimentRepo) GetByKey(_ context.Context, tenantID uuid.UUID, key string) (domain.Experiment, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	key = domain.NormalizeKey(key)
	for _, e := range r.s.Experiments {
		if e.TenantID == tenantID && e.Key == key {
			return e, nil
		}
	}
	return domain.Experiment{}, domain.ErrNotFound
}

func (r *ExperimentRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.Experiment, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Experiment{}
	for _, e := range r.s.Experiments {
		if e.TenantID == tenantID {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *ExperimentRepo) SaveAssignment(_ context.Context, a domain.Assignment) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Assignments = append(r.s.Assignments, a)
	return nil
}

func (r *ExperimentRepo) GetAssignment(_ context.Context, tenantID, experimentID uuid.UUID, subjectID string) (domain.Assignment, bool, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, a := range r.s.Assignments {
		if a.TenantID == tenantID && a.ExperimentID == experimentID && a.SubjectID == subjectID {
			return a, true, nil
		}
	}
	return domain.Assignment{}, false, nil
}

type EventRepo struct{ s *Store }

func (r *EventRepo) Save(_ context.Context, e domain.LiveOpsEvent) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Events {
		if r.s.Events[i].ID == e.ID {
			r.s.Events[i] = e
			return nil
		}
	}
	r.s.Events = append(r.s.Events, e)
	return nil
}

func (r *EventRepo) List(_ context.Context, tenantID uuid.UUID, status string) ([]domain.LiveOpsEvent, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.LiveOpsEvent{}
	for _, e := range r.s.Events {
		if e.TenantID == tenantID && (status == "" || e.Status == status) {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *EventRepo) GetByKey(_ context.Context, tenantID uuid.UUID, key string) (domain.LiveOpsEvent, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	key = domain.NormalizeKey(key)
	for _, e := range r.s.Events {
		if e.TenantID == tenantID && e.Key == key {
			return e, nil
		}
	}
	return domain.LiveOpsEvent{}, domain.ErrNotFound
}

type ChangeRepo struct{ s *Store }

func (r *ChangeRepo) Save(_ context.Context, c domain.ChangeRequest) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Changes {
		if r.s.Changes[i].ID == c.ID {
			r.s.Changes[i] = c
			return nil
		}
	}
	r.s.Changes = append(r.s.Changes, c)
	return nil
}

func (r *ChangeRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.ChangeRequest, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, c := range r.s.Changes {
		if c.TenantID == tenantID && c.ID == id {
			return c, nil
		}
	}
	return domain.ChangeRequest{}, domain.ErrNotFound
}

func (r *ChangeRepo) ListPending(_ context.Context, tenantID uuid.UUID) ([]domain.ChangeRequest, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.ChangeRequest{}
	for _, c := range r.s.Changes {
		if c.TenantID == tenantID && c.Status == "pending" {
			out = append(out, c)
		}
	}
	return out, nil
}

type RollbackRepo struct{ s *Store }

func (r *RollbackRepo) Save(_ context.Context, rec domain.RollbackRecord) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Rollbacks = append(r.s.Rollbacks, rec)
	return nil
}

func (r *RollbackRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.RollbackRecord, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.RollbackRecord{}
	for _, rec := range r.s.Rollbacks {
		if rec.TenantID == tenantID {
			out = append(out, rec)
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
	out := []domain.OutboxMessage{}
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

type Cache struct{ s *Store }

func (c *Cache) Get(_ context.Context, key string) (string, bool, error) {
	c.s.mu.RLock()
	defer c.s.mu.RUnlock()
	e, ok := c.s.Cache[key]
	if !ok || (!e.ExpiresAt.IsZero() && time.Now().UTC().After(e.ExpiresAt)) {
		return "", false, nil
	}
	return e.Value, true, nil
}

func (c *Cache) Set(_ context.Context, key, value string, ttl time.Duration) error {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	c.s.Cache[key] = cacheEntry{Value: value, ExpiresAt: time.Now().UTC().Add(ttl)}
	return nil
}

type MockMetrics struct{ N int }

func (m *MockMetrics) Ingest(_ context.Context, _ uuid.UUID, _ string, _ float64, _ map[string]string) error {
	m.N++
	return nil
}

type MockAI struct{}

func (MockAI) SuggestWinner(_ context.Context, _ uuid.UUID, _ string, rates map[string]float64) (string, error) {
	return domain.PickWinner(keysOf(rates), rates), nil
}

func keysOf(m map[string]float64) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
