package memory

import (
	"context"
	"math"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/nexora/ai-platform-service/internal/domain"
)

type Store struct {
	mu       sync.RWMutex
	Features map[string]domain.FeatureRecord
	Models   map[string]domain.ModelCard // tenant|key|version
	Prompts  []domain.PromptTemplate
	Memory   []domain.ConversationMemory
	AgentRuns []domain.AgentRun
	Rules    []domain.AutomationRule
	AutoRuns []domain.AutomationRun
	Drifts   []domain.DriftReport
	Outbox   []domain.OutboxMessage
}

func NewStore() *Store {
	return &Store{Features: make(map[string]domain.FeatureRecord), Models: make(map[string]domain.ModelCard)}
}

func featKey(tenantID uuid.UUID, entityType string, entityID uuid.UUID, name string, version int) string {
	return tenantID.String() + "|" + entityType + "|" + entityID.String() + "|" + name + "|" + itoa(version)
}

func modelKey(tenantID uuid.UUID, key, version string) string {
	return tenantID.String() + "|" + key + "|" + version
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

type Repos struct {
	Features   *FeatureRepo
	Models     *ModelRepo
	Prompts    *PromptRepo
	Memory     *MemoryRepo
	Agents     *AgentRepo
	Automation *AutomationRepo
	Drift      *DriftRepo
	Outbox     *OutboxRepo
}

func NewRepos(s *Store) *Repos {
	return &Repos{
		Features: &FeatureRepo{s: s}, Models: &ModelRepo{s: s}, Prompts: &PromptRepo{s: s},
		Memory: &MemoryRepo{s: s}, Agents: &AgentRepo{s: s}, Automation: &AutomationRepo{s: s},
		Drift: &DriftRepo{s: s}, Outbox: &OutboxRepo{s: s},
	}
}

type FeatureRepo struct{ s *Store }

func (r *FeatureRepo) Upsert(_ context.Context, f domain.FeatureRecord) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Features[featKey(f.TenantID, f.EntityType, f.EntityID, f.Name, f.Version)] = f
	return nil
}

func (r *FeatureRepo) Get(_ context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID, name string, version int) (domain.FeatureRecord, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	f, ok := r.s.Features[featKey(tenantID, entityType, entityID, name, version)]
	if !ok {
		return domain.FeatureRecord{}, domain.ErrNotFound
	}
	return f, nil
}

func (r *FeatureRepo) ListByEntity(_ context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID) ([]domain.FeatureRecord, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	prefix := tenantID.String() + "|" + entityType + "|" + entityID.String() + "|"
	out := make([]domain.FeatureRecord, 0)
	for k, f := range r.s.Features {
		if strings.HasPrefix(k, prefix) {
			out = append(out, f)
		}
	}
	return out, nil
}

type ModelRepo struct{ s *Store }

func (r *ModelRepo) Save(_ context.Context, m domain.ModelCard) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Models[modelKey(m.TenantID, m.Key, m.Version)] = m
	return nil
}

func (r *ModelRepo) Get(_ context.Context, tenantID uuid.UUID, key, version string) (domain.ModelCard, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	m, ok := r.s.Models[modelKey(tenantID, key, version)]
	if !ok {
		return domain.ModelCard{}, domain.ErrNotFound
	}
	return m, nil
}

func (r *ModelRepo) GetProd(_ context.Context, tenantID uuid.UUID, key string) (domain.ModelCard, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var found domain.ModelCard
	ok := false
	for _, m := range r.s.Models {
		if m.TenantID == tenantID && m.Key == key && m.Stage == domain.StageProd {
			if !ok || m.UpdatedAt.After(found.UpdatedAt) {
				found = m
				ok = true
			}
		}
	}
	if !ok {
		return domain.ModelCard{}, domain.ErrNotFound
	}
	return found, nil
}

func (r *ModelRepo) List(_ context.Context, tenantID uuid.UUID, key string) ([]domain.ModelCard, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.ModelCard, 0)
	for _, m := range r.s.Models {
		if m.TenantID == tenantID && (key == "" || m.Key == key) {
			out = append(out, m)
		}
	}
	return out, nil
}

type PromptRepo struct{ s *Store }

func (r *PromptRepo) Save(_ context.Context, p domain.PromptTemplate) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Prompts {
		if r.s.Prompts[i].ID == p.ID {
			r.s.Prompts[i] = p
			return nil
		}
	}
	r.s.Prompts = append(r.s.Prompts, p)
	return nil
}

func (r *PromptRepo) GetActive(_ context.Context, tenantID uuid.UUID, key, locale string) (domain.PromptTemplate, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, p := range r.s.Prompts {
		if p.TenantID == tenantID && p.Key == key && p.Active && (locale == "" || p.Locale == "" || p.Locale == locale) {
			return p, nil
		}
	}
	return domain.PromptTemplate{}, domain.ErrNotFound
}

func (r *PromptRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.PromptTemplate, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.PromptTemplate, 0)
	for _, p := range r.s.Prompts {
		if p.TenantID == tenantID {
			out = append(out, p)
		}
	}
	return out, nil
}

type MemoryRepo struct{ s *Store }

func (r *MemoryRepo) Append(_ context.Context, m domain.ConversationMemory) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Memory = append(r.s.Memory, m)
	return nil
}

