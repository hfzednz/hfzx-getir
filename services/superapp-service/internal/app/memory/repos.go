package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/nexora/superapp-service/internal/domain"
)

type Store struct {
	mu            sync.RWMutex
	Modules       map[uuid.UUID]domain.Module
	ModuleByKey   map[string]uuid.UUID
	Manifests     map[string]domain.ModuleManifest
	Installs      map[string]domain.Install
	Permissions   map[string]domain.PermissionGrant
	Listings      map[uuid.UUID]domain.StoreListing
	Ratings       map[uuid.UUID]domain.StoreRating
	Widgets       map[uuid.UUID]domain.WidgetPlacement
	Monetization  map[uuid.UUID]domain.MonetizationRule
	Launches      map[uuid.UUID]domain.LaunchEvent
	Outbox        map[uuid.UUID]domain.OutboxMessage
}

func NewStore() *Store {
	return &Store{
		Modules: map[uuid.UUID]domain.Module{}, ModuleByKey: map[string]uuid.UUID{},
		Manifests: map[string]domain.ModuleManifest{}, Installs: map[string]domain.Install{},
		Permissions: map[string]domain.PermissionGrant{}, Listings: map[uuid.UUID]domain.StoreListing{},
		Ratings: map[uuid.UUID]domain.StoreRating{}, Widgets: map[uuid.UUID]domain.WidgetPlacement{},
		Monetization: map[uuid.UUID]domain.MonetizationRule{}, Launches: map[uuid.UUID]domain.LaunchEvent{},
		Outbox: map[uuid.UUID]domain.OutboxMessage{},
	}
}

func mk(tenantID uuid.UUID, key string) string { return tenantID.String() + ":" + key }
func manKey(tenantID, moduleID uuid.UUID, version string) string {
	return tenantID.String() + ":" + moduleID.String() + ":" + version
}
func instKey(tenantID uuid.UUID, subject string, moduleID uuid.UUID) string {
	return tenantID.String() + ":" + subject + ":" + moduleID.String()
}
func permKey(tenantID uuid.UUID, subject string, moduleID uuid.UUID, perm string) string {
	return tenantID.String() + ":" + subject + ":" + moduleID.String() + ":" + perm
}

type Repos struct {
	Modules      *ModuleRepo
	Manifests    *ManifestRepo
	Installs     *InstallRepo
	Permissions  *PermissionRepo
	Listings     *ListingRepo
	Ratings      *RatingRepo
	Widgets      *WidgetRepo
	Monetization *MonetizationRepo
	Launches     *LaunchRepo
	Outbox       *OutboxRepo
	LiveOps      *MockLiveOps
	AI           *MockAI
	Metrics      *MockMetrics
}

func NewRepos(s *Store) *Repos {
	return &Repos{
		Modules: &ModuleRepo{s: s}, Manifests: &ManifestRepo{s: s}, Installs: &InstallRepo{s: s},
		Permissions: &PermissionRepo{s: s}, Listings: &ListingRepo{s: s}, Ratings: &RatingRepo{s: s},
		Widgets: &WidgetRepo{s: s}, Monetization: &MonetizationRepo{s: s}, Launches: &LaunchRepo{s: s},
		Outbox: &OutboxRepo{s: s}, LiveOps: &MockLiveOps{enabled: true}, AI: &MockAI{}, Metrics: &MockMetrics{},
	}
}

type ModuleRepo struct{ s *Store }

func (r *ModuleRepo) Save(_ context.Context, m domain.Module) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Modules[m.ID] = m
	r.s.ModuleByKey[mk(m.TenantID, m.Key)] = m.ID
	return nil
}
func (r *ModuleRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Module, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	m, ok := r.s.Modules[id]
	if !ok || m.TenantID != tenantID {
		return domain.Module{}, domain.ErrNotFound
	}
	return m, nil
}
func (r *ModuleRepo) GetByKey(_ context.Context, tenantID uuid.UUID, key string) (domain.Module, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	id, ok := r.s.ModuleByKey[mk(tenantID, key)]
	if !ok {
		return domain.Module{}, domain.ErrNotFound
	}
	return r.s.Modules[id], nil
}
func (r *ModuleRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.Module, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Module{}
	for _, m := range r.s.Modules {
		if m.TenantID == tenantID {
			out = append(out, m)
		}
	}
	return out, nil
}

type ManifestRepo struct{ s *Store }

func (r *ManifestRepo) Save(_ context.Context, m domain.ModuleManifest) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Manifests[manKey(m.TenantID, m.ModuleID, m.Version)] = m
	return nil
}
func (r *ManifestRepo) Get(_ context.Context, tenantID, moduleID uuid.UUID, version string) (domain.ModuleManifest, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	m, ok := r.s.Manifests[manKey(tenantID, moduleID, version)]
	if !ok {
		return domain.ModuleManifest{}, domain.ErrNotFound
	}
	return m, nil
}
func (r *ManifestRepo) Latest(_ context.Context, tenantID, moduleID uuid.UUID) (domain.ModuleManifest, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var best domain.ModuleManifest
	found := false
	for _, m := range r.s.Manifests {
		if m.TenantID == tenantID && m.ModuleID == moduleID {
			if !found || m.CreatedAt.After(best.CreatedAt) ||
				(m.CreatedAt.Equal(best.CreatedAt) && m.Version > best.Version) {
				best = m
				found = true
			}
		}
	}
	if !found {
		return domain.ModuleManifest{}, domain.ErrNotFound
	}
	return best, nil
}

