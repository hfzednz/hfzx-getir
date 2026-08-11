package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/app"
	"github.com/nexora/catalog-service/internal/app/memory"
	"github.com/nexora/catalog-service/internal/domain"
)

func testDeps(t *testing.T) (*app.Deps, *memory.Store) {
	t.Helper()
	store := memory.NewStore()
	products, variants, skus, categories, brands, attributes, locales, seo, media, bundles, relations, versions, workflow, importJobs, compliance := memory.NewRepos(store)
	d := &app.Deps{
		Products:    products,
		Variants:    variants,
		SKUs:        skus,
		Categories:  categories,
		Brands:      brands,
		Attributes:  attributes,
		Locales:     locales,
		SEO:         seo,
		Media:       media,
		Bundles:     bundles,
		Relations:   relations,
		Versions:    versions,
		Workflow:    workflow,
		ImportJobs:  importJobs,
		Compliance:  compliance,
		Search:      &memory.SearchIndexer{S: store},
		Events:      &memory.EventPublisher{S: store},
		MediaClient: memory.MediaClient{},
		AI:          memory.AIClient{},
		Clock:       memory.FixedTime(time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)),
		IDs:         memory.IDGen{},
	}
	return d, store
}

var testTenant = uuid.MustParse("11111111-1111-1111-1111-111111111111")
var testActor = uuid.MustParse("22222222-2222-2222-2222-222222222222")

func TestFindByBarcode(t *testing.T) {
	d, _ := testDeps(t)
	ctx := context.Background()

	p, err := d.CreateProduct(ctx, app.CreateProductInput{
		TenantID: testTenant, Slug: "widget-a", SKUCode: "WIDGET-A",
	})
	if err != nil {
		t.Fatalf("CreateProduct: %v", err)
	}
	v, err := d.CreateVariant(ctx, app.CreateVariantInput{
		TenantID: testTenant, ProductID: p.ID, Name: "Default", OptionValues: map[string]any{"size": "M"},
	})
	if err != nil {
		t.Fatalf("CreateVariant: %v", err)
	}
	barcode := "4006381333931" // valid EAN-13
	if _, err := d.AddSKUIdentifier(ctx, app.AddSKUIdentifierInput{
		TenantID: testTenant, VariantID: v.ID, Type: domain.SKUTypeEAN, Value: barcode, IsPrimary: true,
	}); err != nil {
		t.Fatalf("AddSKUIdentifier: %v", err)
	}

	sku, varOut, prodOut, err := d.FindByBarcode(ctx, testTenant, domain.SKUTypeEAN, barcode)
	if err != nil {
		t.Fatalf("FindByBarcode: %v", err)
	}
	if sku.Value != barcode {
		t.Fatalf("expected barcode %q, got %q", barcode, sku.Value)
	}
	if varOut.ID != v.ID || prodOut.ID != p.ID {
		t.Fatal("expected matching variant and product")
	}
}

func TestIllegalTransition(t *testing.T) {
	d, _ := testDeps(t)
	ctx := context.Background()

	p, err := d.CreateProduct(ctx, app.CreateProductInput{
		TenantID: testTenant, Slug: "illegal-transition", SKUCode: "ILLEGAL-1",
	})
	if err != nil {
		t.Fatalf("CreateProduct: %v", err)
	}
	_, err = d.TransitionProductStatus(ctx, testTenant, p.ID, domain.ProductStatusPublished)
	if err == nil {
		t.Fatal("expected illegal transition draft → published")
	}
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestPublishCreatesVersion(t *testing.T) {
	d, _ := testDeps(t)
	ctx := context.Background()

	p, err := d.CreateProduct(ctx, app.CreateProductInput{
		TenantID: testTenant, Slug: "publish-me", SKUCode: "PUB-1",
	})
	if err != nil {
		t.Fatalf("CreateProduct: %v", err)
	}
	if _, err := d.SubmitProduct(ctx, testTenant, p.ID, testActor, ""); err != nil {
		t.Fatalf("SubmitProduct: %v", err)
	}
	if _, err := d.ApproveProduct(ctx, testTenant, p.ID, testActor, ""); err != nil {
		t.Fatalf("ApproveProduct: %v", err)
	}
	p, ver, err := d.PublishProduct(ctx, testTenant, p.ID, testActor, "go live")
	if err != nil {
		t.Fatalf("PublishProduct: %v", err)
	}
	if p.Status != domain.ProductStatusPublished {
		t.Fatalf("expected published, got %s", p.Status)
	}
	if ver.ID == uuid.Nil || ver.VersionNumber != 1 {
		t.Fatalf("expected version 1, got %+v", ver)
	}
	versions, err := d.ListProductVersions(ctx, testTenant, p.ID, 10, 0)
	if err != nil || len(versions) != 1 {
		t.Fatalf("ListProductVersions: len=%d err=%v", len(versions), err)
	}
}

func TestCreateCategory(t *testing.T) {
	d, _ := testDeps(t)
	ctx := context.Background()

	root, err := d.CreateCategory(ctx, app.CreateCategoryInput{
		TenantID: testTenant, Name: "Grocery", Slug: "grocery", Kind: domain.CategoryKindDepartment,
	})
	if err != nil {
		t.Fatalf("CreateCategory root: %v", err)
	}
	if root.Depth != 0 || root.Path == "" {
		t.Fatalf("unexpected root path/depth: %+v", root)
	}

	child, err := d.CreateCategory(ctx, app.CreateCategoryInput{
		TenantID: testTenant, ParentID: &root.ID, Name: "Dairy", Slug: "dairy",
	})
	if err != nil {
		t.Fatalf("CreateCategory child: %v", err)
	}
	if child.Depth != 1 || child.ParentID == nil || *child.ParentID != root.ID {
		t.Fatalf("unexpected child: %+v", child)
	}
}

func TestImportValidateColumns(t *testing.T) {
	d, _ := testDeps(t)
	ctx := context.Background()

	csvMissing := "slug,sku_code\nfoo,bar\n"
	job, err := d.ValidateImportCSV(ctx, testTenant, strings.NewReader(csvMissing))
	if err != nil {
		t.Fatalf("ValidateImportCSV: %v", err)
	}
	if job.Status != domain.ImportJobStatusFailed {
		t.Fatalf("expected failed status, got %s", job.Status)
	}
	foundTitle := false
	for _, e := range job.Errors {
		if msg, _ := e["error"].(string); strings.Contains(msg, "title") {
			foundTitle = true
		}
	}
	if !foundTitle {
		t.Fatalf("expected missing title column error, got %+v", job.Errors)
	}

	csvOK := "slug,sku_code,title,lang\nmy-product,SKU-1,My Product,en\n"
	job2, err := d.ValidateImportCSV(ctx, testTenant, strings.NewReader(csvOK))
	if err != nil {
		t.Fatalf("ValidateImportCSV ok: %v", err)
	}
	if job2.Status != domain.ImportJobStatusCompleted {
		t.Fatalf("expected completed, got %s errors=%+v", job2.Status, job2.Errors)
	}
	if job2.TotalRows != 1 || job2.SuccessRows != 1 {
		t.Fatalf("expected 1 success row, got total=%d success=%d", job2.TotalRows, job2.SuccessRows)
	}
}
