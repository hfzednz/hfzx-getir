package app

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/search-service/internal/domain"
)

// IndexDocument upserts a product into lexical + vector indexes.
func (d *Deps) IndexDocument(ctx context.Context, doc domain.ProductDocument) error {
	if doc.TenantID == uuid.Nil || doc.ProductID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	now := d.now()
	doc.IndexedAt = now
	if doc.Title == "" {
		return domain.ErrInvalidArgument
	}

	// Enrich from ports when missing
	if d.Inventory != nil {
		if ok, err := d.Inventory.Available(ctx, doc.TenantID, doc.ProductID, nil); err == nil {
			doc.Available = ok
		}
	}
	if d.Pricing != nil && doc.PriceMinor == 0 {
		if p, c, cur, err := d.Pricing.PriceHint(ctx, doc.TenantID, doc.ProductID); err == nil {
			doc.PriceMinor, doc.CompareAtMinor, doc.Currency = p, c, cur
			if c > 0 && p < c {
				doc.DiscountPct = float64(c-p) / float64(c) * 100
			}
		}
	}
	if d.Reviews != nil && doc.ReviewCount == 0 {
		if avg, n, err := d.Reviews.RatingHint(ctx, doc.TenantID, doc.ProductID); err == nil {
			doc.RatingAvg, doc.ReviewCount = avg, n
		}
	}

	text := doc.Title + " " + doc.Description + " " + doc.BrandName + " " + strings.Join(doc.Tags, " ")
	if d.Embed != nil {
		if vec, err := d.Embed.EmbedText(ctx, doc.TenantID, text, "tr-TR"); err == nil {
			doc.Embedding = vec
			d.emit(ctx, doc.TenantID, doc.ProductID, domain.EventEmbeddingGenerated, map[string]any{"dims": len(vec)})
			if d.Vectors != nil {
				_ = d.Vectors.Upsert(ctx, doc.TenantID, doc.ProductID, vec)
			}
		}
	}

	if err := d.Docs.Upsert(ctx, doc); err != nil {
		return err
	}
	if d.Lexical != nil {
		if err := d.Lexical.Index(ctx, doc); err != nil {
			return err
		}
	}
	_ = d.Suggests.Upsert(ctx, doc.TenantID, domain.SuggestCandidate{
		Text: doc.Title, ProductID: &doc.ProductID, Weight: doc.Popularity + 1,
	})
	if doc.BrandName != "" {
		_ = d.Suggests.Upsert(ctx, doc.TenantID, domain.SuggestCandidate{
			Text: doc.BrandName, Weight: 0.5,
		})
	}

	d.emit(ctx, doc.TenantID, doc.ProductID, domain.EventIndexUpdated, map[string]any{
		"version": doc.Version, "title": doc.Title,
	})
	d.emit(ctx, doc.TenantID, doc.ProductID, domain.EventProductRankUpdated, map[string]any{
		"popularity": doc.Popularity, "rating": doc.RatingAvg,
	})
	return nil
}

// ReindexFromCatalog pulls one product via catalog port and indexes it.
func (d *Deps) ReindexFromCatalog(ctx context.Context, tenantID, productID uuid.UUID) error {
	if d.Catalog == nil {
		return domain.ErrInvalidArgument
	}
	doc, err := d.Catalog.FetchProduct(ctx, tenantID, productID)
	if err != nil {
		return err
	}
	doc.TenantID = tenantID
	doc.ProductID = productID
	return d.IndexDocument(ctx, doc)
}

// StartIndexJob creates a job record (full/batch orchestration stub).
func (d *Deps) StartIndexJob(ctx context.Context, tenantID uuid.UUID, mode string) (domain.IndexJob, error) {
	var zero domain.IndexJob
	switch mode {
	case domain.IndexFull, domain.IndexIncremental, domain.IndexRealtime, domain.IndexBatch:
	default:
		return zero, domain.ErrInvalidArgument
	}
	now := d.now()
	job := domain.IndexJob{
		ID: d.newID(), TenantID: tenantID, Mode: mode, Status: "running",
		CreatedAt: now, UpdatedAt: now,
	}
	docs, err := d.Docs.List(ctx, tenantID, 10000)
	if err != nil {
		return zero, err
	}
	job.DocsTotal = len(docs)
	for _, doc := range docs {
		if err := d.IndexDocument(ctx, doc); err != nil {
			job.Status = "failed"
			job.Error = err.Error()
			job.UpdatedAt = d.now()
			_ = d.Jobs.Save(ctx, job)
			return job, err
		}
		job.DocsDone++
	}
	job.Status = "done"
	job.UpdatedAt = d.now()
	_ = d.Jobs.Save(ctx, job)
	return job, nil
}

