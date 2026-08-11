// Package search provides an OpenSearch indexer stub.
package search

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/nexora/inventory-service/internal/app/ports"
)

// Indexer is an OpenSearch-backed indexer stub with in-process fallback.
type Indexer struct {
	URL  string
	log  *slog.Logger
	mu   sync.RWMutex
	docs map[uuid.UUID]ports.SearchDocument
}

// NewIndexer returns a search indexer. Without OPENSEARCH_URL it indexes in-process only.
func NewIndexer(url string, log *slog.Logger) *Indexer {
	if log == nil {
		log = slog.Default()
	}
	return &Indexer{URL: url, log: log, docs: make(map[uuid.UUID]ports.SearchDocument)}
}

func (i *Indexer) IndexStock(_ context.Context, doc ports.SearchDocument) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.docs[doc.BalanceID] = doc
	if i.URL != "" {
		i.log.Debug("opensearch.index.stub", "balanceId", doc.BalanceID, "url", i.URL)
	}
	return nil
}

func (i *Indexer) DeleteStock(_ context.Context, _, balanceID uuid.UUID) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	delete(i.docs, balanceID)
	return nil
}

func (i *Indexer) Search(_ context.Context, q ports.SearchQuery) (ports.SearchResult, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	query := strings.ToLower(strings.TrimSpace(q.Query))
	hits := make([]ports.SearchDocument, 0)
	for _, doc := range i.docs {
		if doc.TenantID != q.TenantID {
			continue
		}
		if q.WarehouseID != nil && doc.WarehouseID != *q.WarehouseID {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(doc.SKUCode), query) {
			continue
		}
		hits = append(hits, doc)
	}
	total := len(hits)
	if q.Offset >= total {
		return ports.SearchResult{Total: total}, nil
	}
	limit := q.Limit
	if limit <= 0 {
		limit = 20
	}
	end := q.Offset + limit
	if end > total {
		end = total
	}
	return ports.SearchResult{Total: total, Hits: hits[q.Offset:end]}, nil
}

func (i *Indexer) ReindexAll(_ context.Context, tenantID uuid.UUID) error {
	i.log.Info("opensearch.reindex.stub", "tenantId", tenantID, "url", i.URL)
	return nil
}

var _ ports.SearchIndexer = (*Indexer)(nil)
