package memory

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/search-service/internal/domain"
)

type Repos struct {
	Docs     *DocumentRepo
	Synonyms *SynonymRepo
	Boosts   *BoostRepo
	Jobs     *IndexJobRepo
	Trends   *TrendRepo
	Suggests *SuggestRepo
	Outbox   *OutboxRepo
	Lexical  *LexicalIndex
	Vectors  *VectorStore
}

func NewRepos(s *Store) *Repos {
	return &Repos{
		Docs: &DocumentRepo{s: s}, Synonyms: &SynonymRepo{s: s}, Boosts: &BoostRepo{s: s},
		Jobs: &IndexJobRepo{s: s}, Trends: &TrendRepo{s: s}, Suggests: &SuggestRepo{s: s},
		Outbox: &OutboxRepo{s: s}, Lexical: &LexicalIndex{s: s}, Vectors: &VectorStore{s: s},
	}
}

type DocumentRepo struct{ s *Store }

func (r *DocumentRepo) Upsert(_ context.Context, d domain.ProductDocument) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Docs[docKey(d.TenantID, d.ProductID)] = d
	return nil
}

func (r *DocumentRepo) Get(_ context.Context, tenantID, productID uuid.UUID) (domain.ProductDocument, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	d, ok := r.s.Docs[docKey(tenantID, productID)]
	if !ok {
		return domain.ProductDocument{}, domain.ErrNotFound
	}
	return d, nil
}

func (r *DocumentRepo) List(_ context.Context, tenantID uuid.UUID, limit int) ([]domain.ProductDocument, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.ProductDocument, 0)
	for _, d := range r.s.Docs {
		if d.TenantID == tenantID {
			out = append(out, d)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *DocumentRepo) Delete(_ context.Context, tenantID, productID uuid.UUID) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	delete(r.s.Docs, docKey(tenantID, productID))
	return nil
}

type SynonymRepo struct{ s *Store }

func (r *SynonymRepo) Save(_ context.Context, s domain.SynonymRule) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Synonyms {
		if r.s.Synonyms[i].ID == s.ID {
			r.s.Synonyms[i] = s
			return nil
		}
	}
	r.s.Synonyms = append(r.s.Synonyms, s)
	return nil
}

func (r *SynonymRepo) List(_ context.Context, tenantID uuid.UUID, locale string) ([]domain.SynonymRule, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.SynonymRule, 0)
	for _, s := range r.s.Synonyms {
		if s.TenantID == tenantID && (locale == "" || s.Locale == "" || s.Locale == locale) {
			out = append(out, s)
		}
	}
	return out, nil
}

type BoostRepo struct{ s *Store }

func (r *BoostRepo) Save(_ context.Context, b domain.BoostRule) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Boosts {
		if r.s.Boosts[i].ID == b.ID {
			r.s.Boosts[i] = b
			return nil
		}
	}
	r.s.Boosts = append(r.s.Boosts, b)
	return nil
}

func (r *BoostRepo) ListActive(_ context.Context, tenantID uuid.UUID, now time.Time) ([]domain.BoostRule, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.BoostRule, 0)
	for _, b := range r.s.Boosts {
		if b.TenantID != tenantID || !b.Active {
			continue
		}
		if b.StartsAt != nil && now.Before(*b.StartsAt) {
			continue
		}
		if b.EndsAt != nil && now.After(*b.EndsAt) {
			continue
		}
		out = append(out, b)
	}
	return out, nil
}

type IndexJobRepo struct{ s *Store }

func (r *IndexJobRepo) Save(_ context.Context, j domain.IndexJob) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Jobs[j.ID] = j
	return nil
}

func (r *IndexJobRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.IndexJob, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	j, ok := r.s.Jobs[id]
	if !ok || j.TenantID != tenantID {
		return domain.IndexJob{}, domain.ErrNotFound
	}
	return j, nil
}

