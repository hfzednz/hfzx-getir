// Package media provides a media-service HTTP client with synthesized CDN fallback.
package media

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/app/ports"
	"github.com/nexora/catalog-service/internal/domain"
)

// Client resolves media assets from media-service.
type Client struct {
	BaseURL string
	HTTP    *http.Client
	log     *slog.Logger
}

// NewClient returns a media-service client.
func NewClient(baseURL string, log *slog.Logger) *Client {
	if log == nil {
		log = slog.Default()
	}
	baseURL = strings.TrimRight(baseURL, "/")
	c := &Client{
		BaseURL: baseURL,
		HTTP:    &http.Client{Timeout: 5 * time.Second},
		log:     log,
	}
	if baseURL != "" {
		log.Info("media.client.http", "baseURL", baseURL)
	} else {
		log.Info("media.client.unconfigured", "note", "MEDIA_SERVICE_URL empty")
	}
	return c
}

// GetAsset GETs media-service asset metadata. Empty BaseURL errors; HTTP failure falls back to synthesized CDN URL.
func (c *Client) GetAsset(ctx context.Context, tenantID, assetID uuid.UUID) (ports.MediaAsset, error) {
	if c.BaseURL == "" {
		return ports.MediaAsset{}, fmt.Errorf("media service not configured")
	}

	fallback := ports.MediaAsset{
		ID:     assetID,
		Kind:   domain.MediaKindImage,
		CDNURL: fmt.Sprintf("%s/v1/media/assets/%s/cdn", c.BaseURL, assetID),
	}

	path := fmt.Sprintf("/v1/media/assets/%s", assetID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		c.log.Warn("media.get.fallback", "err", err, "assetId", assetID)
		return fallback, nil
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Tenant-Id", tenantID.String())

	res, err := c.HTTP.Do(req)
	if err != nil {
		c.log.Warn("media.get.fallback", "err", err, "assetId", assetID)
		return fallback, nil
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 300 {
		c.log.Warn("media.get.fallback", "status", res.StatusCode, "assetId", assetID, "body", string(raw))
		return fallback, nil
	}

	var body struct {
		ID          string `json:"id"`
		Kind        string `json:"kind"`
		CDNURL      string `json:"cdnUrl"`
		ContentType string `json:"contentType"`
		AltText     string `json:"altText"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		c.log.Warn("media.get.fallback", "err", err, "assetId", assetID)
		return fallback, nil
	}

	out := fallback
	if body.ID != "" {
		if id, err := uuid.Parse(body.ID); err == nil {
			out.ID = id
		}
	}
	if body.Kind != "" {
		k := domain.MediaKind(body.Kind)
		if k.Valid() {
			out.Kind = k
		}
	}
	if body.CDNURL != "" {
		out.CDNURL = body.CDNURL
	}
	if body.AltText != "" {
		out.AltText = body.AltText
	}
	return out, nil
}

var _ ports.MediaClient = (*Client)(nil)
