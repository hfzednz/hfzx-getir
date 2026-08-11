package domain

import (
	"math"
	"time"
)

// BayesianAverage computes a Bayesian mean from weighted star samples.
func BayesianAverage(sumWeighted, weightSum, priorMean, confidence float64) float64 {
	if priorMean <= 0 {
		priorMean = BayesianPriorMean
	}
	if confidence <= 0 {
		confidence = BayesianConfidence
	}
	if weightSum < 0 {
		weightSum = 0
	}
	return (confidence*priorMean + sumWeighted) / (confidence + weightSum)
}

// TimeDecayWeight returns an exponential decay weight for age.
func TimeDecayWeight(createdAt, now time.Time, lambda float64) float64 {
	if lambda <= 0 {
		lambda = DefaultDecayLambda
	}
	days := now.Sub(createdAt).Hours() / 24.0
	if days < 0 {
		days = 0
	}
	return math.Exp(-lambda * days)
}

// RecomputeAggregates rebuilds aggregate stats from ratings.
func RecomputeAggregates(ag *RatingAggregate, ratings []Rating, now time.Time) {
	if ag == nil {
		return
	}
	var sum, wSum, decaySum, decayW float64
	var verifiedSum float64
	verifiedN := 0
	for _, r := range ratings {
		w := r.Weight
		if w <= 0 {
			w = 1
		}
		sum += r.Stars * w
		wSum += w
		dw := TimeDecayWeight(r.CreatedAt, now, DefaultDecayLambda) * w
		decaySum += r.Stars * dw
		decayW += dw
		if r.Verified {
			verifiedSum += r.Stars
			verifiedN++
		}
	}
	ag.Count = len(ratings)
	ag.SumStars = sum
	if wSum > 0 {
		ag.AvgStars = sum / wSum
	} else {
		ag.AvgStars = 0
	}
	ag.BayesianAvg = BayesianAverage(sum, wSum, BayesianPriorMean, BayesianConfidence)
	if decayW > 0 {
		ag.TimeDecayAvg = decaySum / decayW
	} else {
		ag.TimeDecayAvg = ag.AvgStars
	}
	ag.VerifiedCount = verifiedN
	if verifiedN > 0 {
		ag.VerifiedAvg = verifiedSum / float64(verifiedN)
	} else {
		ag.VerifiedAvg = 0
	}
	ag.UpdatedAt = now
}

// ComputeTrustScore derives a 0–100 reviewer trust score.
func ComputeTrustScore(t *TrustScore) float64 {
	if t == nil {
		return 0
	}
	score := 40.0
	score += math.Min(25, float64(t.VerifiedPurchases)*2.5)
	score += math.Min(20, float64(t.PublishedReviews)*1.5)
	score += math.Min(15, float64(t.HelpfulReceived)*0.5)
	score -= math.Min(40, float64(t.RejectedReviews)*8)
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score
}

// TrustWeight maps trust score to rating weight (0.2–1.5).
func TrustWeight(trust float64) float64 {
	w := 0.2 + (trust/100.0)*1.3
	if w < 0.2 {
		return 0.2
	}
	if w > 1.5 {
		return 1.5
	}
	return w
}

// ReputationFromAggregate maps rating aggregates + penalties into 0–100 reputation.
func ReputationFromAggregate(bayesian float64, reviewCount int, rejectRate float64) float64 {
	// Map 1–5 bayesian → 0–100 with volume dampening.
	base := ((bayesian - 1) / 4.0) * 100.0
	if reviewCount < 5 {
		base = base*0.7 + 50*0.3 // shrink toward mid until enough volume
	}
	base -= rejectRate * 30
	if base < 0 {
		base = 0
	}
	if base > 100 {
		base = 100
	}
	return base
}

// HeuristicSpamScore returns 0..1 spam likelihood from text heuristics.
func HeuristicSpamScore(body string, recentDupCount int, velocityPerHour int) float64 {
	score := 0.0
	if len(body) < 8 {
		score += 0.2
	}
	if recentDupCount > 0 {
		score += math.Min(0.5, float64(recentDupCount)*0.25)
	}
	if velocityPerHour > 10 {
		score += 0.4
	} else if velocityPerHour > 5 {
		score += 0.2
	}
	if score > 1 {
		score = 1
	}
	return score
}
