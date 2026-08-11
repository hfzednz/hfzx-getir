package memory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/review-service/internal/app/ports"
	"github.com/nexora/review-service/internal/domain"
)

// Repos bundles memory repositories.
type Repos struct {
	Reviews      *ReviewRepo
	Ratings      *RatingRepo
	Media        *MediaRepo
	Interactions *InteractionRepo
	Quality      *QualityRepo
	Moderation   *ModerationRepo
	Trust        *TrustRepo
	Reputation   *ReputationRepo
	Outbox       *OutboxRepo
	Search       *SearchRepo
}

// NewRepos wires memory repos against a shared store.
func NewRepos(s *Store) *Repos {
	return &Repos{
		Reviews: &ReviewRepo{s: s}, Ratings: &RatingRepo{s: s}, Media: &MediaRepo{s: s},
		Interactions: &InteractionRepo{s: s}, Quality: &QualityRepo{s: s},
		Moderation: &ModerationRepo{s: s}, Trust: &TrustRepo{s: s}, Reputation: &ReputationRepo{s: s},
		Outbox: &OutboxRepo{s: s}, Search: &SearchRepo{s: s},
	}
}

// --- Reviews ---

type ReviewRepo struct{ s *Store }

func (r *ReviewRepo) Save(_ context.Context, rev domain.Review) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Reviews[rev.ID] = rev
	if rev.IdempotencyKey != "" {
		r.s.Idem[rev.TenantID.String()+"|"+rev.IdempotencyKey] = rev.ID
	}
	h := hash(rev.Body)
	r.s.BodyHashes[h] = append(r.s.BodyHashes[h], rev.CreatedAt)
	return nil
}

func (r *ReviewRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Review, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	rev, ok := r.s.Reviews[id]
	if !ok || rev.TenantID != tenantID {
		return domain.Review{}, domain.ErrNotFound
	}
	return rev, nil
}

func (r *ReviewRepo) GetByIdempotency(_ context.Context, tenantID uuid.UUID, key string) (domain.Review, bool, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	id, ok := r.s.Idem[tenantID.String()+"|"+key]
	if !ok {
		return domain.Review{}, false, nil
	}
	rev, ok := r.s.Reviews[id]
	if !ok {
		return domain.Review{}, false, nil
	}
	return rev, true, nil
}

func (r *ReviewRepo) ListByTarget(_ context.Context, tenantID uuid.UUID, targetType string, targetID uuid.UUID, status string, limit int) ([]domain.Review, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.Review, 0)
	for _, rev := range r.s.Reviews {
		if rev.TenantID != tenantID || rev.TargetType != targetType || rev.TargetID != targetID {
			continue
		}
		if status != "" && rev.Status != status {
			continue
		}
		out = append(out, rev)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (r *ReviewRepo) ListByAuthor(_ context.Context, tenantID, authorID uuid.UUID, limit int) ([]domain.Review, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.Review, 0)
	for _, rev := range r.s.Reviews {
		if rev.TenantID == tenantID && rev.AuthorID == authorID {
			out = append(out, rev)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *ReviewRepo) SaveRevision(_ context.Context, rev domain.ReviewRevision) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Revisions = append(r.s.Revisions, rev)
	return nil
}

func (r *ReviewRepo) ListRevisions(_ context.Context, tenantID, reviewID uuid.UUID) ([]domain.ReviewRevision, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.ReviewRevision, 0)
	for _, rev := range r.s.Revisions {
		if rev.TenantID == tenantID && rev.ReviewID == reviewID {
			out = append(out, rev)
		}
	}
	return out, nil
}

func (r *ReviewRepo) CountRecentByAuthor(_ context.Context, tenantID, authorID uuid.UUID, since time.Time) (int, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	n := 0
	for _, rev := range r.s.Reviews {
		if rev.TenantID == tenantID && rev.AuthorID == authorID && !rev.CreatedAt.Before(since) {
			n++
		}
	}
	return n, nil
}

func (r *ReviewRepo) CountDupBody(_ context.Context, _ uuid.UUID, bodyHash string, since time.Time) (int, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	n := 0
	for _, t := range r.s.BodyHashes[bodyHash] {
		if !t.Before(since) {
			n++
		}
	}
	// subtract the current insert which may already be counted once after Save —
	// callers typically count before save; if after, n-1 is fine via >=1 signal.
	return n, nil
}

func hash(body string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(body))))
	return hex.EncodeToString(sum[:])
}

// --- Ratings ---

type RatingRepo struct{ s *Store }

func (r *RatingRepo) Save(_ context.Context, rt domain.Rating) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Ratings = append(r.s.Ratings, rt)
	return nil
}

func (r *RatingRepo) ListByTarget(_ context.Context, tenantID uuid.UUID, targetType string, targetID uuid.UUID) ([]domain.Rating, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.Rating, 0)
	for _, rt := range r.s.Ratings {
		if rt.TenantID == tenantID && rt.TargetType == targetType && rt.TargetID == targetID {
			out = append(out, rt)
		}
	}
	return out, nil
}

func (r *RatingRepo) GetAggregate(_ context.Context, tenantID uuid.UUID, targetType string, targetID uuid.UUID, scheme string) (domain.RatingAggregate, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	ag, ok := r.s.Aggregates[aggKey(tenantID, targetType, targetID, scheme)]
	if !ok {
		return domain.RatingAggregate{}, domain.ErrNotFound
	}
	return ag, nil
}

func (r *RatingRepo) SaveAggregate(_ context.Context, a domain.RatingAggregate) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Aggregates[aggKey(a.TenantID, a.TargetType, a.TargetID, a.Scheme)] = a
	return nil
}

