package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/platform-ops-service/internal/app/ports"
	"github.com/nexora/platform-ops-service/internal/domain"
)

var _ ports.Registry = (*Registry)(nil)

// Registry is an in-process super-admin directory (tenants, companies, audit).
type Registry struct {
	mu         sync.RWMutex
	tenants    map[string]domain.PlatformTenant
	companies  map[string]domain.PlatformCompany
	proposals  map[string]domain.DualControlProposal
	audit      []domain.PlatformAuditEntry
	people     map[string]domain.PlatformPerson
}

func NewRegistry() *Registry {
	return &Registry{
		tenants:   map[string]domain.PlatformTenant{},
		companies: map[string]domain.PlatformCompany{},
		proposals: map[string]domain.DualControlProposal{},
		people:    map[string]domain.PlatformPerson{},
	}
}

func (r *Registry) ListTenants(_ context.Context) ([]domain.PlatformTenant, []domain.DualControlProposal, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]domain.PlatformTenant, 0, len(r.tenants))
	for _, t := range r.tenants {
		items = append(items, t)
	}
	props := make([]domain.DualControlProposal, 0, len(r.proposals))
	for _, p := range r.proposals {
		props = append(props, p)
	}
	return items, props, nil
}

func (r *Registry) GetTenant(_ context.Context, id string) (domain.PlatformTenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tenants[id]
	if !ok {
		return domain.PlatformTenant{}, domain.ErrNotFound
	}
	return t, nil
}

func (r *Registry) SaveTenant(_ context.Context, t domain.PlatformTenant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t.Slug != "" {
		for id, existing := range r.tenants {
			if existing.Slug == t.Slug && id != t.ID {
				return domain.ErrConflict
			}
		}
	}
	r.tenants[t.ID] = t
	if c, ok := r.companies[t.CompanyID]; ok {
		n := 0
		for _, x := range r.tenants {
			if x.CompanyID == t.CompanyID {
				n++
			}
		}
		c.TenantCount = n
		r.companies[t.CompanyID] = c
	}
	return nil
}

func (r *Registry) ListCompanies(_ context.Context) ([]domain.PlatformCompany, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.PlatformCompany, 0, len(r.companies))
	for _, c := range r.companies {
		out = append(out, c)
	}
	return out, nil
}

func (r *Registry) GetCompany(_ context.Context, id string) (domain.PlatformCompany, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.companies[id]
	if !ok {
		return domain.PlatformCompany{}, domain.ErrNotFound
	}
	return c, nil
}

func (r *Registry) SaveCompany(_ context.Context, c domain.PlatformCompany) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.companies[c.ID] = c
	return nil
}

func (r *Registry) DeleteCompany(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.companies[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.companies, id)
	return nil
}

func (r *Registry) SaveProposal(_ context.Context, p domain.DualControlProposal) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.proposals[p.ID] = p
	return nil
}

func (r *Registry) GetProposal(_ context.Context, id string) (domain.DualControlProposal, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.proposals[id]
	if !ok {
		return domain.DualControlProposal{}, domain.ErrNotFound
	}
	return p, nil
}

func (r *Registry) AppendAudit(_ context.Context, e domain.PlatformAuditEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.audit = append([]domain.PlatformAuditEntry{e}, r.audit...)
	if len(r.audit) > 500 {
		r.audit = r.audit[:500]
	}
	return nil
}

func (r *Registry) ListAudit(_ context.Context, q string) ([]domain.PlatformAuditEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	q = strings.ToLower(strings.TrimSpace(q))
	out := make([]domain.PlatformAuditEntry, 0, len(r.audit))
	for _, e := range r.audit {
		if q == "" || strings.Contains(strings.ToLower(e.ActorEmail+e.Action+e.Resource+e.IP+e.SessionID), q) {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *Registry) ListPeople(_ context.Context) ([]domain.PlatformPerson, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.PlatformPerson, 0, len(r.people))
	for _, p := range r.people {
		out = append(out, p)
	}
	return out, nil
}

func (r *Registry) SavePerson(_ context.Context, p domain.PlatformPerson) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.people[p.ID] = p
	return nil
}

func NewID() string { return uuid.NewString() }

func Now() time.Time { return time.Now().UTC() }
