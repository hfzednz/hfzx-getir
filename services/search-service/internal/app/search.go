package app

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/search-service/internal/domain"
)

// Search executes hybrid discovery search.
func (d *Deps) Search(ctx context.Context, q domain.SearchQuery) (domain.SearchResult, error) {
	start := time.Now()
	var out domain.SearchResult
	if q.TenantID == uuid.Nil {
		return out, domain.ErrInvalidArgument
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 || q.PageSize > 100 {
		q.PageSize = 20
	}
	if !domain.ValidSort(q.Sort) {
		return out, domain.ErrInvalidArgument
	}
	if q.Sort == "" {
		q.Sort = domain.SortRelevance
	}

	q.Normalized = domain.NormalizeQuery(q.RawQuery)
	out.Intent = domain.DetectIntent(q.Normalized)
	out.QueryID = d.newID()
	out.RewrittenQuery = q.Normalized

	// Synonym expansion + optional LLM rewrite
	expanded := q.Normalized
	if syns, err := d.Synonyms.List(ctx, q.TenantID, q.Locale); err == nil {
		expanded = applySynonyms(expanded, syns)
	}
	if d.LLM != nil && q.Normalized != "" {
		if rw, err := d.LLM.RewriteQuery(ctx, q.TenantID, expanded, q.Locale); err == nil && rw != "" {
			out.RewrittenQuery = rw
			expanded = rw
		}
	}

	tokens := domain.Tokenize(expanded)
	limit := q.PageSize * q.Page * 3
	if limit < 50 {
		limit = 50
	}

	var lexicalIDs []uuid.UUID
	lexicalScores := map[uuid.UUID]float64{}
	if d.Lexical != nil {
		ids, scores, err := d.Lexical.Search(ctx, q.TenantID, tokens, true, limit)
		if err != nil {
			return out, err
		}
		lexicalIDs, lexicalScores = ids, scores
	}

	vectorScores := map[uuid.UUID]float64{}
	var vectorIDs []uuid.UUID
	useHybrid := q.Hybrid || q.Sort == domain.SortAI
	if useHybrid && d.Embed != nil && d.Vectors != nil && expanded != "" {
		vec, err := d.Embed.EmbedText(ctx, q.TenantID, expanded, q.Locale)
		if err == nil && len(vec) > 0 {
			ids, scores, err := d.Vectors.Search(ctx, q.TenantID, vec, limit)
			if err == nil {
				vectorIDs, vectorScores = ids, scores
			}
		}
	}

	fused := domain.ReciprocalRankFusion([][]uuid.UUID{lexicalIDs, vectorIDs}, 60)
	if len(fused) == 0 {
		// browse / empty query: list docs
		docs, err := d.Docs.List(ctx, q.TenantID, limit)
		if err != nil {
			return out, err
		}
		for _, doc := range docs {
			fused[doc.ProductID] = doc.Popularity + 0.01
			lexicalScores[doc.ProductID] = doc.Popularity
		}
	}

	boosts, _ := d.Boosts.ListActive(ctx, q.TenantID, d.now())
	pinSet := map[uuid.UUID]bool{}
	sponsorSet := map[uuid.UUID]bool{}
	boostWeight := map[uuid.UUID]float64{}
	for _, b := range boosts {
		for _, pid := range b.ProductIDs {
			switch b.Kind {
			case "pin":
				pinSet[pid] = true
				boostWeight[pid] += 5 + b.Weight
			case "sponsor":
				sponsorSet[pid] = true
				boostWeight[pid] += 2 + b.Weight
			case "boost":
				boostWeight[pid] += b.Weight
			case "demote":
				boostWeight[pid] -= b.Weight
			}
		}
	}

	behaviorBonus := map[uuid.UUID]float64{}
	if q.Personalize && q.UserID != nil && d.Recs != nil {
		if ids, err := d.Recs.ForYou(ctx, q.TenantID, q.UserID, 20); err == nil {
			for i, id := range ids {
				behaviorBonus[id] = 1.0 / float64(i+1)
			}
		}
	}

	type scored struct {
		id  uuid.UUID
		hit domain.RankedHit
	}
	candidates := make([]scored, 0, len(fused))
	for pid, rrf := range fused {
		doc, err := d.Docs.Get(ctx, q.TenantID, pid)
		if err != nil {
			continue
		}
		if !passFilters(doc, q.Filters) {
			continue
		}
		lex := lexicalScores[pid]
		vec := vectorScores[pid]
		inv := 0.0
		if doc.Available {
			inv = 1
		}
		merch := boostWeight[pid]
		beh := behaviorBonus[pid]
		final := domain.ComputeRankScore(lex+rrf, vec, doc.Popularity, 0, 0, inv, doc.ProfitScore, doc.FreshnessScore, beh, merch)
		if q.Sort == domain.SortAI {
			final += beh * 0.2
		}
		hit := domain.RankedHit{
			ProductID: pid, VariantID: doc.VariantID, Score: final,
			LexicalScore: lex, VectorScore: vec, BehaviorScore: beh, MerchBoost: merch,
			Sponsored: sponsorSet[pid], Pinned: pinSet[pid],
		}
		if hit.Pinned {
			hit.Reasons = append(hit.Reasons, "pinned")
		}
		if hit.Sponsored {
			hit.Reasons = append(hit.Reasons, "sponsored")
		}
		candidates = append(candidates, scored{id: pid, hit: hit})
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		a, b := candidates[i].hit, candidates[j].hit
		if a.Pinned != b.Pinned {
			return a.Pinned
		}
		return compareSort(a, b, q.Sort, func(id uuid.UUID) domain.ProductDocument {
			doc, _ := d.Docs.Get(ctx, q.TenantID, id)
			return doc
		}, candidates[i].id, candidates[j].id)
	})

	// Did-you-mean from suggests when few hits
	if len(candidates) < 3 && q.Normalized != "" && d.Suggests != nil {
		if cands, err := d.Suggests.Suggest(ctx, q.TenantID, q.Normalized, 5); err == nil {
			bestDist := 3
			for _, c := range cands {
				dist := domain.Levenshtein(q.Normalized, domain.NormalizeQuery(c.Text))
				if dist > 0 && dist < bestDist {
					bestDist = dist
					out.DidYouMean = c.Text
				}
			}
		}
	}

	out.Total = len(candidates)
	out.ZeroResult = out.Total == 0
	startIdx := (q.Page - 1) * q.PageSize
	if startIdx > len(candidates) {
		startIdx = len(candidates)
	}
	endIdx := startIdx + q.PageSize
	if endIdx > len(candidates) {
		endIdx = len(candidates)
	}
	for _, c := range candidates[startIdx:endIdx] {
		out.Hits = append(out.Hits, c.hit)
	}

	if q.IncludeFacets {
		ids := make([]uuid.UUID, 0, len(candidates))
		for _, c := range candidates {
			ids = append(ids, c.id)
		}
		out.Facets = buildFacets(ctx, d, q.TenantID, ids)
	}

	_ = d.Trends.Bump(ctx, q.TenantID, "search", q.Normalized, nil, 1, d.now())
	for _, h := range out.Hits {
		pid := h.ProductID
		_ = d.Trends.Bump(ctx, q.TenantID, "product", pid.String(), &pid, 0.2, d.now())
	}

	out.TookMs = time.Since(start).Milliseconds()
	d.emit(ctx, q.TenantID, out.QueryID, domain.EventSearchPerformed, map[string]any{
		"query": q.Normalized, "total": out.Total, "zeroResult": out.ZeroResult,
		"intent": out.Intent, "tookMs": out.TookMs,
	})
	return out, nil
}

