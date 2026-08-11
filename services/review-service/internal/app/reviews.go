package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/review-service/internal/domain"
)

// CreateReviewInput creates a review (+ optional rating & quality dims).
type CreateReviewInput struct {
	AuthorID       uuid.UUID
	TargetType     string
	TargetID       uuid.UUID
	OrderID        *uuid.UUID
	Locale         string
	Title          string
	Body           string
	Anonymous      bool
	IdempotencyKey string
	Scheme         string
	RatingValue    *float64
	Quality        map[string]float64
	MediaRefs      []string
	Tags           []string
}

// CreateReviewResult is the create response.
type CreateReviewResult struct {
	Review domain.Review
	Rating *domain.Rating
	Case   domain.ModerationCase
}

// CreateReview submits content into moderation.
func (d *Deps) CreateReview(ctx context.Context, tenantID uuid.UUID, in CreateReviewInput) (CreateReviewResult, error) {
	var out CreateReviewResult
	if in.IdempotencyKey != "" {
		if existing, ok, err := d.Reviews.GetByIdempotency(ctx, tenantID, in.IdempotencyKey); err != nil {
			return out, err
		} else if ok {
			out.Review = existing
			return out, nil
		}
	}

	now := d.now()
	r := domain.Review{
		ID: d.newID(), TenantID: tenantID, AuthorID: in.AuthorID,
		TargetType: in.TargetType, TargetID: in.TargetID, OrderID: in.OrderID,
		Locale: strings.TrimSpace(in.Locale), Title: strings.TrimSpace(in.Title),
		Body: strings.TrimSpace(in.Body), Anonymous: in.Anonymous,
		Status: domain.ReviewStatusPendingModeration, Revision: 1,
		IdempotencyKey: in.IdempotencyKey, Tags: in.Tags,
		CreatedAt: now, UpdatedAt: now,
	}
	if r.Locale == "" {
		r.Locale = "tr-TR"
	}
	if err := r.ValidateCreate(); err != nil {
		return out, err
	}

	if in.OrderID != nil && d.Orders != nil {
		purchased, delivered, err := d.Orders.VerifyPurchase(ctx, tenantID, in.AuthorID, *in.OrderID, in.TargetID, in.TargetType)
		if err != nil {
			return out, err
		}
		r.VerifiedPurchase = purchased
		r.VerifiedDelivery = delivered
	}

	trust, _ := d.Trust.Get(ctx, tenantID, in.AuthorID)
	if trust.ReviewerID == uuid.Nil {
		trust = domain.TrustScore{TenantID: tenantID, ReviewerID: in.AuthorID, Score: 40, UpdatedAt: now}
	}

	bodyHash := hashBody(r.Body)
	since := now.Add(-1 * time.Hour)
	velocity, _ := d.Reviews.CountRecentByAuthor(ctx, tenantID, in.AuthorID, since)
	dups, _ := d.Reviews.CountDupBody(ctx, tenantID, bodyHash, now.Add(-24*time.Hour))
	fraudScore := domain.HeuristicSpamScore(r.Body, dups, velocity)
	fraudSignals := []string{}
	if dups > 0 {
		fraudSignals = append(fraudSignals, "duplicate_body")
	}
	if velocity > 5 {
		fraudSignals = append(fraudSignals, "high_velocity")
	}

	aiScore := 0.0
	labels := []string{}
	pii := false
	sentiment := 0.0
	topics := []string{}
	if d.AI != nil {
		res, err := d.AI.Analyze(ctx, tenantID, r.Title, r.Body, r.Locale)
		if err == nil {
			aiScore = res.UnsafeScore
			labels = res.Labels
			pii = res.PIIFound
			sentiment = res.Sentiment
			if res.MaskedBody != "" && pii {
				r.Body = res.MaskedBody
			}
		}
		if tops, sent, err := d.AI.ExtractTopics(ctx, tenantID, r.Body); err == nil {
			topics = tops
			if sentiment == 0 {
				sentiment = sent
			}
		}
	}
	r.Sentiment = sentiment
	r.Topics = topics

	mod := domain.ModerationCase{
		ID: d.newID(), ReviewID: r.ID, TenantID: tenantID,
		Status: domain.ModerationPending, AIScore: aiScore, Labels: labels,
		FraudScore: fraudScore, FraudSignals: fraudSignals, PIIMasked: pii,
		CreatedAt: now, UpdatedAt: now,
	}

	// Auto decisions
	switch {
	case fraudScore >= 0.85 || aiScore >= 0.9:
		mod.AutoDecision = "reject"
		mod.Status = domain.ModerationRejected
		r.Status = domain.ReviewStatusRejected
		trust.RejectedReviews++
	case fraudScore < 0.3 && aiScore < 0.25 && (r.VerifiedPurchase || trust.Score >= 60):
		mod.AutoDecision = "approve"
		mod.Status = domain.ModerationApproved
		r.Status = domain.ReviewStatusPublished
		pub := now
		r.PublishedAt = &pub
		trust.PublishedReviews++
		if r.VerifiedPurchase {
			trust.VerifiedPurchases++
		}
	default:
		mod.AutoDecision = "queue"
	}
	trust.Score = domain.ComputeTrustScore(&trust)
	trust.Badges = deriveBadges(&trust)
	trust.UpdatedAt = now

	if err := d.Reviews.Save(ctx, r); err != nil {
		return out, err
	}
	_ = d.Reviews.SaveRevision(ctx, domain.ReviewRevision{
		ID: d.newID(), ReviewID: r.ID, TenantID: tenantID, Revision: 1,
		Title: r.Title, Body: r.Body, Locale: r.Locale, CreatedAt: now, CreatedBy: in.AuthorID,
	})
	if err := d.Moderation.Save(ctx, mod); err != nil {
		return out, err
	}
	_ = d.Trust.Save(ctx, trust)

	var rating *domain.Rating
	if in.RatingValue != nil {
		scheme := in.Scheme
		if scheme == "" {
			scheme = domain.SchemeStars5
		}
		stars, err := domain.NormalizeStars(scheme, *in.RatingValue)
		if err != nil {
			return out, err
		}
		rt := domain.Rating{
			ID: d.newID(), TenantID: tenantID, AuthorID: in.AuthorID,
			TargetType: in.TargetType, TargetID: in.TargetID, ReviewID: &r.ID,
			Scheme: scheme, Value: *in.RatingValue, Stars: stars,
			Verified: r.VerifiedPurchase, Weight: domain.TrustWeight(trust.Score),
			CreatedAt: now,
		}
		if err := d.Ratings.Save(ctx, rt); err != nil {
			return out, err
		}
		rating = &rt
		d.emit(ctx, tenantID, rt.ID, domain.EventRatingSubmitted, map[string]any{
			"targetType": in.TargetType, "targetId": in.TargetID.String(), "stars": stars,
		})
		if r.Status == domain.ReviewStatusPublished {
			if err := d.refreshAggregates(ctx, tenantID, in.TargetType, in.TargetID, scheme); err != nil {
				return out, err
			}
		}
	}

	for dim, val := range in.Quality {
		if !domain.ValidDimension(dim) || val < 1 || val > 5 {
			continue
		}
		_ = d.Quality.Save(ctx, domain.QualityScore{
			ID: d.newID(), ReviewID: r.ID, TenantID: tenantID,
			Dimension: dim, Value: val, CreatedAt: now,
		})
	}

	for _, ref := range in.MediaRefs {
		kind := "image"
		ok := true
		if d.MediaClient != nil {
			var err error
			ok, kind, err = d.MediaClient.ValidateRef(ctx, tenantID, ref)
			if err != nil || !ok {
				continue
			}
		}
		_ = d.Media.Save(ctx, domain.ReviewMedia{
			ID: d.newID(), ReviewID: r.ID, TenantID: tenantID, MediaRef: ref,
			Kind: kind, ModerationOK: r.Status == domain.ReviewStatusPublished,
			CreatedAt: now,
		})
		d.emit(ctx, tenantID, r.ID, domain.EventMediaAttached, map[string]any{"mediaRef": ref})
	}

	d.emit(ctx, tenantID, r.ID, domain.EventReviewCreated, map[string]any{
		"status": r.Status, "targetType": r.TargetType, "targetId": r.TargetID.String(),
	})
	d.emit(ctx, tenantID, mod.ID, domain.EventModerationQueued, map[string]any{
		"reviewId": r.ID.String(), "autoDecision": mod.AutoDecision,
	})
	if r.Status == domain.ReviewStatusPublished {
		d.emit(ctx, tenantID, r.ID, domain.EventReviewApproved, map[string]any{"auto": true})
		if d.Search != nil {
			_ = d.Search.IndexReview(ctx, r)
		}
		_ = d.refreshReputation(ctx, tenantID, r.TargetType, r.TargetID)
	}
	if r.Status == domain.ReviewStatusRejected {
		d.emit(ctx, tenantID, r.ID, domain.EventReviewRejected, map[string]any{"auto": true})
	}
	d.emit(ctx, tenantID, in.AuthorID, domain.EventTrustScoreUpdated, map[string]any{"score": trust.Score})

	out.Review = r
	out.Rating = rating
	out.Case = mod
	return out, nil
}

