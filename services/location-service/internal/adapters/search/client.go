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
	"time"

	"github.com/google/uuid"
	"github.com/nexora/location-service/internal/app/ports"
	"github.com/nexora/location-service/internal/domain"
)

const (
	poiIndex     = "nexora-location-pois"
	addressIndex = "nexora-location-addresses"
)

// Client indexes POI/address documents into OpenSearch with geo_point when URL is set.
type Client struct {
	URL        string
	log        *slog.Logger
	HTTPClient *http.Client
}

// NewClient returns a GeoSearchIndexer. Empty URL no-ops writes and returns empty search.
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
		log.Info("opensearch.geo.http", "url", url, "indexes", []string{poiIndex, addressIndex})
	} else {
		log.Info("opensearch.geo.noop", "note", "OPENSEARCH_URL empty")
	}
	return c
}

func (c *Client) IndexPOI(ctx context.Context, p domain.POI) error {
	if c.URL == "" {
		return nil
	}
	if !p.Active {
		return c.DeletePOI(ctx, p.TenantID, p.ID)
	}
	body, err := json.Marshal(map[string]any{
		"id":       p.ID.String(),
		"tenantId": p.TenantID.String(),
		"kind":     string(p.Kind),
		"refId":    p.RefID,
		"name":     p.Name,
		"active":   p.Active,
		"location": map[string]float64{"lat": p.Lat, "lon": p.Lng},
		"meta":     p.Meta,
	})
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/%s/_doc/%s", poiIndex, p.TenantID.String()+"_"+p.ID.String())
	return c.doJSON(ctx, http.MethodPut, path, body, nil)
}

func (c *Client) DeletePOI(ctx context.Context, tenantID, poiID uuid.UUID) error {
	if c.URL == "" {
		return nil
	}
	path := fmt.Sprintf("/%s/_doc/%s", poiIndex, tenantID.String()+"_"+poiID.String())
	if err := c.doJSON(ctx, http.MethodDelete, path, nil, nil); err != nil {
		c.log.Warn("opensearch.poi.delete", "err", err, "poiId", poiID)
	}
	return nil
}

func (c *Client) IndexAddress(ctx context.Context, a domain.NormalizedAddress) error {
	if c.URL == "" {
		return nil
	}
	body, err := json.Marshal(map[string]any{
		"id":       a.ID.String(),
		"tenantId": a.TenantID.String(),
		"line1":    a.Line1,
		"building": a.Building,
		"city":     a.Components.City,
		"country":  a.Components.Country,
		"placeId":  a.PlaceID,
		"location": map[string]float64{"lat": a.Lat, "lon": a.Lng},
	})
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/%s/_doc/%s", addressIndex, a.TenantID.String()+"_"+a.ID.String())
	return c.doJSON(ctx, http.MethodPut, path, body, nil)
}

func (c *Client) SearchPOI(ctx context.Context, tenantID uuid.UUID, query string, lat, lng, radiusM float64, limit int) ([]uuid.UUID, error) {
	if c.URL == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 20
	}
	must := []map[string]any{
		{"term": map[string]any{"tenantId.keyword": tenantID.String()}},
		{"term": map[string]any{"active": true}},
	}
	if strings.TrimSpace(query) != "" {
		must = append(must, map[string]any{
			"multi_match": map[string]any{
				"query":  query,
				"fields": []string{"name^2", "refId", "kind"},
			},
		})
	}
	filter := []map[string]any{}
	if domain.ValidLatLng(lat, lng) && radiusM > 0 {
		filter = append(filter, map[string]any{
			"geo_distance": map[string]any{
				"distance": fmt.Sprintf("%.0fm", radiusM),
				"location": map[string]float64{"lat": lat, "lon": lng},
			},
		})
	}
	boolQ := map[string]any{"must": must}
	if len(filter) > 0 {
		boolQ["filter"] = filter
	}
	payload, _ := json.Marshal(map[string]any{
		"size":    limit,
		"query":   map[string]any{"bool": boolQ},
		"_source": false,
	})
	var raw struct {
		Hits struct {
			Hits []struct {
				ID string `json:"_id"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/"+poiIndex+"/_search", payload, &raw); err != nil {
		c.log.Warn("opensearch.poi.search", "err", err)
		return nil, err
	}
	out := make([]uuid.UUID, 0, len(raw.Hits.Hits))
	for _, h := range raw.Hits.Hits {
		parts := strings.SplitN(h.ID, "_", 2)
		idStr := h.ID
		if len(parts) == 2 {
			idStr = parts[1]
		}
		id, err := uuid.Parse(idStr)
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

var _ ports.GeoSearchIndexer = (*Client)(nil)