func applySynonyms(q string, rules []domain.SynonymRule) string {
	out := q
	for _, r := range rules {
		if !r.Active {
			continue
		}
		term := domain.NormalizeQuery(r.Term)
		if term == "" || !strings.Contains(out, term) {
			continue
		}
		for _, s := range r.Synonyms {
			s = domain.NormalizeQuery(s)
			if s != "" && !strings.Contains(out, s) {
				out = out + " " + s
			}
		}
	}
	return out
}

func passFilters(doc domain.ProductDocument, f domain.SearchFilters) bool {
	if f.AvailableOnly && !doc.Available {
		return false
	}
	if f.PriceMin != nil && doc.PriceMinor < *f.PriceMin {
		return false
	}
	if f.PriceMax != nil && doc.PriceMinor > *f.PriceMax {
		return false
	}
	if f.DiscountMin != nil && doc.DiscountPct < *f.DiscountMin {
		return false
	}
	if f.RatingMin != nil && doc.RatingAvg < *f.RatingMin {
		return false
	}
	if f.ReviewMin != nil && doc.ReviewCount < *f.ReviewMin {
		return false
	}
	if f.MaxETAMin != nil && doc.DeliveryETAMin > *f.MaxETAMin {
		return false
	}
	if f.WarehouseID != nil {
		ok := false
		for _, w := range doc.WarehouseIDs {
			if w == *f.WarehouseID {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	if len(f.BrandIDs) > 0 {
		ok := false
		for _, b := range f.BrandIDs {
			if doc.BrandID == b {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	if len(f.CategoryIDs) > 0 {
		ok := false
		for _, want := range f.CategoryIDs {
			for _, have := range doc.CategoryIDs {
				if want == have {
					ok = true
					break
				}
			}
		}
		if !ok {
			return false
		}
	}
	for k, v := range f.Attrs {
		if doc.Attributes[k] != v {
			return false
		}
	}
	return true
}

func compareSort(a, b domain.RankedHit, sortMode string, docFn func(uuid.UUID) domain.ProductDocument, idA, idB uuid.UUID) bool {
	da, db := docFn(idA), docFn(idB)
	switch sortMode {
	case domain.SortPopularity:
		return da.Popularity > db.Popularity
	case domain.SortNewest:
		return da.IndexedAt.After(db.IndexedAt)
	case domain.SortPriceAsc:
		return da.PriceMinor < db.PriceMinor
	case domain.SortPriceDesc:
		return da.PriceMinor > db.PriceMinor
	case domain.SortRating:
		return da.RatingAvg > db.RatingAvg
	case domain.SortDiscount:
		return da.DiscountPct > db.DiscountPct
	case domain.SortDeliverySpeed:
		return da.DeliveryETAMin < db.DeliveryETAMin
	default:
		return a.Score > b.Score
	}
}

func buildFacets(ctx context.Context, d *Deps, tenantID uuid.UUID, productIDs []uuid.UUID) map[string][]domain.FacetBucket {
	brands := map[string]int{}
	cats := map[string]int{}
	for _, id := range productIDs {
		doc, err := d.Docs.Get(ctx, tenantID, id)
		if err != nil {
			continue
		}
		if doc.BrandName != "" {
			brands[doc.BrandName]++
		}
		for _, p := range doc.CategoryPath {
			cats[p]++
		}
	}
	return map[string][]domain.FacetBucket{
		"brand":    toBuckets(brands),
		"category": toBuckets(cats),
	}
}

func toBuckets(m map[string]int) []domain.FacetBucket {
	out := make([]domain.FacetBucket, 0, len(m))
	for k, v := range m {
		out = append(out, domain.FacetBucket{Key: k, Count: v})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Count > out[j].Count })
	return out
}

// Autocomplete returns prefix suggestions.
func (d *Deps) Autocomplete(ctx context.Context, tenantID uuid.UUID, prefix string, limit int) ([]domain.SuggestCandidate, error) {
	if tenantID == uuid.Nil {
		return nil, domain.ErrInvalidArgument
	}
	prefix = domain.NormalizeQuery(prefix)
	if limit <= 0 {
		limit = 10
	}
	return d.Suggests.Suggest(ctx, tenantID, prefix, limit)
}

// RecordSuggestionClick analytics.
func (d *Deps) RecordSuggestionClick(ctx context.Context, tenantID uuid.UUID, text string, productID *uuid.UUID) error {
	id := d.newID()
	d.emit(ctx, tenantID, id, domain.EventSuggestionClicked, map[string]any{
		"text": text, "productId": productID,
	})
	_ = d.Trends.Bump(ctx, tenantID, "search", domain.NormalizeQuery(text), productID, 2, d.now())
	return nil
}
