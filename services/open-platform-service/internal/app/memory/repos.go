package memory

import (
	"context"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/nexora/open-platform-service/internal/domain"
)

type Store struct {
	mu           sync.RWMutex
	Apps         map[uuid.UUID]domain.DeveloperApp
	Keys         map[uuid.UUID]domain.ApiKey
	Catalog      map[string]domain.CatalogEntry
	Versions     map[uuid.UUID]domain.ApiVersion
	Policies     map[uuid.UUID]domain.GatewayPolicy
	Webhooks     map[uuid.UUID]domain.WebhookEndpoint
	Deliveries   map[uuid.UUID]domain.WebhookDelivery
	SDKs         map[uuid.UUID]domain.SdkPackage
	Integrations map[uuid.UUID]domain.IntegrationConnector
	Usage        map[string]domain.UsageCounter
	Outbox       map[uuid.UUID]domain.OutboxMessage
}

func NewStore() *Store {
	return &Store{
		Apps: map[uuid.UUID]domain.DeveloperApp{}, Keys: map[uuid.UUID]domain.ApiKey{},
		Catalog: map[string]domain.CatalogEntry{}, Versions: map[uuid.UUID]domain.ApiVersion{},
		Policies: map[uuid.UUID]domain.GatewayPolicy{}, Webhooks: map[uuid.UUID]domain.WebhookEndpoint{},
		Deliveries: map[uuid.UUID]domain.WebhookDelivery{}, SDKs: map[uuid.UUID]domain.SdkPackage{},
		Integrations: map[uuid.UUID]domain.IntegrationConnector{}, Usage: map[string]domain.UsageCounter{},
		Outbox: map[uuid.UUID]domain.OutboxMessage{},
	}
}

func ck(tenantID uuid.UUID, key string) string { return tenantID.String() + ":" + key }
func uk(tenantID, appID uuid.UUID, day string) string {
	return tenantID.String() + ":" + appID.String() + ":" + day
}

type Repos struct {
	Apps         *AppRepo
	Keys         *KeyRepo
	Catalog      *CatalogRepo
	Versions     *VersionRepo
	Policies     *PolicyRepo
	Webhooks     *WebhookRepo
	Deliveries   *DeliveryRepo
	SDKs         *SdkRepo
	Integrations *IntegrationRepo
	Usage        *UsageRepo
	Outbox       *OutboxRepo
	HTTP         *MockHTTP
	Identity     *MockIdentity
	Metrics      *MockMetrics
}

func NewRepos(s *Store) *Repos {
	return &Repos{
		Apps: &AppRepo{s: s}, Keys: &KeyRepo{s: s}, Catalog: &CatalogRepo{s: s},
		Versions: &VersionRepo{s: s}, Policies: &PolicyRepo{s: s}, Webhooks: &WebhookRepo{s: s},
		Deliveries: &DeliveryRepo{s: s}, SDKs: &SdkRepo{s: s}, Integrations: &IntegrationRepo{s: s},
		Usage: &UsageRepo{s: s}, Outbox: &OutboxRepo{s: s},
		HTTP: &MockHTTP{}, Identity: &MockIdentity{}, Metrics: &MockMetrics{},
	}
}

type AppRepo struct{ s *Store }

func (r *AppRepo) Save(_ context.Context, a domain.DeveloperApp) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Apps[a.ID] = a
	return nil
}
func (r *AppRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.DeveloperApp, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	a, ok := r.s.Apps[id]
	if !ok || a.TenantID != tenantID {
		return domain.DeveloperApp{}, domain.ErrNotFound
	}
	return a, nil
}
func (r *AppRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.DeveloperApp, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.DeveloperApp{}
	for _, a := range r.s.Apps {
		if a.TenantID == tenantID {
			out = append(out, a)
		}
	}
	return out, nil
}

type KeyRepo struct{ s *Store }

func (r *KeyRepo) Save(_ context.Context, k domain.ApiKey) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Keys[k.ID] = k
	return nil
}
func (r *KeyRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.ApiKey, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	k, ok := r.s.Keys[id]
	if !ok || k.TenantID != tenantID {
		return domain.ApiKey{}, domain.ErrNotFound
	}
	return k, nil
}
func (r *KeyRepo) ListByApp(_ context.Context, tenantID, appID uuid.UUID) ([]domain.ApiKey, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.ApiKey{}
	for _, k := range r.s.Keys {
		if k.TenantID == tenantID && k.AppID == appID {
			out = append(out, k)
		}
	}
	return out, nil
}

type CatalogRepo struct{ s *Store }

func (r *CatalogRepo) Save(_ context.Context, e domain.CatalogEntry) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Catalog[ck(e.TenantID, e.Key)] = e
	return nil
}
func (r *CatalogRepo) GetByKey(_ context.Context, tenantID uuid.UUID, key string) (domain.CatalogEntry, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	e, ok := r.s.Catalog[ck(tenantID, key)]
	if !ok {
		return domain.CatalogEntry{}, domain.ErrNotFound
	}
	return e, nil
}
func (r *CatalogRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.CatalogEntry, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.CatalogEntry{}
	for _, e := range r.s.Catalog {
		if e.TenantID == tenantID {
			out = append(out, e)
		}
	}
	return out, nil
}

type VersionRepo struct{ s *Store }

