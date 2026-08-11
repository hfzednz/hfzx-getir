package memory

import (
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/nexora/search-service/internal/domain"
)

type Store struct {
	mu        sync.RWMutex
	Docs      map[string]domain.ProductDocument
	Synonyms  []domain.SynonymRule
	Boosts    []domain.BoostRule
	Jobs      map[uuid.UUID]domain.IndexJob
	Trends    map[string]domain.TrendEntry
	Suggests  []domain.SuggestCandidate
	Outbox    []domain.OutboxMessage
	Lexical   map[string]map[uuid.UUID]float64 // token -> product scores
	Vectors   map[string][]float64             // tenant|product -> vec
	TenantSug map[uuid.UUID][]domain.SuggestCandidate
}

func NewStore() *Store {
	return &Store{
		Docs: make(map[string]domain.ProductDocument),
		Jobs: make(map[uuid.UUID]domain.IndexJob),
		Trends: make(map[string]domain.TrendEntry),
		Lexical: make(map[string]map[uuid.UUID]float64),
		Vectors: make(map[string][]float64),
		TenantSug: make(map[uuid.UUID][]domain.SuggestCandidate),
	}
}

func docKey(tenantID, productID uuid.UUID) string {
	return tenantID.String() + "|" + productID.String()
}

func trendKey(tenantID uuid.UUID, kind, key string) string {
	return tenantID.String() + "|" + kind + "|" + key
}

func vecKey(tenantID, productID uuid.UUID) string {
	return tenantID.String() + "|" + productID.String()
}

func tokenKey(tenantID uuid.UUID, token string) string {
	return tenantID.String() + "|" + strings.ToLower(token)
}
