// Package memory provides in-memory port implementations for tests and dev mode.
package memory

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/app/ports"
	"github.com/nexora/catalog-service/internal/domain"
)

// Store is a shared in-memory database for all repositories.
type Store struct {
	mu sync.RWMutex

	Products           map[uuid.UUID]domain.Product
	ProductSlug        map[string]uuid.UUID // tenant:slug -> id
	Variants           map[uuid.UUID]domain.Variant
	SKUs               map[uuid.UUID]domain.SKUIdentifier
	SKUByValue         map[string]uuid.UUID // tenant:type:value
	Categories         map[uuid.UUID]domain.Category
	CategorySlug       map[string]uuid.UUID
	ProductCategories  map[string]domain.ProductCategory // product:category
	Brands             map[uuid.UUID]domain.Brand
	BrandSlug          map[string]uuid.UUID
	AttributeDefs      map[uuid.UUID]domain.AttributeDef
	AttributeDefCode   map[string]uuid.UUID
	ProductAttributes  map[uuid.UUID]domain.ProductAttribute
	Locales            map[uuid.UUID]domain.ProductLocale
	LocaleKey          map[string]uuid.UUID // tenant:product:lang
	SEO                map[uuid.UUID]domain.SEO
	SEOKey             map[string]uuid.UUID
	Media              map[uuid.UUID]domain.ProductMedia
	Bundles            map[uuid.UUID]domain.Bundle
	BundleByProduct    map[uuid.UUID]uuid.UUID
	BundleItems        map[uuid.UUID][]domain.BundleItem
	Relations          map[uuid.UUID]domain.ProductRelation
	Versions           map[uuid.UUID]domain.ProductVersion
	VersionCount       map[string]int // tenant:product -> max version
	Workflow           []domain.ApprovalAction
	ImportJobs         map[uuid.UUID]domain.ImportJob
	Compliance         map[uuid.UUID]domain.ProductCompliance
	ComplianceByProduct map[uuid.UUID]uuid.UUID
	Suppliers          map[uuid.UUID]domain.Supplier
	SupplierCode       map[string]uuid.UUID
	SupplierProducts   map[uuid.UUID]domain.SupplierProduct

	SearchDocs map[uuid.UUID]ports.SearchDocument
	Events     []publishedEvent
}

type publishedEvent struct {
	Topic   string
	Key     string
	Payload any
}

// NewStore returns an empty in-memory store.
func NewStore() *Store {
	return &Store{
		Products:            make(map[uuid.UUID]domain.Product),
		ProductSlug:         make(map[string]uuid.UUID),
		Variants:            make(map[uuid.UUID]domain.Variant),
		SKUs:                make(map[uuid.UUID]domain.SKUIdentifier),
		SKUByValue:          make(map[string]uuid.UUID),
		Categories:          make(map[uuid.UUID]domain.Category),
		CategorySlug:        make(map[string]uuid.UUID),
		ProductCategories:   make(map[string]domain.ProductCategory),
		Brands:              make(map[uuid.UUID]domain.Brand),
		BrandSlug:           make(map[string]uuid.UUID),
		AttributeDefs:       make(map[uuid.UUID]domain.AttributeDef),
		AttributeDefCode:    make(map[string]uuid.UUID),
		ProductAttributes:   make(map[uuid.UUID]domain.ProductAttribute),
		Locales:             make(map[uuid.UUID]domain.ProductLocale),
		LocaleKey:           make(map[string]uuid.UUID),
		SEO:                 make(map[uuid.UUID]domain.SEO),
		SEOKey:              make(map[string]uuid.UUID),
		Media:               make(map[uuid.UUID]domain.ProductMedia),
		Bundles:             make(map[uuid.UUID]domain.Bundle),
		BundleByProduct:     make(map[uuid.UUID]uuid.UUID),
		BundleItems:         make(map[uuid.UUID][]domain.BundleItem),
		Relations:           make(map[uuid.UUID]domain.ProductRelation),
		Versions:            make(map[uuid.UUID]domain.ProductVersion),
		VersionCount:        make(map[string]int),
		ImportJobs:          make(map[uuid.UUID]domain.ImportJob),
		Compliance:          make(map[uuid.UUID]domain.ProductCompliance),
		ComplianceByProduct: make(map[uuid.UUID]uuid.UUID),
		Suppliers:           make(map[uuid.UUID]domain.Supplier),
		SupplierCode:        make(map[string]uuid.UUID),
		SupplierProducts:    make(map[uuid.UUID]domain.SupplierProduct),
		SearchDocs:          make(map[uuid.UUID]ports.SearchDocument),
	}
}

func tenantKey(tenantID uuid.UUID, parts ...string) string {
	b := strings.Builder{}
	b.WriteString(tenantID.String())
	for _, p := range parts {
		b.WriteByte(':')
		b.WriteString(p)
	}
	return b.String()
}

// Clock is a fixed clock for tests.
type Clock struct{ T time.Time }

func (c *Clock) Now() time.Time {
	if c.T.IsZero() {
		return time.Now().UTC()
	}
	return c.T.UTC()
}

