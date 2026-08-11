package app_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nexora/review-service/internal/app"
	"github.com/nexora/review-service/internal/app/memory"
	"github.com/nexora/review-service/internal/domain"
)

func testDeps() *app.Deps {
	store := memory.NewStore()
	repos := memory.NewRepos(store)
	return &app.Deps{
		Reviews: repos.Reviews, Ratings: repos.Ratings, Media: repos.Media,
		Interactions: repos.Interactions, Quality: repos.Quality, Moderation: repos.Moderation,
		Trust: repos.Trust, Reputation: repos.Reputation, Outbox: repos.Outbox,
		Orders: memory.MockOrders{}, MediaClient: memory.MockMedia{}, AI: memory.MockAI{},
		Search: repos.Search, Clock: app.SystemClock{}, IDs: app.UUIDGen{},
	}
}

func TestCreateReviewAutoApproveVerified(t *testing.T) {
	d := testDeps()
	tenant := uuid.New()
	author := uuid.New()
	product := uuid.New()
	order := uuid.New()
	val := 5.0
	res, err := d.CreateReview(context.Background(), tenant, app.CreateReviewInput{
		AuthorID: author, TargetType: domain.TargetProduct, TargetID: product,
		OrderID: &order, Title: "Great", Body: "great delivery freshness packaging",
		RatingValue: &val, Scheme: domain.SchemeStars5,
		Quality: map[string]float64{domain.DimFreshness: 5, domain.DimOverall: 5},
		MediaRefs: []string{"media://img-1"},
		IdempotencyKey: "idem-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Review.Status != domain.ReviewStatusPublished {
		t.Fatalf("want published got %s (fraud=%v ai=%v)", res.Review.Status, res.Case.FraudScore, res.Case.AIScore)
	}
	if !res.Review.VerifiedPurchase {
		t.Fatal("expected verified purchase")
	}
	// idempotent
	res2, err := d.CreateReview(context.Background(), tenant, app.CreateReviewInput{
		AuthorID: author, TargetType: domain.TargetProduct, TargetID: product,
		Body: "great delivery", IdempotencyKey: "idem-1",
	})
	if err != nil || res2.Review.ID != res.Review.ID {
		t.Fatalf("idempotency failed: %v %+v", err, res2.Review)
	}
	ag, err := d.Ratings.GetAggregate(context.Background(), tenant, domain.TargetProduct, product, domain.SchemeStars5)
	if err != nil || ag.Count < 1 {
		t.Fatalf("aggregate: %v %+v", err, ag)
	}
	rep, err := d.Reputation.Get(context.Background(), tenant, domain.TargetProduct, product)
	if err != nil || rep.Score <= 0 {
		t.Fatalf("reputation: %v %+v", err, rep)
	}
}

func TestCreateReviewRejectHate(t *testing.T) {
	d := testDeps()
	res, err := d.CreateReview(context.Background(), uuid.New(), app.CreateReviewInput{
		AuthorID: uuid.New(), TargetType: domain.TargetCourier, TargetID: uuid.New(),
		Body: "I hate and kill this service",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Review.Status != domain.ReviewStatusRejected {
		t.Fatalf("want rejected got %s", res.Review.Status)
	}
}

func TestModerationDecideAndVote(t *testing.T) {
	d := testDeps()
	tenant := uuid.New()
	author := uuid.New()
	target := uuid.New()
	// low trust unverified -> queue
	res, err := d.CreateReview(context.Background(), tenant, app.CreateReviewInput{
		AuthorID: author, TargetType: domain.TargetBrand, TargetID: target,
		Body: "average product okay taste",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Review.Status != domain.ReviewStatusPendingModeration {
		t.Fatalf("want pending got %s auto=%s", res.Review.Status, res.Case.AutoDecision)
	}
	moderator := uuid.New()
	rev, err := d.DecideModeration(context.Background(), tenant, res.Review.ID, moderator, true, "ok")
	if err != nil || rev.Status != domain.ReviewStatusPublished {
		t.Fatalf("decide: %v %s", err, rev.Status)
	}
	voter := uuid.New()
	rev, err = d.VoteHelpful(context.Background(), tenant, rev.ID, voter, true)
	if err != nil || rev.HelpfulCount != 1 {
		t.Fatalf("vote: %v %+v", err, rev)
	}
	_, err = d.AddComment(context.Background(), tenant, rev.ID, voter, "thanks for the tip", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.ReportReview(context.Background(), tenant, rev.ID, uuid.New(), "spam", "looks fake")
	if err != nil {
		t.Fatal(err)
	}
}

func TestSubmitRatingThumbs(t *testing.T) {
	d := testDeps()
	tenant := uuid.New()
	target := uuid.New()
	rt, ag, err := d.SubmitRating(context.Background(), tenant, app.SubmitRatingInput{
		AuthorID: uuid.New(), TargetType: domain.TargetWarehouse, TargetID: target,
		Scheme: domain.SchemeThumbs, Value: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if rt.Stars != 5 || ag.Count != 1 {
		t.Fatalf("rating %+v agg %+v", rt, ag)
	}
}

func TestDeleteReviewGDPR(t *testing.T) {
	d := testDeps()
	tenant := uuid.New()
	author := uuid.New()
	order := uuid.New()
	res, err := d.CreateReview(context.Background(), tenant, app.CreateReviewInput{
		AuthorID: author, TargetType: domain.TargetExperience, TargetID: uuid.New(),
		OrderID: &order, Body: "excellent overall experience delivery",
	})
	if err != nil {
		t.Fatal(err)
	}
	rev, err := d.DeleteReview(context.Background(), tenant, res.Review.ID, author, false)
	if err != nil || rev.Status != domain.ReviewStatusDeleted || rev.Body != "" {
		t.Fatalf("delete: %v %+v", err, rev)
	}
}
