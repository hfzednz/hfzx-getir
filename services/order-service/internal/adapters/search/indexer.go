// Package search provides an OpenSearch-backed order indexer with in-process fallback.
package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/order-service/internal/app/ports"
)

const indexName = "nexora-orders"

// Indexer indexes orders in OpenSearch when URL is set; always keeps in-process mem as fallback.
type Indexer struct {
	URL        string
	log        *slog.Logger
	HTTPClient *http.Client
	mu         sync.RWMutex
	docs       map[uuid.UUID]ports.SearchDocument
}

// NewIndexer returns a search indexer. Without OPENSEARCH_URL it indexes in-process only.
func NewIndexer(url string, log *slog.Logger) *Indexer {
	if log == nil {
		log = slog.Default()
	}
	url = strings.TrimRight(url, "/")
	i := &Indexer{
		URL:        url,
		log:        log,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		docs:       make(map[uuid.UUID]ports.SearchDocument),
	}
	if url != "" {
		log.Info("opensearch.indexer.http", "url", url, "index", indexName)
	} else {
		log.Info("opensearch.indexer.memory", "note", "OPENSEARCH_URL empty")
	}
	return i
}

func (i *Indexer) docID(tenantID, orderID uuid.UUID) string {
	return tenantID.String() + "_" + orderID.String()
}

