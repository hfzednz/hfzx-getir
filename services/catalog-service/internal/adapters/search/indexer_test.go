package search

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/app/ports"
)

func TestSearchMemoryBilingualLocales(t *testing.T) {
	idx := NewIndexer("", slog.Default())
	tenant := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	ctx := context.Background()

	docs := []ports.SearchDocument{
		{
			ProductID: uuid.New(), TenantID: tenant, SKU: "MILK-1L", Title: "Fresh Milk", Brand: "Nexora",
			Locales: map[string]map[string]string{
				"en": {"title": "Fresh Milk", "description": "1 litre whole milk"},
				"tr": {"title": "Taze Süt", "description": "1 litre tam yağlı süt"},
			},
		},
		{
			ProductID: uuid.New(), TenantID: tenant, SKU: "BREAD-1", Title: "Village Bread", Brand: "Fırın",
			Locales: map[string]map[string]string{
				"en": {"title": "Village Bread", "description": "Fresh baked bread"},
				"tr": {"title": "Köy Ekmeği", "description": "Taze fırın ekmeği"},
			},
		},
		{
			ProductID: uuid.New(), TenantID: tenant, SKU: "YOG-1", Title: "Strained Yogurt", Brand: "Nexora",
			Locales: map[string]map[string]string{
				"en": {"title": "Strained Yogurt", "description": "Creamy yogurt"},
				"tr": {"title": "Süzme Yoğurt", "description": "Kremamsı yoğurt"},
			},
		},
	}
	for _, doc := range docs {
		if err := idx.IndexProduct(ctx, doc); err != nil {
			t.Fatal(err)
		}
	}

	pairs := [][2]string{
		{"süt", "MILK-1L"},
		{"milk", "MILK-1L"},
		{"SÜT", "MILK-1L"},
		{"ekmek", "BREAD-1"},
		{"bread", "BREAD-1"},
		{"yoğurt", "YOG-1"},
		{"yogurt", "YOG-1"},
	}
	for _, pair := range pairs {
		res, err := idx.Search(ctx, ports.SearchQuery{TenantID: tenant, Query: pair[0], Limit: 20})
		if err != nil {
			t.Fatalf("%q: %v", pair[0], err)
		}
		found := false
		for _, hit := range res.Hits {
			if hit.SKU == pair[1] {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%q did not find %s (hits=%d)", pair[0], pair[1], len(res.Hits))
		}
	}
}
