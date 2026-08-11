package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/recommendation-service/internal/domain"
)

// IngestSignal records a behavior event and updates co-occurrence on purchase/cart.
func (d *Deps) IngestSignal(ctx context.Context, s domain.BehaviorSignal) (domain.BehaviorSignal, error) {
	if s.TenantID == uuid.Nil || s.UserID == uuid.Nil || s.ProductID == uuid.Nil {
		return s, domain.ErrInvalidArgument
	}
	if s.Weight <= 0 {
		s.Weight = domain.SignalWeight(s.Kind)
	}
	if s.ID == uuid.Nil {
		s.ID = d.newID()
	}
	s.CreatedAt = d.now()
	if err := d.Signals.Save(ctx, s); err != nil {
		return s, err
	}

	if s.Kind == domain.SignalPurchase || s.Kind == domain.SignalCart {
		recent, _ := d.Signals.ListByUser(ctx, s.TenantID, s.UserID, 50)
		for _, other := range recent {
			if other.ProductID == s.ProductID {
				continue
			}
			if other.Kind == domain.SignalPurchase || other.Kind == domain.SignalCart || other.Kind == domain.SignalView {
				_ = d.CoOccur.Bump(ctx, s.TenantID, s.ProductID, other.ProductID, 1, d.now())
			}
		}
	}

	d.emit(ctx, s.TenantID, s.ID, domain.EventSignalIngested, map[string]any{
		"kind": s.Kind, "productId": s.ProductID.String(), "userId": s.UserID.String(),
	})
	return s, nil
}

// UpsertFeatures stores content features for recommendations.
func (d *Deps) UpsertFeatures(ctx context.Context, f domain.ProductFeatures) error {
	if f.ProductID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	return d.Features.Upsert(ctx, f)
}

// Recommend builds a rail for the requested strategy.
func (d *Deps) Recommend(ctx context.Context, req domain.RecommendRequest) (domain.RecommendationRail, error) {
	var rail domain.RecommendationRail
	if req.TenantID == uuid.Nil {
		return rail, domain.ErrInvalidArgument
	}
	if !domain.ValidStrategy(req.Strategy) {
		return rail, domain.ErrInvalidArgument
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Strategy == "" {
		req.Strategy = domain.StrategyHybrid
	}
	if req.Context == "" {
		req.Context = "home"
	}

	exclude := map[uuid.UUID]struct{}{}
	for _, id := range req.ExcludeIDs {
		exclude[id] = struct{}{}
	}
	for _, id := range req.CartProductIDs {
		exclude[id] = struct{}{}
	}
	if req.ProductID != nil {
		exclude[*req.ProductID] = struct{}{}
	}

	var items []domain.RecommendationItem
	var err error
	switch req.Strategy {
	case domain.StrategyFBT:
		items, err = d.fbt(ctx, req, exclude)
	case domain.StrategyContent:
		items, err = d.contentBased(ctx, req, exclude)
	case domain.StrategyCollaborative:
		items, err = d.collaborative(ctx, req, exclude)
	case domain.StrategyUpsell:
		items, err = d.upsell(ctx, req, exclude)
	case domain.StrategyCrossSell:
		items, err = d.crossSell(ctx, req, exclude)
	case domain.StrategyTrending:
		items, err = d.trending(ctx, req, exclude)
	case domain.StrategyPersonalized:
		items, err = d.personalized(ctx, req, exclude)
	default: // hybrid
		items, err = d.hybrid(ctx, req, exclude)
	}
	if err != nil {
		return rail, err
	}

	feats := map[uuid.UUID]domain.ProductFeatures{}
	for _, it := range items {
		if f, e := d.Features.Get(ctx, it.ProductID); e == nil {
			feats[it.ProductID] = f
		}
	}
	items = domain.DiversifyByCategory(items, feats, req.Limit)
	for i := range items {
		if items[i].Strategy == "" {
			items[i].Strategy = req.Strategy
		}
		if items[i].Reason == "" {
			items[i].Reason = req.Strategy
		}
	}

	rail = domain.RecommendationRail{
		ID: d.newID(), TenantID: req.TenantID, Name: req.Context + ":" + req.Strategy,
		Strategy: req.Strategy, Context: req.Context, Items: items, UserID: req.UserID, CreatedAt: d.now(),
	}
	productIDs := make([]string, 0, len(items))
	for _, it := range items {
		productIDs = append(productIDs, it.ProductID.String())
	}
	d.emit(ctx, req.TenantID, rail.ID, domain.EventRecommendationShown, map[string]any{
		"strategy": req.Strategy, "context": req.Context, "productIds": productIDs,
	})
	return rail, nil
}

func (d *Deps) fbt(ctx context.Context, req domain.RecommendRequest, exclude map[uuid.UUID]struct{}) ([]domain.RecommendationItem, error) {
	scores := map[uuid.UUID]float64{}
	seeds := req.CartProductIDs
	if req.ProductID != nil {
		seeds = append(seeds, *req.ProductID)
	}
	for _, seed := range seeds {
		rows, err := d.CoOccur.TopFor(ctx, req.TenantID, seed, req.Limit*3)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			other := row.ProductB
			if other == seed {
				other = row.ProductA
			}
			scores[other] += float64(row.Count)
		}
	}
	items := domain.TopN(scores, exclude, req.Limit)
	for i := range items {
		items[i].Strategy = domain.StrategyFBT
		items[i].Reason = "frequently_bought_together"
	}
	return items, nil
}

