package app

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/review-service/internal/domain"
)

// SubmitRatingInput is a standalone rating without text review.
type SubmitRatingInput struct {
	AuthorID   uuid.UUID
	TargetType string
	TargetID   uuid.UUID
	OrderID    *uuid.UUID
	Scheme     string
	Value      float64
}

// SubmitRating records a rating and refreshes aggregates.
func (d *Deps) SubmitRating(ctx context.Context, tenantID uuid.UUID, in SubmitRatingInput) (domain.Rating, domain.RatingAggregate, error) {
	var zero domain.Rating
	var ag domain.RatingAggregate
	if !domain.ValidTarget(in.TargetType) || in.AuthorID == uuid.Nil || in.TargetID == uuid.Nil {
		return zero, ag, domain.ErrInvalidArgument
	}
	scheme := in.Scheme
	if scheme == "" {
		scheme = domain.SchemeStars5
	}
	stars, err := domain.NormalizeStars(scheme, in.Value)
	if err != nil {
		return zero, ag, err
	}
	now := d.now()
	verified := false
	if in.OrderID != nil && d.Orders != nil {
		purchased, _, err := d.Orders.VerifyPurchase(ctx, tenantID, in.AuthorID, *in.OrderID, in.TargetID, in.TargetType)
		if err != nil {
			return zero, ag, err
		}
		verified = purchased
	}
	trust, _ := d.Trust.Get(ctx, tenantID, in.AuthorID)
	if trust.ReviewerID == uuid.Nil {
		trust = domain.TrustScore{TenantID: tenantID, ReviewerID: in.AuthorID, Score: 40, UpdatedAt: now}
		_ = d.Trust.Save(ctx, trust)
	}
	rt := domain.Rating{
		ID: d.newID(), TenantID: tenantID, AuthorID: in.AuthorID,
		TargetType: in.TargetType, TargetID: in.TargetID,
		Scheme: scheme, Value: in.Value, Stars: stars,
		Verified: verified, Weight: domain.TrustWeight(trust.Score), CreatedAt: now,
	}
	if err := d.Ratings.Save(ctx, rt); err != nil {
		return zero, ag, err
	}
	if err := d.refreshAggregates(ctx, tenantID, in.TargetType, in.TargetID, scheme); err != nil {
		return zero, ag, err
	}
	ag, _ = d.Ratings.GetAggregate(ctx, tenantID, in.TargetType, in.TargetID, scheme)
	_ = d.refreshReputation(ctx, tenantID, in.TargetType, in.TargetID)
	d.emit(ctx, tenantID, rt.ID, domain.EventRatingSubmitted, map[string]any{
		"targetType": in.TargetType, "targetId": in.TargetID.String(), "stars": stars,
	})
	return rt, ag, nil
}

// VoteHelpful records a helpful/not-helpful vote.
func (d *Deps) VoteHelpful(ctx context.Context, tenantID, reviewID, voterID uuid.UUID, helpful bool) (domain.Review, error) {
	r, err := d.Reviews.Get(ctx, tenantID, reviewID)
	if err != nil {
		return r, err
	}
	if r.Status != domain.ReviewStatusPublished {
		return r, domain.ErrInvariant
	}
	if existing, ok, _ := d.Interactions.GetVote(ctx, reviewID, voterID); ok {
		// flip counts if changed
		if existing.Helpful == helpful {
			return r, nil
		}
		if existing.Helpful {
			r.HelpfulCount--
			r.NotHelpfulCount++
		} else {
			r.NotHelpfulCount--
			r.HelpfulCount++
		}
		existing.Helpful = helpful
		_ = d.Interactions.SaveVote(ctx, existing)
	} else {
		_ = d.Interactions.SaveVote(ctx, domain.ReviewVote{
			ID: d.newID(), ReviewID: reviewID, TenantID: tenantID, VoterID: voterID,
			Helpful: helpful, CreatedAt: d.now(),
		})
		if helpful {
			r.HelpfulCount++
		} else {
			r.NotHelpfulCount++
		}
	}
	r.UpdatedAt = d.now()
	if err := d.Reviews.Save(ctx, r); err != nil {
		return r, err
	}
	if helpful {
		trust, _ := d.Trust.Get(ctx, tenantID, r.AuthorID)
		if trust.ReviewerID != uuid.Nil {
			trust.HelpfulReceived++
			trust.Score = domain.ComputeTrustScore(&trust)
			trust.Badges = deriveBadges(&trust)
			trust.UpdatedAt = d.now()
			_ = d.Trust.Save(ctx, trust)
			d.emit(ctx, tenantID, r.AuthorID, domain.EventTrustScoreUpdated, map[string]any{"score": trust.Score})
		}
	}
	return r, nil
}