// UpdateReview edits body/title and re-queues moderation.
func (d *Deps) UpdateReview(ctx context.Context, tenantID, reviewID, actorID uuid.UUID, title, body string) (domain.Review, error) {
	r, err := d.Reviews.Get(ctx, tenantID, reviewID)
	if err != nil {
		return r, err
	}
	if r.AuthorID != actorID {
		return r, domain.ErrForbidden
	}
	if r.Status == domain.ReviewStatusDeleted {
		return r, domain.ErrIllegalTransition
	}
	now := d.now()
	r.Title = strings.TrimSpace(title)
	r.Body = strings.TrimSpace(body)
	r.Revision++
	r.Status = domain.ReviewStatusPendingModeration
	r.UpdatedAt = now
	if err := r.ValidateCreate(); err != nil {
		return r, err
	}
	_ = d.Reviews.SaveRevision(ctx, domain.ReviewRevision{
		ID: d.newID(), ReviewID: r.ID, TenantID: tenantID, Revision: r.Revision,
		Title: r.Title, Body: r.Body, Locale: r.Locale, CreatedAt: now, CreatedBy: actorID,
	})
	if err := d.Reviews.Save(ctx, r); err != nil {
		return r, err
	}
	mod := domain.ModerationCase{
		ID: d.newID(), ReviewID: r.ID, TenantID: tenantID,
		Status: domain.ModerationPending, AutoDecision: "queue",
		CreatedAt: now, UpdatedAt: now,
	}
	if d.AI != nil {
		if res, err := d.AI.Analyze(ctx, tenantID, r.Title, r.Body, r.Locale); err == nil {
			mod.AIScore = res.UnsafeScore
			mod.Labels = res.Labels
			mod.PIIMasked = res.PIIFound
			r.Sentiment = res.Sentiment
			_ = d.Reviews.Save(ctx, r)
		}
	}
	_ = d.Moderation.Save(ctx, mod)
	d.emit(ctx, tenantID, r.ID, domain.EventReviewUpdated, map[string]any{"revision": r.Revision})
	return r, nil
}

