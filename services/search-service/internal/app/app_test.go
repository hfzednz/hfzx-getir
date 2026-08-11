package app_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nexora/search-service/internal/app"
	"github.com/nexora/search-service/internal/app/memory"
	"github.com/nexora/search-service/internal/domain"
)

func testDeps() *app.Deps {
	store := memory.NewStore()
	repos := memory.NewRepos(store)
	return &app.Deps{
		Docs: repos.Docs, Synonyms: repos.Synonyms, Boosts: repos.Boosts,
		Jobs: repos.Jobs, Trends: repos.Trends, Suggests: repos.Suggests, Outbox: repos.Outbox,
		Lexical: repos.Lexical, Vectors: repos.Vectors,
		Embed: memory.MockEmbed{}, LLM: memory.MockLLM{},
		Catalog: &memory.MockCatalog{}, Inventory: memory.MockInventory{},
		Pricing: memory.MockPricing{}, Reviews: memory.MockReviews{},
		Recs: &memory.MockRecs{}, Clock: app.SystemClock{}, IDs: app.UUIDGen{},
	}
}

func TestIndexAndHybridSearch(t *testing.T) {
	d := testDeps()
	tenant := uuid.New()
	p1 := uuid.New()
	p2 := uuid.New()
	_ = d.IndexDocument(context.Background(), domain.ProductDocument{
		TenantID: tenant, ProductID: p1, Title: "Organic Milk 1L", Description: "fresh dairy milk",
		BrandName: "Nexora Farms", Tags: []string{"dairy", "organic"}, Available: true,
		PriceMinor: 4500, Popularity: 10, RatingAvg: 4.8, ReviewCount: 20,
		Attributes: map[string]string{"organic": "true"}, CategoryPath: []string{"Dairy"},
	})
	_ = d.IndexDocument(context.Background(), domain.ProductDocument{
		TenantID: tenant, ProductID: p2, Title: "Almond Drink", Description: "plant based",
		BrandName: "Green", Available: true, PriceMinor: 5500, Popularity: 3,
	})
	_, err := d.UpsertSynonym(context.Background(), domain.SynonymRule{
		TenantID: tenant, Term: "süt", Synonyms: []string{"milk"}, Locale: "tr-TR",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.UpsertBoost(context.Background(), domain.BoostRule{
		TenantID: tenant, Name: "pin milk", Kind: "pin", ProductIDs: []uuid.UUID{p1}, Weight: 2,
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := d.Search(context.Background(), domain.SearchQuery{
		TenantID: tenant, RawQuery: "milk", Hybrid: true, IncludeFacets: true,
		Filters: domain.SearchFilters{AvailableOnly: true}, Page: 1, PageSize: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total < 1 || res.ZeroResult {
		t.Fatalf("expected hits %+v", res)
	}
	if len(res.Hits) == 0 || !res.Hits[0].Pinned {
		t.Fatalf("expected pinned first %+v", res.Hits)
	}

	ac, err := d.Autocomplete(context.Background(), tenant, "org", 5)
	if err != nil || len(ac) == 0 {
		t.Fatalf("autocomplete %v %v", err, ac)
	}

	img, err := d.ImageSearch(context.Background(), tenant, "media://milk.jpg", 5)
	if err != nil {
		t.Fatal(err)
	}
	if img.Total < 1 {
		t.Fatalf("image search empty")
	}
}

func TestVoiceAndTrends(t *testing.T) {
	d := testDeps()
	tenant := uuid.New()
	_ = d.IndexDocument(context.Background(), domain.ProductDocument{
		TenantID: tenant, ProductID: uuid.New(), Title: "Banana Bundle", Available: true, Popularity: 5,
	})
	res, err := d.VoiceSearch(context.Background(), tenant, "banana deal", "en", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Intent != domain.IntentDeal {
		t.Fatalf("intent %s", res.Intent)
	}
	trends, err := d.ListTrends(context.Background(), tenant, "search", 10)
	if err != nil || len(trends) == 0 {
		t.Fatalf("trends %v %v", err, trends)
	}
}
