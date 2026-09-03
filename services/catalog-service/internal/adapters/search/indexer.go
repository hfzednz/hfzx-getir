// Package search provides an OpenSearch-backed catalog indexer with in-process fallback.
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
	"github.com/nexora/catalog-service/internal/app/ports"
	"github.com/nexora/catalog-service/internal/domain"
)

const indexName = "nexora-catalog-products"

// Indexer indexes catalog products in OpenSearch when URL is set; otherwise in-process.
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

func (i *Indexer) docID(tenantID, productID uuid.UUID) string {
	return tenantID.String() + "_" + productID.String()
}

func (i *Indexer) IndexProduct(ctx context.Context, doc ports.SearchDocument) error {
	i.mu.Lock()
	i.docs[doc.ProductID] = doc
	i.mu.Unlock()
	if i.URL == "" {
		return nil
	}
	body, err := json.Marshal(map[string]any{
		"productId":   doc.ProductID.String(),
		"tenantId":    doc.TenantID.String(),
		"sku":         doc.SKU,
		"barcodes":    doc.Barcodes,
		"title":       doc.Title,
		"brand":       doc.Brand,
		"categoryIds": uuidsToStrings(doc.CategoryIDs),
		"attributes":  doc.Attributes,
		"status":      string(doc.Status),
		"locales":     doc.Locales,
	})
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/%s/_doc/%s", indexName, i.docID(doc.TenantID, doc.ProductID))
	return i.doJSON(ctx, http.MethodPut, path, body, nil)
}