// AddComment adds a reply to a review.
func (d *Deps) AddComment(ctx context.Context, tenantID, reviewID, authorID uuid.UUID, body string, parentID *uuid.UUID) (domain.ReviewComment, error) {
	var zero domain.ReviewComment
	body = strings.TrimSpace(body)
	if body == "" || len(body) > 2000 {
		return zero, domain.ErrInvalidArgument
	}
	if _, err := d.Reviews.Get(ctx, tenantID, reviewID); err != nil {
		return zero, err
	}
	now := d.now()
	c := domain.ReviewComment{
		ID: d.newID(), ReviewID: reviewID, TenantID: tenantID, AuthorID: authorID,
		ParentID: parentID, Body: body, Status: "published", CreatedAt: now, UpdatedAt: now,
	}
	if err := d.Interactions.SaveComment(ctx, c); err != nil {
		return zero, err
	}
	return c, nil
}

// ReportReview flags a review for re-moderation.
func (d *Deps) ReportReview(ctx context.Context, tenantID, reviewID, reporterID uuid.UUID, reason, details string) (domain.ReviewReport, error) {
	var zero domain.ReviewReport
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return zero, domain.ErrInvalidArgument
	}
	r, err := d.Reviews.Get(ctx, tenantID, reviewID)
	if err != nil {
		return zero, err
	}
	now := d.now()
	rep := domain.ReviewReport{
		ID: d.newID(), ReviewID: reviewID, TenantID: tenantID, ReporterID: reporterID,
		Reason: reason, Details: details, CreatedAt: now,
	}
	if err := d.Interactions.SaveReport(ctx, rep); err != nil {
		return zero, err
	}
	r.ReportCount++
	r.UpdatedAt = now
	if r.ReportCount >= 3 && r.Status == domain.ReviewStatusPublished {
		r.Status = domain.ReviewStatusHidden
	}
	_ = d.Reviews.Save(ctx, r)
	d.emit(ctx, tenantID, reviewID, domain.EventReviewReported, map[string]any{"reason": reason})
	return rep, nil
}

// PinReview pins/unpins a review (admin).
func (d *Deps) PinReview(ctx context.Context, tenantID, reviewID uuid.UUID, pinned bool) (domain.Review, error) {
	r, err := d.Reviews.Get(ctx, tenantID, reviewID)
	if err != nil {
		return r, err
	}
	r.Pinned = pinned
	r.UpdatedAt = d.now()
	return r, d.Reviews.Save(ctx, r)
}

// SummarizeTarget uses AI to summarize published review bodies.
func (d *Deps) SummarizeTarget(ctx context.Context, tenantID uuid.UUID, targetType string, targetID uuid.UUID, limit int) (string, error) {
	if d.AI == nil {
		return "", domain.ErrInvalidArgument
	}
	if limit <= 0 {
		limit = 20
	}
	list, err := d.Reviews.ListByTarget(ctx, tenantID, targetType, targetID, domain.ReviewStatusPublished, limit)
	if err != nil {
		return "", err
	}
	bodies := make([]string, 0, len(list))
	for _, r := range list {
		bodies = append(bodies, r.Body)
	}
	return d.AI.Summarize(ctx, tenantID, bodies)
}

// AdminStats returns lightweight dashboard counters.
func (d *Deps) AdminStats(ctx context.Context, tenantID uuid.UUID) (map[string]any, error) {
	pending, err := d.Moderation.ListPending(ctx, tenantID, 500)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"pendingModeration": len(pending),
		"tenantId":          tenantID.String(),
	}, nil
}
