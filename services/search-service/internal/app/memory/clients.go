package memory

import (
	"context"
	"hash/fnv"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/search-service/internal/domain"
)

type MockEmbed struct{}

func (MockEmbed) EmbedText(_ context.Context, _ uuid.UUID, text, _ string) ([]float64, error) {
	return hashEmbed(text, 32), nil
}

func (MockEmbed) EmbedImage(_ context.Context, _ uuid.UUID, imageRef string) ([]float64, error) {
	return hashEmbed("img:"+imageRef, 32), nil
}

func hashEmbed(text string, dims int) []float64 {
	v := make([]float64, dims)
	tokens := strings.Fields(strings.ToLower(text))
	if len(tokens) == 0 {
		tokens = []string{text}
	}
	for _, t := range tokens {
		h := fnv.New64a()
		_, _ = h.Write([]byte(t))
		x := h.Sum64()
		for i := 0; i < dims; i++ {
			bit := float64((x>>uint(i%64))&1)*2 - 1
			v[i] += bit
		}
	}
	var norm float64
	for _, x := range v {
		norm += x * x
	}
	norm = math.Sqrt(norm)
	if norm > 0 {
		for i := range v {
			v[i] /= norm
		}
	}
	return v
}

type MockLLM struct{}

func (MockLLM) RewriteQuery(_ context.Context, _ uuid.UUID, query, _ string) (string, error) {
	q := strings.TrimSpace(query)
	if strings.Contains(q, "süt") {
		return q + " milk dairy", nil
	}
	return q, nil
}

func (MockLLM) SummarizeResults(_ context.Context, _ uuid.UUID, query string, titles []string) (string, error) {
	return "Top matches for \"" + query + "\": " + strings.Join(titles, ", "), nil
}

type MockCatalog struct {
	Docs map[uuid.UUID]domain.ProductDocument
}

func (m *MockCatalog) FetchProduct(_ context.Context, tenantID, productID uuid.UUID) (domain.ProductDocument, error) {
	if m.Docs != nil {
		if d, ok := m.Docs[productID]; ok {
			d.TenantID = tenantID
			d.ProductID = productID
			return d, nil
		}
	}
	return domain.ProductDocument{
		TenantID: tenantID, ProductID: productID, Title: "Catalog Product",
		Description: "from catalog port", Available: true, Popularity: 1,
	}, nil
}

type MockInventory struct{}

func (MockInventory) Available(_ context.Context, _, _ uuid.UUID, _ *uuid.UUID) (bool, error) {
	return true, nil
}

type MockPricing struct{}

func (MockPricing) PriceHint(_ context.Context, _, _ uuid.UUID) (int64, int64, string, error) {
	return 4990, 5990, "TRY", nil
}

type MockReviews struct{}

func (MockReviews) RatingHint(_ context.Context, _, _ uuid.UUID) (float64, int, error) {
	return 4.5, 12, nil
}

type MockRecs struct {
	IDs []uuid.UUID
}

func (m *MockRecs) ForYou(_ context.Context, _ uuid.UUID, _ *uuid.UUID, limit int) ([]uuid.UUID, error) {
	if limit > len(m.IDs) {
		limit = len(m.IDs)
	}
	return m.IDs[:limit], nil
}
