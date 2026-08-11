package domain_test

import (
	"testing"
	"time"

	"github.com/nexora/review-service/internal/domain"
)

func TestBayesianAverage(t *testing.T) {
	avg := domain.BayesianAverage(5*4.0, 5, 4.0, 20)
	// (20*4 + 20) / (20+5) = 100/25 = 4
	if avg < 3.99 || avg > 4.01 {
		t.Fatalf("expected ~4.0 got %v", avg)
	}
}

func TestTimeDecayWeight(t *testing.T) {
	now := time.Now().UTC()
	w0 := domain.TimeDecayWeight(now, now, 0.01)
	if w0 < 0.99 || w0 > 1.01 {
		t.Fatalf("fresh weight want ~1 got %v", w0)
	}
	old := now.Add(-100 * 24 * time.Hour)
	wOld := domain.TimeDecayWeight(old, now, 0.01)
	if wOld >= w0 {
		t.Fatalf("old weight should decay: %v >= %v", wOld, w0)
	}
}

func TestNormalizeStars(t *testing.T) {
	s, err := domain.NormalizeStars(domain.SchemeThumbs, 1)
	if err != nil || s != 5 {
		t.Fatalf("thumbs up -> 5, got %v %v", s, err)
	}
	s, err = domain.NormalizeStars(domain.SchemeThumbs, 0)
	if err != nil || s != 1 {
		t.Fatalf("thumbs down -> 1, got %v %v", s, err)
	}
	_, err = domain.NormalizeStars(domain.SchemeStars5, 6)
	if err == nil {
		t.Fatal("expected invalid stars")
	}
}

func TestRecomputeAggregates(t *testing.T) {
	now := time.Now().UTC()
	ag := domain.RatingAggregate{}
	ratings := []domain.Rating{
		{Stars: 5, Weight: 1, Verified: true, CreatedAt: now},
		{Stars: 3, Weight: 1, Verified: false, CreatedAt: now},
	}
	domain.RecomputeAggregates(&ag, ratings, now)
	if ag.Count != 2 || ag.AvgStars != 4 {
		t.Fatalf("agg %+v", ag)
	}
	if ag.BayesianAvg <= 0 || ag.VerifiedCount != 1 {
		t.Fatalf("bayes/verified %+v", ag)
	}
}

func TestComputeTrustAndReputation(t *testing.T) {
	ts := &domain.TrustScore{VerifiedPurchases: 4, PublishedReviews: 10, HelpfulReceived: 20}
	score := domain.ComputeTrustScore(ts)
	if score < 40 || score > 100 {
		t.Fatalf("trust %v", score)
	}
	rep := domain.ReputationFromAggregate(4.5, 20, 0)
	if domain.ReputationTier(rep) == "poor" {
		t.Fatalf("expected better tier for 4.5, score=%v", rep)
	}
}

func TestHeuristicSpam(t *testing.T) {
	s := domain.HeuristicSpamScore("ok", 2, 12)
	if s < 0.5 {
		t.Fatalf("expected elevated spam score got %v", s)
	}
}
