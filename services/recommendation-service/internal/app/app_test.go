package app_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nexora/recommendation-service/internal/app"
	"github.com/nexora/recommendation-service/internal/app/memory"
	"github.com/nexora/recommendation-service/internal/domain"
)

func testDeps() *app.Deps {
	store := memory.NewStore()
	repos := memory.NewRepos(store)
	return &app.Deps{
		Features: repos.Features, Signals: repos.Signals, CoOccur: repos.CoOccur, Outbox: repos.Outbox,
		Catalog: memory.MockCatalog{}, Trends: &memory.MockTrends{},
		Clock: app.SystemClock{}, IDs: app.UUIDGen{},
	}
}

func TestFBTAndHybrid(t *testing.T) {
	d := testDeps()
	tenant := uuid.New()
	user := uuid.New()
	p1, p2, p3 := uuid.New(), uuid.New(), uuid.New()
	cat := uuid.New()
	_ = d.UpsertFeatures(context.Background(), domain.ProductFeatures{ProductID: p1, CategoryID: cat, Tags: []string{"milk"}, PriceMinor: 4000, Popularity: 5, RatingAvg: 4})
	_ = d.UpsertFeatures(context.Background(), domain.ProductFeatures{ProductID: p2, CategoryID: cat, Tags: []string{"milk", "organic"}, PriceMinor: 6000, Popularity: 4, RatingAvg: 4.5})
	_ = d.UpsertFeatures(context.Background(), domain.ProductFeatures{ProductID: p3, CategoryID: uuid.New(), Tags: []string{"bread"}, PriceMinor: 2000, Popularity: 8, RatingAvg: 4.2})

	_, _ = d.IngestSignal(context.Background(), domain.BehaviorSignal{TenantID: tenant, UserID: user, ProductID: p1, Kind: domain.SignalPurchase})
	_, _ = d.IngestSignal(context.Background(), domain.BehaviorSignal{TenantID: tenant, UserID: user, ProductID: p3, Kind: domain.SignalPurchase})

	rail, err := d.Recommend(context.Background(), domain.RecommendRequest{
		TenantID: tenant, ProductID: &p1, Strategy: domain.StrategyFBT, Limit: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rail.Items) == 0 {
		t.Fatal("expected fbt items")
	}

	sim, err := d.Recommend(context.Background(), domain.RecommendRequest{
		TenantID: tenant, ProductID: &p1, Strategy: domain.StrategyContent, Limit: 5,
	})
	if err != nil || len(sim.Items) == 0 {
		t.Fatalf("content %v %+v", err, sim)
	}

	up, err := d.Recommend(context.Background(), domain.RecommendRequest{
		TenantID: tenant, ProductID: &p1, Strategy: domain.StrategyUpsell, Limit: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, it := range up.Items {
		if it.ProductID == p2 {
			found = true
		}
	}
	if !found {
		t.Fatalf("upsell should include pricier same-category p2: %+v", up.Items)
	}

	hy, err := d.Recommend(context.Background(), domain.RecommendRequest{
		TenantID: tenant, UserID: &user, ProductID: &p1, Strategy: domain.StrategyHybrid, Limit: 5,
	})
	if err != nil || len(hy.Items) == 0 {
		t.Fatalf("hybrid %v %+v", err, hy)
	}
}