func (i *Indexer) DeleteProduct(ctx context.Context, tenantID, productID uuid.UUID) error {
	i.mu.Lock()
	delete(i.docs, productID)
	i.mu.Unlock()
	if i.URL == "" {
		return nil
	}
	path := fmt.Sprintf("/%s/_doc/%s", indexName, i.docID(tenantID, productID))
	if err := i.doJSON(ctx, http.MethodDelete, path, nil, nil); err != nil {
		i.log.Warn("opensearch.delete", "err", err, "productId", productID)
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
				"fields": []string{"title^3", "sku^2", "brand", "barcodes", "locales.tr.title^3", "locales.en.title^3", "locales.tr.description", "locales.en.description"},
			},
		})
	}
	if q.Status != nil {
		must = append(must, map[string]any{"term": map[string]any{"status.keyword": string(*q.Status)}})
	}
	if q.CategoryID != nil {
		must = append(must, map[string]any{"term": map[string]any{"categoryIds.keyword": q.CategoryID.String()}})
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

func (i *Indexer) Suggest(ctx context.Context, tenantID uuid.UUID, prefix string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 10
	}
	if i.URL == "" {
		return i.suggestMemory(tenantID, prefix, limit)
	}
	payload, _ := json.Marshal(map[string]any{
		"size": limit,
		"query": map[string]any{
			"bool": map[string]any{
				"must": []map[string]any{
					{"term": map[string]any{"tenantId.keyword": tenantID.String()}},
					{"prefix": map[string]any{"title": strings.ToLower(prefix)}},
				},
			},
		},
		"_source": []string{"title"},
	})
	var raw struct {
		Hits struct {
			Hits []struct {
				Source struct {
					Title string `json:"title"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := i.doJSON(ctx, http.MethodPost, "/"+indexName+"/_search", payload, &raw); err != nil {
		return i.suggestMemory(tenantID, prefix, limit)
	}
	out := make([]string, 0, len(raw.Hits.Hits))
	for _, h := range raw.Hits.Hits {
		if h.Source.Title != "" {
			out = append(out, h.Source.Title)
		}
	}
	return out, nil
}

func (i *Indexer) ReindexAll(ctx context.Context, tenantID uuid.UUID) error {
	if i.URL == "" {
		i.log.Info("opensearch.reindex.memory", "tenantId", tenantID)
		return nil
	}
	i.mu.RLock()
	docs := make([]ports.SearchDocument, 0)
	for _, doc := range i.docs {
		if doc.TenantID == tenantID {
			docs = append(docs, doc)
		}
	}
	i.mu.RUnlock()
	for _, doc := range docs {
		if err := i.IndexProduct(ctx, doc); err != nil {
			return err
		}
	}
	return nil
}

func foldSearch(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case 'ı', 'İ', 'I':
			b.WriteByte('i')
		case 'ğ', 'Ğ':
			b.WriteByte('g')
		case 'ü', 'Ü':
			b.WriteByte('u')
		case 'ş', 'Ş':
			b.WriteByte('s')
		case 'ö', 'Ö':
			b.WriteByte('o')
		case 'ç', 'Ç':
			b.WriteByte('c')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func queryVariants(q string) []string {
	out := []string{q}
	if strings.HasSuffix(q, "s") && len(q) > 3 {
		out = append(out, strings.TrimSuffix(q, "s"))
	}
	for _, suf := range []string{"ler", "lar", "lari", "leri"} {
		if strings.HasSuffix(q, suf) && len(q) > len(suf)+2 {
			out = append(out, strings.TrimSuffix(q, suf))
		}
	}
	for _, syn := range grocerySynonyms[q] {
		out = append(out, syn)
	}
	return out
}

// grocerySynonyms expands folded TR↔EN grocery queries so "süt" matches milk
// even when only the English title is indexed (and the reverse).
var grocerySynonyms = map[string][]string{
	"sut":    {"milk"},
	"milk":   {"sut"},
	"ekmek":  {"bread"},
	"bread":  {"ekmek"},
	"yogurt": {"yogurt"},
}

func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	if a == "" {
		return len(b)
	}
	if b == "" {
		return len(a)
	}
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := 0; j <= len(b); j++ {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := cur[j-1] + 1
			sub := prev[j-1] + cost
			cur[j] = del
			if ins < cur[j] {
				cur[j] = ins
			}
			if sub < cur[j] {
				cur[j] = sub
			}
		}
		prev, cur = cur, prev
	}
	return prev[len(b)]
}

func docSearchBlob(doc ports.SearchDocument) string {
	parts := []string{doc.Title, doc.SKU, doc.Brand, strings.Join(doc.Barcodes, " ")}
	for _, loc := range doc.Locales {
		for _, v := range loc {
			parts = append(parts, v)
		}
	}
	for _, v := range doc.Attributes {
		switch t := v.(type) {
		case string:
			parts = append(parts, t)
		}
	}
	return foldSearch(strings.Join(parts, " "))
}

func blobMatches(blob, query string) bool {
	for _, q := range queryVariants(query) {
		if q == "" {
			continue
		}
		if strings.Contains(blob, q) {
			return true
		}
		for _, tok := range strings.Fields(blob) {
			if tokenMatches(q, tok) {
				return true
			}
		}
	}
	return false
}

func tokenMatches(q, tok string) bool {
	if tok == "" || q == "" {
		return false
	}
	if strings.Contains(tok, q) || (len(tok) >= 3 && strings.Contains(q, tok)) {
		return true
	}
	if len(q) >= 4 && levenshtein(q, tok) <= 1 {
		return true
	}
	n := len(q)
	if len(tok) < n {
		n = len(tok)
	}
	if n >= 4 {
		same := 0
		for i := 0; i < n && q[i] == tok[i]; i++ {
			same++
		}
		if same >= n-1 {
			return true
		}
	}
	return false
}

func (i *Indexer) searchMemory(q ports.SearchQuery) (ports.SearchResult, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	query := foldSearch(q.Query)
	hits := make([]ports.SearchDocument, 0)
	for _, doc := range i.docs {
		if doc.TenantID != q.TenantID {
			continue
		}
		if q.Status != nil && doc.Status != *q.Status {
			continue
		}
		if query != "" && !blobMatches(docSearchBlob(doc), query) {
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

func (i *Indexer) suggestMemory(tenantID uuid.UUID, prefix string, limit int) ([]string, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	prefix = strings.ToLower(prefix)
	out := make([]string, 0)
	for _, doc := range i.docs {
		if doc.TenantID == tenantID && strings.HasPrefix(strings.ToLower(doc.Title), prefix) {
			out = append(out, doc.Title)
		}
		if len(out) >= limit {
			break
		}
	}
	return out, nil
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
	if v, ok := src["productId"].(string); ok {
		doc.ProductID, _ = uuid.Parse(v)
	}
	if v, ok := src["tenantId"].(string); ok {
		doc.TenantID, _ = uuid.Parse(v)
	}
	if v, ok := src["sku"].(string); ok {
		doc.SKU = v
	}
	if v, ok := src["title"].(string); ok {
		doc.Title = v
	}
	if v, ok := src["brand"].(string); ok {
		doc.Brand = v
	}
	if v, ok := src["status"].(string); ok {
		doc.Status = domain.ProductStatus(v)
	}
	return doc
}

var _ ports.SearchIndexer = (*Indexer)(nil)