func (i *Indexer) IndexOrder(ctx context.Context, doc ports.SearchDocument) error {
	i.mu.Lock()
	i.docs[doc.OrderID] = doc
	i.mu.Unlock()
	if i.URL == "" {
		return nil
	}
	body, err := json.Marshal(map[string]any{
		"orderId":             doc.OrderID.String(),
		"tenantId":            doc.TenantID.String(),
		"customerPrincipalId": doc.CustomerPrincipalID.String(),
		"status":              doc.Status,
		"type":                doc.Type,
		"currency":            doc.Currency,
		"totalMinor":          doc.TotalMinor,
		"priority":            doc.Priority,
		"idempotencyKey":      doc.IdempotencyKey,
		"skuCodes":            doc.SKUCodes,
		"warehouseIds":        uuidsToStrings(doc.WarehouseIDs),
		"createdAt":           doc.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":           doc.UpdatedAt.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/%s/_doc/%s", indexName, i.docID(doc.TenantID, doc.OrderID))
	return i.doJSON(ctx, http.MethodPut, path, body, nil)
}

func (i *Indexer) DeleteOrder(ctx context.Context, tenantID, orderID uuid.UUID) error {
	i.mu.Lock()
	delete(i.docs, orderID)
	i.mu.Unlock()
	if i.URL == "" {
		return nil
	}
	path := fmt.Sprintf("/%s/_doc/%s", indexName, i.docID(tenantID, orderID))
	if err := i.doJSON(ctx, http.MethodDelete, path, nil, nil); err != nil {
		i.log.Warn("opensearch.delete", "err", err, "orderId", orderID)
	}
	return nil
}

func (i *Indexer) Search(ctx context.Context, q ports.SearchQuery) (ports.SearchResult, error) {
	if i.URL == "" {
		return i.searchMemory(q)
	}
	limit := q.Limit
	if limit <= 0 {
		limit = 20
	}
	must := []map[string]any{
		{"term": map[string]any{"tenantId.keyword": q.TenantID.String()}},
	}
	if q.Query != "" {
		must = append(must, map[string]any{
			"multi_match": map[string]any{
				"query":  q.Query,
				"fields": []string{"orderId^3", "idempotencyKey^2", "status", "skuCodes", "type"},
			},
		})
	}
	if q.Status != nil {
		must = append(must, map[string]any{"term": map[string]any{"status.keyword": string(*q.Status)}})
	}
	if q.CustomerID != nil {
		must = append(must, map[string]any{"term": map[string]any{"customerPrincipalId.keyword": q.CustomerID.String()}})
	}
	payload, _ := json.Marshal(map[string]any{
		"from": q.Offset,
		"size": limit,
		"query": map[string]any{
			"bool": map[string]any{"must": must},
		},
	})
	var raw struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source map[string]any `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := i.doJSON(ctx, http.MethodPost, "/"+indexName+"/_search", payload, &raw); err != nil {
		i.log.Warn("opensearch.search.fallback", "err", err)
		return i.searchMemory(q)
	}
	out := ports.SearchResult{Total: raw.Hits.Total.Value, Hits: make([]ports.SearchDocument, 0, len(raw.Hits.Hits))}
	for _, h := range raw.Hits.Hits {
		out.Hits = append(out.Hits, mapToDoc(h.Source))
	}
	return out, nil
}

func (i *Indexer) searchMemory(q ports.SearchQuery) (ports.SearchResult, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	query := strings.ToLower(strings.TrimSpace(q.Query))
	hits := make([]ports.SearchDocument, 0)
	for _, doc := range i.docs {
		if doc.TenantID != q.TenantID {
			continue
		}
		if q.Status != nil && doc.Status != string(*q.Status) {
			continue
		}
		if q.CustomerID != nil && doc.CustomerPrincipalID != *q.CustomerID {
			continue
		}
		if query != "" {
			match := strings.Contains(strings.ToLower(doc.OrderID.String()), query) ||
				strings.Contains(strings.ToLower(doc.IdempotencyKey), query) ||
				strings.Contains(strings.ToLower(doc.Status), query) ||
				strings.Contains(strings.ToLower(doc.Type), query)
			if !match {
				for _, sku := range doc.SKUCodes {
					if strings.Contains(strings.ToLower(sku), query) {
						match = true
						break
					}
				}
			}
			if !match {
				continue
			}
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

func (i *Indexer) doJSON(ctx context.Context, method, path string, body []byte, out any) error {
	var rdr io.Reader
	if body != nil {
		rdr = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, i.URL+path, rdr)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := i.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if res.StatusCode >= 300 && res.StatusCode != http.StatusNotFound {
		return fmt.Errorf("opensearch %s %s: status %d: %s", method, path, res.StatusCode, string(raw))
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

func uuidsToStrings(ids []uuid.UUID) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.String()
	}
	return out
}

func mapToDoc(src map[string]any) ports.SearchDocument {
	var doc ports.SearchDocument
	if v, ok := src["orderId"].(string); ok {
		doc.OrderID, _ = uuid.Parse(v)
	}
	if v, ok := src["tenantId"].(string); ok {
		doc.TenantID, _ = uuid.Parse(v)
	}
	if v, ok := src["customerPrincipalId"].(string); ok {
		doc.CustomerPrincipalID, _ = uuid.Parse(v)
	}
	if v, ok := src["status"].(string); ok {
		doc.Status = v
	}
	if v, ok := src["type"].(string); ok {
		doc.Type = v
	}
	if v, ok := src["currency"].(string); ok {
		doc.Currency = v
	}
	if v, ok := src["totalMinor"].(float64); ok {
		doc.TotalMinor = int64(v)
	}
	if v, ok := src["priority"].(float64); ok {
		doc.Priority = int(v)
	}
	if v, ok := src["idempotencyKey"].(string); ok {
		doc.IdempotencyKey = v
	}
	if arr, ok := src["skuCodes"].([]any); ok {
		for _, item := range arr {
			if s, ok := item.(string); ok {
				doc.SKUCodes = append(doc.SKUCodes, s)
			}
		}
	}
	if arr, ok := src["warehouseIds"].([]any); ok {
		for _, item := range arr {
			if s, ok := item.(string); ok {
				if id, err := uuid.Parse(s); err == nil {
					doc.WarehouseIDs = append(doc.WarehouseIDs, id)
				}
			}
		}
	}
	if v, ok := src["createdAt"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			doc.CreatedAt = t
		}
	}
	if v, ok := src["updatedAt"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			doc.UpdatedAt = t
		}
	}
	return doc
}

var _ ports.SearchIndexer = (*Indexer)(nil)
