package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nexora/recommendation-service/internal/domain"
)

func TestContentSimilarityAndTopN(t *testing.T) {
	cat := uuid.New()
	a := domain.ProductFeatures{ProductID: uuid.New(), CategoryID: cat, Tags: []string{"dairy", "organic"}}
	b := domain.ProductFeatures{ProductID: uuid.New(), CategoryID: cat, Tags: []string{"dairy", "milk"}}
	if domain.ContentSimilarity(a, b) < 0.4 {
		t.Fatal("expected similarity")
	}
	scores := map[uuid.UUID]float64{a.ProductID: 2, b.ProductID: 1}
	items := domain.TopN(scores, map[uuid.UUID]struct{}{}, 1)
	if len(items) != 1 || items[0].ProductID != a.ProductID {
		t.Fatalf("%+v", items)
	}
}

func TestBlendAndSignalWeight(t *testing.T) {
	id := uuid.New()
	m := domain.BlendScores([]map[uuid.UUID]float64{{id: 1}, {id: 1}}, []float64{0.5, 0.5})
	if m[id] != 1 {
		t.Fatalf("%v", m[id])
	}
	if domain.SignalWeight(domain.SignalPurchase) <= domain.SignalWeight(domain.SignalView) {
		t.Fatal("purchase should outweigh view")
	}
}