// --- Media / Interactions / Quality / Moderation / Trust / Reputation / Outbox / Search ---

type MediaRepo struct{ s *Store }

func (r *MediaRepo) Save(_ context.Context, m domain.ReviewMedia) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Media = append(r.s.Media, m)
	return nil
}

func (r *MediaRepo) ListByReview(_ context.Context, tenantID, reviewID uuid.UUID) ([]domain.ReviewMedia, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.ReviewMedia, 0)
	for _, m := range r.s.Media {
		if m.TenantID == tenantID && m.ReviewID == reviewID {
			out = append(out, m)
		}
	}
	return out, nil
}

type InteractionRepo struct{ s *Store }

func (r *InteractionRepo) SaveVote(_ context.Context, v domain.ReviewVote) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Votes[voteKey(v.ReviewID, v.VoterID)] = v
	return nil
}

func (r *InteractionRepo) GetVote(_ context.Context, reviewID, voterID uuid.UUID) (domain.ReviewVote, bool, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	v, ok := r.s.Votes[voteKey(reviewID, voterID)]
	return v, ok, nil
}

func (r *InteractionRepo) SaveComment(_ context.Context, c domain.ReviewComment) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Comments = append(r.s.Comments, c)
	return nil
}

func (r *InteractionRepo) ListComments(_ context.Context, tenantID, reviewID uuid.UUID) ([]domain.ReviewComment, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.ReviewComment, 0)
	for _, c := range r.s.Comments {
		if c.TenantID == tenantID && c.ReviewID == reviewID {
			out = append(out, c)
		}
	}
	return out, nil
}

func (r *InteractionRepo) SaveReport(_ context.Context, rep domain.ReviewReport) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Reports = append(r.s.Reports, rep)
	return nil
}

type QualityRepo struct{ s *Store }

func (r *QualityRepo) Save(_ context.Context, q domain.QualityScore) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Quality = append(r.s.Quality, q)
	return nil
}

func (r *QualityRepo) ListByReview(_ context.Context, tenantID, reviewID uuid.UUID) ([]domain.QualityScore, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.QualityScore, 0)
	for _, q := range r.s.Quality {
		if q.TenantID == tenantID && q.ReviewID == reviewID {
			out = append(out, q)
		}
	}
	return out, nil
}

type ModerationRepo struct{ s *Store }

func (r *ModerationRepo) Save(_ context.Context, m domain.ModerationCase) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Moderation[m.ReviewID] = m
	return nil
}

func (r *ModerationRepo) GetByReview(_ context.Context, tenantID, reviewID uuid.UUID) (domain.ModerationCase, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	m, ok := r.s.Moderation[reviewID]
	if !ok || m.TenantID != tenantID {
		return domain.ModerationCase{}, domain.ErrNotFound
	}
	return m, nil
}

func (r *ModerationRepo) ListPending(_ context.Context, tenantID uuid.UUID, limit int) ([]domain.ModerationCase, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.ModerationCase, 0)
	for _, m := range r.s.Moderation {
		if m.TenantID == tenantID && m.Status == domain.ModerationPending {
			out = append(out, m)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

type TrustRepo struct{ s *Store }

func (r *TrustRepo) Get(_ context.Context, tenantID, reviewerID uuid.UUID) (domain.TrustScore, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	t, ok := r.s.Trust[trustKey(tenantID, reviewerID)]
	if !ok {
		return domain.TrustScore{}, domain.ErrNotFound
	}
	return t, nil
}

func (r *TrustRepo) Save(_ context.Context, t domain.TrustScore) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Trust[trustKey(t.TenantID, t.ReviewerID)] = t
	return nil
}

type ReputationRepo struct{ s *Store }

func (r *ReputationRepo) Get(_ context.Context, tenantID uuid.UUID, targetType string, targetID uuid.UUID) (domain.ReputationScore, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	rep, ok := r.s.Reputation[repKey(tenantID, targetType, targetID)]
	if !ok {
		return domain.ReputationScore{}, domain.ErrNotFound
	}
	return rep, nil
}

func (r *ReputationRepo) Save(_ context.Context, rep domain.ReputationScore) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Reputation[repKey(rep.TenantID, rep.TargetType, rep.TargetID)] = rep
	return nil
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

type SearchRepo struct{ s *Store }

func (r *SearchRepo) IndexReview(_ context.Context, rev domain.Review) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.SearchIndex[rev.ID] = rev
	return nil
}

func (r *SearchRepo) DeleteReview(_ context.Context, _, reviewID uuid.UUID) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	delete(r.s.SearchIndex, reviewID)
	return nil
}

func (r *SearchRepo) Search(_ context.Context, tenantID uuid.UUID, query string, targetType string, limit int) ([]uuid.UUID, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	q := strings.ToLower(query)
	out := make([]uuid.UUID, 0)
	for id, rev := range r.s.SearchIndex {
		if rev.TenantID != tenantID {
			continue
		}
		if targetType != "" && rev.TargetType != targetType {
			continue
		}
		if q == "" || strings.Contains(strings.ToLower(rev.Body), q) || strings.Contains(strings.ToLower(rev.Title), q) {
			out = append(out, id)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

var (
	_ ports.ReviewRepo      = (*ReviewRepo)(nil)
	_ ports.RatingRepo      = (*RatingRepo)(nil)
	_ ports.SearchIndexer   = (*SearchRepo)(nil)
	_ ports.OutboxRepository = (*OutboxRepo)(nil)
)