func (d *Deps) contentBased(ctx context.Context, req domain.RecommendRequest, exclude map[uuid.UUID]struct{}) ([]domain.RecommendationItem, error) {
	if req.ProductID == nil {
		return d.trending(ctx, req, exclude)
	}
	seed, err := d.Features.Get(ctx, *req.ProductID)
	if err != nil {
		if d.Catalog != nil {
			seed, err = d.Catalog.Features(ctx, req.TenantID, *req.ProductID)
			if err != nil {
				return nil, err
			}
			_ = d.Features.Upsert(ctx, seed)
		} else {
			return nil, err
		}
	}
	all, err := d.Features.ListAll(ctx, 500)
	if err != nil {
		return nil, err
	}
	scores := map[uuid.UUID]float64{}
	for _, f := range all {
		if f.ProductID == seed.ProductID {
			continue
		}
		scores[f.ProductID] = domain.ContentSimilarity(seed, f)
	}
	items := domain.TopN(scores, exclude, req.Limit)
	for i := range items {
		items[i].Strategy = domain.StrategyContent
		items[i].Reason = "similar_content"
	}
	return items, nil
}

func (d *Deps) collaborative(ctx context.Context, req domain.RecommendRequest, exclude map[uuid.UUID]struct{}) ([]domain.RecommendationItem, error) {
	scores := map[uuid.UUID]float64{}
	if req.UserID == nil {
		return d.trending(ctx, req, exclude)
	}
	mySignals, err := d.Signals.ListByUser(ctx, req.TenantID, *req.UserID, 100)
	if err != nil {
		return nil, err
	}
	neighborWeight := map[uuid.UUID]float64{}
	for _, s := range mySignals {
		users, _ := d.Signals.UsersWhoInteracted(ctx, req.TenantID, s.ProductID, 50)
		for _, u := range users {
			if u == *req.UserID {
				continue
			}
			neighborWeight[u] += s.Weight
		}
	}
	for u, w := range neighborWeight {
		sigs, _ := d.Signals.ListByUser(ctx, req.TenantID, u, 50)
		for _, s := range sigs {
			scores[s.ProductID] += w * s.Weight
		}
	}
	items := domain.TopN(scores, exclude, req.Limit)
	for i := range items {
		items[i].Strategy = domain.StrategyCollaborative
		items[i].Reason = "users_like_you"
	}
	return items, nil
}

func (d *Deps) upsell(ctx context.Context, req domain.RecommendRequest, exclude map[uuid.UUID]struct{}) ([]domain.RecommendationItem, error) {
	if req.ProductID == nil {
		return nil, domain.ErrInvalidArgument
	}
	seed, err := d.Features.Get(ctx, *req.ProductID)
	if err != nil {
		return nil, err
	}
	all, err := d.Features.ListAll(ctx, 500)
	if err != nil {
		return nil, err
	}
	scores := map[uuid.UUID]float64{}
	for _, f := range all {
		if f.CategoryID != seed.CategoryID || f.ProductID == seed.ProductID {
			continue
		}
		if f.PriceMinor > seed.PriceMinor {
			scores[f.ProductID] = float64(f.PriceMinor-seed.PriceMinor)/float64(seed.PriceMinor+1) + f.RatingAvg/5
		}
	}
	items := domain.TopN(scores, exclude, req.Limit)
	for i := range items {
		items[i].Strategy = domain.StrategyUpsell
		items[i].Reason = "upsell"
	}
	return items, nil
}

func (d *Deps) crossSell(ctx context.Context, req domain.RecommendRequest, exclude map[uuid.UUID]struct{}) ([]domain.RecommendationItem, error) {
	items, err := d.fbt(ctx, req, exclude)
	if err != nil {
		return nil, err
	}
	if len(items) > 0 {
		for i := range items {
			items[i].Strategy = domain.StrategyCrossSell
			items[i].Reason = "cross_sell"
		}
		return items, nil
	}
	// fallback: different category content
	if req.ProductID == nil {
		return d.trending(ctx, req, exclude)
	}
	seed, err := d.Features.Get(ctx, *req.ProductID)
	if err != nil {
		return nil, err
	}
	all, _ := d.Features.ListAll(ctx, 500)
	scores := map[uuid.UUID]float64{}
	for _, f := range all {
		if f.CategoryID == seed.CategoryID {
			continue
		}
		scores[f.ProductID] = f.Popularity + f.RatingAvg
	}
	items = domain.TopN(scores, exclude, req.Limit)
	for i := range items {
		items[i].Strategy = domain.StrategyCrossSell
		items[i].Reason = "cross_category"
	}
	return items, nil
}

