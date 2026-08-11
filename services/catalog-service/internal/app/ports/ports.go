// Package ports defines application-layer dependency interfaces (hexagonal ports).
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/domain"
)

// Clock abstracts time for deterministic tests.
type Clock interface {
	Now() time.Time
}

// IDGen abstracts UUID generation.
type IDGen interface {
	New() uuid.UUID
}

// EventPublisher publishes domain events (Kafka adapters).
type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload any) error
}

// SearchDocument is the OpenSearch projection for a product.
type SearchDocument struct {
	ProductID   uuid.UUID
	TenantID    uuid.UUID
	SKU         string
	Barcodes    []string
	Title       string
	Brand       string
	CategoryIDs []uuid.UUID
	Attributes  map[string]any
	Status      domain.ProductStatus
	Locales     map[string]map[string]string
}

// SearchQuery filters catalog search.
type SearchQuery struct {
	TenantID    uuid.UUID
	Query       string
	Status      *domain.ProductStatus
	CategoryID  *uuid.UUID
	BrandID     *uuid.UUID
	Limit       int
	Offset      int
}

// SearchResult is a page of search hits.
type SearchResult struct {
	Total int
	Hits  []SearchDocument
}

// SearchIndexer indexes and queries catalog documents (OpenSearch adapter).
type SearchIndexer interface {
	IndexProduct(ctx context.Context, doc SearchDocument) error
	DeleteProduct(ctx context.Context, tenantID, productID uuid.UUID) error
	Search(ctx context.Context, q SearchQuery) (SearchResult, error)
	Suggest(ctx context.Context, tenantID uuid.UUID, prefix string, limit int) ([]string, error)
	ReindexAll(ctx context.Context, tenantID uuid.UUID) error
}

// MediaAsset is a resolved media-service asset.
type MediaAsset struct {
	ID     uuid.UUID
	Kind   domain.MediaKind
	CDNURL string
	AltText string
}

// MediaClient resolves media asset metadata from media-service.
type MediaClient interface {
	GetAsset(ctx context.Context, tenantID, assetID uuid.UUID) (MediaAsset, error)
}

// AIDescribeResult is a stub AI description response.
type AIDescribeResult struct {
	Title       string
	Description string
	Keywords    []string
}

// AITranslateResult is a stub translation response.
type AITranslateResult struct {
	Lang        string
	Title       string
	Description string
}

// AICategorizeResult is a stub categorization response.
type AICategorizeResult struct {
	CategoryIDs []uuid.UUID
	Confidence  float64
}

// AIQualityResult is a stub quality score response.
type AIQualityResult struct {
	Score   float64
	Issues  []string
	Summary string
}

// AIClient provides AI enrichment ports (stub adapters).
type AIClient interface {
	Describe(ctx context.Context, tenantID, productID uuid.UUID) (AIDescribeResult, error)
	Translate(ctx context.Context, tenantID, productID uuid.UUID, lang string) (AITranslateResult, error)
	Categorize(ctx context.Context, tenantID, productID uuid.UUID) (AICategorizeResult, error)
	QualityScore(ctx context.Context, tenantID, productID uuid.UUID) (AIQualityResult, error)
}

// ProductFilter lists products with optional filters.
type ProductFilter struct {
	TenantID uuid.UUID
	Status   *domain.ProductStatus
	BrandID  *uuid.UUID
	Query    string
	Limit    int
	Offset   int
}

// ProductRepository persists master products.
type ProductRepository interface {
	Create(ctx context.Context, p domain.Product) error
	Update(ctx context.Context, p domain.Product) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Product, error)
	GetBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (domain.Product, error)
	List(ctx context.Context, f ProductFilter) ([]domain.Product, int, error)
	Delete(ctx context.Context, tenantID, id uuid.UUID, at time.Time) error
}

// VariantRepository persists product variants.
type VariantRepository interface {
	Create(ctx context.Context, v domain.Variant) error
	Update(ctx context.Context, v domain.Variant) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Variant, error)
	ListByProduct(ctx context.Context, tenantID, productID uuid.UUID) ([]domain.Variant, error)
	Delete(ctx context.Context, tenantID, id uuid.UUID, at time.Time) error
}

// SKUIdentifierRepository persists barcode / code identifiers.
type SKUIdentifierRepository interface {
	Create(ctx context.Context, s domain.SKUIdentifier) error
	Update(ctx context.Context, s domain.SKUIdentifier) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.SKUIdentifier, error)
	ListByVariant(ctx context.Context, tenantID, variantID uuid.UUID) ([]domain.SKUIdentifier, error)
	FindByValue(ctx context.Context, tenantID uuid.UUID, typ domain.SKUIdentifierType, value string) (domain.SKUIdentifier, error)
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
}