// DeleteReview soft-deletes a review (GDPR/KVKK right to delete content).
func (d *Deps) DeleteReview(ctx context.Context, tenantID, reviewID, actorID uuid.UUID, admin bool) (domain.Review, error) {
	r, err := d.Reviews.Get(ctx, tenantID, reviewID)
	if err != nil {
		return r, err
	}
	if !admin && r.AuthorID != actorID {
		return r, domain.ErrForbidden
	}
	now := d.now()
	r.Status = domain.ReviewStatusDeleted
	r.DeletedAt = &now
	r.UpdatedAt = now
	r.Body = ""
	r.Title = ""
	if err := d.Reviews.Save(ctx, r); err != nil {
		return r, err
	}
	if d.Search != nil {
		_ = d.Search.DeleteReview(ctx, tenantID, reviewID)
	}
	d.emit(ctx, tenantID, r.ID, domain.EventReviewDeleted, nil)
	_ = d.refreshReputation(ctx, tenantID, r.TargetType, r.TargetID)
	return r, nil
}

// DecideModeration applies a human moderation decision.
func (d *Deps) DecideModeration(ctx context.Context, tenantID, reviewID, moderatorID uuid.UUID, approve bool, note string) (domain.Review, error) {
	r, err := d.Reviews.Get(ctx, tenantID, reviewID)
	if err != nil {
		return r, err
	}
	mod, err := d.Moderation.GetByReview(ctx, tenantID, reviewID)
	if err != nil {
		return r, err
	}
	now := d.now()
	mod.DecidedBy = &moderatorID
	mod.DecisionNote = note
	mod.DecidedAt = &now
	mod.UpdatedAt = now
	trust, _ := d.Trust.Get(ctx, tenantID, r.AuthorID)
	if trust.ReviewerID == uuid.Nil {
		trust = domain.TrustScore{TenantID: tenantID, ReviewerID: r.AuthorID, Score: 40}
	}
	if approve {
		if !domain.CanTransition(r.Status, domain.ReviewStatusPublished) && r.Status != domain.ReviewStatusPendingModeration {
			return r, domain.ErrIllegalTransition
		}
		mod.Status = domain.ModerationApproved
		r.Status = domain.ReviewStatusPublished
		r.PublishedAt = &now
		trust.PublishedReviews++
		if r.VerifiedPurchase {
			trust.VerifiedPurchases++
		}
		d.emit(ctx, tenantID, r.ID, domain.EventReviewApproved, map[string]any{"moderatorId": moderatorID.String()})
		if d.Search != nil {
			_ = d.Search.IndexReview(ctx, r)
		}
	} else {
		mod.Status = domain.ModerationRejected
		r.Status = domain.ReviewStatusRejected
		trust.RejectedReviews++
		d.emit(ctx, tenantID, r.ID, domain.EventReviewRejected, map[string]any{"moderatorId": moderatorID.String()})
		if d.Search != nil {
			_ = d.Search.DeleteReview(ctx, tenantID, reviewID)
		}
	}
	r.UpdatedAt = now
	trust.Score = domain.ComputeTrustScore(&trust)
	trust.Badges = deriveBadges(&trust)
	trust.UpdatedAt = now
	if err := d.Reviews.Save(ctx, r); err != nil {
		return r, err
	}
	_ = d.Moderation.Save(ctx, mod)
	_ = d.Trust.Save(ctx, trust)
	_ = d.refreshAggregates(ctx, tenantID, r.TargetType, r.TargetID, domain.SchemeStars5)
	_ = d.refreshReputation(ctx, tenantID, r.TargetType, r.TargetID)
	d.emit(ctx, tenantID, r.AuthorID, domain.EventTrustScoreUpdated, map[string]any{"score": trust.Score})
	return r, nil
}