// IDGen generates random UUIDs (same as production).
type IDGen struct{}

func (IDGen) New() uuid.UUID { return uuid.New() }

// EventPublisher records published events.
type EventPublisher struct{ S *Store }

func (p *EventPublisher) Publish(_ context.Context, topic, key string, payload any) error {
	p.S.mu.Lock()
	defer p.S.mu.Unlock()
	p.S.Events = append(p.S.Events, publishedEvent{Topic: topic, Key: key, Payload: payload})
	return nil
}

// SearchIndexer is an in-memory search index.
type SearchIndexer struct{ S *Store }

func (i *SearchIndexer) IndexProduct(_ context.Context, doc ports.SearchDocument) error {
	i.S.mu.Lock()
	defer i.S.mu.Unlock()
	i.S.SearchDocs[doc.ProductID] = doc
	return nil
}

func (i *SearchIndexer) DeleteProduct(_ context.Context, _, productID uuid.UUID) error {
	i.S.mu.Lock()
	defer i.S.mu.Unlock()
	delete(i.S.SearchDocs, productID)
	return nil
}

func (i *SearchIndexer) Search(_ context.Context, q ports.SearchQuery) (ports.SearchResult, error) {
	i.S.mu.RLock()
	defer i.S.mu.RUnlock()
	hits := make([]ports.SearchDocument, 0)
	query := strings.ToLower(strings.TrimSpace(q.Query))
	for _, doc := range i.S.SearchDocs {
		if doc.TenantID != q.TenantID {
			continue
		}
		if q.Status != nil && doc.Status != *q.Status {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(doc.Title), query) &&
			!strings.Contains(strings.ToLower(doc.SKU), query) {
			continue
		}
		hits = append(hits, doc)
	}
	sort.Slice(hits, func(a, b int) bool { return hits[a].ProductID.String() < hits[b].ProductID.String() })
	total := len(hits)
	if q.Offset >= len(hits) {
		return ports.SearchResult{Total: total}, nil
	}
	end := q.Offset + q.Limit
	if end > len(hits) {
		end = len(hits)
	}
	return ports.SearchResult{Total: total, Hits: hits[q.Offset:end]}, nil
}

func (i *SearchIndexer) Suggest(_ context.Context, tenantID uuid.UUID, prefix string, limit int) ([]string, error) {
	i.S.mu.RLock()
	defer i.S.mu.RUnlock()
	prefix = strings.ToLower(prefix)
	out := make([]string, 0)
	seen := map[string]struct{}{}
	for _, doc := range i.S.SearchDocs {
		if doc.TenantID != tenantID {
			continue
		}
		if strings.HasPrefix(strings.ToLower(doc.Title), prefix) {
			if _, ok := seen[doc.Title]; !ok {
				out = append(out, doc.Title)
				seen[doc.Title] = struct{}{}
			}
		}
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (i *SearchIndexer) ReindexAll(_ context.Context, _ uuid.UUID) error { return nil }

// MediaClient stub resolves CDN URLs locally.
type MediaClient struct{}

func (MediaClient) GetAsset(_ context.Context, _, assetID uuid.UUID) (ports.MediaAsset, error) {
	return ports.MediaAsset{
		ID:     assetID,
		Kind:   domain.MediaKindImage,
		CDNURL: "https://cdn.nexora.local/media/" + assetID.String(),
	}, nil
}

// AIClient stub returns placeholder AI responses.
type AIClient struct{}

func (AIClient) Describe(_ context.Context, _, _ uuid.UUID) (ports.AIDescribeResult, error) {
	return ports.AIDescribeResult{Title: "AI title", Description: "AI description"}, nil
}
func (AIClient) Translate(_ context.Context, _, _ uuid.UUID, lang string) (ports.AITranslateResult, error) {
	return ports.AITranslateResult{Lang: lang, Title: "translated"}, nil
}
func (AIClient) Categorize(_ context.Context, _, _ uuid.UUID) (ports.AICategorizeResult, error) {
	return ports.AICategorizeResult{Confidence: 0.8}, nil
}
func (AIClient) QualityScore(_ context.Context, _, _ uuid.UUID) (ports.AIQualityResult, error) {
	return ports.AIQualityResult{Score: 0.9}, nil
}

// Repos wiring helpers.
func NewRepos(s *Store) (
	*ProductRepo, *VariantRepo, *SKURepo, *CategoryRepo, *BrandRepo,
	*AttributeRepo, *LocaleRepo, *SEORepo, *MediaRepo, *BundleRepo,
	*RelationRepo, *VersionRepo, *WorkflowRepo, *ImportJobRepo, *ComplianceRepo,
) {
	return &ProductRepo{S: s}, &VariantRepo{S: s}, &SKURepo{S: s}, &CategoryRepo{S: s}, &BrandRepo{S: s},
		&AttributeRepo{S: s}, &LocaleRepo{S: s}, &SEORepo{S: s}, &MediaRepo{S: s}, &BundleRepo{S: s},
		&RelationRepo{S: s}, &VersionRepo{S: s}, &WorkflowRepo{S: s}, &ImportJobRepo{S: s}, &ComplianceRepo{S: s}
}
