package memory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/notification-service/internal/app/ports"
	"github.com/nexora/notification-service/internal/domain"
)

// Repos bundles all memory repositories.
type Repos struct {
	Templates   *TemplateRepo
	Messages    *MessageRepo
	Preferences *PreferenceRepo
	Devices     *DeviceRepo
	Inbox       *InboxRepo
	Schedules   *ScheduleRepo
	Deliveries  *DeliveryRepo
	Outbox      *OutboxRepo
}

// NewRepos wires repos to a shared store.
func NewRepos(s *Store) *Repos {
	return &Repos{
		Templates:   &TemplateRepo{S: s},
		Messages:    &MessageRepo{S: s},
		Preferences: &PreferenceRepo{S: s},
		Devices:     &DeviceRepo{S: s},
		Inbox:       &InboxRepo{S: s},
		Schedules:   &ScheduleRepo{S: s},
		Deliveries:  &DeliveryRepo{S: s},
		Outbox:      &OutboxRepo{S: s},
	}
}

// TemplateRepo is an in-memory TemplateRepo.
type TemplateRepo struct{ S *Store }

func (r *TemplateRepo) Upsert(_ context.Context, t domain.Template) (domain.Template, error) {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.templates[t.ID] = t
	r.S.templateIdx[tplKey(t.TenantID, t.Key, t.Channel, t.Locale)] = t.ID
	return t, nil
}

func (r *TemplateRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Template, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	t, ok := r.S.templates[id]
	if !ok || t.TenantID != tenantID {
		return domain.Template{}, domain.ErrNotFound
	}
	return t, nil
}

func (r *TemplateRepo) GetByKey(_ context.Context, tenantID uuid.UUID, key string, channel domain.Channel, locale string) (domain.Template, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.templateIdx[tplKey(tenantID, key, channel, locale)]
	if !ok {
		return domain.Template{}, domain.ErrNotFound
	}
	t, ok := r.S.templates[id]
	if !ok {
		return domain.Template{}, domain.ErrNotFound
	}
	return t, nil
}

func (r *TemplateRepo) Approve(_ context.Context, tenantID, id uuid.UUID, now time.Time) (domain.Template, error) {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	t, ok := r.S.templates[id]
	if !ok || t.TenantID != tenantID {
		return domain.Template{}, domain.ErrNotFound
	}
	t.Status = domain.TemplateActive
	t.UpdatedAt = now
	r.S.templates[id] = t
	return t, nil
}

func (r *TemplateRepo) List(_ context.Context, tenantID uuid.UUID, key string) ([]domain.Template, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Template
	for _, t := range r.S.templates {
		if t.TenantID != tenantID {
			continue
		}
		if key != "" && t.Key != key {
			continue
		}
		out = append(out, t)
	}
	return out, nil
}

var _ ports.TemplateRepo = (*TemplateRepo)(nil)

// MessageRepo is an in-memory MessageRepo.
type MessageRepo struct{ S *Store }

func (r *MessageRepo) Create(_ context.Context, m domain.Message) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if m.IdempotencyKey != "" {
		if _, ok := r.S.msgIdem[idemKey(m.TenantID, m.IdempotencyKey)]; ok {
			return domain.ErrIdempotencyConflict
		}
		r.S.msgIdem[idemKey(m.TenantID, m.IdempotencyKey)] = m.ID
	}
	r.S.messages[m.ID] = m
	return nil
}

func (r *MessageRepo) Update(_ context.Context, m domain.Message) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.messages[m.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.messages[m.ID] = m
	return nil
}

func (r *MessageRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Message, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	m, ok := r.S.messages[id]
	if !ok || m.TenantID != tenantID {
		return domain.Message{}, domain.ErrNotFound
	}
	return m, nil
}

func (r *MessageRepo) GetByIdempotency(_ context.Context, tenantID uuid.UUID, key string) (domain.Message, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.msgIdem[idemKey(tenantID, key)]
	if !ok {
		return domain.Message{}, domain.ErrNotFound
	}
	m, ok := r.S.messages[id]
	if !ok {
		return domain.Message{}, domain.ErrNotFound
	}
	return m, nil
}

