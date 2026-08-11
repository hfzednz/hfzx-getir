package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/search-service/internal/domain"
)

type Clock interface{ Now() time.Time }
type IDGen interface{ New() uuid.UUID }

type DocumentRepo interface {
	Upsert(ctx context.Context, d domain.ProductDocument) error
	Get(ctx context.Context, tenantID, productID uuid.UUID) (domain.ProductDocument, error)
	List(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.ProductDocument, error)
	Delete(ctx context.Context, tenantID, productID uuid.UUID) error
}

type SynonymRepo interface {
	Save(ctx context.Context, s domain.SynonymRule) error
	List(ctx context.Context, tenantID uuid.UUID, locale string) ([]domain.SynonymRule, error)
}

type BoostRepo interface {
	Save(ctx context.Context, b domain.BoostRule) error
	ListActive(ctx context.Context, tenantID uuid.UUID, now time.Time) ([]domain.BoostRule, error)
}

type IndexJobRepo interface {
	Save(ctx context.Context, j domain.IndexJob) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.IndexJob, error)
	List(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.IndexJob, error)
}

type TrendRepo interface {
	Save(ctx context.Context, t domain.TrendEntry) error
	List(ctx context.Context, tenantID uuid.UUID, kind string, limit int) ([]domain.TrendEntry, error)
	Bump(ctx context.Context, tenantID uuid.UUID, kind, key string, entityID *uuid.UUID, delta float64, now time.Time) error
}

type SuggestRepo interface {
	Upsert(ctx context.Context, tenantID uuid.UUID, c domain.SuggestCandidate) error
	Suggest(ctx context.Context, tenantID uuid.UUID, prefix string, limit int) ([]domain.SuggestCandidate, error)
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
	Update(ctx context.Context, m domain.OutboxMessage) error
}

type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, payload map[string]any) error
}

type LexicalIndex interface {
	Index(ctx context.Context, d domain.ProductDocument) error
	Delete(ctx context.Context, tenantID, productID uuid.UUID) error
	Search(ctx context.Context, tenantID uuid.UUID, tokens []string, fuzzy bool, limit int) ([]uuid.UUID, map[uuid.UUID]float64, error)
}

type VectorStore interface {
	Upsert(ctx context.Context, tenantID, productID uuid.UUID, vector []float64) error
	Search(ctx context.Context, tenantID uuid.UUID, vector []float64, limit int) ([]uuid.UUID, map[uuid.UUID]float64, error)
	Delete(ctx context.Context, tenantID, productID uuid.UUID) error
}

type EmbeddingClient interface {
	EmbedText(ctx context.Context, tenantID uuid.UUID, text, locale string) ([]float64, error)
	EmbedImage(ctx context.Context, tenantID uuid.UUID, imageRef string) ([]float64, error)
}

type LLMClient interface {
	RewriteQuery(ctx context.Context, tenantID uuid.UUID, query, locale string) (string, error)
	SummarizeResults(ctx context.Context, tenantID uuid.UUID, query string, titles []string) (string, error)
}

type CatalogReadClient interface {
	FetchProduct(ctx context.Context, tenantID, productID uuid.UUID) (domain.ProductDocument, error)
}

type InventoryClient interface {
	Available(ctx context.Context, tenantID, productID uuid.UUID, warehouseID *uuid.UUID) (bool, error)
}

type PricingClient interface {
	PriceHint(ctx context.Context, tenantID, productID uuid.UUID) (priceMinor, compareAt int64, currency string, err error)
}

type ReviewClient interface {
	RatingHint(ctx context.Context, tenantID, productID uuid.UUID) (avg float64, count int, err error)
}

type RecommendationClient interface {
	ForYou(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID, limit int) ([]uuid.UUID, error)
}
