package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nexora/search-service/internal/domain"
)

func TestNormalizeAndIntent(t *testing.T) {
	if domain.NormalizeQuery("  Milk   Tea ") != "milk tea" {
		t.Fatal("normalize")
	}
	if domain.DetectIntent("compare a vs b") != domain.IntentCompare {
		t.Fatal("intent compare")
	}
	if domain.DetectIntent("best deal indirim") != domain.IntentDeal {
		t.Fatal("intent deal")
	}
}

func TestLevenshteinAndRRF(t *testing.T) {
	if domain.Levenshtein("kitten", "sitting") != 3 {
		t.Fatal("levenshtein")
	}
	a := uuid.New()
	b := uuid.New()
	scores := domain.ReciprocalRankFusion([][]uuid.UUID{{a, b}, {b, a}}, 60)
	if scores[a] <= 0 || scores[b] <= 0 {
		t.Fatal("rrf")
	}
}

func TestCosineAndRank(t *testing.T) {
	sim := domain.CosineSimilarity([]float64{1, 0}, []float64{1, 0})
	if sim < 0.99 {
		t.Fatalf("cosine %v", sim)
	}
	s := domain.ComputeRankScore(1, 1, 1, 1, 1, 1, 1, 1, 1, 1)
	if s <= 0 {
		t.Fatal("rank")
	}
}

func TestTokenizeStopStem(t *testing.T) {
	toks := domain.Tokenize("the milk and running")
	if len(toks) == 0 {
		t.Fatal("tokens")
	}
}