type InstallRepo struct{ s *Store }

func (r *InstallRepo) Save(_ context.Context, i domain.Install) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Installs[instKey(i.TenantID, i.SubjectID, i.ModuleID)] = i
	return nil
}
func (r *InstallRepo) Get(_ context.Context, tenantID uuid.UUID, subjectID string, moduleID uuid.UUID) (domain.Install, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	i, ok := r.s.Installs[instKey(tenantID, subjectID, moduleID)]
	if !ok {
		return domain.Install{}, domain.ErrNotFound
	}
	return i, nil
}
func (r *InstallRepo) ListBySubject(_ context.Context, tenantID uuid.UUID, subjectID string) ([]domain.Install, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Install{}
	for _, i := range r.s.Installs {
		if i.TenantID == tenantID && i.SubjectID == subjectID {
			out = append(out, i)
		}
	}
	return out, nil
}

type PermissionRepo struct{ s *Store }

func (r *PermissionRepo) Save(_ context.Context, g domain.PermissionGrant) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Permissions[permKey(g.TenantID, g.SubjectID, g.ModuleID, g.Permission)] = g
	return nil
}
func (r *PermissionRepo) List(_ context.Context, tenantID uuid.UUID, subjectID string, moduleID uuid.UUID) ([]domain.PermissionGrant, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.PermissionGrant{}
	for _, g := range r.s.Permissions {
		if g.TenantID == tenantID && g.SubjectID == subjectID && g.ModuleID == moduleID {
			out = append(out, g)
		}
	}
	return out, nil
}
func (r *PermissionRepo) Has(_ context.Context, tenantID uuid.UUID, subjectID string, moduleID uuid.UUID, perm string) (bool, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	g, ok := r.s.Permissions[permKey(tenantID, subjectID, moduleID, perm)]
	return ok && g.Granted, nil
}

type ListingRepo struct{ s *Store }

func (r *ListingRepo) Save(_ context.Context, l domain.StoreListing) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Listings[l.ModuleID] = l
	return nil
}
func (r *ListingRepo) GetByModule(_ context.Context, tenantID, moduleID uuid.UUID) (domain.StoreListing, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	l, ok := r.s.Listings[moduleID]
	if !ok || l.TenantID != tenantID {
		return domain.StoreListing{}, domain.ErrNotFound
	}
	return l, nil
}
func (r *ListingRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.StoreListing, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.StoreListing{}
	for _, l := range r.s.Listings {
		if l.TenantID == tenantID {
			out = append(out, l)
		}
	}
	return out, nil
}

type RatingRepo struct{ s *Store }

func (r *RatingRepo) Save(_ context.Context, rating domain.StoreRating) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Ratings[rating.ID] = rating
	return nil
}
func (r *RatingRepo) ListByModule(_ context.Context, tenantID, moduleID uuid.UUID) ([]domain.StoreRating, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.StoreRating{}
	for _, rating := range r.s.Ratings {
		if rating.TenantID == tenantID && rating.ModuleID == moduleID {
			out = append(out, rating)
		}
	}
	return out, nil
}

type WidgetRepo struct{ s *Store }

func (r *WidgetRepo) Save(_ context.Context, w domain.WidgetPlacement) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Widgets[w.ID] = w
	return nil
}
func (r *WidgetRepo) ListBySubject(_ context.Context, tenantID uuid.UUID, subjectID string) ([]domain.WidgetPlacement, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.WidgetPlacement{}
	for _, w := range r.s.Widgets {
		if w.TenantID == tenantID && w.SubjectID == subjectID {
			out = append(out, w)
		}
	}
	return out, nil
}

type MonetizationRepo struct{ s *Store }

func (r *MonetizationRepo) Save(_ context.Context, rule domain.MonetizationRule) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Monetization[rule.ModuleID] = rule
	return nil
}
func (r *MonetizationRepo) GetByModule(_ context.Context, tenantID, moduleID uuid.UUID) (domain.MonetizationRule, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	rule, ok := r.s.Monetization[moduleID]
	if !ok || rule.TenantID != tenantID {
		return domain.MonetizationRule{}, domain.ErrNotFound
	}
	return rule, nil
}

type LaunchRepo struct{ s *Store }

func (r *LaunchRepo) Save(_ context.Context, e domain.LaunchEvent) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Launches[e.ID] = e
	return nil
}

type OutboxRepo struct{ s *Store }

func (r *OutboxRepo) Enqueue(_ context.Context, m domain.OutboxMessage) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Outbox[m.ID] = m
	return nil
}
func (r *OutboxRepo) ListPending(_ context.Context, limit int) ([]domain.OutboxMessage, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.OutboxMessage{}
	for _, m := range r.s.Outbox {
		if m.Status == domain.OutboxStatusPending {
			out = append(out, m)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}
func (r *OutboxRepo) Update(_ context.Context, m domain.OutboxMessage) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Outbox[m.ID] = m
	return nil
}

type MockLiveOps struct{ enabled bool }

func (m *MockLiveOps) ModuleEnabled(context.Context, uuid.UUID, string, string) (bool, error) {
	return m.enabled, nil
}

type MockAI struct{}

func (MockAI) RecommendModules(_ context.Context, _ uuid.UUID, _ string, limit int) ([]string, error) {
	out := []string{"qc", "food", "pharmacy"}
	if limit > 0 && len(out) > limit {
		return out[:limit], nil
	}
	return out, nil
}

type MockMetrics struct{}

func (MockMetrics) Record(context.Context, string, map[string]string, float64) error { return nil }