func (d *Deps) trending(ctx context.Context, req domain.RecommendRequest, exclude map[uuid.UUID]struct{}) ([]domain.RecommendationItem, error) {
	scores := map[uuid.UUID]float64{}
	if d.Trends != nil {
		ids, err := d.Trends.TrendingProductIDs(ctx, req.TenantID, req.Limit*2)
		if err == nil {
			for i, id := range ids {
				scores[id] = float64(len(ids)-i) + 1
			}
		}
	}
	if len(scores) == 0 {
		all, err := d.Features.ListAll(ctx, 200)
		if err != nil {
			return nil, err
		}
		for _, f := range all {
			scores[f.ProductID] = f.Popularity + f.RatingAvg
		}
	}
	items := domain.TopN(scores, exclude, req.Limit)
	for i := range items {
		items[i].Strategy = domain.StrategyTrending
		items[i].Reason = "trending"
	}
	return items, nil
}

func (d *Deps) personalized(ctx context.Context, req domain.RecommendRequest, exclude map[uuid.UUID]struct{}) ([]domain.RecommendationItem, error) {
	cf, err := d.collaborative(ctx, req, exclude)
	if err != nil {
		return nil, err
	}
	tr, err := d.trending(ctx, req, exclude)
	if err != nil {
		return nil, err
	}
	scores := map[uuid.UUID]float64{}
	for _, it := range cf {
		scores[it.ProductID] += it.Score * 0.7
	}
	for _, it := range tr {
		scores[it.ProductID] += it.Score * 0.3
	}
	// boost from user history categories
	if req.UserID != nil {
		sigs, _ := d.Signals.ListByUser(ctx, req.TenantID, *req.UserID, 50)
		for _, s := range sigs {
			if f, e := d.Features.Get(ctx, s.ProductID); e == nil {
				all, _ := d.Features.ListAll(ctx, 200)
				for _, cand := range all {
					if cand.CategoryID == f.CategoryID {
						scores[cand.ProductID] += 0.2 * s.Weight
					}
				}
			}
		}
	}
	items := domain.TopN(scores, exclude, req.Limit)
	for i := range items {
		items[i].Strategy = domain.StrategyPersonalized
		items[i].Reason = "for_you"
	}
	return items, nil
}

func (d *Deps) hybrid(ctx context.Context, req domain.RecommendRequest, exclude map[uuid.UUID]struct{}) ([]domain.RecommendationItem, error) {
	cf, _ := d.collaborative(ctx, req, exclude)
	cb, _ := d.contentBased(ctx, req, exclude)
	fb, _ := d.fbt(ctx, req, exclude)
	tr, _ := d.trending(ctx, req, exclude)
	m1, m2, m3, m4 := map[uuid.UUID]float64{}, map[uuid.UUID]float64{}, map[uuid.UUID]float64{}, map[uuid.UUID]float64{}
	for _, it := range cf {
		m1[it.ProductID] = it.Score
	}
	for _, it := range cb {
		m2[it.ProductID] = it.Score
	}
	for _, it := range fb {
		m3[it.ProductID] = it.Score
	}
	for _, it := range tr {
		m4[it.ProductID] = it.Score
	}
	blended := domain.BlendScores([]map[uuid.UUID]float64{m1, m2, m3, m4}, []float64{0.35, 0.25, 0.25, 0.15})
	items := domain.TopN(blended, exclude, req.Limit)
	for i := range items {
		items[i].Strategy = domain.StrategyHybrid
		items[i].Reason = "hybrid"
	}
	return items, nil
}

// RecordClick recommendation click analytics.
func (d *Deps) RecordClick(ctx context.Context, tenantID, railID, productID uuid.UUID, userID *uuid.UUID) error {
	d.emit(ctx, tenantID, railID, domain.EventRecommendationClicked, map[string]any{
		"productId": productID.String(), "userId": userID,
	})
	if userID != nil {
		_, _ = d.IngestSignal(ctx, domain.BehaviorSignal{
			TenantID: tenantID, UserID: *userID, ProductID: productID, Kind: domain.SignalClick,
		})
	}
	return nil
}

// AdminStats counters.
func (d *Deps) AdminStats(ctx context.Context, tenantID uuid.UUID) (map[string]any, error) {
	_ = tenantID
	all, _ := d.Features.ListAll(ctx, 100000)
	return map[string]any{"features": len(all)}, nil
}