func (d *Deps) refreshAggregates(ctx context.Context, tenantID uuid.UUID, targetType string, targetID uuid.UUID, scheme string) error {
	ratings, err := d.Ratings.ListByTarget(ctx, tenantID, targetType, targetID)
	if err != nil {
		return err
	}
	filtered := make([]domain.Rating, 0, len(ratings))
	for _, rt := range ratings {
		if scheme == "" || rt.Scheme == scheme || (scheme == domain.SchemeStars5) {
			filtered = append(filtered, rt)
		}
	}
	ag := domain.RatingAggregate{
		TenantID: tenantID, TargetType: targetType, TargetID: targetID, Scheme: scheme,
	}
	if ag.Scheme == "" {
		ag.Scheme = domain.SchemeStars5
	}
	domain.RecomputeAggregates(&ag, filtered, d.now())
	return d.Ratings.SaveAggregate(ctx, ag)
}

func (d *Deps) refreshReputation(ctx context.Context, tenantID uuid.UUID, targetType string, targetID uuid.UUID) error {
	ag, err := d.Ratings.GetAggregate(ctx, tenantID, targetType, targetID, domain.SchemeStars5)
	if err != nil && err != domain.ErrNotFound {
		return err
	}
	bayes := ag.BayesianAvg
	if bayes == 0 && ag.Count == 0 {
		bayes = domain.BayesianPriorMean
	}
	score := domain.ReputationFromAggregate(bayes, ag.Count, 0)
	rep := domain.ReputationScore{
		TenantID: tenantID, TargetType: targetType, TargetID: targetID,
		Score: score, Tier: domain.ReputationTier(score), ReviewCount: ag.Count, UpdatedAt: d.now(),
	}
	if err := d.Reputation.Save(ctx, rep); err != nil {
		return err
	}
	d.emit(ctx, tenantID, targetID, domain.EventReputationUpdated, map[string]any{
		"targetType": targetType, "score": score, "tier": rep.Tier,
	})
	return nil
}

func hashBody(body string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(body))))
	return hex.EncodeToString(sum[:])
}

func deriveBadges(t *domain.TrustScore) []string {
	badges := []string{}
	if t.VerifiedPurchases >= 1 {
		badges = append(badges, "verified_buyer")
	}
	if t.PublishedReviews >= 10 && t.Score >= 70 {
		badges = append(badges, "trusted_reviewer")
	}
	if t.PublishedReviews >= 50 && t.HelpfulReceived >= 100 {
		badges = append(badges, "top_contributor")
	}
	if t.Score >= 85 {
		badges = append(badges, "expert_reviewer")
	}
	return badges
}
