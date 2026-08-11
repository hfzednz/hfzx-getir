package memory

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/location-service/internal/app/ports"
	"github.com/nexora/location-service/internal/domain"
)

// Repos bundles in-memory port implementations.
type Repos struct {
	Addresses *AddressRepo
	POIs      *POIRepo
	History   *HistoryRepo
	Cache     *CacheRepo
	Heat      *HeatRepo
	Outbox    *OutboxRepo
}

// NewRepos wires all memory repositories to a shared store.
func NewRepos(s *Store) *Repos {
	return &Repos{
		Addresses: &AddressRepo{S: s},
		POIs:      &POIRepo{S: s},
		History:   &HistoryRepo{S: s},
		Cache:     &CacheRepo{S: s},
		Heat:      &HeatRepo{S: s},
		Outbox:    &OutboxRepo{S: s},
	}
}

// AddressRepo is an in-memory AddressRepo.
type AddressRepo struct{ S *Store }

func (r *AddressRepo) Upsert(_ context.Context, a domain.NormalizedAddress) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.addresses[a.ID] = a
	return nil
}

func (r *AddressRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.NormalizedAddress, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	a, ok := r.S.addresses[id]
	if !ok || a.TenantID != tenantID {
		return domain.NormalizedAddress{}, fmt.Errorf("%w: address", domain.ErrNotFound)
	}
	return a, nil
}

var _ ports.AddressRepo = (*AddressRepo)(nil)

// POIRepo is an in-memory POIRepo.
type POIRepo struct{ S *Store }

func (r *POIRepo) Upsert(_ context.Context, p domain.POI) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	cp := p
	if p.Meta != nil {
		cp.Meta = make(map[string]any, len(p.Meta))
		for k, v := range p.Meta {
			cp.Meta[k] = v
		}
	}
	r.S.pois[p.ID] = cp
	return nil
}

func (r *POIRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.POI, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	p, ok := r.S.pois[id]
	if !ok || p.TenantID != tenantID {
		return domain.POI{}, fmt.Errorf("%w: poi", domain.ErrNotFound)
	}
	return p, nil
}

func (r *POIRepo) Nearby(_ context.Context, q domain.NearbyQuery) ([]domain.POI, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	type scored struct {
		p domain.POI
		d float64
	}
	var hits []scored
	for _, p := range r.S.pois {
		if p.TenantID != q.TenantID || !p.Active {
			continue
		}
		if q.Kind != nil && p.Kind != *q.Kind {
			continue
		}
		d := domain.HaversineDistanceMeters(q.Center, domain.LatLng{Lat: p.Lat, Lng: p.Lng})
		if d <= q.RadiusM {
			hits = append(hits, scored{p: p, d: d})
		}
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].d < hits[j].d })
	out := make([]domain.POI, 0, min(limit, len(hits)))
	for i := 0; i < len(hits) && i < limit; i++ {
		out = append(out, hits[i].p)
	}
	return out, nil
}

func (r *POIRepo) NearestOfKind(_ context.Context, tenantID uuid.UUID, kind domain.POIKind, lat, lng float64, limit int) ([]domain.POI, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	if limit <= 0 {
		limit = 1
	}
	center := domain.LatLng{Lat: lat, Lng: lng}
	type scored struct {
		p domain.POI
		d float64
	}
	var hits []scored
	for _, p := range r.S.pois {
		if p.TenantID != tenantID || !p.Active || p.Kind != kind {
			continue
		}
		d := domain.HaversineDistanceMeters(center, domain.LatLng{Lat: p.Lat, Lng: p.Lng})
		hits = append(hits, scored{p: p, d: d})
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].d < hits[j].d })
	out := make([]domain.POI, 0, min(limit, len(hits)))
	for i := 0; i < len(hits) && i < limit; i++ {
		out = append(out, hits[i].p)
	}
	return out, nil
}