func (r *MessageRepo) ListFailed(_ context.Context, tenantID uuid.UUID, limit int) ([]domain.Message, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Message
	for _, m := range r.S.messages {
		if m.TenantID == tenantID && m.Status == domain.MessageFailed {
			out = append(out, m)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *MessageRepo) CountByStatus(_ context.Context, tenantID uuid.UUID) (map[domain.MessageStatus]int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := map[domain.MessageStatus]int{}
	for _, m := range r.S.messages {
		if m.TenantID == tenantID {
			out[m.Status]++
		}
	}
	return out, nil
}

var _ ports.MessageRepo = (*MessageRepo)(nil)

// PreferenceRepo is an in-memory PreferenceRepo.
type PreferenceRepo struct{ S *Store }

func (r *PreferenceRepo) Get(_ context.Context, tenantID, principalID uuid.UUID) (domain.Preference, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	p, ok := r.S.preferences[prinKey(tenantID, principalID)]
	if !ok {
		return domain.Preference{}, domain.ErrNotFound
	}
	return clonePref(p), nil
}

func (r *PreferenceRepo) Upsert(_ context.Context, p domain.Preference) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.preferences[prinKey(p.TenantID, p.PrincipalID)] = clonePref(p)
	return nil
}

func (r *PreferenceRepo) RecordConsent(_ context.Context, c domain.Consent) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.consents = append(r.S.consents, c)
	return nil
}

func (r *PreferenceRepo) ListConsents(_ context.Context, tenantID, principalID uuid.UUID) ([]domain.Consent, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Consent
	for _, c := range r.S.consents {
		if c.TenantID == tenantID && c.PrincipalID == principalID {
			out = append(out, c)
		}
	}
	return out, nil
}

func clonePref(p domain.Preference) domain.Preference {
	cp := p
	cp.ChannelOptOut = map[domain.Channel]bool{}
	for k, v := range p.ChannelOptOut {
		cp.ChannelOptOut[k] = v
	}
	return cp
}

var _ ports.PreferenceRepo = (*PreferenceRepo)(nil)

// DeviceRepo is an in-memory DeviceRepo.
type DeviceRepo struct{ S *Store }

func (r *DeviceRepo) Upsert(_ context.Context, d domain.Device) (domain.Device, error) {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	tok := tenantKey(d.TenantID, d.Token)
	if existingID, ok := r.S.deviceToken[tok]; ok {
		ex := r.S.devices[existingID]
		ex.PrincipalID = d.PrincipalID
		ex.Platform = d.Platform
		ex.Locale = d.Locale
		ex.Active = true
		ex.UpdatedAt = d.UpdatedAt
		r.S.devices[existingID] = ex
		return ex, nil
	}
	r.S.devices[d.ID] = d
	r.S.deviceToken[tok] = d.ID
	return d, nil
}

func (r *DeviceRepo) ListActive(_ context.Context, tenantID, principalID uuid.UUID) ([]domain.Device, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Device
	for _, d := range r.S.devices {
		if d.TenantID == tenantID && d.PrincipalID == principalID && d.Active {
			out = append(out, d)
		}
	}
	return out, nil
}

func (r *DeviceRepo) Deactivate(_ context.Context, tenantID, id uuid.UUID) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	d, ok := r.S.devices[id]
	if !ok || d.TenantID != tenantID {
		return domain.ErrNotFound
	}
	d.Active = false
	r.S.devices[id] = d
	return nil
}

var _ ports.DeviceRepo = (*DeviceRepo)(nil)

// InboxRepo is an in-memory InboxRepo.
type InboxRepo struct{ S *Store }

func (r *InboxRepo) Create(_ context.Context, item domain.InboxItem) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.inbox[item.ID] = item
	return nil
}

func (r *InboxRepo) List(_ context.Context, tenantID, principalID uuid.UUID, includeArchived bool) ([]domain.InboxItem, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.InboxItem
	for _, it := range r.S.inbox {
		if it.TenantID != tenantID || it.PrincipalID != principalID {
			continue
		}
		if it.Archived && !includeArchived {
			continue
		}
		out = append(out, it)
	}
	return out, nil
}

func (r *InboxRepo) MarkRead(_ context.Context, tenantID, id uuid.UUID, now time.Time) (domain.InboxItem, error) {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	it, ok := r.S.inbox[id]
	if !ok || it.TenantID != tenantID {
		return domain.InboxItem{}, domain.ErrNotFound
	}
	it.Read = true
	it.ReadAt = &now
	r.S.inbox[id] = it
	return it, nil
}