func (r *VersionRepo) Save(_ context.Context, v domain.ApiVersion) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Versions[v.ID] = v
	return nil
}
func (r *VersionRepo) ListByCatalog(_ context.Context, tenantID uuid.UUID, catalogKey string) ([]domain.ApiVersion, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.ApiVersion{}
	for _, v := range r.s.Versions {
		if v.TenantID == tenantID && v.CatalogKey == catalogKey {
			out = append(out, v)
		}
	}
	return out, nil
}

type PolicyRepo struct{ s *Store }

func (r *PolicyRepo) Save(_ context.Context, p domain.GatewayPolicy) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Policies[p.ID] = p
	return nil
}
func (r *PolicyRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.GatewayPolicy, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.GatewayPolicy{}
	for _, p := range r.s.Policies {
		if p.TenantID == tenantID {
			out = append(out, p)
		}
	}
	return out, nil
}

type WebhookRepo struct{ s *Store }

func (r *WebhookRepo) Save(_ context.Context, w domain.WebhookEndpoint) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Webhooks[w.ID] = w
	return nil
}
func (r *WebhookRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.WebhookEndpoint, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	w, ok := r.s.Webhooks[id]
	if !ok || w.TenantID != tenantID {
		return domain.WebhookEndpoint{}, domain.ErrNotFound
	}
	return w, nil
}
func (r *WebhookRepo) ListByApp(_ context.Context, tenantID, appID uuid.UUID) ([]domain.WebhookEndpoint, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.WebhookEndpoint{}
	for _, w := range r.s.Webhooks {
		if w.TenantID == tenantID && w.AppID == appID {
			out = append(out, w)
		}
	}
	return out, nil
}
func (r *WebhookRepo) ListActiveForEvent(_ context.Context, tenantID uuid.UUID, eventType string) ([]domain.WebhookEndpoint, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.WebhookEndpoint{}
	for _, w := range r.s.Webhooks {
		if w.TenantID != tenantID || !w.Active {
			continue
		}
		for _, e := range w.Events {
			if e == eventType || e == "*" || (strings.HasSuffix(e, ".*") && strings.HasPrefix(eventType, strings.TrimSuffix(e, ".*"))) {
				out = append(out, w)
				break
			}
		}
	}
	return out, nil
}

type DeliveryRepo struct{ s *Store }

func (r *DeliveryRepo) Save(_ context.Context, d domain.WebhookDelivery) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Deliveries[d.ID] = d
	return nil
}
func (r *DeliveryRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.WebhookDelivery, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	d, ok := r.s.Deliveries[id]
	if !ok || d.TenantID != tenantID {
		return domain.WebhookDelivery{}, domain.ErrNotFound
	}
	return d, nil
}
func (r *DeliveryRepo) ListPending(_ context.Context, limit int) ([]domain.WebhookDelivery, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.WebhookDelivery{}
	for _, d := range r.s.Deliveries {
		if d.Status == domain.DeliveryPending {
			out = append(out, d)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}
func (r *DeliveryRepo) ListByEndpoint(_ context.Context, tenantID, endpointID uuid.UUID) ([]domain.WebhookDelivery, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.WebhookDelivery{}
	for _, d := range r.s.Deliveries {
		if d.TenantID == tenantID && d.EndpointID == endpointID {
			out = append(out, d)
		}
	}
	return out, nil
}

type SdkRepo struct{ s *Store }

func (r *SdkRepo) Save(_ context.Context, s domain.SdkPackage) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.SDKs[s.ID] = s
	return nil
}
func (r *SdkRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.SdkPackage, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.SdkPackage{}
	for _, s := range r.s.SDKs {
		if s.TenantID == tenantID {
			out = append(out, s)
		}
	}
	return out, nil
}

type IntegrationRepo struct{ s *Store }

func (r *IntegrationRepo) Save(_ context.Context, i domain.IntegrationConnector) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Integrations[i.ID] = i
	return nil
}
func (r *IntegrationRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.IntegrationConnector, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.IntegrationConnector{}
	for _, i := range r.s.Integrations {
		if i.TenantID == tenantID {
			out = append(out, i)
		}
	}
	return out, nil
}

type UsageRepo struct{ s *Store }

func (r *UsageRepo) Save(_ context.Context, u domain.UsageCounter) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Usage[uk(u.TenantID, u.AppID, u.Day)] = u
	return nil
}
func (r *UsageRepo) Get(_ context.Context, tenantID, appID uuid.UUID, day string) (domain.UsageCounter, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	u, ok := r.s.Usage[uk(tenantID, appID, day)]
	if !ok {
		return domain.UsageCounter{}, domain.ErrNotFound
	}
	return u, nil
}
func (r *UsageRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.UsageCounter, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.UsageCounter{}
	for _, u := range r.s.Usage {
		if u.TenantID == tenantID {
			out = append(out, u)
		}
	}
	return out, nil
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

type MockHTTP struct{}

func (MockHTTP) Post(context.Context, string, map[string]string, []byte) (int, error) {
	return 200, nil
}

type MockIdentity struct{}

func (MockIdentity) RegisterOAuthClient(_ context.Context, _ uuid.UUID, name string, _ []string) (string, error) {
	return "oauth_" + strings.ReplaceAll(strings.ToLower(name), " ", "_"), nil
}

type MockMetrics struct{}

func (MockMetrics) Record(context.Context, string, map[string]string, float64) error { return nil }