func (r *POIRepo) InBBox(_ context.Context, tenantID uuid.UUID, bbox domain.BBox, kind *domain.POIKind, limit int) ([]domain.POI, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	if limit <= 0 {
		limit = 50
	}
	out := make([]domain.POI, 0)
	for _, p := range r.S.pois {
		if p.TenantID != tenantID || !p.Active {
			continue
		}
		if kind != nil && p.Kind != *kind {
			continue
		}
		if bbox.Contains(domain.LatLng{Lat: p.Lat, Lng: p.Lng}) {
			out = append(out, p)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *POIRepo) Count(_ context.Context, tenantID uuid.UUID) (int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	n := 0
	for _, p := range r.S.pois {
		if p.TenantID == tenantID && p.Active {
			n++
		}
	}
	return n, nil
}

var _ ports.POIRepo = (*POIRepo)(nil)

// HistoryRepo is an in-memory HistoryRepo with per-subject cap.
type HistoryRepo struct{ S *Store }

func (r *HistoryRepo) Ingest(_ context.Context, h domain.LocationHistory) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.history = append(r.S.history, h)
	// Cap per subject: keep newest MaxHistoryPerSubject.
	type key struct {
		t uuid.UUID
		s domain.SubjectType
		i string
	}
	counts := map[key]int{}
	for i := len(r.S.history) - 1; i >= 0; i-- {
		row := r.S.history[i]
		k := key{row.TenantID, row.SubjectType, row.SubjectID}
		counts[k]++
		if counts[k] > domain.MaxHistoryPerSubject {
			r.S.history = append(r.S.history[:i], r.S.history[i+1:]...)
		}
	}
	return nil
}

func (r *HistoryRepo) List(_ context.Context, tenantID uuid.UUID, subjectType domain.SubjectType, subjectID string, limit int) ([]domain.LocationHistory, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	if limit <= 0 {
		limit = domain.MaxHistoryPerSubject
	}
	out := make([]domain.LocationHistory, 0)
	for i := len(r.S.history) - 1; i >= 0; i-- {
		h := r.S.history[i]
		if h.TenantID == tenantID && h.SubjectType == subjectType && h.SubjectID == subjectID {
			out = append(out, h)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

var _ ports.HistoryRepo = (*HistoryRepo)(nil)

// CacheRepo is an in-memory CacheRepo.
type CacheRepo struct{ S *Store }

func (r *CacheRepo) GetGeocode(_ context.Context, queryHash string) (domain.GeocodeResult, bool, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	e, ok := r.S.geocode[queryHash]
	if !ok {
		return domain.GeocodeResult{}, false, nil
	}
	if !e.ExpiresAt.IsZero() && time.Now().UTC().After(e.ExpiresAt) {
		return domain.GeocodeResult{}, false, nil
	}
	return e.Result, true, nil
}

func (r *CacheRepo) SetGeocode(_ context.Context, queryHash string, result domain.GeocodeResult, expiresAt time.Time) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.geocode[queryHash] = geocodeEntry{Result: result, ExpiresAt: expiresAt}
	return nil
}

func manifestKey(tenantID uuid.UUID, region string) string {
	return tenantID.String() + "|" + region
}

func (r *CacheRepo) GetOfflineManifest(_ context.Context, tenantID uuid.UUID, region string) (domain.OfflineManifest, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	m, ok := r.S.manifests[manifestKey(tenantID, region)]
	if !ok {
		return domain.OfflineManifest{}, fmt.Errorf("%w: offline manifest", domain.ErrNotFound)
	}
	return m, nil
}

func (r *CacheRepo) UpsertOfflineManifest(_ context.Context, m domain.OfflineManifest) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	key := manifestKey(m.TenantID, m.Region)
	if existing, ok := r.S.manifests[key]; ok {
		m.ID = existing.ID
	}
	r.S.manifests[key] = m
	return nil
}

var _ ports.CacheRepo = (*CacheRepo)(nil)

// HeatRepo is an in-memory HeatRepo.
type HeatRepo struct{ S *Store }

func heatKey(tenantID uuid.UUID, grid string) string {
	return tenantID.String() + "|" + grid
}

func (r *HeatRepo) Upsert(_ context.Context, c domain.HeatCell) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	key := heatKey(c.TenantID, c.GridCell)
	if existing, ok := r.S.heat[key]; ok {
		c.ID = existing.ID
	}
	r.S.heat[key] = c
	return nil
}

func (r *HeatRepo) List(_ context.Context, tenantID uuid.UUID, limit int) ([]domain.HeatCell, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	if limit <= 0 {
		limit = 100
	}
	out := make([]domain.HeatCell, 0)
	for _, c := range r.S.heat {
		if c.TenantID == tenantID {
			out = append(out, c)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

var _ ports.HeatRepo = (*HeatRepo)(nil)

// OutboxRepo is an in-memory OutboxRepository.
type OutboxRepo struct{ S *Store }

func (r *OutboxRepo) Enqueue(_ context.Context, m domain.OutboxMessage) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.outbox = append(r.S.outbox, m)
	return nil
}

func (r *OutboxRepo) Update(_ context.Context, m domain.OutboxMessage) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	for i := range r.S.outbox {
		if r.S.outbox[i].ID == m.ID {
			r.S.outbox[i] = m
			return nil
		}
	}
	return fmt.Errorf("%w: outbox", domain.ErrNotFound)
}

func (r *OutboxRepo) ListPending(_ context.Context, limit int) ([]domain.OutboxMessage, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	if limit <= 0 {
		limit = 100
	}
	out := make([]domain.OutboxMessage, 0)
	for _, m := range r.S.outbox {
		if m.Status == domain.OutboxStatusPending {
			out = append(out, m)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

var _ ports.OutboxRepository = (*OutboxRepo)(nil)