func (r *InboxRepo) Archive(_ context.Context, tenantID, id uuid.UUID) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	it, ok := r.S.inbox[id]
	if !ok || it.TenantID != tenantID {
		return domain.ErrNotFound
	}
	it.Archived = true
	r.S.inbox[id] = it
	return nil
}

var _ ports.InboxRepo = (*InboxRepo)(nil)

// ScheduleRepo is an in-memory ScheduleRepo.
type ScheduleRepo struct{ S *Store }

func (r *ScheduleRepo) Create(_ context.Context, s domain.Schedule) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.schedules[s.ID] = s
	return nil
}

func (r *ScheduleRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Schedule, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	s, ok := r.S.schedules[id]
	if !ok || s.TenantID != tenantID {
		return domain.Schedule{}, domain.ErrNotFound
	}
	return s, nil
}

func (r *ScheduleRepo) Update(_ context.Context, s domain.Schedule) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.schedules[s.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.schedules[s.ID] = s
	return nil
}

func (r *ScheduleRepo) ListDue(_ context.Context, now time.Time, limit int) ([]domain.Schedule, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Schedule
	for _, s := range r.S.schedules {
		if s.Status == domain.SchedulePending && !s.SendAt.After(now) {
			out = append(out, s)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *ScheduleRepo) Cancel(_ context.Context, tenantID, id uuid.UUID, now time.Time) (domain.Schedule, error) {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	s, ok := r.S.schedules[id]
	if !ok || s.TenantID != tenantID {
		return domain.Schedule{}, domain.ErrNotFound
	}
	if s.Status != domain.SchedulePending {
		return domain.Schedule{}, domain.ErrConflict
	}
	s.Status = domain.ScheduleCancelled
	s.UpdatedAt = now
	r.S.schedules[id] = s
	return s, nil
}

var _ ports.ScheduleRepo = (*ScheduleRepo)(nil)

// DeliveryRepo is an in-memory DeliveryRepo.
type DeliveryRepo struct{ S *Store }

func (r *DeliveryRepo) CreateAttempt(_ context.Context, a domain.DeliveryAttempt) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.attempts = append(r.S.attempts, a)
	return nil
}

func (r *DeliveryRepo) ListAttempts(_ context.Context, tenantID, messageID uuid.UUID) ([]domain.DeliveryAttempt, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.DeliveryAttempt
	for _, a := range r.S.attempts {
		if a.TenantID == tenantID && a.MessageID == messageID {
			out = append(out, a)
		}
	}
	return out, nil
}

func (r *DeliveryRepo) CreateEvent(_ context.Context, e domain.DeliveryEvent) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.events = append(r.S.events, e)
	return nil
}

func (r *DeliveryRepo) MoveToDLQ(_ context.Context, item domain.DLQItem) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.dlq = append(r.S.dlq, item)
	return nil
}

func (r *DeliveryRepo) ListDLQ(_ context.Context, tenantID uuid.UUID, limit int) ([]domain.DLQItem, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.DLQItem
	for _, item := range r.S.dlq {
		if item.TenantID == tenantID {
			out = append(out, item)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *DeliveryRepo) UpsertRoute(_ context.Context, route domain.ProviderRoute) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.routes[route.ID] = route
	return nil
}

func (r *DeliveryRepo) ListRoutes(_ context.Context, tenantID uuid.UUID) ([]domain.ProviderRoute, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.ProviderRoute
	for _, route := range r.S.routes {
		if route.TenantID == tenantID {
			out = append(out, route)
		}
	}
	return out, nil
}

var _ ports.DeliveryRepo = (*DeliveryRepo)(nil)

// OutboxRepo is an in-memory OutboxRepository.
type OutboxRepo struct{ S *Store }

func (r *OutboxRepo) Enqueue(_ context.Context, m domain.OutboxMessage) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.outbox = append(r.S.outbox, m)
	return nil
}

func (r *OutboxRepo) ListPending(_ context.Context, limit int) ([]domain.OutboxMessage, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.OutboxMessage
	for _, m := range r.S.outbox {
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
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	for i := range r.S.outbox {
		if r.S.outbox[i].ID == m.ID {
			r.S.outbox[i] = m
			return nil
		}
	}
	return domain.ErrNotFound
}

var _ ports.OutboxRepository = (*OutboxRepo)(nil)
