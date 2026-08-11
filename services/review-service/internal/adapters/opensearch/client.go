package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/review-service/internal/app/ports"
	"github.com/nexora/review-service/internal/domain"
)

const indexPrefix = "nexora-reviews-"

// Client indexes published reviews in OpenSearch when URL is set.
type Client struct {
	URL        string
	log        *slog.Logger
	HTTPClient *http.Client
}

// NewClient returns an OpenSearch SearchIndexer. Empty URL no-ops writes and returns empty search.
func NewClient(url string, log *slog.Logger) *Client {
	if log == nil {
		log = slog.Default()
	}
	url = strings.TrimRight(url, "/")
	c := &Client{
		URL:        url,
		log:        log,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
	if url != "" {
		log.Info("opensearch.reviews.http", "url", url)
	} else {
		log.Info("opensearch.reviews.noop", "note", "OPENSEARCH_URL empty")
	}
	return c
}

func (c *Client) indexName(tenantID uuid.UUID) string {
	return indexPrefix + tenantID.String()
}

func (c *Client) IndexReview(ctx context.Context, r domain.Review) error {
	if c.URL == "" {
		return nil
	}
	body, err := json.Marshal(map[string]any{
		"id":               r.ID.String(),
		"tenantId":         r.TenantID.String(),
		"authorId":         r.AuthorID.String(),
		"targetType":       r.TargetType,
		"targetId":         r.TargetID.String(),
		"title":            r.Title,
		"body":             r.Body,
		"status":           r.Status,
		"sentiment":        r.Sentiment,
		"tags":             r.Tags,
		"topics":           r.Topics,
		"verifiedPurchase": r.VerifiedPurchase,
		"helpfulCount":     r.HelpfulCount,
		"createdAt":        r.CreatedAt.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/%s/_doc/%s", c.indexName(r.TenantID), r.ID.String())
	return c.doJSON(ctx, http.MethodPut, path, body, nil)
}

func (c *Client) DeleteReview(ctx context.Context, tenantID, reviewID uuid.UUID) error {
	if c.URL == "" {
		return nil
	}
	path := fmt.Sprintf("/%s/_doc/%s", c.indexName(tenantID), reviewID.String())
	if err := c.doJSON(ctx, http.MethodDelete, path, nil, nil); err != nil {
		c.log.Warn("opensearch.delete", "err", err, "reviewId", reviewID)
	}
	return nil
}

func (c *Client) Search(ctx context.Context, tenantID uuid.UUID, query string, targetType string, limit int) ([]uuid.UUID, error) {
	if c.URL == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 20
	}
	must := []map[string]any{
		{"term": map[string]any{"tenantId.keyword": tenantID.String()}},
		{"term": map[string]any{"status.keyword": domain.ReviewStatusPublished}},
	}
	if targetType != "" {
		must = append(must, map[string]any{"term": map[string]any{"targetType.keyword": targetType}})
	}
	if strings.TrimSpace(query) != "" {
		must = append(must, map[string]any{
			"multi_match": map[string]any{
				"query":  query,
				"fields": []string{"title^2", "body", "tags", "topics"},
			},
		})
	}
	payload, _ := json.Marshal(map[string]any{
		"size": limit,
		"query": map[string]any{
			"bool": map[string]any{"must": must},
		},
		"_source": false,
	})
	var raw struct {
		Hits struct {
			Hits []struct {
				ID string `json:"_id"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/"+c.indexName(tenantID)+"/_search", payload, &raw); err != nil {
		c.log.Warn("opensearch.search", "err", err)
		return nil, err
	}
	out := make([]uuid.UUID, 0, len(raw.Hits.Hits))
	for _, h := range raw.Hits.Hits {
		id, err := uuid.Parse(h.ID)
		if err != nil {
			continue
		}
		out = append(out, id)
	}
	return out, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, body []byte, out any) error {
	var rdr io.Reader
	if body != nil {
		rdr = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.URL+path, rdr)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.HTTPClient.Do(req)
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

var _ ports.SearchIndexer = (*Client)(nil)
