package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/recommendation-service/internal/domain"
)

type Store struct {
	mu       sync.RWMutex
	Features map[uuid.UUID]domain.ProductFeatures
	Signals  []domain.BehaviorSignal
	CoOccur  map[string]domain.CoOccurrence
	Outbox   []domain.OutboxMessage
}

func NewStore() *Store {
	return &Store{
		Features: make(map[uuid.UUID]domain.ProductFeatures),
		CoOccur:  make(map[string]domain.CoOccurrence),
	}
}

func pairKey(tenantID, a, b uuid.UUID) string {
	x, y := a.String(), b.String()
	if x > y {
		x, y = y, x
	}
	return tenantID.String() + "|" + x + "|" + y
}

type Repos struct {
	Features *FeatureRepo
	Signals  *SignalRepo
	CoOccur  *CoOccurRepo
	Outbox   *OutboxRepo
}

func NewRepos(s *Store) *Repos {
	return &Repos{
		Features: &FeatureRepo{s: s}, Signals: &SignalRepo{s: s},
		CoOccur: &CoOccurRepo{s: s}, Outbox: &OutboxRepo{s: s},
	}
}

type FeatureRepo struct{ s *Store }

func (r *FeatureRepo) Upsert(_ context.Context, f domain.ProductFeatures) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Features[f.ProductID] = f
	return nil
}

func (r *FeatureRepo) Get(_ context.Context, productID uuid.UUID) (domain.ProductFeatures, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	f, ok := r.s.Features[productID]
	if !ok {
		return domain.ProductFeatures{}, domain.ErrNotFound
	}
	return f, nil
}

func (r *FeatureRepo) List(_ context.Context, ids []uuid.UUID) ([]domain.ProductFeatures, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.ProductFeatures, 0, len(ids))
	for _, id := range ids {
		if f, ok := r.s.Features[id]; ok {
			out = append(out, f)
		}
	}
	return out, nil
}

func (r *FeatureRepo) ListAll(_ context.Context, limit int) ([]domain.ProductFeatures, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.ProductFeatures, 0, len(r.s.Features))
	for _, f := range r.s.Features {
		out = append(out, f)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

type SignalRepo struct{ s *Store }

func (r *SignalRepo) Save(_ context.Context, s domain.BehaviorSignal) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Signals = append(r.s.Signals, s)
	return nil
}

func (r *SignalRepo) ListByUser(_ context.Context, tenantID, userID uuid.UUID, limit int) ([]domain.BehaviorSignal, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.BehaviorSignal, 0)
	for i := len(r.s.Signals) - 1; i >= 0; i-- {
		s := r.s.Signals[i]
		if s.TenantID == tenantID && s.UserID == userID {
			out = append(out, s)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *SignalRepo) UsersWhoInteracted(_ context.Context, tenantID, productID uuid.UUID, limit int) ([]uuid.UUID, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	seen := map[uuid.UUID]struct{}{}
	out := make([]uuid.UUID, 0)
	for _, s := range r.s.Signals {
		if s.TenantID == tenantID && s.ProductID == productID {
			if _, ok := seen[s.UserID]; ok {
				continue
			}
			seen[s.UserID] = struct{}{}
			out = append(out, s.UserID)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

type CoOccurRepo struct{ s *Store }

func (r *CoOccurRepo) Bump(_ context.Context, tenantID, a, b uuid.UUID, delta int, now time.Time) error {
	if a == b {
		return nil
	}
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	k := pairKey(tenantID, a, b)
	c := r.s.CoOccur[k]
	c.TenantID = tenantID
	c.ProductA, c.ProductB = a, b
	if a.String() > b.String() {
		c.ProductA, c.ProductB = b, a
	}
	c.Count += delta
	c.UpdatedAt = now
	r.s.CoOccur[k] = c
	return nil
}

func (r *CoOccurRepo) TopFor(_ context.Context, tenantID, productID uuid.UUID, limit int) ([]domain.CoOccurrence, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.CoOccurrence, 0)
	for _, c := range r.s.CoOccur {
		if c.TenantID == tenantID && (c.ProductA == productID || c.ProductB == productID) {
			out = append(out, c)
		}
	}
	// simple selection sort by count
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Count > out[i].Count {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
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

type MockCatalog struct{}

func (MockCatalog) Features(_ context.Context, _, productID uuid.UUID) (domain.ProductFeatures, error) {
	return domain.ProductFeatures{ProductID: productID, Popularity: 1}, nil
}

type MockTrends struct {
	IDs []uuid.UUID
}

func (m *MockTrends) TrendingProductIDs(_ context.Context, _ uuid.UUID, limit int) ([]uuid.UUID, error) {
	if limit > len(m.IDs) {
		limit = len(m.IDs)
	}
	return m.IDs[:limit], nil
}