// CategoryRepository persists taxonomy nodes and product memberships.
type CategoryRepository interface {
	Create(ctx context.Context, c domain.Category) error
	Update(ctx context.Context, c domain.Category) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Category, error)
	GetBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (domain.Category, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]domain.Category, error)
	ListChildren(ctx context.Context, tenantID, parentID uuid.UUID) ([]domain.Category, error)
	AssignProduct(ctx context.Context, pc domain.ProductCategory) error
	RemoveProduct(ctx context.Context, productID, categoryID uuid.UUID) error
	ListProductCategories(ctx context.Context, productID uuid.UUID) ([]domain.ProductCategory, error)
	ListProductsInCategory(ctx context.Context, tenantID, categoryID uuid.UUID, limit, offset int) ([]uuid.UUID, error)
}

// BrandRepository persists brands.
type BrandRepository interface {
	Create(ctx context.Context, b domain.Brand) error
	Update(ctx context.Context, b domain.Brand) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Brand, error)
	GetBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (domain.Brand, error)
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Brand, int, error)
}

// AttributeRepository persists attribute definitions and product values.
type AttributeRepository interface {
	CreateDef(ctx context.Context, d domain.AttributeDef) error
	UpdateDef(ctx context.Context, d domain.AttributeDef) error
	GetDefByID(ctx context.Context, tenantID, id uuid.UUID) (domain.AttributeDef, error)
	GetDefByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.AttributeDef, error)
	ListDefs(ctx context.Context, tenantID uuid.UUID) ([]domain.AttributeDef, error)
	UpsertProductAttribute(ctx context.Context, a domain.ProductAttribute) error
	ListProductAttributes(ctx context.Context, tenantID, productID uuid.UUID) ([]domain.ProductAttribute, error)
	DeleteProductAttribute(ctx context.Context, tenantID, id uuid.UUID) error
}

// LocaleRepository persists localized product content.
type LocaleRepository interface {
	Upsert(ctx context.Context, l domain.ProductLocale) error
	Get(ctx context.Context, tenantID, productID uuid.UUID, lang string) (domain.ProductLocale, error)
	ListByProduct(ctx context.Context, tenantID, productID uuid.UUID) ([]domain.ProductLocale, error)
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
}

// SEORepository persists SEO metadata.
type SEORepository interface {
	Upsert(ctx context.Context, s domain.SEO) error
	Get(ctx context.Context, tenantID uuid.UUID, entityType domain.SEOEntityType, entityID uuid.UUID, lang string) (domain.SEO, error)
	ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType domain.SEOEntityType, entityID uuid.UUID) ([]domain.SEO, error)
}

// MediaRepository persists product media references.
type MediaRepository interface {
	Create(ctx context.Context, m domain.ProductMedia) error
	Update(ctx context.Context, m domain.ProductMedia) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.ProductMedia, error)
	ListByProduct(ctx context.Context, tenantID, productID uuid.UUID) ([]domain.ProductMedia, error)
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
}

// BundleRepository persists bundle composition.
type BundleRepository interface {
	Upsert(ctx context.Context, b domain.Bundle) error
	GetByProduct(ctx context.Context, tenantID, productID uuid.UUID) (domain.Bundle, error)
	SetItems(ctx context.Context, bundleID uuid.UUID, items []domain.BundleItem) error
	ListItems(ctx context.Context, bundleID uuid.UUID) ([]domain.BundleItem, error)
}

// RelationRepository persists product relations.
type RelationRepository interface {
	Upsert(ctx context.Context, r domain.ProductRelation) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	ListBySource(ctx context.Context, tenantID, sourceProductID uuid.UUID, typ *domain.RelationType) ([]domain.ProductRelation, error)
}

// VersionRepository persists immutable product versions.
type VersionRepository interface {
	Create(ctx context.Context, v domain.ProductVersion) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.ProductVersion, error)
	GetLatest(ctx context.Context, tenantID, productID uuid.UUID) (domain.ProductVersion, error)
	ListByProduct(ctx context.Context, tenantID, productID uuid.UUID, limit, offset int) ([]domain.ProductVersion, error)
	NextVersionNumber(ctx context.Context, tenantID, productID uuid.UUID) (int, error)
}

// WorkflowRepository persists approval audit trail.
type WorkflowRepository interface {
	CreateAction(ctx context.Context, a domain.ApprovalAction) error
	ListByProduct(ctx context.Context, tenantID, productID uuid.UUID, limit int) ([]domain.ApprovalAction, error)
}

// ImportJobRepository persists async import/export jobs.
type ImportJobRepository interface {
	Create(ctx context.Context, j domain.ImportJob) error
	Update(ctx context.Context, j domain.ImportJob) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.ImportJob, error)
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.ImportJob, error)
}

// ComplianceRepository persists product compliance flags.
type ComplianceRepository interface {
	Upsert(ctx context.Context, c domain.ProductCompliance) error
	GetByProduct(ctx context.Context, tenantID, productID uuid.UUID) (domain.ProductCompliance, error)
}

// SupplierRepository persists supplier masters and product links (metadata only).
type SupplierRepository interface {
	Create(ctx context.Context, s domain.Supplier) error
	Update(ctx context.Context, s domain.Supplier) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Supplier, error)
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Supplier, error)
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Supplier, int, error)
	LinkProduct(ctx context.Context, sp domain.SupplierProduct) error
	ListProducts(ctx context.Context, tenantID, supplierID uuid.UUID) ([]domain.SupplierProduct, error)
}