func (r *MemoryRepo) ListSession(_ context.Context, tenantID, sessionID uuid.UUID, limit int) ([]domain.ConversationMemory, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.ConversationMemory, 0)
	for _, m := range r.s.Memory {
		if m.TenantID == tenantID && m.SessionID == sessionID {
			out = append(out, m)
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out, nil
}

type AgentRepo struct{ s *Store }

func (r *AgentRepo) SaveRun(_ context.Context, run domain.AgentRun) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.AgentRuns = append(r.s.AgentRuns, run)
	return nil
}

func (r *AgentRepo) GetRun(_ context.Context, tenantID, id uuid.UUID) (domain.AgentRun, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, run := range r.s.AgentRuns {
		if run.TenantID == tenantID && run.ID == id {
			return run, nil
		}
	}
	return domain.AgentRun{}, domain.ErrNotFound
}

func (r *AgentRepo) ListRuns(_ context.Context, tenantID uuid.UUID, kind string, limit int) ([]domain.AgentRun, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.AgentRun, 0)
	for i := len(r.s.AgentRuns) - 1; i >= 0; i-- {
		run := r.s.AgentRuns[i]
		if run.TenantID == tenantID && (kind == "" || run.Kind == kind) {
			out = append(out, run)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

type AutomationRepo struct{ s *Store }

func (r *AutomationRepo) SaveRule(_ context.Context, rule domain.AutomationRule) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Rules {
		if r.s.Rules[i].ID == rule.ID {
			r.s.Rules[i] = rule
			return nil
		}
	}
	r.s.Rules = append(r.s.Rules, rule)
	return nil
}

func (r *AutomationRepo) ListRules(_ context.Context, tenantID uuid.UUID) ([]domain.AutomationRule, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.AutomationRule, 0)
	for _, rule := range r.s.Rules {
		if rule.TenantID == tenantID {
			out = append(out, rule)
		}
	}
	return out, nil
}

func (r *AutomationRepo) SaveRun(_ context.Context, run domain.AutomationRun) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.AutoRuns = append(r.s.AutoRuns, run)
	return nil
}

type DriftRepo struct{ s *Store }

func (r *DriftRepo) Save(_ context.Context, d domain.DriftReport) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Drifts = append(r.s.Drifts, d)
	return nil
}

func (r *DriftRepo) List(_ context.Context, tenantID uuid.UUID, modelKey string, limit int) ([]domain.DriftReport, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.DriftReport, 0)
	for i := len(r.s.Drifts) - 1; i >= 0; i-- {
		d := r.s.Drifts[i]
		if d.TenantID == tenantID && (modelKey == "" || d.ModelKey == modelKey) {
			out = append(out, d)
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

// HeuristicRuntime implements InferenceRuntime without GPU.
type HeuristicRuntime struct{}

func (HeuristicRuntime) Predict(_ context.Context, model domain.ModelCard, features map[string]float64, inputs map[string]any) (map[string]float64, map[string]any, error) {
	preds := map[string]float64{}
	outs := map[string]any{"framework": model.Framework, "artifact": model.ArtifactURI}
	switch model.Key {
	case "demand_forecast":
		base := features["avg_daily_sales"]
		if base == 0 {
			base = 20
		}
		h := features["horizon_days"]
		if h == 0 {
			h = 7
		}
		preds["units"] = base * h * (1 + 0.05*features["promo_lift"])
		preds["confidence"] = 0.75
	case "fraud_score":
		score := 0.1*features["velocity"] + 0.3*features["device_risk"] + 0.4*features["amount_z"]
		if score > 1 {
			score = 1
		}
		preds["fraud_probability"] = score
		preds["risk"] = score
	case "pricing_suggest":
		cost := features["unit_cost"]
		if cost == 0 {
			cost = 10
		}
		preds["suggested_price"] = cost * (1.2 + 0.1*features["demand_index"])
		preds["elasticity"] = 1.1
		outs["humanGated"] = true
	case "embed_text":
		// not used here
		preds["ok"] = 1
	default:
		var sum float64
		for _, v := range features {
			sum += v
		}
		preds["score"] = math.Tanh(sum / 10)
	}
	_ = inputs
	return preds, outs, nil
}

// MockLLM is a deterministic provider.
type MockLLM struct{}

func (MockLLM) Complete(_ context.Context, provider, system, user string, tools []string) (string, []domain.ToolCall, int, int, error) {
	content := "[" + provider + "] " + strings.TrimSpace(system[:min(40, len(system))]) + " → " + strings.TrimSpace(user)
	if len(content) > 400 {
		content = content[:400]
	}
	var tcs []domain.ToolCall
	if len(tools) > 0 && strings.Contains(strings.ToLower(user), "order") {
		tcs = append(tcs, domain.ToolCall{Name: tools[0], Arguments: map[string]any{"q": "order"}})
	}
	return content, tcs, len(system)/4 + 1, len(content)/4 + 1, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MockRAG returns canned grounding.
type MockRAG struct{}

func (MockRAG) Retrieve(_ context.Context, _ uuid.UUID, query string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 3
	}
	out := []string{"Policy: refunds within 24h via CRM.", "Catalog note: freshness SLA 15m.", "Query: " + query}
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// MockEmbed hash-like embedding.
type MockEmbed struct{}

func (MockEmbed) Embed(_ context.Context, _ uuid.UUID, text string) ([]float64, error) {
	v := make([]float64, 16)
	for i, r := range text {
		v[i%16] += float64(r%7) / 7
	}
	var n float64
	for _, x := range v {
		n += x * x
	}
	n = math.Sqrt(n)
	if n > 0 {
		for i := range v {
			v[i] /= n
		}
	}
	return v, nil
}