func (r *IndexJobRepo) List(_ context.Context, tenantID uuid.UUID, limit int) ([]domain.IndexJob, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.IndexJob, 0)
	for _, j := range r.s.Jobs {
		if j.TenantID == tenantID {
			out = append(out, j)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

type TrendRepo struct{ s *Store }

func (r *TrendRepo) Save(_ context.Context, t domain.TrendEntry) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Trends[trendKey(t.TenantID, t.Kind, t.Key)] = t
	return nil
}

func (r *TrendRepo) List(_ context.Context, tenantID uuid.UUID, kind string, limit int) ([]domain.TrendEntry, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.TrendEntry, 0)
	for _, t := range r.s.Trends {
		if t.TenantID == tenantID && (kind == "" || t.Kind == kind) {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (r *TrendRepo) Bump(_ context.Context, tenantID uuid.UUID, kind, key string, entityID *uuid.UUID, delta float64, now time.Time) error {
	if key == "" {
		return nil
	}
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	k := trendKey(tenantID, kind, key)
	t := r.s.Trends[k]
	t.TenantID = tenantID
	t.Kind = kind
	t.Key = key
	t.EntityID = entityID
	t.Score += delta
	t.UpdatedAt = now
	r.s.Trends[k] = t
	return nil
}

type SuggestRepo struct{ s *Store }

func (r *SuggestRepo) Upsert(_ context.Context, tenantID uuid.UUID, c domain.SuggestCandidate) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	list := r.s.TenantSug[tenantID]
	for i := range list {
		if strings.EqualFold(list[i].Text, c.Text) {
			list[i] = c
			r.s.TenantSug[tenantID] = list
			return nil
		}
	}
	r.s.TenantSug[tenantID] = append(list, c)
	return nil
}

func (r *SuggestRepo) Suggest(_ context.Context, tenantID uuid.UUID, prefix string, limit int) ([]domain.SuggestCandidate, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	prefix = strings.ToLower(prefix)
	out := make([]domain.SuggestCandidate, 0)
	for _, c := range r.s.TenantSug[tenantID] {
		if prefix == "" || strings.HasPrefix(strings.ToLower(c.Text), prefix) || strings.Contains(strings.ToLower(c.Text), prefix) {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Weight > out[j].Weight })
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

type LexicalIndex struct{ s *Store }

func (r *LexicalIndex) Index(_ context.Context, d domain.ProductDocument) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	text := d.Title + " " + d.Description + " " + d.BrandName + " " + strings.Join(d.Tags, " ")
	tokens := domain.Tokenize(text)
	for _, tok := range tokens {
		k := tokenKey(d.TenantID, tok)
		m := r.s.Lexical[k]
		if m == nil {
			m = map[uuid.UUID]float64{}
			r.s.Lexical[k] = m
		}
		m[d.ProductID] += 1.0
	}
	return nil
}

func (r *LexicalIndex) Delete(_ context.Context, tenantID, productID uuid.UUID) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for k, m := range r.s.Lexical {
		if strings.HasPrefix(k, tenantID.String()+"|") {
			delete(m, productID)
		}
	}
	return nil
}

func (r *LexicalIndex) Search(_ context.Context, tenantID uuid.UUID, tokens []string, fuzzy bool, limit int) ([]uuid.UUID, map[uuid.UUID]float64, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	scores := map[uuid.UUID]float64{}
	for _, tok := range tokens {
		k := tokenKey(tenantID, tok)
		for pid, s := range r.s.Lexical[k] {
			scores[pid] += s
		}
		if fuzzy {
			// fuzzy: match tokens with edit distance <= 1 against indexed keys (bounded)
			prefix := tenantID.String() + "|"
			n := 0
			for key, m := range r.s.Lexical {
				if !strings.HasPrefix(key, prefix) {
					continue
				}
				indexed := strings.TrimPrefix(key, prefix)
				if domain.Levenshtein(tok, indexed) == 1 {
					for pid, s := range m {
						scores[pid] += s * 0.5
					}
				}
				n++
				if n > 500 {
					break
				}
			}
		}
	}
	type pair struct {
		id uuid.UUID
		s  float64
	}
	arr := make([]pair, 0, len(scores))
	for id, s := range scores {
		arr = append(arr, pair{id, s})
	}
	sort.Slice(arr, func(i, j int) bool { return arr[i].s > arr[j].s })
	ids := make([]uuid.UUID, 0, len(arr))
	for i, p := range arr {
		if limit > 0 && i >= limit {
			break
		}
		ids = append(ids, p.id)
	}
	return ids, scores, nil
}

type VectorStore struct{ s *Store }

func (r *VectorStore) Upsert(_ context.Context, tenantID, productID uuid.UUID, vector []float64) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Vectors[vecKey(tenantID, productID)] = vector
	return nil
}

func (r *VectorStore) Delete(_ context.Context, tenantID, productID uuid.UUID) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	delete(r.s.Vectors, vecKey(tenantID, productID))
	return nil
}

func (r *VectorStore) Search(_ context.Context, tenantID uuid.UUID, vector []float64, limit int) ([]uuid.UUID, map[uuid.UUID]float64, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	scores := map[uuid.UUID]float64{}
	prefix := tenantID.String() + "|"
	for k, v := range r.s.Vectors {
		if !strings.HasPrefix(k, prefix) {
			continue
		}
		pid, err := uuid.Parse(strings.TrimPrefix(k, prefix))
		if err != nil {
			continue
		}
		scores[pid] = domain.CosineSimilarity(vector, v)
	}
	type pair struct {
		id uuid.UUID
		s  float64
	}
	arr := make([]pair, 0, len(scores))
	for id, s := range scores {
		arr = append(arr, pair{id, s})
	}
	sort.Slice(arr, func(i, j int) bool { return arr[i].s > arr[j].s })
	ids := make([]uuid.UUID, 0)
	for i, p := range arr {
		if limit > 0 && i >= limit {
			break
		}
		ids = append(ids, p.id)
	}
	return ids, scores, nil
}