// UpsertSynonym saves a synonym rule.
func (d *Deps) UpsertSynonym(ctx context.Context, s domain.SynonymRule) (domain.SynonymRule, error) {
	if s.TenantID == uuid.Nil || strings.TrimSpace(s.Term) == "" || len(s.Synonyms) == 0 {
		return s, domain.ErrInvalidArgument
	}
	if s.ID == uuid.Nil {
		s.ID = d.newID()
	}
	s.Term = domain.NormalizeQuery(s.Term)
	s.Active = true
	s.UpdatedAt = d.now()
	return s, d.Synonyms.Save(ctx, s)
}

// UpsertBoost saves merchandising rule.
func (d *Deps) UpsertBoost(ctx context.Context, b domain.BoostRule) (domain.BoostRule, error) {
	if b.TenantID == uuid.Nil || b.Name == "" {
		return b, domain.ErrInvalidArgument
	}
	switch b.Kind {
	case "pin", "boost", "demote", "sponsor":
	default:
		return b, domain.ErrInvalidArgument
	}
	if b.ID == uuid.Nil {
		b.ID = d.newID()
	}
	b.Active = true
	b.UpdatedAt = d.now()
	return b, d.Boosts.Save(ctx, b)
}

// ListTrends returns trend leaderboard.
func (d *Deps) ListTrends(ctx context.Context, tenantID uuid.UUID, kind string, limit int) ([]domain.TrendEntry, error) {
	if limit <= 0 {
		limit = 20
	}
	return d.Trends.List(ctx, tenantID, kind, limit)
}

// RefreshTrends emits TrendingUpdated.
func (d *Deps) RefreshTrends(ctx context.Context, tenantID uuid.UUID) error {
	d.emit(ctx, tenantID, d.newID(), domain.EventTrendingUpdated, map[string]any{"scope": "tenant"})
	return nil
}

// VoiceSearch parses STT text then searches (speech-to-text is external).
func (d *Deps) VoiceSearch(ctx context.Context, tenantID uuid.UUID, transcript string, locale string, userID *uuid.UUID) (domain.SearchResult, error) {
	return d.Search(ctx, domain.SearchQuery{
		TenantID: tenantID, RawQuery: transcript, Locale: locale, UserID: userID,
		Hybrid: true, Personalize: userID != nil, Sort: domain.SortAI, Page: 1, PageSize: 20,
	})
}

// ImageSearch embeds image ref and runs vector search.
func (d *Deps) ImageSearch(ctx context.Context, tenantID uuid.UUID, imageRef string, limit int) (domain.SearchResult, error) {
	var out domain.SearchResult
	out.QueryID = d.newID()
	out.Intent = domain.IntentFind
	if d.Embed == nil || d.Vectors == nil {
		return out, domain.ErrInvalidArgument
	}
	if limit <= 0 {
		limit = 20
	}
	vec, err := d.Embed.EmbedImage(ctx, tenantID, imageRef)
	if err != nil {
		return out, err
	}
	ids, scores, err := d.Vectors.Search(ctx, tenantID, vec, limit)
	if err != nil {
		return out, err
	}
	for _, id := range ids {
		doc, err := d.Docs.Get(ctx, tenantID, id)
		if err != nil {
			continue
		}
		out.Hits = append(out.Hits, domain.RankedHit{
			ProductID: id, VariantID: doc.VariantID, Score: scores[id], VectorScore: scores[id],
			Reasons: []string{"visual_similarity"},
		})
	}
	out.Total = len(out.Hits)
	out.ZeroResult = out.Total == 0
	d.emit(ctx, tenantID, out.QueryID, domain.EventSearchPerformed, map[string]any{
		"mode": "image", "total": out.Total,
	})
	return out, nil
}

// AdminStats dashboard counters.
func (d *Deps) AdminStats(ctx context.Context, tenantID uuid.UUID) (map[string]any, error) {
	docs, _ := d.Docs.List(ctx, tenantID, 100000)
	zeroish := 0
	trends, _ := d.Trends.List(ctx, tenantID, "search", 100)
	for _, t := range trends {
		if t.Score < 0.01 {
			zeroish++
		}
	}
	return map[string]any{
		"indexedDocuments": len(docs),
		"trendKeys":        len(trends),
		"tenantId":         tenantID.String(),
	}, nil
}

// SummarizeSearch LLM summary of hit titles.
func (d *Deps) SummarizeSearch(ctx context.Context, tenantID uuid.UUID, query string, hits []domain.RankedHit) (string, error) {
	if d.LLM == nil {
		return "", domain.ErrInvalidArgument
	}
	titles := make([]string, 0, len(hits))
	for _, h := range hits {
		if doc, err := d.Docs.Get(ctx, tenantID, h.ProductID); err == nil {
			titles = append(titles, doc.Title)
		}
	}
	return d.LLM.SummarizeResults(ctx, tenantID, query, titles)
}
