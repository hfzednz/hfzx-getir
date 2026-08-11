package domain

import (
	"sort"
	"time"

	"github.com/google/uuid"
)

const (
	StrategyCollaborative = "collaborative"
	StrategyContent       = "content"
	StrategyHybrid        = "hybrid"
	StrategyFBT           = "fbt"
	StrategyUpsell        = "upsell"
	StrategyCrossSell     = "cross_sell"
	StrategyTrending      = "trending"
	StrategyPersonalized  = "personalized"
)

const (
	SignalView     = "view"
	SignalClick    = "click"
	SignalCart     = "cart"
	SignalPurchase = "purchase"
	SignalWishlist = "wishlist"
	SignalSearch   = "search"
)

type ProductFeatures struct {
	ProductID   uuid.UUID
	CategoryID  uuid.UUID
	BrandID     uuid.UUID
	Tags        []string
	PriceMinor  int64
	Popularity  float64
	RatingAvg   float64
}

type BehaviorSignal struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	UserID    uuid.UUID
	ProductID uuid.UUID
	Kind      string
	Weight    float64
	CreatedAt time.Time
}

type CoOccurrence struct {
	TenantID  uuid.UUID
	ProductA  uuid.UUID
	ProductB  uuid.UUID
	Count     int
	UpdatedAt time.Time
}

type RecommendationItem struct {
	ProductID uuid.UUID
	Score     float64
	Strategy  string
	Reason    string
}

type RecommendationRail struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Name      string
	Strategy  string
	Context   string // home|pdp|cart|checkout|empty_search
	Items     []RecommendationItem
	UserID    *uuid.UUID
	CreatedAt time.Time
}

type RecommendRequest struct {
	TenantID       uuid.UUID
	UserID         *uuid.UUID
	ProductID      *uuid.UUID // seed for similar/fbt/upsell
	CartProductIDs []uuid.UUID
	Strategy       string
	Context        string
	Limit          int
	ExcludeIDs     []uuid.UUID
}

func ValidStrategy(s string) bool {
	switch s {
	case "", StrategyCollaborative, StrategyContent, StrategyHybrid, StrategyFBT,
		StrategyUpsell, StrategyCrossSell, StrategyTrending, StrategyPersonalized:
		return true
	default:
		return false
	}
}

func SignalWeight(kind string) float64 {
	switch kind {
	case SignalPurchase:
		return 5
	case SignalCart:
		return 3
	case SignalWishlist:
		return 2.5
	case SignalClick:
		return 2
	case SignalView:
		return 1
	case SignalSearch:
		return 0.5
	default:
		return 1
	}
}

// ContentSimilarity Jaccard on tags + category/brand match.
func ContentSimilarity(a, b ProductFeatures) float64 {
	score := 0.0
	if a.CategoryID != uuid.Nil && a.CategoryID == b.CategoryID {
		score += 0.4
	}
	if a.BrandID != uuid.Nil && a.BrandID == b.BrandID {
		score += 0.2
	}
	if len(a.Tags) == 0 || len(b.Tags) == 0 {
		return score
	}
	set := map[string]struct{}{}
	for _, t := range a.Tags {
		set[t] = struct{}{}
	}
	inter := 0
	for _, t := range b.Tags {
		if _, ok := set[t]; ok {
			inter++
		}
	}
	union := len(a.Tags) + len(b.Tags) - inter
	if union > 0 {
		score += 0.4 * float64(inter) / float64(union)
	}
	return score
}

// BlendScores merges strategy score maps with weights.
func BlendScores(maps []map[uuid.UUID]float64, weights []float64) map[uuid.UUID]float64 {
	out := map[uuid.UUID]float64{}
	for i, m := range maps {
		w := 1.0
		if i < len(weights) {
			w = weights[i]
		}
		for id, s := range m {
			out[id] += s * w
		}
	}
	return out
}

// TopN returns highest scoring ids.
func TopN(scores map[uuid.UUID]float64, exclude map[uuid.UUID]struct{}, limit int) []RecommendationItem {
	type pair struct {
		id uuid.UUID
		s  float64
	}
	arr := make([]pair, 0, len(scores))
	for id, s := range scores {
		if _, skip := exclude[id]; skip {
			continue
		}
		arr = append(arr, pair{id, s})
	}
	sort.Slice(arr, func(i, j int) bool { return arr[i].s > arr[j].s })
	if limit <= 0 {
		limit = 10
	}
	out := make([]RecommendationItem, 0, limit)
	for i, p := range arr {
		if i >= limit {
			break
		}
		out = append(out, RecommendationItem{ProductID: p.id, Score: p.s})
	}
	return out
}

// DiversifyByCategory prefers spreading categories when features known.
func DiversifyByCategory(items []RecommendationItem, feats map[uuid.UUID]ProductFeatures, limit int) []RecommendationItem {
	if limit <= 0 {
		limit = len(items)
	}
	seenCat := map[uuid.UUID]int{}
	out := make([]RecommendationItem, 0, limit)
	rest := make([]RecommendationItem, 0)
	for _, it := range items {
		f := feats[it.ProductID]
		if seenCat[f.CategoryID] >= 2 {
			rest = append(rest, it)
			continue
		}
		seenCat[f.CategoryID]++
		out = append(out, it)
		if len(out) >= limit {
			return out
		}
	}
	for _, it := range rest {
		out = append(out, it)
		if len(out) >= limit {
			break
		}
	}
	return out
}
